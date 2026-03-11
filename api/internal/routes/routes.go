package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/config"
)

// mock returns a handler that echoes method and path. Replace with real handlers later.
func mock(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"method": c.Request.Method, "path": c.Request.URL.Path, "status": "mock"})
}

// repoScope mounts routes for a single repo (by owner/slug or by id): get/patch/delete, tags, tag model, collaborators, transfer, lineage, fork.
func repoScope(g *gin.RouterGroup) {
	h := mock
	g.GET("", h)
	g.PATCH("", h)
	g.DELETE("", h)
	g.GET("/tags", h)
	g.GET("/tags/:tag", h)
	g.GET("/tags/:tag/model", h)
	g.PUT("/tags/:tag", h)
	g.DELETE("/tags/:tag", h)
	g.GET("/collaborators", h)
	g.PUT("/collaborators/:username", h)
	g.DELETE("/collaborators/:username", h)
	g.POST("/transfer", h)
	g.GET("/lineage", h)
	g.POST("/fork", h)
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
	repos.GET("", mock)           // list (?scope, q, owner)
	repos.GET("/:owner", mock)    // list by owner
	repos.POST("", mock)          // create
	repoScope(repos.Group("/:owner/:slug")) // get/patch/delete repo, tags, collaborators, transfer, lineage, fork

	// Repo by id (alias): same routes as owner/slug, under /repo/:id. Add ResolveRepoID middleware when lookup exists.
	repoScope(v0.Group("/repo/:id"))

	// Search
	v0.GET("/search", mock)
}
