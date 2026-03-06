package routes

import (
	"net/http"

	"opendi.org/model-hub/api/internal/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter constructs the HTTP handler for the API using gin. For now, all
// endpoints are mock implementations that return simple JSON payloads
// describing the requested endpoint. This lets the frontend and CLI be wired
// up before the real persistence and auth logic is implemented.
func NewRouter(cfg *config.Config, db *gorm.DB) http.Handler {
	_ = cfg
	_ = db

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Auth endpoints.
	router.GET("/v0/auth/me", ginMock("GET /v0/auth/me"))
	router.POST("/v0/auth/logout", ginMock("POST /v0/auth/logout"))
	router.POST("/v0/auth/token/refresh", ginMock("POST /v0/auth/token/refresh"))
	router.GET("/v0/auth/login/google/start", ginMock("GET /v0/auth/login/google/start"))
	router.GET("/v0/auth/login/google/callback", ginMock("GET /v0/auth/login/google/callback"))
	router.POST("/v0/auth/cli/login", ginMock("POST /v0/auth/cli/login"))
	router.POST("/v0/auth/cli/poll", ginMock("POST /v0/auth/cli/poll"))

	// Repository listing and creation.
	router.GET("/v0/repositories", ginMock("GET /v0/repositories"))
	router.POST("/v0/repositories", ginMock("POST /v0/repositories"))

	// Owner-scoped repositories and all nested paths: treat as generic mock.
	router.Any("/v0/repositories/*path", ginMockDynamic())

	// ID-based alias prefix /v0/repo/:id and its subpaths are also mocked.
	router.Any("/v0/repo/*path", ginMockDynamic())

	// Search.
	router.GET("/v0/search", ginMock("GET /v0/search"))

	return router
}

// ginMock returns a handler that always responds with a fixed endpoint label.
func ginMock(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"endpoint": endpoint,
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"query":    c.Request.URL.RawQuery,
			"status":   "mock",
		})
	}
}

// ginMockDynamic echoes back the method and path, useful for all nested repo/tag
// routes during the mock phase.
func ginMockDynamic() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"endpoint": c.Request.Method + " " + c.Request.URL.Path,
			"method":   c.Request.Method,
			"path":     c.Request.URL.Path,
			"query":    c.Request.URL.RawQuery,
			"status":   "mock",
		})
	}
}
