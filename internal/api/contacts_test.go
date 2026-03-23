package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFindContacts(t *testing.T) {
	tests := []struct {
		name       string
		params     FindContactParams
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantCount  int
		wantQuery  string
	}{
		{
			name:       "success by email",
			params:     FindContactParams{Email: "bob@example.com"},
			statusCode: http.StatusOK,
			body:       `[{"id":"cnt_abc123","email":"bob@example.com","subscribed":true,"mailingLists":{}}]`,
			wantCount:  1,
			wantQuery:  "email=bob%40example.com",
		},
		{
			name:       "success by userId",
			params:     FindContactParams{UserID: "user_123"},
			statusCode: http.StatusOK,
			body:       `[{"id":"cnt_abc123","email":"bob@example.com","subscribed":true,"mailingLists":{}}]`,
			wantCount:  1,
			wantQuery:  "userId=user_123",
		},
		{
			name:       "empty result",
			params:     FindContactParams{Email: "none@example.com"},
			statusCode: http.StatusOK,
			body:       `[]`,
			wantCount:  0,
		},
		{
			name:       "unauthorized",
			params:     FindContactParams{Email: "bob@example.com"},
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "invalid json",
			params:     FindContactParams{Email: "bob@example.com"},
			statusCode: http.StatusOK,
			body:       `not json`,
			wantErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotQuery = r.URL.RawQuery
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key")
			contacts, err := client.FindContacts(tt.params)

			if tt.wantQuery != "" && gotQuery != tt.wantQuery {
				t.Errorf("query = %q, want %q", gotQuery, tt.wantQuery)
			}

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
			if len(contacts) != tt.wantCount {
				t.Errorf("len(contacts) = %d, want %d", len(contacts), tt.wantCount)
			}
		})
	}
}
