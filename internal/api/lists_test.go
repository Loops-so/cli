package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListMailingLists(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantCount  int
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `[{"id":"list_1","name":"Newsletter","description":"Weekly updates","isPublic":true},{"id":"list_2","name":"Announcements","description":"","isPublic":false}]`,
			wantCount:  2,
		},
		{
			name:       "empty list",
			statusCode: http.StatusOK,
			body:       `[]`,
			wantCount:  0,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
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
			lists, err := client.ListMailingLists()

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
			if len(lists) != tt.wantCount {
				t.Errorf("len(lists) = %d, want %d", len(lists), tt.wantCount)
			}
		})
	}
}

func TestListMailingLists_ResponseData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id":"list_1","name":"Newsletter","description":"Weekly updates","isPublic":true}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", false)
	lists, err := client.ListMailingLists()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	l := lists[0]
	if l.ID != "list_1" {
		t.Errorf("ID = %q, want %q", l.ID, "list_1")
	}
	if l.Name != "Newsletter" {
		t.Errorf("Name = %q, want %q", l.Name, "Newsletter")
	}
	if l.Description != "Weekly updates" {
		t.Errorf("Description = %q, want %q", l.Description, "Weekly updates")
	}
	if !l.IsPublic {
		t.Errorf("IsPublic = false, want true")
	}
}
