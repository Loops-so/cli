package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const listTransactionalResponse = `{
	"pagination": {
		"totalResults": 2,
		"returnedResults": 2,
		"perPage": 20,
		"totalPages": 1,
		"nextCursor": "",
		"nextPage": 0
	},
	"data": [
		{
			"id": "abc123",
			"name": "Welcome Email",
			"lastUpdated": "2024-01-15T10:30:00Z",
			"dataVariables": ["name", "email"]
		},
		{
			"id": "def456",
			"name": "Password Reset",
			"lastUpdated": "2024-01-14T08:15:00Z",
			"dataVariables": ["resetLink"]
		}
	]
}`

func TestListTransactional(t *testing.T) {
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
			body:       listTransactionalResponse,
			wantCount:  2,
		},
		{
			name:       "empty list",
			statusCode: http.StatusOK,
			body:       `{"pagination":{"totalResults":0},"data":[]}`,
			wantCount:  0,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"success":false,"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			body:       `{"success":false,"error":"Invalid perPage value"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "Invalid perPage value"},
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

			client := NewClient(server.URL, "test-key")
			emails, pagination, err := client.ListTransactional(PaginationParams{})

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
			if len(emails) != tt.wantCount {
				t.Errorf("len(emails) = %d, want %d", len(emails), tt.wantCount)
			}
			if pagination == nil {
				t.Fatal("expected pagination, got nil")
			}
		})
	}
}

func TestListTransactional_ResponseData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(listTransactionalResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	emails, pagination, err := client.ListTransactional(PaginationParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if emails[0].ID != "abc123" {
		t.Errorf("ID = %q, want %q", emails[0].ID, "abc123")
	}
	if emails[0].Name != "Welcome Email" {
		t.Errorf("Name = %q, want %q", emails[0].Name, "Welcome Email")
	}
	if emails[0].LastUpdated != "2024-01-15T10:30:00Z" {
		t.Errorf("LastUpdated = %q, want %q", emails[0].LastUpdated, "2024-01-15T10:30:00Z")
	}
	if len(emails[0].DataVariables) != 2 || emails[0].DataVariables[0] != "name" {
		t.Errorf("DataVariables = %v, want [name email]", emails[0].DataVariables)
	}
	if pagination.TotalResults != 2 {
		t.Errorf("TotalResults = %d, want 2", pagination.TotalResults)
	}
}

func TestListTransactional_QueryParams(t *testing.T) {
	tests := []struct {
		name       string
		params     PaginationParams
		wantPerPage string
		wantCursor  string
	}{
		{
			name:   "no params",
			params: PaginationParams{},
		},
		{
			name:        "per-page only",
			params:      PaginationParams{PerPage: "50"},
			wantPerPage: "50",
		},
		{
			name:       "cursor only",
			params:     PaginationParams{Cursor: "abc"},
			wantCursor: "abc",
		},
		{
			name:        "both params",
			params:      PaginationParams{PerPage: "10", Cursor: "xyz"},
			wantPerPage: "10",
			wantCursor:  "xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPerPage, gotCursor string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPerPage = r.URL.Query().Get("perPage")
				gotCursor = r.URL.Query().Get("cursor")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"pagination":{},"data":[]}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key")
			client.ListTransactional(tt.params)

			if gotPerPage != tt.wantPerPage {
				t.Errorf("perPage = %q, want %q", gotPerPage, tt.wantPerPage)
			}
			if gotCursor != tt.wantCursor {
				t.Errorf("cursor = %q, want %q", gotCursor, tt.wantCursor)
			}
		})
	}
}

func TestSendTransactional(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
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
			body:       `{"message":"Transactional email not found"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusNotFound, Message: "Transactional email not found"},
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Recipient email is required"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "Recipient email is required"},
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
			err := client.SendTransactional(SendTransactionalRequest{
				Email:           "test@example.com",
				TransactionalID: "abc123",
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

func TestSendTransactional_RequestBody(t *testing.T) {
	addToAudience := true
	tests := []struct {
		name         string
		req          SendTransactionalRequest
		wantEmail    string
		wantID       string
		wantAudience *bool
		wantVars     map[string]any
	}{
		{
			name:      "required fields only",
			req:       SendTransactionalRequest{Email: "a@b.com", TransactionalID: "abc"},
			wantEmail: "a@b.com",
			wantID:    "abc",
		},
		{
			name:         "with add-to-audience",
			req:          SendTransactionalRequest{Email: "a@b.com", TransactionalID: "abc", AddToAudience: &addToAudience},
			wantEmail:    "a@b.com",
			wantID:       "abc",
			wantAudience: &addToAudience,
		},
		{
			name:      "with data variables",
			req:       SendTransactionalRequest{Email: "a@b.com", TransactionalID: "abc", DataVariables: map[string]any{"name": "Alice"}},
			wantEmail: "a@b.com",
			wantID:    "abc",
			wantVars:  map[string]any{"name": "Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got SendTransactionalRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				json.Unmarshal(b, &got)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key")
			client.SendTransactional(tt.req)

			if got.Email != tt.wantEmail {
				t.Errorf("email = %q, want %q", got.Email, tt.wantEmail)
			}
			if got.TransactionalID != tt.wantID {
				t.Errorf("transactionalId = %q, want %q", got.TransactionalID, tt.wantID)
			}
			if tt.wantAudience != nil {
				if got.AddToAudience == nil || *got.AddToAudience != *tt.wantAudience {
					t.Errorf("addToAudience = %v, want %v", got.AddToAudience, tt.wantAudience)
				}
			}
			if tt.wantVars != nil {
				if got.DataVariables["name"] != tt.wantVars["name"] {
					t.Errorf("dataVariables = %v, want %v", got.DataVariables, tt.wantVars)
				}
			}
		})
	}
}
