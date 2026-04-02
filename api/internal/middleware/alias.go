package middleware

import (
	"errors"
	"net/http"
	"strings"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/models/hub"
)

// ResolveRepositoryByOwnerSlug is middleware that resolves :owner/:slug route
// parameters to a Repository and stores it in context for handler access.
//
// On success:
//   - Sets context key "repository" to the resolved *hub.Repository (preloaded with Owner)
//
// On error:
//   - Sends 404 if repository not found
//   - Sends 500 if database error occurs
//   - Does NOT call next handler on error
func ResolveRepositoryByOwnerSlug(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		owner := strings.ToLower(strings.TrimSpace(c.Param("owner")))
		slug := c.Param("slug")

		var repo hub.Repository
		if err := db.Preload("Owner").Where("hub_repositories.slug = ?", slug).
			Joins("INNER JOIN hub_users ON hub_users.id = hub_repositories.owner_id").
			Where("hub_users.username = ?", owner).
			First(&repo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			c.Abort()
			return
		}

		c.Set("repository", &repo)
		c.Next()
	}
}

// ResolveRepositoryByID is middleware that resolves a repository ID from context
// or route parameters. Primarily used when you already have a repo ID and need
// to fetch the full repository details.
//
// On success:
//   - Sets context key "repository" to the resolved *hub.Repository (preloaded with Owner)
//
// On error:
//   - Sends 404 if repository not found
//   - Sends 500 if database error occurs
func ResolveRepositoryByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		parsed, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil || parsed == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repository ID"})
			c.Abort()
			return
		}
		repoID := uint(parsed)

		var repo hub.Repository
		if err := db.Preload("Owner").First(&repo, repoID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "repository not found"})
				c.Abort()
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			c.Abort()
			return
		}

		c.Set("repository", &repo)
		c.Next()
	}
}

// GetRepository retrieves the repository from context set by middleware.
// Returns nil if not present.
func GetRepository(c *gin.Context) *hub.Repository {
	repo, exists := c.Get("repository")
	if !exists {
		return nil
	}
	if r, ok := repo.(*hub.Repository); ok {
		return r
	}
	return nil
}
