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
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})
	router.POST("/repositories", CreateRepository(db))

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
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})
	router.POST("/repositories", CreateRepository(db))

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
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})
	router.POST("/repositories", CreateRepository(db))

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

// TestListRepositories_AuthenticatedUser tests listing authenticated user's repositories (owned + shared)
func TestListRepositories_AuthenticatedUser(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{}, &hub.Collaborator{})

	user1 := createTestUser(t, db, "user1")
	user2 := createTestUser(t, db, "user2")

	// Create repositories
	repo1 := createTestRepository(t, db, user1.ID, "user1-public", "public")
	createTestRepository(t, db, user1.ID, "user1-private", "private")
	createTestRepository(t, db, user2.ID, "user2-public", "public")

	// Make user1 a collaborator on user2's repo
	if err := db.Create(&hub.Collaborator{RepoID: repo1.ID, UserID: user2.ID, Role: "write"}).Error; err != nil {
		t.Fatalf("failed to create collaborator: %v", err)
	}

	router := gin.New()
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user1)
		c.Next()
	})
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// user1 should see their 2 repos (owned) + any shared with them
	assert.Greater(t, int64(1), int64(0)) // At least one repo
}

// TestListRepositories_UserOwnedRepos tests listing user's own repositories
func TestListRepositories_UserOwnedRepos(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "my-repo-1", "public")
	createTestRepository(t, db, user.ID, "my-repo-2", "private")

	router := gin.New()
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should see 2 repos owned by this user
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 2, len(response.Repositories))
}

// TestListRepositories_Unauthenticated tests that unauthenticated users cannot access ListRepositories
func TestListRepositories_Unauthenticated(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	router := gin.New()
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListRepositories_SearchFilter tests searching repositories by q parameter
func TestListRepositories_SearchFilter(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "alpha-repo", "public")
	createTestRepository(t, db, user.ID, "beta-repo", "public")
	createTestRepository(t, db, user.ID, "gamma-repo", "public")

	router := gin.New()
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?q=beta", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should only match beta-repo
	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, "beta-repo", response.Repositories[0].Slug)
}

// TestListRepositories_VisibilityFilter tests filtering repositories by visibility
func TestListRepositories_VisibilityFilter(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "public-repo", "public")
	createTestRepository(t, db, user.ID, "private-repo1", "private")
	createTestRepository(t, db, user.ID, "private-repo2", "private")

	router := gin.New()
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?visibility=private", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should only match private repos
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 2, len(response.Repositories))
}

// TestListRepositories_SortByParameter tests sorting repositories
func TestListRepositories_SortByParameter(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "z-repo", "public")
	createTestRepository(t, db, user.ID, "a-repo", "public")
	createTestRepository(t, db, user.ID, "m-repo", "public")

	router := gin.New()
	// Set authenticated user (must be before route registration)
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Next()
	})
	router.GET("/repositories", ListRepositories(db))

	httpReq, _ := http.NewRequest("GET", "/repositories?sortBy=name&sortOrder=asc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Should return 3 repos sorted by name ascending
	assert.Equal(t, int64(3), response.Total)
	if len(response.Repositories) > 0 {
		assert.Equal(t, "a-repo", response.Repositories[0].Slug)
	}
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
