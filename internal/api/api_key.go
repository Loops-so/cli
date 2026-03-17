package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIKeyResponse struct {
	TeamName string `json:"teamName"`
}

func (c *Client) GetAPIKey() (*APIKeyResponse, error) {
	req, err := c.newRequest(http.MethodGet, "/api-key")
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromResponse(resp)
	}

	var result APIKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
