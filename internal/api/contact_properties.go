package api

import (
	"bytes"
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

func (c *Client) CreateContactProperty(name, propType string) error {
	b, err := json.Marshal(map[string]string{"name": name, "type": propType})
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	req, err := c.newRequest(http.MethodPost, "/contacts/properties", bytes.NewReader(b))
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errorFromResponse(resp)
	}

	return nil
}
