package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantErr    string
		wantTeam   string
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `{"success":true,"teamName":"Acme"}`,
			wantTeam:   "Acme",
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"success":false,"error":"Invalid API key"}`,
			wantErr:    "invalid API key",
		},
		{
			name:       "unexpected status",
			statusCode: http.StatusInternalServerError,
			body:       ``,
			wantErr:    "unexpected status: 500",
		},
		{
			name:       "invalid json",
			statusCode: http.StatusOK,
			body:       `not json`,
			wantErr:    "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key")
			result, err := client.GetAPIKey()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TeamName != tt.wantTeam {
				t.Errorf("TeamName = %q, want %q", result.TeamName, tt.wantTeam)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
