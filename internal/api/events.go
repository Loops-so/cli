package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type SendEventRequest struct {
	Email             string         `json:"-"`
	UserID            string         `json:"-"`
	EventName         string         `json:"-"`
	EventProperties   map[string]any `json:"-"`
	MailingLists      map[string]bool `json:"-"`
	ContactProperties map[string]any `json:"-"`
	IdempotencyKey    string         `json:"-"`
}

func (c *Client) SendEvent(req SendEventRequest) error {
	body := make(map[string]any)
	for k, v := range req.ContactProperties {
		body[k] = v
	}
	body["eventName"] = req.EventName
	if req.Email != "" {
		body["email"] = req.Email
	}
	if req.UserID != "" {
		body["userId"] = req.UserID
	}
	if len(req.EventProperties) > 0 {
		body["eventProperties"] = req.EventProperties
	}
	if len(req.MailingLists) > 0 {
		body["mailingLists"] = req.MailingLists
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := c.newRequest(http.MethodPost, "/events/send", bytes.NewReader(b))
	if err != nil {
		return err
	}
	if req.IdempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)
	}

	resp, err := c.do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errorFromResponse(resp)
	}

	return nil
}
