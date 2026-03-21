package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListContactProperties(t *testing.T) {
	tests := []struct {
		name       string
		customOnly bool
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantCount  int
		wantQuery  string
	}{
		{
			name:       "success all",
			customOnly: false,
			statusCode: http.StatusOK,
			body:       `[{"key":"firstName","label":"First name","type":"string"},{"key":"score","label":"Score","type":"number"}]`,
			wantCount:  2,
		},
		{
			name:       "success custom",
			customOnly: true,
			statusCode: http.StatusOK,
			body:       `[{"key":"score","label":"Score","type":"number"}]`,
			wantCount:  1,
			wantQuery:  "list=custom",
		},
		{
			name:       "empty",
			customOnly: false,
			statusCode: http.StatusOK,
			body:       `[]`,
			wantCount:  0,
		},
		{
			name:       "unauthorized",
			customOnly: false,
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "invalid json",
			customOnly: false,
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
			props, err := client.ListContactProperties(tt.customOnly)

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
			if len(props) != tt.wantCount {
				t.Errorf("len(props) = %d, want %d", len(props), tt.wantCount)
			}
		})
	}
}

func TestCreateContactProperty(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantBody   map[string]string
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       `{"success":true}`,
			wantBody:   map[string]string{"name": "age", "type": "number"},
		},
		{
			name:       "failure",
			statusCode: http.StatusBadRequest,
			body:       `{"message":"Property already exists"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "Property already exists"},
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBody map[string]string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key")
			err := client.CreateContactProperty("age", "number")

			if tt.wantBody != nil {
				for k, v := range tt.wantBody {
					if gotBody[k] != v {
						t.Errorf("body[%q] = %q, want %q", k, gotBody[k], v)
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

func TestListContactProperties_ResponseData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"key":"favoriteColor","label":"Favorite color","type":"string"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	props, err := client.ListContactProperties(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p := props[0]
	if p.Key != "favoriteColor" {
		t.Errorf("Key = %q, want %q", p.Key, "favoriteColor")
	}
	if p.Label != "Favorite color" {
		t.Errorf("Label = %q, want %q", p.Label, "Favorite color")
	}
	if p.Type != "string" {
		t.Errorf("Type = %q, want %q", p.Type, "string")
	}
}
