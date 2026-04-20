package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidateModel_ReturnsStructuredDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/v0/validate", ValidateModel())

	body := []byte(`{
  "meta": {
    "uuid": "not-a-valid-uuid",
    "name": 42,
    "version": false
  },
  "runnableModels": [],
  "diagrams": [],
  "controls": [],
  "inputOutputValues": [],
  "evaluatableAssets": []
}`)
	req, _ := http.NewRequest("POST", "/v0/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := payload["error"]; !ok {
		t.Fatalf("expected error field in response")
	}
	details, ok := payload["details"].([]interface{})
	if !ok || len(details) == 0 {
		t.Fatalf("expected non-empty details array; got: %#v", payload["details"])
	}

	first, ok := details[0].(map[string]interface{})
	if !ok {
		t.Fatalf("details[0] is not an object: %#v", details[0])
	}
	if _, ok := first["instancePath"]; !ok {
		t.Fatalf("details[0] missing instancePath: %#v", first)
	}
	if _, ok := first["message"]; !ok {
		t.Fatalf("details[0] missing message: %#v", first)
	}
}
