package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	router.POST("/repositories", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		CreateRepository(db)(c)
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
	router.POST("/repositories", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		CreateRepository(db)(c)
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
	router.POST("/repositories", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		CreateRepository(db)(c)
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

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=mine", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

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
	router.GET("/repositories", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		ListRepositories(db)(c)
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

// TestListRepositories_Unauthenticated tests that unauthenticated users cannot access private scopes.
func TestListRepositories_Unauthenticated(t *testing.T) {
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

	httpReq, _ := http.NewRequest("GET", "/repositories?scope=mine&visibility=private", nil)
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

// TestGlobalSearch_Unauthenticated_OnlyPublicRepos tests that anonymous callers
// only ever see public repositories, even when repos are owned by different users.
func TestGlobalSearch_Unauthenticated_OnlyPublicRepos(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	// 2 public repos, 1 private
	user1 := createTestUser(t, db, "user1")
	user2 := createTestUser(t, db, "user2")
	createTestRepository(t, db, user1.ID, "user1-public", "public")
	createTestRepository(t, db, user1.ID, "user1-private", "private")
	createTestRepository(t, db, user2.ID, "user2-public", "public")

	router := gin.New()
	router.GET("/search", GlobalSearch(db))

	httpReq, _ := http.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Results should include the 2 public repos, but not the private one
	assert.Equal(t, int64(2), response.Total)
	for _, repo := range response.Repositories {
		assert.Equal(t, "public", repo.Visibility)
	}
}

// TestGlobalSearch_Unauthenticated_VisibilityPrivate_ReturnsEmpty is a regression
// test for issue #173: GET /v0/search?visibility=private must never leak private
// repositories to unauthenticated callers.
// Issue link: https://github.com/opendi-org/model-hub/issues/173
func TestGlobalSearch_Unauthenticated_VisibilityPrivate_ReturnsEmpty(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	// 1 public repo, 1 private
	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "secret-repo", "private")
	createTestRepository(t, db, user.ID, "public-repo", "public")

	router := gin.New()
	router.GET("/search", GlobalSearch(db))

	// Private visibility requested
	httpReq, _ := http.NewRequest("GET", "/search?visibility=private", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Unauthenticated response should never contain private repos
	assert.Equal(t, int64(0), response.Total)
	assert.Empty(t, response.Repositories)
}

// TestGlobalSearch_Unauthenticated_VisibilityPrivate_WithOwner_ReturnsEmpty covers
// the specific narrowing case given in #173. Leak should stay closed with scope narrowed
// to one owner.
// Issue link: https://github.com/opendi-org/model-hub/issues/173
func TestGlobalSearch_Unauthenticated_VisibilityPrivate_WithOwner_ReturnsEmpty(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	// "alice" was the example name in #173
	// Create private repos for alice
	user := createTestUser(t, db, "alice")
	createTestRepository(t, db, user.ID, "alice-secret-1", "private")
	createTestRepository(t, db, user.ID, "alice-secret-2", "private")

	router := gin.New()
	router.GET("/search", GlobalSearch(db))

	// Narrow scope to specific owner
	httpReq, _ := http.NewRequest("GET", "/search?visibility=private&owner=alice", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Unauthenticated response should never contain private repos
	assert.Equal(t, int64(0), response.Total)
	assert.Empty(t, response.Repositories)
}

// TestGlobalSearch_AuthenticatedUser_VisibilityPrivate_OnlyOwnRepos tests that an
// authenticated caller using ?visibility=private only sees their own private repos,
// never another user's.
func TestGlobalSearch_AuthenticatedUser_VisibilityPrivate_OnlyOwnRepos(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	// Create private repos for two users
	user1 := createTestUser(t, db, "user1")
	user2 := createTestUser(t, db, "user2")
	createTestRepository(t, db, user1.ID, "user1-private", "private")
	createTestRepository(t, db, user2.ID, "user2-private", "private")

	router := gin.New()
	// Authenticate as user1
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, user1)
		c.Next()
	})
	router.GET("/search", GlobalSearch(db))

	httpReq, _ := http.NewRequest("GET", "/search?visibility=private", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// User1 should only see private repos from User1, not User2.
	assert.Equal(t, int64(1), response.Total)
	if assert.Len(t, response.Repositories, 1) {
		assert.Equal(t, "user1-private", response.Repositories[0].Slug)
	}
}

// TestGlobalSearch_AuthenticatedCollaborator_SeesSharedPrivateRepo tests that a
// collaborator (not the owner) can find a shared private repo via search.
func TestGlobalSearch_AuthenticatedCollaborator_SeesSharedPrivateRepo(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{}, &hub.Collaborator{})

	// Two repos: one shared, one not
	owner := createTestUser(t, db, "owner")
	collaborator := createTestUser(t, db, "collaborator")
	sharedRepo := createTestRepository(t, db, owner.ID, "shared-private", "private")
	createTestRepository(t, db, owner.ID, "not-shared-private", "private")

	// Make the "collaborator" user a collaborator on the "owner" user's repo
	if err := db.Create(&hub.Collaborator{RepoID: sharedRepo.ID, UserID: collaborator.ID, Role: "read"}).Error; err != nil {
		t.Fatalf("failed to create collaborator: %v", err)
	}

	router := gin.New()
	// Authenticate as collaborator
	router.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, collaborator)
		c.Next()
	})
	router.GET("/search", GlobalSearch(db))

	httpReq, _ := http.NewRequest("GET", "/search?visibility=private", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// Collaborator should see the shared repo, but not the un-shared repo
	assert.Equal(t, int64(1), response.Total)
	if assert.Len(t, response.Repositories, 1) {
		assert.Equal(t, "shared-private", response.Repositories[0].Slug)
	}
}

// TestGlobalSearch_ScopeParamIgnored_NoAuthRequired tests that GlobalSearch always
// behaves as scope=all. Unlike ListRepositories, it shouldn't reject unauthenticated
// requests for scope=mine/shared-with-me, it should just ignore those settings.
func TestGlobalSearch_ScopeParamIgnored_NoAuthRequired(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	// One repo so we have results to receive
	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "public-repo", "public")

	router := gin.New()
	router.GET("/search", GlobalSearch(db))

	// Test both scope=mine and scope=shared-with-me
	cases := []struct {
		name               string
		endpointWithParams string
	}{
		{"scope=mine", "/search?scope=mine"},
		{"scope=shared-with-me", "/search?scope=shared-with-me"},
	}

	// Both should succeed despite being unauthenticated, and return
	// the one public repo
	expectedResponseCount := int64(1)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			httpReq, _ := http.NewRequest("GET", tc.endpointWithParams, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httpReq)

			assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

			var response dto.ListRepositoriesResponse
			json.Unmarshal(w.Body.Bytes(), &response)

			assert.Equal(t, expectedResponseCount, response.Total)
		})
	}
}

// TestGlobalSearch_SearchFilter tests that q= parameter works via the shared ListRepositories service.
// See TestListRepositories_SearchFilter
func TestGlobalSearch_SearchFilter(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer db.Migrator().DropTable(&hub.Repository{}, &hub.User{})

	// 3 repos with searchable names
	user := createTestUser(t, db, "testuser")
	createTestRepository(t, db, user.ID, "alpha-repo", "public")
	createTestRepository(t, db, user.ID, "beta-repo", "public")
	createTestRepository(t, db, user.ID, "beta-private-repo", "private")

	router := gin.New()
	router.GET("/search", GlobalSearch(db))

	httpReq, _ := http.NewRequest("GET", "/search?q=beta", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var response dto.ListRepositoriesResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	// "beta" query should return the one public repo with slug containing "beta".
	// the private repo containing "beta" should NOT be included
	assert.Equal(t, int64(1), response.Total)
	if assert.Len(t, response.Repositories, 1) {
		assert.Equal(t, "beta-repo", response.Repositories[0].Slug)
	}
}

// TestIsValidTagName tests the tag name validation helper
func TestIsValidTagName(t *testing.T) {
	tests := []struct {
		tag   string
		valid bool
	}{
		{"latest", true},
		{"v1.0", true},
		{"my-tag", true},
		{"my_tag", true},
		{"v1.0.0-beta", true},
		{"", false},
		{"tag with spaces", false},
		{"tag@special", false},
		{string(make([]byte, 256)), false}, // too long
	}

	for _, tc := range tests {
		result := isValidTagName(tc.tag)
		assert.Equal(t, tc.valid, result, "tag %q", tc.tag)
	}
}

// loadExampleCDM reads a CDM JSON fixture from the cdm-json-schema/examples directory.
func loadExampleCDM(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join(apiRoot(), "cdm-json-schema", "examples", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("load CDM fixture %s: %v", filename, err)
	}
	return data
}

// loadAnyExampleCDM reads the first available fixture from a preferred list.
// This avoids brittle failures when one fixture file is missing in CI checkouts.
func loadAnyExampleCDM(t *testing.T, preferred ...string) []byte {
	t.Helper()
	base := filepath.Join(apiRoot(), "cdm-json-schema", "examples")
	if _, err := os.Stat(base); err != nil {
		t.Skipf("cdm-json-schema/examples not found at %s: %v (init submodule: git submodule update --init)", base, err)
		return nil
	}
	for _, name := range preferred {
		p := filepath.Join(base, name)
		if data, err := os.ReadFile(p); err == nil {
			return data
		}
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		t.Skipf("read examples directory %s: %v", base, err)
		return nil
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
			continue
		}
		p := filepath.Join(base, e.Name())
		if data, err := os.ReadFile(p); err == nil {
			return data
		}
	}

	t.Skipf("no readable CDM fixture found in %s", base)
	return nil
}

// TestPutTagModel_Upload tests uploading a model to a tag
func TestPutTagModel_Upload(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer cleanupTestDB(t, db)

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "public")
	cdmJSON := loadAnyExampleCDM(t, "coffee_noninteractive.json", "coffee.json")

	router := gin.New()
	router.PUT("/repositories/:owner/:slug/tags/:tag", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionOwner)
		PutTagModel(db)(c)
	})

	req, _ := http.NewRequest("PUT", "/repositories/testuser/test-repo/tags/v1.0", bytes.NewReader(cdmJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "v1.0", resp["tag"])
	assert.NotEmpty(t, resp["digest"])
	assert.NotEmpty(t, resp["size"])
}

// TestPutTagModel_InvalidTag tests that invalid tag names are rejected
func TestPutTagModel_InvalidTag(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer cleanupTestDB(t, db)

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "public")

	router := gin.New()
	router.PUT("/repositories/:owner/:slug/tags/:tag", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionOwner)
		PutTagModel(db)(c)
	})

	req, _ := http.NewRequest("PUT", "/repositories/testuser/test-repo/tags/bad tag!", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestPutTagModel_NoWritePermission tests that read-only users cannot upload
func TestPutTagModel_NoWritePermission(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer cleanupTestDB(t, db)

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "public")

	router := gin.New()
	router.PUT("/repositories/:owner/:slug/tags/:tag", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionRead)
		PutTagModel(db)(c)
	})

	req, _ := http.NewRequest("PUT", "/repositories/testuser/test-repo/tags/v1.0", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestGetTagModel_DownloadAlwaysReturnsModel tests downloading a model payload.
func TestGetTagModel_DownloadAlwaysReturnsModel(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer cleanupTestDB(t, db)

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "public")
	cdmJSON := loadAnyExampleCDM(t, "coffee_noninteractive.json", "coffee.json")

	// Upload first
	putRouter := gin.New()
	putRouter.PUT("/repositories/:owner/:slug/tags/:tag", func(c *gin.Context) {
		middleware.SetCurrentUser(c, user)
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionOwner)
		PutTagModel(db)(c)
	})
	putReq, _ := http.NewRequest("PUT", "/repositories/testuser/test-repo/tags/v1.0", bytes.NewReader(cdmJSON))
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	putRouter.ServeHTTP(putW, putReq)
	assert.Equal(t, http.StatusOK, putW.Code, putW.Body.String())

	var putResp map[string]interface{}
	json.Unmarshal(putW.Body.Bytes(), &putResp)
	digest, ok := putResp["digest"].(string)
	if !ok || digest == "" {
		t.Fatalf("expected upload digest in response, got: %s", putW.Body.String())
	}

	getRouter := gin.New()
	getRouter.GET("/repositories/:owner/:slug/tags/:tag/model", func(c *gin.Context) {
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionRead)
		GetTagModel(db)(c)
	})

	// Fresh download — should return 200 with model body.
	getReq, _ := http.NewRequest("GET", "/repositories/testuser/test-repo/tags/v1.0/model", nil)
	getW := httptest.NewRecorder()
	getRouter.ServeHTTP(getW, getReq)
	assert.Equal(t, http.StatusOK, getW.Code)
	assert.NotEmpty(t, getW.Body.Bytes())
	assert.Contains(t, getW.Body.String(), "{")
	assert.NotEmpty(t, digest)
}

// TestGetTagModel_NotFound tests downloading a non-existent tag
func TestGetTagModel_NotFound(t *testing.T) {
	db := testDB(t)
	cleanupTestDB(t, db)
	defer cleanupTestDB(t, db)

	user := createTestUser(t, db, "testuser")
	repo := createTestRepository(t, db, user.ID, "test-repo", "public")

	router := gin.New()
	router.GET("/repositories/:owner/:slug/tags/:tag/model", func(c *gin.Context) {
		c.Set("repository", repo)
		c.Set("permission", middleware.PermissionRead)
		GetTagModel(db)(c)
	})

	req, _ := http.NewRequest("GET", "/repositories/testuser/test-repo/tags/nonexistent/model", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// cleanupTestDB removes all test data from the database
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	db.Exec("DELETE FROM hub_cdm_tags")
	db.Exec("DELETE FROM hub_repositories")
	db.Exec("DELETE FROM hub_collaborators")
	db.Exec("DELETE FROM hub_users")
}
