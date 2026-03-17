package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type TransactionalEmail struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	LastUpdated   string   `json:"lastUpdated"`
	DataVariables []string `json:"dataVariables"`
}

func (c *Client) ListTransactional(params PaginationParams) ([]TransactionalEmail, *Pagination, error) {
	q := url.Values{}
	if params.PerPage != "" {
		q.Set("perPage", params.PerPage)
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}

	path := "/transactional"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	req, err := c.newRequest(http.MethodGet, path)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errorFromResponse(resp)
	}

	var result struct {
		Pagination Pagination           `json:"pagination"`
		Data       []TransactionalEmail `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, &result.Pagination, nil
}
