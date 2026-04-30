package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/config"
)

func TestRegisterRoutes_AuthMeRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := &gorm.DB{}

	r := gin.New()
	cfg := &config.Config{
		JWTSecret: "test-secret",
		DevMode:   true,
	}
	RegisterRoutes(r, db, cfg)

	req := httptest.NewRequest(http.MethodGet, "/v0/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d body=%s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}

