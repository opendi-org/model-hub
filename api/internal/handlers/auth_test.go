package handlers

import "testing"

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
