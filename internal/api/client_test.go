package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_RetryResetsBody(t *testing.T) {
	origSleep := sleep
	sleep = func(time.Duration) {}
	defer func() { sleep = origSleep }()

	var bodies []string
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		n := attempts.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", false)
	req, _ := client.newRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"hello":"world"}`)))
	resp, err := client.do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if attempts.Load() != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts.Load())
	}
	for i, body := range bodies {
		if body != `{"hello":"world"}` {
			t.Errorf("attempt %d body = %q, want non-empty JSON", i+1, body)
		}
	}
}

func TestNewRequest(t *testing.T) {
	client := NewClient("https://example.com/api/v1", "test-key", false)

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
			req, err := client.newRequest(tt.method, tt.path, nil)
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

			if got := req.Header.Get("User-Agent"); got != "loops-go/dev" {
				t.Errorf("User-Agent = %q, want %q", got, "loops-go/dev")
			}
		})
	}
}

func TestWithUserAgent(t *testing.T) {
	client := NewClient("https://example.com/api/v1", "test-key", false).WithUserAgent("loops-cli/1.2.3")
	req, err := client.newRequest(http.MethodGet, "/api-key", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := req.Header.Get("User-Agent"); got != "loops-cli/1.2.3" {
		t.Errorf("User-Agent = %q, want %q", got, "loops-cli/1.2.3")
	}
}

func TestNewRequest_InvalidURL(t *testing.T) {
	client := NewClient("://bad-url", "test-key", false)
	_, err := client.newRequest(http.MethodGet, "/path", nil)
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestErrorFromResponse(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		body        string
		wantMessage string
	}{
		{
			name:        "reads error field",
			statusCode:  http.StatusBadRequest,
			body:        `{"error":"something went wrong"}`,
			wantMessage: "something went wrong",
		},
		{
			name:        "falls back to message field",
			statusCode:  http.StatusBadRequest,
			body:        `{"message":"something went wrong"}`,
			wantMessage: "something went wrong",
		},
		{
			name:        "prefers error over message",
			statusCode:  http.StatusBadRequest,
			body:        `{"error":"error field","message":"message field"}`,
			wantMessage: "error field",
		},
		{
			name:        "falls back to generic when body is empty",
			statusCode:  http.StatusBadRequest,
			body:        ``,
			wantMessage: "unexpected status: 400",
		},
		{
			name:        "falls back to generic when fields are absent",
			statusCode:  http.StatusBadRequest,
			body:        `{"success":false}`,
			wantMessage: "unexpected status: 400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			apiErr := errorFromResponse(resp)
			if apiErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantMessage)
			}
		})
	}
}

func TestDo_Retries(t *testing.T) {
	origSleep := sleep
	sleep = func(time.Duration) {}
	defer func() { sleep = origSleep }()

	tests := []struct {
		name         string
		responses    []int
		wantStatus   int
		wantAttempts int32
	}{
		{
			name:         "success on first attempt",
			responses:    []int{200},
			wantStatus:   200,
			wantAttempts: 1,
		},
		{
			name:         "retries on 429 then succeeds",
			responses:    []int{429, 200},
			wantStatus:   200,
			wantAttempts: 2,
		},
		{
			name:         "retries on 500 then succeeds",
			responses:    []int{500, 200},
			wantStatus:   200,
			wantAttempts: 2,
		},
		{
			name:         "exhausts retries on persistent 429",
			responses:    []int{429, 429, 429},
			wantStatus:   429,
			wantAttempts: 3,
		},
		{
			name:         "no retry on 401",
			responses:    []int{401},
			wantStatus:   401,
			wantAttempts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			idx := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempts.Add(1)
				w.WriteHeader(tt.responses[idx])
				idx++
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			req, _ := client.newRequest(http.MethodGet, "/", nil)
			resp, err := client.do(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if attempts.Load() != tt.wantAttempts {
				t.Errorf("attempts = %d, want %d", attempts.Load(), tt.wantAttempts)
			}
		})
	}
}
