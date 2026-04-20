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

func TestGetEmailMessage(t *testing.T) {
	body := `{
		"success": true,
		"emailMessageId": "em_abc123",
		"campaignId": "cmp_xyz789",
		"subject": "Hello",
		"previewText": "Preview",
		"fromName": "Acme",
		"fromEmail": "hello",
		"replyToEmail": "support@acme.com",
		"lmx": "<Paragraph>Hi</Paragraph><Paragraph>Body text.</Paragraph>",
		"contentRevisionId": "rev_1",
		"updatedAt": "2026-04-20T10:00:00Z"
	}`

	tests := []struct {
		name       string
		id         string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantID     string
	}{
		{
			name:       "success",
			id:         "em_abc123",
			statusCode: http.StatusOK,
			body:       body,
			wantID:     "em_abc123",
		},
		{
			name:       "not found",
			id:         "em_missing",
			statusCode: http.StatusNotFound,
			body:       `{"success":false,"message":"Email message not found"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusNotFound, Message: "Email message not found"},
		},
		{
			name:       "mjml conflict",
			id:         "em_mjml",
			statusCode: http.StatusConflict,
			body:       `{"success":false,"message":"Email message uses MJML format"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusConflict, Message: "Email message uses MJML format"},
		},
		{
			name:       "invalid json",
			id:         "em_abc123",
			statusCode: http.StatusOK,
			body:       `not json`,
			wantErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			result, err := client.GetEmailMessage(tt.id)

			if tt.wantAPIErr != nil {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.wantAPIErr.StatusCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantAPIErr.StatusCode)
				}
				if apiErr.Message != tt.wantAPIErr.Message {
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
			if want := "/email-messages/" + tt.id; gotPath != want {
				t.Errorf("path = %q, want %q", gotPath, want)
			}
			if result.EmailMessageID != tt.wantID {
				t.Errorf("EmailMessageID = %q, want %q", result.EmailMessageID, tt.wantID)
			}
			if result.CampaignID == nil || *result.CampaignID != "cmp_xyz789" {
				t.Errorf("CampaignID = %v, want cmp_xyz789", result.CampaignID)
			}
			if result.Subject != "Hello" {
				t.Errorf("Subject = %q, want Hello", result.Subject)
			}
			if result.LMX != "<Paragraph>Hi</Paragraph><Paragraph>Body text.</Paragraph>" {
				t.Errorf("LMX = %q", result.LMX)
			}
			if result.ContentRevisionID == nil || *result.ContentRevisionID != "rev_1" {
				t.Errorf("ContentRevisionID = %v, want rev_1", result.ContentRevisionID)
			}
		})
	}
}

const updateEmailMessageResponse = `{
	"success": true,
	"emailMessageId": "em_abc123",
	"campaignId": "cmp_xyz789",
	"subject": "Updated",
	"previewText": "new preview",
	"fromName": "Acme",
	"fromEmail": "hello",
	"replyToEmail": "support@acme.com",
	"lmx": "<Paragraph>Hi</Paragraph>",
	"contentRevisionId": "rev_2",
	"updatedAt": "2026-04-20T11:00:00Z",
	"warnings": [
		{"rule":"unknown_attr","severity":"warning","message":"unknown","path":"body.0"}
	]
}`

func TestUpdateEmailMessage(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantAPIErr *APIError
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       updateEmailMessageResponse,
		},
		{
			name:       "revision conflict",
			statusCode: http.StatusConflict,
			body:       `{"success":false,"message":"Revision mismatch"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusConflict, Message: "Revision mismatch"},
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       `{"success":false,"message":"Email message not found"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusNotFound, Message: "Email message not found"},
		},
		{
			name:       "lmx compile failure",
			statusCode: http.StatusUnprocessableEntity,
			body:       `{"success":false,"message":"LMX failed to compile"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnprocessableEntity, Message: "LMX failed to compile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotMethod string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				gotMethod = r.Method
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			req := UpdateEmailMessageRequest{
				EmailMessageFields: EmailMessageFields{Subject: "Updated"},
				Set:                map[string]bool{"subject": true},
			}
			result, err := client.UpdateEmailMessage("em_abc123", req)

			if tt.wantAPIErr != nil {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.wantAPIErr.StatusCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantAPIErr.StatusCode)
				}
				if apiErr.Message != tt.wantAPIErr.Message {
					t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantAPIErr.Message)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotMethod != http.MethodPost {
				t.Errorf("method = %q, want POST", gotMethod)
			}
			if want := "/email-messages/em_abc123"; gotPath != want {
				t.Errorf("path = %q, want %q", gotPath, want)
			}
			if result.EmailMessageID != "em_abc123" {
				t.Errorf("EmailMessageID = %q, want em_abc123", result.EmailMessageID)
			}
			if result.ContentRevisionID == nil || *result.ContentRevisionID != "rev_2" {
				t.Errorf("ContentRevisionID = %v, want rev_2", result.ContentRevisionID)
			}
			if len(result.Warnings) != 1 || result.Warnings[0].Rule != "unknown_attr" {
				t.Errorf("Warnings = %v, want [unknown_attr]", result.Warnings)
			}
		})
	}
}

func TestUpdateEmailMessage_RequestBody(t *testing.T) {
	tests := []struct {
		name       string
		req        UpdateEmailMessageRequest
		wantFields map[string]any
		absent     []string
	}{
		{
			name: "only subject set",
			req: UpdateEmailMessageRequest{
				EmailMessageFields: EmailMessageFields{
					Subject:     "Hello",
					PreviewText: "ignored-not-in-set",
				},
				Set: map[string]bool{"subject": true},
			},
			wantFields: map[string]any{"subject": "Hello"},
			absent:     []string{"previewText", "fromName", "fromEmail", "replyToEmail", "lmx", "expectedRevisionId"},
		},
		{
			name: "blank string is sent when set",
			req: UpdateEmailMessageRequest{
				EmailMessageFields: EmailMessageFields{PreviewText: ""},
				Set:                map[string]bool{"previewText": true},
			},
			wantFields: map[string]any{"previewText": ""},
			absent:     []string{"subject"},
		},
		{
			name: "expected revision id included when non-empty",
			req: UpdateEmailMessageRequest{
				EmailMessageFields: EmailMessageFields{Subject: "Hi"},
				Set:                map[string]bool{"subject": true},
				ExpectedRevisionID: "rev_1",
			},
			wantFields: map[string]any{"subject": "Hi", "expectedRevisionId": "rev_1"},
		},
		{
			name: "all six content fields",
			req: UpdateEmailMessageRequest{
				EmailMessageFields: EmailMessageFields{
					Subject: "s", PreviewText: "p", FromName: "n",
					FromEmail: "u", ReplyToEmail: "r@x.com", LMX: "<p/>",
				},
				Set: map[string]bool{
					"subject": true, "previewText": true, "fromName": true,
					"fromEmail": true, "replyToEmail": true, "lmx": true,
				},
			},
			wantFields: map[string]any{
				"subject": "s", "previewText": "p", "fromName": "n",
				"fromEmail": "u", "replyToEmail": "r@x.com", "lmx": "<p/>",
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
				w.Write([]byte(updateEmailMessageResponse))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			if _, err := client.UpdateEmailMessage("em_abc123", tt.req); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for k, want := range tt.wantFields {
				if body[k] != want {
					t.Errorf("body[%q] = %v, want %v", k, body[k], want)
				}
			}
			for _, k := range tt.absent {
				if _, present := body[k]; present {
					t.Errorf("body[%q] should not be present, got %v", k, body[k])
				}
			}
		})
	}
}
