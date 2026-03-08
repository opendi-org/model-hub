package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/config"
	"opendi.org/model-hub/api/internal/handlers"
)

// mock returns a handler that echoes method and path. Replace with real handlers later.
func mock(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"method": c.Request.Method, "path": c.Request.URL.Path, "status": "mock"})
}

// repoScope mounts routes for a single repo (by owner/slug or by id): get/patch/delete, tags, tag model, collaborators, transfer, lineage, fork.
func repoScope(g *gin.RouterGroup, db *gorm.DB) {
	g.GET("", handlers.GetRepository(db))
	g.PATCH("", handlers.UpdateRepository(db))
	g.DELETE("", handlers.DeleteRepository(db))
	g.GET("/tags", mock)
	g.GET("/tags/:tag", mock)
	g.GET("/tags/:tag/model", mock)
	g.PUT("/tags/:tag", mock)
	g.DELETE("/tags/:tag", mock)
	g.GET("/collaborators", mock)
	g.PUT("/collaborators/:username", mock)
	g.DELETE("/collaborators/:username", mock)
	g.POST("/transfer", mock)
	g.GET("/lineage", mock)
	g.POST("/fork", mock)
}

// RegisterRoutes mounts all v0 endpoints from endpoints.md.
func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	v0 := r.Group("/v0")
	_ = db
	_ = cfg

	// Auth: me, logout, token refresh, Google OAuth, CLI login/poll
	auth := v0.Group("/auth")
	auth.GET("/me", mock)
	auth.POST("/logout", mock)
	auth.POST("/token/refresh", mock)
	auth.GET("/login/google/start", mock)
	auth.GET("/login/google/callback", mock)
	auth.POST("/cli/login", mock)
	auth.POST("/cli/poll", mock)

	// Repositories
	repos := v0.Group("/repositories")
	repos.GET("", handlers.ListRepositories(db))        // list (?scope, q, owner) - optional auth
	repos.GET("/:owner", handlers.ListRepositories(db)) // list by owner - optional auth
	repos.POST("", handlers.CreateRepository(db))       // create (auth checked in handler)
	repoScope(repos.Group("/:owner/:slug"), db)         // get/patch/delete repo, tags, collaborators, transfer, lineage, fork

	// Repo by id (alias): same routes as owner/slug, under /repo/:id. Add ResolveRepoID middleware when lookup exists.
	repoScope(v0.Group("/repo/:id"), db)

	// Search
	v0.GET("/search", mock)
}
