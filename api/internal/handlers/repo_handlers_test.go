package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/database"
	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/middleware"
	"opendi.org/model-hub/api/internal/models/hub"
)

// apiRoot returns the api/ directory (parent of internal/), relative to this test file.
func apiRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// testDB opens a Postgres DB and runs migrations. Skips the test if DB is not available.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		host := os.Getenv("DB_HOSTNAME")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USERNAME")
		if user == "" {
			user = "postgres"
		}
		pass := os.Getenv("DB_PASSWORD")
		if pass == "" {
			pass = "postgres"
		}
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "modelhub_test"
		}
		dsn = "host=" + host + " port=" + port + " user=" + user + " password=" + pass + " dbname=" + dbname + " sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("database not available: %v (set DB_* or TEST_DSN to run)", err)
		return nil
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// createTestUser creates a test user in the database and returns it
func createTestUser(t *testing.T, db *gorm.DB, username string) *hub.User {
	user := &hub.User{
		Username:  username,
		Email:     username + "@example.com",
		AvatarURL: "https://example.com/avatar.png",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

// createTestRepository creates a test repository in the database
func createTestRepository(t *testing.T, db *gorm.DB, ownerID uint, slug string, visibility string) *hub.Repository {
	repo := &hub.Repository{
		OwnerID:     ownerID,
		Slug:        slug,
		Description: "Test repository",
		Visibility:  visibility,
	}
	if err := db.Create(repo).Error; err != nil {
		t.Fatalf("failed to create test repository: %v", err)
	}
	return repo
}

// setupTestRouter sets up a Gin route for testing
func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add test middleware to set current user in context
	r.Use(func(c *gin.Context) {
		// Check if there's a user_id in the request header for testing
		if userID := c.GetString("test_user_id"); userID != "" {
			c.Set("_auth_user", &hub.User{
				Model: gorm.Model{ID: 0}, // Will be set by test
			})
		}
	})

	repos := r.Group("/v0/repositories")
	{
		repos.POST("/", CreateRepository(db))
		repos.GET("/", ListRepositories(db))
	}

	return r
}

// ── Test Cases ────────────────────────────────────────────────────────────────

// TestCreateRepository_ValidRequest tests creating a repository with valid data
func TestCreateRepository_ValidRequest(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")

	router := gin.New()
	router.POST("/repositories", CreateRepository(db))

	// Set authenticated user
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})

	req := dto.CreateRepositoryRequest{
		Slug:        "test-repo",
		Description: "A test repository",
		Visibility:  "public",
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/repositories", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response dto.RepositoryListItem
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "test-repo", response.Slug)
	assert.Equal(t, "A test repository", response.Description)
	assert.Equal(t, "public", response.Visibility)
	assert.Equal(t, "testuser", response.Owner)
}

// TestCreateRepository_InvalidSlug tests that invalid slugs are rejected
func TestCreateRepository_InvalidSlug(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")

	router := gin.New()
	router.POST("/repositories", CreateRepository(db))

	// Set authenticated user
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})

	tests := []string{
		"test repo", // spaces not allowed
		"test@repo", // special chars not allowed
		"",          // empty
	}

	for _, invalidSlug := range tests {
		req := dto.CreateRepositoryRequest{
			Slug:       invalidSlug,
			Visibility: "public",
		}

		body, _ := json.Marshal(req)
		httpReq, _ := http.NewRequest("POST", "/repositories", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, httpReq)
		assert.Equal(t, http.StatusBadRequest, w.Code, "invalid slug %q should be rejected", invalidSlug)
	}
}

// TestCreateRepository_DuplicateSlug tests that duplicate slugs under same owner are rejected
func TestCreateRepository_DuplicateSlug(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "existing-repo", "private")

	router := gin.New()
	router.POST("/repositories", CreateRepository(db))
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})

	req := dto.CreateRepositoryRequest{
		Slug:       "existing-repo",
		Visibility: "public",
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/repositories", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestCreateRepository_Unauthorized tests that unauthenticated users cannot create repos
func TestCreateRepository_Unauthorized(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	router := gin.New()
	router.POST("/repositories", CreateRepository(db))

	req := dto.CreateRepositoryRequest{
		Slug:       "test-repo",
		Visibility: "public",
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/repositories", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListRepositories_AllScope tests listing all public repositories
func TestListRepositories_AllScope(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{}, &hub.Collaborator{})

	user1 := createTestUser(t, db, "user1")
	user2 := createTestUser(t, db, "user2")

	// Create some repositories
	createTestRepository(t, db, user1.ID, "public-repo", "public")
	createTestRepository(t, db, user1.ID, "private-repo", "private")
	createTestRepository(t, db, user2.ID, "public-repo2", "public")

	router := gin.New()
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=all", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should only see 2 public repos when not authenticated
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 2, len(response.Repositories))
}

// TestListRepositories_MineScope tests listing user's own repositories
func TestListRepositories_MineScope(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "my-repo-1", "public")
	createTestRepository(t, db, user.ID, "my-repo-2", "private")

	router := gin.New()
	router.GET("/repositories", ListRepositories(db))
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=mine", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should see 2 repos owned by this user
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 2, len(response.Repositories))
}

// TestListRepositories_MineScope_Unauthenticated tests that unauthenticated users cannot view "mine" scope
func TestListRepositories_MineScope_Unauthenticated(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	router := gin.New()
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=mine", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListRepositories_SearchFilter tests searching repositories by name
func TestListRepositories_SearchFilter(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "alpha-repo", "public")
	createTestRepository(t, db, user.ID, "beta-repo", "public")
	createTestRepository(t, db, user.ID, "gamma-repo", "public")

	router := gin.New()
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=all&q=beta", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should only match beta-repo
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, "beta-repo", response.Repositories[0].Slug)
}

// TestListRepositories_OwnerFilter tests filtering repositories by owner
func TestListRepositories_OwnerFilter(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user1 := createTestUser(t, db, "user1")
	user2 := createTestUser(t, db, "user2")
	createTestRepository(t, db, user1.ID, "repo1", "public")
	createTestRepository(t, db, user2.ID, "repo2", "public")

	router := gin.New()
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=all&owner=user1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should only match user1's repo
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, "user1", response.Repositories[0].Owner)
}

// TestListRepositories_InvalidScope tests that invalid scope returns error
func TestListRepositories_InvalidScope(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	router := gin.New()
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateRepository_ValidRequest tests updating repo details
func TestUpdateRepository_ValidRequest(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "private")

	router := gin.New()
	// Mock middleware that sets repository and permission in context
	router.PATCH("/repositories/:owner/:slug", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionOwner)
		UpdateRepository(db)(c)
	})

	req := dto.UpdateRepositoryRequest{
		Slug:        "updated-repo",
		Description: "Updated description",
		Visibility:  "public",
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("PATCH", "/repositories/testuser/test-repo", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.RepositoryListItem
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "updated-repo", response.Slug)
	assert.Equal(t, "Updated description", response.Description)
	assert.Equal(t, "public", response.Visibility)
}

// TestUpdateRepository_InvalidSlug tests that invalid slug is rejected on update
func TestUpdateRepository_InvalidSlug(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "private")

	router := gin.New()
	router.PATCH("/repositories/:owner/:slug", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionOwner)
		UpdateRepository(db)(c)
	})

	req := dto.UpdateRepositoryRequest{
		Slug:       "invalid$slug",
		Visibility: "public",
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("PATCH", "/repositories/testuser/test-repo", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateRepository_NoPermission tests that non-owners cannot update
func TestUpdateRepository_NoPermission(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	owner := createTestUser(t, db, "owner")
	other := createTestUser(t, db, "other")
	repo := createTestRepository(t, db, owner.ID, "test-repo", "private")

	router := gin.New()
	router.PATCH("/repositories/:owner/:slug", func(c *gin.Context) {
		middleware.SetCurrentUser(c, other)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionRead)
		UpdateRepository(db)(c)
	})

	req := dto.UpdateRepositoryRequest{
		Slug:       "test-repo",
		Visibility: "public",
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("PATCH", "/repositories/owner/test-repo", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestDeleteRepository_ValidRequest tests deleting a repository
func TestDeleteRepository_ValidRequest(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "private")

	router := gin.New()
	router.DELETE("/repositories/:owner/:slug", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionOwner)
		DeleteRepository(db)(c)
	})

	httpReq, _ := http.NewRequest("DELETE", "/repositories/testuser/test-repo", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify repo is soft-deleted
	var repo2 hub.Repository
	err := db.Unscoped().First(&repo2, repo.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, repo2.DeletedAt)
}

// TestDeleteRepository_NoPermission tests that non-owners cannot delete
func TestDeleteRepository_NoPermission(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	owner := createTestUser(t, db, "owner")
	other := createTestUser(t, db, "other")
	repo := createTestRepository(t, db, owner.ID, "test-repo", "private")

	router := gin.New()
	router.DELETE("/repositories/:owner/:slug", func(c *gin.Context) {
		middleware.SetCurrentUser(c, other)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionWrite)
		DeleteRepository(db)(c)
	})

	httpReq, _ := http.NewRequest("DELETE", "/repositories/owner/test-repo", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestIsValidSlug tests the slug validation function
func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		slug  string
		valid bool
	}{
		{"valid-slug", true},
		{"valid_slug", true},
		{"valid123", true},
		{"123", true},
		{"a", true},
		{"a-_B-_0", true},
		{"", false},
		{"slug with spaces", false},
		{"slug@special", false},
		{"slug#chars", false},
		{"slug.dot", false},
		{string(make([]byte, 256)), false}, // too long
	}

	for _, tc := range tests {
		result := isValidSlug(tc.slug)
		assert.Equal(t, tc.valid, result, "slug %q", tc.slug)
	}
}

// cleanupTestDB removes all test data from the database
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("DELETE FROM hub_repositories")
	db.Exec("DELETE FROM hub_collaborators")
	db.Exec("DELETE FROM hub_users")
}
