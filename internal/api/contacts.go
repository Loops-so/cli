package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Contact struct {
	ID           string          `json:"id"`
	Email        string          `json:"email"`
	FirstName    *string         `json:"firstName"`
	LastName     *string         `json:"lastName"`
	Source       string          `json:"source"`
	Subscribed   bool            `json:"subscribed"`
	UserGroup    string          `json:"userGroup"`
	UserID       *string         `json:"userId"`
	MailingLists map[string]bool `json:"mailingLists"`
	OptInStatus  *string         `json:"optInStatus"`
}

type FindContactParams struct {
	Email  string
	UserID string
}

func (c *Client) FindContacts(params FindContactParams) ([]Contact, error) {
	req, err := c.newRequest(http.MethodGet, "/contacts/find", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if params.Email != "" {
		q.Set("email", params.Email)
	}
	if params.UserID != "" {
		q.Set("userId", params.UserID)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromResponse(resp)
	}

	var result []Contact
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
