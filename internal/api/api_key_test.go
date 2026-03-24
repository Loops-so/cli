package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAPIKey(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		body          string
		wantAPIErr    *APIError
		wantErrMsg    string
		wantTeam      string
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `{"teamName":"Acme"}`,
			wantTeam:   "Acme",
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"success":false,"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "unexpected status",
			statusCode: http.StatusInternalServerError,
			body:       ``,
			wantAPIErr: &APIError{StatusCode: http.StatusInternalServerError},
		},
		{
			name:       "invalid json",
			statusCode: http.StatusOK,
			body:       `not json`,
			wantErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			result, err := client.GetAPIKey()

			if tt.wantAPIErr != nil {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.wantAPIErr.StatusCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantAPIErr.StatusCode)
				}
				if tt.wantAPIErr.Message != "" && apiErr.Message != tt.wantAPIErr.Message {
					t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantAPIErr.Message)
				}
				return
			}

			if tt.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrMsg)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErrMsg)
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
