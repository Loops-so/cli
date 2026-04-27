package cmd

import (
	"net/http"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunCampaignsUpdate(t *testing.T) {
	body := `{
		"success": true,
		"campaignId": "cmp_abc123",
		"emailMessageId": "em_abc123",
		"name": "Renamed",
		"status": "Draft",
		"createdAt": "2026-04-01T10:00:00Z",
		"updatedAt": "2026-04-25T10:00:00Z"
	}`

	t.Run("returns campaign on success", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		c, err := runCampaignsUpdate(cfg(t), "cmp_abc123", api.UpdateCampaignRequest{Name: "Renamed"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.CampaignID != "cmp_abc123" {
			t.Errorf("CampaignID = %q, want cmp_abc123", c.CampaignID)
		}
		if c.Name != "Renamed" {
			t.Errorf("Name = %q, want Renamed", c.Name)
		}
		if deref(c.EmailMessageID) != "em_abc123" {
			t.Errorf("EmailMessageID = %q, want em_abc123", deref(c.EmailMessageID))
		}
	})

	t.Run("returns error when not in draft", func(t *testing.T) {
		serveJSON(t, http.StatusConflict, `{"success":false,"message":"Campaign is not in draft status"}`)
		_, err := runCampaignsUpdate(cfg(t), "cmp_abc123", api.UpdateCampaignRequest{Name: "Renamed"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
