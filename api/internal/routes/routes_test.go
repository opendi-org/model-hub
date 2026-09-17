package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/config"
	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/services"
	"opendi.org/model-hub/api/internal/testsupport"
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

func TestCLIPoll_RateLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testsupport.TestDB(t)

	testsupport.CleanupTestDB(t, db)
	defer testsupport.CleanupTestDB(t, db)

	router := gin.New()

	limit := 3
	reqPerMin := float64(600)
	cfg := &config.Config{
		JWTSecret:                         "test-secret",
		DevMode:                           true,
		RateLimitCLIPollBurst:             limit,
		RateLimitCLIPollRequestsPerMinute: reqPerMin,
	}
	RegisterRoutes(router, db, cfg)

	auth := services.NewAuthService(db)
	sessionCode := "auth-session-TestCLIPoll_RateLimitExceeded"
	if err := auth.CreateCLISession(sessionCode, time.Now().Add(time.Hour).UTC()); err != nil {
		t.Fatalf("Error creating session: %v", err)
	}

	req := dto.CLIPollRequest{
		Code: sessionCode,
	}
	body, _ := json.Marshal(req)

	for i := 0; i < limit; i++ {
		t.Run(fmt.Sprintf("limit-run-%d", i), func(t *testing.T) {
			httpReq, _ := http.NewRequest("POST", "/v0/auth/cli/poll", bytes.NewReader(body))
			httpReq.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httpReq)
			// Endpoint should be found, but we haven't hit the limit yet
			assert.NotEqual(t, http.StatusTooManyRequests, w.Code, w.Body)
			assert.NotEqual(t, http.StatusNotFound, w.Code, w.Body)
		})
	}

	t.Run("limit-run-exceeded", func(t *testing.T) {
		httpReq, _ := http.NewRequest("POST", "/v0/auth/cli/poll", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)
		assert.Equal(t, http.StatusTooManyRequests, w.Code, w.Body)
	})

	// Wait until we're back within the rate limit
	waitTime := 1.0 / reqPerMin * float64(time.Minute)
	waitBuffer := 10.0 * float64(time.Millisecond)
	time.Sleep(time.Duration(waitTime + waitBuffer))

	t.Run("limit-run-replenished", func(t *testing.T) {
		httpReq, _ := http.NewRequest("POST", "/v0/auth/cli/poll", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)
		// Endpoint should be found, and we're back within the rate limit
		assert.NotEqual(t, http.StatusTooManyRequests, w.Code, w.Body)
		assert.NotEqual(t, http.StatusNotFound, w.Code, w.Body)
	})
}
