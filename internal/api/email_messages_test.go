package api

import (
	"errors"
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
