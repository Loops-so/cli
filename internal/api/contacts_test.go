package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateContact(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name       string
		req        CreateContactRequest
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantID     string
		wantBody   map[string]any
	}{
		{
			name:       "success",
			req:        CreateContactRequest{Email: "bob@example.com"},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantID:     "cnt_abc123",
			wantBody:   map[string]any{"email": "bob@example.com"},
		},
		{
			name: "sends all standard fields",
			req: CreateContactRequest{
				Email:      "bob@example.com",
				FirstName:  "Bob",
				LastName:   "Smith",
				Source:     "api",
				Subscribed: boolPtr(true),
				UserGroup:  "vip",
				UserID:     "user_123",
			},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantID:     "cnt_abc123",
			wantBody: map[string]any{
				"email":      "bob@example.com",
				"firstName":  "Bob",
				"lastName":   "Smith",
				"source":     "api",
				"subscribed": true,
				"userGroup":  "vip",
				"userId":     "user_123",
			},
		},
		{
			name: "merges contact properties at top level",
			req: CreateContactRequest{
				Email:             "bob@example.com",
				ContactProperties: map[string]any{"plan": "pro", "score": float64(42)},
			},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantID:     "cnt_abc123",
			wantBody:   map[string]any{"email": "bob@example.com", "plan": "pro", "score": float64(42)},
		},
		{
			name: "sends mailing lists",
			req: CreateContactRequest{
				Email:        "bob@example.com",
				MailingLists: map[string]bool{"list_abc": true, "list_def": false},
			},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantID:     "cnt_abc123",
		},
		{
			name:       "bad request",
			req:        CreateContactRequest{Email: "notanemail"},
			statusCode: http.StatusBadRequest,
			body:       `{"success":false,"message":"Invalid email address"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "Invalid email address"},
		},
		{
			name:       "conflict",
			req:        CreateContactRequest{Email: "existing@example.com"},
			statusCode: http.StatusConflict,
			body:       `{"success":false,"message":"Contact already exists"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusConflict, Message: "Contact already exists"},
		},
		{
			name:       "unauthorized",
			req:        CreateContactRequest{Email: "bob@example.com"},
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "invalid json response",
			req:        CreateContactRequest{Email: "bob@example.com"},
			statusCode: http.StatusOK,
			body:       `not json`,
			wantErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			id, err := client.CreateContact(tt.req)

			if tt.wantBody != nil {
				for k, v := range tt.wantBody {
					if gotBody[k] != v {
						t.Errorf("body[%q] = %v, want %v", k, gotBody[k], v)
					}
				}
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
			if id != tt.wantID {
				t.Errorf("id = %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestDeleteContact(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		userID     string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantBody   map[string]any
	}{
		{
			name:       "success by email",
			email:      "bob@example.com",
			statusCode: http.StatusOK,
			body:       `{"success":true,"message":"Contact deleted."}`,
			wantBody:   map[string]any{"email": "bob@example.com"},
		},
		{
			name:       "success by userId",
			userID:     "user_123",
			statusCode: http.StatusOK,
			body:       `{"success":true,"message":"Contact deleted."}`,
			wantBody:   map[string]any{"userId": "user_123"},
		},
		{
			name:       "not found",
			email:      "nobody@example.com",
			statusCode: http.StatusNotFound,
			body:       `{"success":false,"message":"Contact not found."}`,
			wantAPIErr: &APIError{StatusCode: http.StatusNotFound, Message: "Contact not found."},
		},
		{
			name:       "unauthorized",
			email:      "bob@example.com",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			err := client.DeleteContact(tt.email, tt.userID)

			if tt.wantBody != nil {
				for k, v := range tt.wantBody {
					if gotBody[k] != v {
						t.Errorf("body[%q] = %v, want %v", k, gotBody[k], v)
					}
				}
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

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateContact(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name       string
		req        UpdateContactRequest
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantBody   map[string]any
	}{
		{
			name:       "success by email",
			req:        UpdateContactRequest{Email: "bob@example.com", FirstName: "Bob"},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantBody:   map[string]any{"email": "bob@example.com", "firstName": "Bob"},
		},
		{
			name:       "success by userId",
			req:        UpdateContactRequest{UserID: "user_123", LastName: "Smith"},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantBody:   map[string]any{"userId": "user_123", "lastName": "Smith"},
		},
		{
			name: "sends all fields",
			req: UpdateContactRequest{
				Email:      "bob@example.com",
				FirstName:  "Bob",
				LastName:   "Smith",
				Subscribed: boolPtr(false),
				UserGroup:  "vip",
			},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantBody: map[string]any{
				"email":      "bob@example.com",
				"firstName":  "Bob",
				"lastName":   "Smith",
				"subscribed": false,
				"userGroup":  "vip",
			},
		},
		{
			name: "merges contact properties at top level",
			req: UpdateContactRequest{
				Email:             "bob@example.com",
				ContactProperties: map[string]any{"plan": "pro"},
			},
			statusCode: http.StatusOK,
			body:       `{"success":true,"id":"cnt_abc123"}`,
			wantBody:   map[string]any{"email": "bob@example.com", "plan": "pro"},
		},
		{
			name:       "bad request",
			req:        UpdateContactRequest{Email: "notanemail"},
			statusCode: http.StatusBadRequest,
			body:       `{"success":false,"message":"Invalid email address"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "Invalid email address"},
		},
		{
			name:       "unauthorized",
			req:        UpdateContactRequest{Email: "bob@example.com"},
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			err := client.UpdateContact(tt.req)

			if tt.wantBody != nil {
				for k, v := range tt.wantBody {
					if gotBody[k] != v {
						t.Errorf("body[%q] = %v, want %v", k, gotBody[k], v)
					}
				}
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
		})
	}
}

func TestCheckContactSuppression(t *testing.T) {
	tests := []struct {
		name         string
		email        string
		userID       string
		statusCode   int
		body         string
		wantAPIErr   *APIError
		wantErrMsg   string
		wantQuery    string
		wantResult   *ContactSuppression
	}{
		{
			name:       "suppressed by email",
			email:      "bob@example.com",
			statusCode: http.StatusOK,
			body:       `{"contact":{"id":"cnt_abc123","email":"bob@example.com","userId":"user_123"},"isSuppressed":true,"removalQuota":{"limit":10,"remaining":8}}`,
			wantQuery:  "email=bob%40example.com",
			wantResult: &ContactSuppression{IsSuppressed: true},
		},
		{
			name:       "not suppressed by userId",
			userID:     "user_123",
			statusCode: http.StatusOK,
			body:       `{"contact":{"id":"cnt_abc123","email":"bob@example.com"},"isSuppressed":false,"removalQuota":{"limit":10,"remaining":10}}`,
			wantQuery:  "userId=user_123",
			wantResult: &ContactSuppression{IsSuppressed: false},
		},
		{
			name:       "decodes quota fields",
			email:      "bob@example.com",
			statusCode: http.StatusOK,
			body:       `{"contact":{"id":"cnt_abc123","email":"bob@example.com"},"isSuppressed":true,"removalQuota":{"limit":10,"remaining":3}}`,
			wantResult: &ContactSuppression{IsSuppressed: true},
		},
		{
			name:       "not found",
			email:      "nobody@example.com",
			statusCode: http.StatusNotFound,
			body:       `{"message":"Contact not found."}`,
			wantAPIErr: &APIError{StatusCode: http.StatusNotFound, Message: "Contact not found."},
		},
		{
			name:       "unauthorized",
			email:      "bob@example.com",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "invalid json",
			email:      "bob@example.com",
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

			client := NewClient(server.URL, "test-key", false)
			result, err := client.CheckContactSuppression(tt.email, tt.userID)

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
			if result.IsSuppressed != tt.wantResult.IsSuppressed {
				t.Errorf("IsSuppressed = %v, want %v", result.IsSuppressed, tt.wantResult.IsSuppressed)
			}
		})
	}
}

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

			client := NewClient(server.URL, "test-key", false)
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
