package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailMessage struct {
	EmailMessageID    string       `json:"emailMessageId"`
	CampaignID        *string      `json:"campaignId"`
	Subject           string       `json:"subject"`
	PreviewText       string       `json:"previewText"`
	FromName          string       `json:"fromName"`
	FromEmail         string       `json:"fromEmail"`
	ReplyToEmail      string       `json:"replyToEmail"`
	LMX               string       `json:"lmx"`
	ContentRevisionID *string      `json:"contentRevisionId"`
	UpdatedAt         string       `json:"updatedAt"`
	Warnings          []LmxWarning `json:"warnings,omitempty"`
}

type UpdateEmailMessageRequest struct {
	EmailMessageFields
	Set                map[string]bool
	ExpectedRevisionID string
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

func (c *Client) UpdateEmailMessage(id string, req UpdateEmailMessageRequest) (*EmailMessage, error) {
	body := map[string]any{}
	if req.Set["subject"] {
		body["subject"] = req.Subject
	}
	if req.Set["previewText"] {
		body["previewText"] = req.PreviewText
	}
	if req.Set["fromName"] {
		body["fromName"] = req.FromName
	}
	if req.Set["fromEmail"] {
		body["fromEmail"] = req.FromEmail
	}
	if req.Set["replyToEmail"] {
		body["replyToEmail"] = req.ReplyToEmail
	}
	if req.Set["lmx"] {
		body["lmx"] = req.LMX
	}
	if req.ExpectedRevisionID != "" {
		body["expectedRevisionId"] = req.ExpectedRevisionID
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := c.newRequest(http.MethodPost, "/email-messages/"+id, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	resp, err := c.do(httpReq)
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
