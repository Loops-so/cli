package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ContactProperty struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

func (c *Client) ListContactProperties(customOnly bool) ([]ContactProperty, error) {
	req, err := c.newRequest(http.MethodGet, "/contacts/properties", nil)
	if err != nil {
		return nil, err
	}

	if customOnly {
		q := req.URL.Query()
		q.Set("list", "custom")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromResponse(resp)
	}

	var result []ContactProperty
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
