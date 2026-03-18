package api

import (
	"bytes"
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

type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Data        string `json:"data"`
}

type SendTransactionalRequest struct {
	Email           string         `json:"email"`
	TransactionalID string         `json:"transactionalId"`
	AddToAudience   *bool          `json:"addToAudience,omitempty"`
	DataVariables   map[string]any `json:"dataVariables,omitempty"`
	Attachments     []Attachment   `json:"attachments,omitempty"`
}

func (c *Client) SendTransactional(req SendTransactionalRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := c.newRequest(http.MethodPost, "/transactional", bytes.NewReader(b))
	if err != nil {
		return err
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

	req, err := c.newRequest(http.MethodGet, path, nil)
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
