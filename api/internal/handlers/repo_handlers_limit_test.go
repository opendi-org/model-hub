package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateModel_RejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v0/validate", ValidateModel())

	oversized := bytes.Repeat([]byte("a"), int(maxJSONBodyBytes)+1)
	req := httptest.NewRequest(http.MethodPost, "/v0/validate", bytes.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected %d, got %d body=%s", http.StatusRequestEntityTooLarge, w.Code, w.Body.String())
	}
}

