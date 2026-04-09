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
func repoScope(g *gin.RouterGroup, db *gorm.DB) {
	g.GET("/tags", mock)
	g.GET("/tags/:tag", mock)
	g.GET("/tags/:tag/model", handlers.GetTagModel(db))
	g.PUT("/tags/:tag", handlers.PutTagModel(db))
	g.DELETE("/tags/:tag", mock)
	g.GET("/collaborators", mock)
	g.PUT("/collaborators/:username", mock)
	g.DELETE("/collaborators/:username", mock)
	g.POST("/transfer", mock)
	g.GET("/lineage", mock)
	g.POST("/fork", mock)
}

// repoManagementScope mounts management routes with proper middleware (requires owner permission)
func repoManagementScope(g *gin.RouterGroup, db *gorm.DB) {
	// Tag routes
	g.GET("/tags/:tag/model",
		middleware.RequireRepositoryPermission(middleware.PermissionRead),
		handlers.GetTagModel(db))
	g.PUT("/tags/:tag",
		middleware.RequireAuthentication(),
		middleware.RequireRepositoryPermission(middleware.PermissionWrite),
		handlers.PutTagModel(db))
	g.DELETE("/tags/:tag",
		middleware.RequireAuthentication(),
		middleware.RequireRepositoryPermission(middleware.PermissionWrite),
		handlers.DeleteTag(db))
	// Collaborator management routes (requires owner)
	g.GET("/collaborators",
		middleware.RequireAuthentication(),
		handlers.ListCollaborators(db))
	g.PUT("/collaborators/:username",
		middleware.RequireAuthentication(),
		middleware.RequireRepositoryPermission(middleware.PermissionOwner),
		handlers.AddCollaborator(db))
	g.DELETE("/collaborators/:username",
		middleware.RequireAuthentication(),
		middleware.RequireRepositoryPermission(middleware.PermissionOwner),
		handlers.RemoveCollaborator(db))

	// Transfer ownership route (requires owner)
	g.POST("/transfer",
		middleware.RequireAuthentication(),
		middleware.RequireRepositoryPermission(middleware.PermissionOwner),
		handlers.TransferRepositoryOwnership(db))

	// Fork repository route (requires read access)
	g.POST("/fork",
		middleware.RequireAuthentication(),
		middleware.RequireRepositoryPermission(middleware.PermissionRead),
		handlers.ForkRepository(db))
}

// RegisterRoutes mounts all v0 endpoints from endpoints.md.
func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	v0 := r.Group("/v0")
	cookieSecure := !cfg.DevMode
	// MVP auth setup:
	// - access token TTL comes from middleware default (single source of truth)
	// - attach auth context + best-effort user hydration on every /v0 request
	v0.Use(
		middleware.AttachAuthContextWithCookieSecure(db, cfg.JWTSecret, 0, &cookieSecure),
		middleware.AuthenticateRequest(),
	)

	// Auth: me, logout, Google OAuth, CLI login/poll
	auth := v0.Group("/auth")
	auth.GET("/me", middleware.RequireAuthentication(), handlers.AuthMe())
	auth.POST("/logout", handlers.AuthLogout())
	auth.GET("/login/google/start", handlers.GoogleStart(cfg))
	auth.GET("/login/google/callback", handlers.GoogleCallback(db, cfg))
	auth.POST("/cli/login", handlers.CLILogin(db))
	auth.POST("/cli/poll", handlers.CLIPoll(db))

	// Repositories - authenticated routes for listing/creating.
	reposAuth := v0.Group("/repositories", middleware.RequireAuthentication())
	reposAuth.GET("", handlers.ListRepositories(db))        // list (?scope, q, owner, visibility) - requires auth
	reposAuth.GET("/:owner", handlers.ListRepositories(db)) // list by owner - requires auth
	reposAuth.POST("", handlers.CreateRepository(db))       // create - requires auth


	// Repository by owner/slug (publicly viewable for public repos).
	// Access is still enforced by ResolveRepositoryByOwnerSlug + CheckRepositoryAccess +
	// RequireRepositoryPermission(read), which returns 404 for private repos when unauthenticated.
	ownerSlugGroup := v0.Group("/repositories/:owner/:slug",
		middleware.ResolveRepositoryByOwnerSlug(db),
		middleware.CheckRepositoryAccess(db),
	)
	ownerSlugGroup.GET("", middleware.RequireRepositoryPermission(middleware.PermissionRead), handlers.GetRepository(db))
	ownerSlugGroup.PATCH("", middleware.RequireAuthentication(), middleware.RequireRepositoryPermission(middleware.PermissionOwner), handlers.UpdateRepository(db))
	ownerSlugGroup.DELETE("", middleware.RequireAuthentication(), middleware.RequireRepositoryPermission(middleware.PermissionOwner), handlers.DeleteRepository(db))
	// Management routes under this scope (tags, collaborators, transfer, fork, lineage)
	repoManagementScope(ownerSlugGroup, db)

	// Repo by id (alias): same routes as owner/slug, under /repo/:id.
	// ResolveRepositoryByID requires repo ID lookup to be implemented.
	idGroup := v0.Group("/repo/:id",
		middleware.ResolveRepositoryByID(db),
		middleware.CheckRepositoryAccess(db),
	)
	idGroup.GET("", middleware.RequireRepositoryPermission(middleware.PermissionRead), handlers.GetRepository(db))
	idGroup.PATCH("", middleware.RequireAuthentication(), middleware.RequireRepositoryPermission(middleware.PermissionOwner), handlers.UpdateRepository(db))
	idGroup.DELETE("", middleware.RequireAuthentication(), middleware.RequireRepositoryPermission(middleware.PermissionOwner), handlers.DeleteRepository(db))
	// Management routes under this scope (tags, collaborators, transfer, fork, lineage)
	repoManagementScope(idGroup, db)

	// Search
	v0.GET("/search", handlers.GlobalSearch(db))
}
