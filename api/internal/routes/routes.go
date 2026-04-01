package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/config"
	"opendi.org/model-hub/api/internal/handlers"
	"opendi.org/model-hub/api/internal/middleware"
)

// mock returns a handler that echoes method and path. Replace with real handlers later.
func mock(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"method": c.Request.Method, "path": c.Request.URL.Path, "status": "mock"})
}

// repoScope mounts additional routes for a single repo (tags, collaborators, transfer, lineage, fork).
// The base GET/PATCH/DELETE routes are mounted separately by the caller to allow middleware injection.
func repoScope(g *gin.RouterGroup) {
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

	// Repositories - requires authentication for searching own projects
	repos := v0.Group("/repositories", middleware.RequireAuthentication())
	repos.GET("", handlers.ListRepositories(db))        // list (?scope, q, owner, visibility) - requires auth
	repos.GET("/:owner", handlers.ListRepositories(db)) // list by owner - requires auth
	repos.POST("", handlers.CreateRepository(db))       // create - requires auth

	// Repository by owner/slug with middleware: resolve repo, check access
	ownerSlugGroup := repos.Group("/:owner/:slug",
		middleware.ResolveRepositoryByOwnerSlug(db),
		middleware.CheckRepositoryAccess(db),
	)
	ownerSlugGroup.GET("", handlers.GetRepository(db))
	ownerSlugGroup.PATCH("", middleware.RequireAuthentication(), handlers.UpdateRepository(db))
	ownerSlugGroup.DELETE("", middleware.RequireAuthentication(), handlers.DeleteRepository(db))
	// Remaining routes under this scope (tags, collaborators, etc.)
	repoScope(ownerSlugGroup)

	// Repo by id (alias): same routes as owner/slug, under /repo/:id.
	// ResolveRepositoryByID requires repo ID lookup to be implemented.
	idGroup := v0.Group("/repo/:id",
		middleware.ResolveRepositoryByID(db),
		middleware.CheckRepositoryAccess(db),
	)
	idGroup.GET("", handlers.GetRepository(db))
	idGroup.PATCH("", middleware.RequireAuthentication(), handlers.UpdateRepository(db))
	idGroup.DELETE("", middleware.RequireAuthentication(), handlers.DeleteRepository(db))
	// Remaining routes under this scope (tags, collaborators, etc.)
	repoScope(idGroup)

	// Search
	v0.GET("/search", handlers.GlobalSearch(db))
}
