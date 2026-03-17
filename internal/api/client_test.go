package api

import (
	"net/http"
	"testing"
)

func TestNewRequest(t *testing.T) {
	client := NewClient("https://example.com/api/v1", "test-key")

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET", http.MethodGet, "/api-key"},
		{"POST", http.MethodPost, "/some-resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := client.newRequest(tt.method, tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if req.Method != tt.method {
				t.Errorf("method = %q, want %q", req.Method, tt.method)
			}

			wantURL := "https://example.com/api/v1" + tt.path
			if req.URL.String() != wantURL {
				t.Errorf("url = %q, want %q", req.URL.String(), wantURL)
			}

			wantAuth := "Bearer test-key"
			if got := req.Header.Get("Authorization"); got != wantAuth {
				t.Errorf("Authorization = %q, want %q", got, wantAuth)
			}
		})
	}
}

func TestNewRequest_InvalidURL(t *testing.T) {
	client := NewClient("://bad-url", "test-key")
	_, err := client.newRequest(http.MethodGet, "/path")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}
