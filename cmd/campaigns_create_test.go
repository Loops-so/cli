package cmd

import (
	"net/http"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunCampaignsCreate(t *testing.T) {
	body := `{
		"success": true,
		"campaignId": "cmp_new",
		"name": "Spring",
		"status": "Draft",
		"createdAt": "2026-04-20T10:00:00Z",
		"updatedAt": "2026-04-20T10:00:00Z",
		"emailMessage": {
			"emailMessageId": "em_new",
			"campaignId": "cmp_new",
			"subject": "Hello",
			"previewText": "",
			"fromName": "",
			"fromEmail": "",
			"replyToEmail": "",
			"lmx": "",
			"contentRevisionId": "rev_1",
			"updatedAt": "2026-04-20T10:00:00Z"
		}
	}`

	t.Run("returns response on success", func(t *testing.T) {
		serveJSON(t, http.StatusCreated, body)
		resp, err := runCampaignsCreate(cfg(t), api.CreateCampaignRequest{Name: "Spring"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CampaignID != "cmp_new" {
			t.Errorf("CampaignID = %q, want cmp_new", resp.CampaignID)
		}
		if resp.EmailMessage == nil || resp.EmailMessage.EmailMessageID != "em_new" {
			t.Errorf("EmailMessage = %v, want em_new", resp.EmailMessage)
		}
	})

	t.Run("returns error on non-201 response", func(t *testing.T) {
		serveJSON(t, http.StatusBadRequest, `{"success":false,"message":"name is required"}`)
		_, err := runCampaignsCreate(cfg(t), api.CreateCampaignRequest{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
