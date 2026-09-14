package testsupport

import (
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/database"
	"opendi.org/model-hub/api/internal/models/hub"
)

// testDB opens a Postgres DB and runs migrations. Skips the test if DB is not available.
func TestDB(t *testing.T) *gorm.DB {
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
func CreateTestUser(t *testing.T, db *gorm.DB, username string) *hub.User {
	t.Helper()
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
func CreateTestRepository(t *testing.T, db *gorm.DB, ownerID uint, slug string, visibility string) *hub.Repository {
	t.Helper()
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

// createTestCollaborator registers a collaborator to a repo and returns the collaborator object
func CreateTestCollaborator(t *testing.T, db *gorm.DB, repoID uint, userID uint, role string) *hub.Collaborator {
	t.Helper()
	collaborator := &hub.Collaborator{
		RepoID: repoID,
		UserID: userID,
		Role:   role,
	}
	if err := db.Create(collaborator).Error; err != nil {
		t.Fatalf("failed to create collaborator: %v", err)
	}
	return collaborator
}

// cleanupTestDB removes all test data from the database. Children must be
// deleted before the parents they reference (hub_cdm_tags and
// hub_collaborators both FK into hub_repositories and hub_users) or the
// deletes fail/leave orphaned rows that later break AutoMigrate's FK setup.
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM hub_cdm_tags",
		"DELETE FROM hub_collaborators",
		"DELETE FROM hub_repositories",
		"DELETE FROM hub_cli_sessions",
		"DELETE FROM hub_users",
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("cleanupTestDB: %s: %v", stmt, err)
		}
	}
}
