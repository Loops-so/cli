package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailMessage struct {
	EmailMessageID    string           `json:"emailMessageId"`
	CampaignID        *string          `json:"campaignId"`
	Subject           string           `json:"subject"`
	PreviewText       string           `json:"previewText"`
	FromName          string           `json:"fromName"`
	FromEmail         string           `json:"fromEmail"`
	ReplyToEmail      string           `json:"replyToEmail"`
	LMX               string           `json:"lmx"`
	ContentRevisionID *string          `json:"contentRevisionId"`
	UpdatedAt         string           `json:"updatedAt"`
	Warnings          []map[string]any `json:"warnings,omitempty"`
}

func (c *Client) GetEmailMessage(id string) (*EmailMessage, error) {
	req, err := c.newRequest(http.MethodGet, "/email-messages/"+id, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromResponse(resp)
	}

	var result EmailMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
