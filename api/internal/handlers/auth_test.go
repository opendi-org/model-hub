package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"opendi.org/model-hub/api/internal/dto"
	"opendi.org/model-hub/api/internal/middleware"
	"opendi.org/model-hub/api/internal/services"
	"opendi.org/model-hub/api/internal/testsupport"
)

func TestFrontendCLIApprovedURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "frontend callback route",
			input:    "http://localhost:3000/auth/callback",
			expected: "http://localhost:3000/auth/cli-approved",
		},
		{
			name:     "frontend nested base path",
			input:    "https://example.com/model-hub/auth/callback",
			expected: "https://example.com/model-hub/auth/cli-approved",
		},
		{
			name:     "invalid redirect url falls back to relative route",
			input:    "not-a-url",
			expected: "/auth/cli-approved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frontendCLIApprovedURL(tt.input)
			if got != tt.expected {
				t.Fatalf("frontendCLIApprovedURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestCLIPoll_PendingSession_ReturnsAccepted Tests that a pending CLI auth session returns
// HTTP status 202, accepted, when polled with CLIPoll
func TestCLIPoll_PendingSession_ReturnsAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testsupport.TestDB(t)

	testsupport.CleanupTestDB(t, db)
	defer testsupport.CleanupTestDB(t, db)

	router := gin.New()
	// Set JWT secret info. Must be set for CLIPoll to work
	// (Pending path technically doesn't hit this dependency now, but if something goes wrong it might try to)
	router.Use(middleware.AttachAuthContextWithCookieSecure(db, "test-secret", time.Hour, nil))
	router.POST("/auth/cli/poll", CLIPoll(db))

	// Create a pending CLI auth session
	auth := services.NewAuthService(db)
	sessionCode := "auth-session-TestCLIPoll_PendingSession_ReturnsAccepted"
	if err := auth.CreateCLISession(sessionCode, time.Now().Add(time.Hour).UTC()); err != nil {
		t.Fatalf("Error creating session: %v", err)
	}

	// Poll the pending session
	req := dto.CLIPollRequest{
		Code: sessionCode,
	}
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/cli/poll", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

// TestCLIPoll_ApprovedSession_ReturnsToken tests that a CLI auth session that has passed through
// pending state and into approved returns HTTP 200 and includes a token in its response
func TestCLIPoll_ApprovedSession_ReturnsToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testsupport.TestDB(t)

	testsupport.CleanupTestDB(t, db)
	defer testsupport.CleanupTestDB(t, db)

	router := gin.New()
	// Set JWT secret info. Must be set for CLIPoll to work
	router.Use(middleware.AttachAuthContextWithCookieSecure(db, "test-secret", time.Hour, nil))
	router.POST("/auth/cli/poll", CLIPoll(db))

	// Create a pending CLI auth session
	auth := services.NewAuthService(db)
	sessionCode := "auth-session-TestCLIPoll_ApprovedSession_ReturnsToken"
	if err := auth.CreateCLISession(sessionCode, time.Now().Add(time.Hour).UTC()); err != nil {
		t.Fatalf("Error creating session: %v", err)
	}

	// Approve the pending session for a test user
	user := testsupport.CreateTestUser(t, db, "user-clipoll")
	if err := auth.ApproveCLISession(sessionCode, user.ID); err != nil {
		t.Fatalf("Error approving session: %v", err)
	}

	// Poll the approved session
	req := dto.CLIPollRequest{
		Code: sessionCode,
	}
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "/auth/cli/poll", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code, w.Body)

	// Session approval should result in a token being provided
	var response dto.TokenResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response.AccessToken, response)
}
