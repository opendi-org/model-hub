package services

import (
	"testing"
	"time"
)

func TestIssueAndVerifyAccessToken(t *testing.T) {
	token, err := IssueAccessToken(123, "alice", "test-secret", time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	claims, err := VerifyAccessToken(token, "test-secret")
	if err != nil {
		t.Fatalf("VerifyAccessToken failed: %v", err)
	}
	if claims.UserID != 123 {
		t.Fatalf("expected uid=123, got %d", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Fatalf("expected username=alice, got %q", claims.Username)
	}
}

