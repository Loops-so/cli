package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendEvent(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantAPIErr *APIError
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `{"success":true}`,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"message":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       `{"message":"Event not found"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusNotFound, Message: "Event not found"},
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Email is required"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "Email is required"},
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
			err := client.SendEvent(SendEventRequest{
				Email:     "test@example.com",
				EventName: "signup",
			})

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

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSendEvent_RequestBody(t *testing.T) {
	tests := []struct {
		name        string
		req         SendEventRequest
		wantPresent map[string]any
		wantAbsent  []string
	}{
		{
			name: "email only",
			req:  SendEventRequest{Email: "a@b.com", EventName: "click"},
			wantPresent: map[string]any{
				"email":     "a@b.com",
				"eventName": "click",
			},
			wantAbsent: []string{"userId"},
		},
		{
			name: "userId only",
			req:  SendEventRequest{UserID: "user-123", EventName: "click"},
			wantPresent: map[string]any{
				"userId":    "user-123",
				"eventName": "click",
			},
			wantAbsent: []string{"email"},
		},
		{
			name: "eventProperties",
			req: SendEventRequest{
				Email:           "a@b.com",
				EventName:       "purchase",
				EventProperties: map[string]any{"amount": 42.0, "plan": "pro"},
			},
			wantPresent: map[string]any{
				"eventName": "purchase",
			},
		},
		{
			name: "mailingLists",
			req: SendEventRequest{
				Email:        "a@b.com",
				EventName:    "signup",
				MailingLists: map[string]bool{"list-abc": true, "list-def": false},
			},
			wantPresent: map[string]any{
				"eventName": "signup",
			},
		},
		{
			name: "contact props merged at top level",
			req: SendEventRequest{
				Email:             "a@b.com",
				EventName:         "signup",
				ContactProperties: map[string]any{"firstName": "Alice", "plan": "starter"},
			},
			wantPresent: map[string]any{
				"firstName": "Alice",
				"plan":      "starter",
			},
		},
		{
			name: "contact props do not override named fields",
			req: SendEventRequest{
				Email:             "a@b.com",
				EventName:         "signup",
				ContactProperties: map[string]any{"email": "override@b.com"},
			},
			wantPresent: map[string]any{
				"email": "a@b.com",
			},
		},
		{
			name:       "eventProperties omitted when nil",
			req:        SendEventRequest{Email: "a@b.com", EventName: "click"},
			wantAbsent: []string{"eventProperties"},
		},
		{
			name:       "mailingLists omitted when nil",
			req:        SendEventRequest{Email: "a@b.com", EventName: "click"},
			wantAbsent: []string{"mailingLists"},
		},
		{
			name: "eventName always present",
			req:  SendEventRequest{Email: "a@b.com", EventName: "my-event"},
			wantPresent: map[string]any{
				"eventName": "my-event",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				json.Unmarshal(b, &body)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			client.SendEvent(tt.req)

			for key, want := range tt.wantPresent {
				got, ok := body[key]
				if !ok {
					t.Errorf("key %q missing from body", key)
					continue
				}
				if want != nil && got != want {
					t.Errorf("body[%q] = %v, want %v", key, got, want)
				}
			}

			for _, key := range tt.wantAbsent {
				if _, ok := body[key]; ok {
					t.Errorf("key %q should be absent from body", key)
				}
			}

			// verify eventProperties structure when set
			if tt.req.EventProperties != nil {
				ep, ok := body["eventProperties"].(map[string]any)
				if !ok {
					t.Errorf("eventProperties is not a map: %T", body["eventProperties"])
				} else {
					for k, v := range tt.req.EventProperties {
						if ep[k] != v {
							t.Errorf("eventProperties[%q] = %v, want %v", k, ep[k], v)
						}
					}
				}
			}

			// verify mailingLists structure when set
			if tt.req.MailingLists != nil {
				ml, ok := body["mailingLists"].(map[string]any)
				if !ok {
					t.Errorf("mailingLists is not a map: %T", body["mailingLists"])
				} else {
					for k, v := range tt.req.MailingLists {
						if ml[k] != v {
							t.Errorf("mailingLists[%q] = %v, want %v", k, ml[k], v)
						}
					}
				}
			}
		})
	}
}

func TestSendEvent_IdempotencyKey(t *testing.T) {
	tests := []struct {
		name           string
		idempotencyKey string
		wantHeader     string
	}{
		{
			name:           "sets header when provided",
			idempotencyKey: "my-key-123",
			wantHeader:     "my-key-123",
		},
		{
			name:           "omits header when empty",
			idempotencyKey: "",
			wantHeader:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotHeader string
			var body map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotHeader = r.Header.Get("Idempotency-Key")
				b, _ := io.ReadAll(r.Body)
				json.Unmarshal(b, &body)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			client.SendEvent(SendEventRequest{
				Email:          "a@b.com",
				EventName:      "click",
				IdempotencyKey: tt.idempotencyKey,
			})

			if gotHeader != tt.wantHeader {
				t.Errorf("Idempotency-Key header = %q, want %q", gotHeader, tt.wantHeader)
			}
			if _, ok := body["idempotencyKey"]; ok {
				t.Error("idempotencyKey should not appear in request body")
			}
		})
	}
}
