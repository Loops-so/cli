package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"
)

const (
	maxRetries = 2
	baseDelay  = 500 * time.Millisecond
)

var sleep = time.Sleep

func isRetryable(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func errorFromResponse(resp *http.Response) *APIError {
	var body struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err == nil {
		if body.Error != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: body.Error}
		}
		if body.Message != "" {
			return &APIError{StatusCode: resp.StatusCode, Message: body.Message}
		}
	}
	return &APIError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("unexpected status: %d", resp.StatusCode)}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<(attempt-1)) * baseDelay
			jitter := time.Duration(rand.Int64N(int64(delay / 2)))
			sleep(delay + jitter)
		}
		resp, err = c.httpClient.Do(req)
		if err != nil {
			if req.Context().Err() != nil {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			continue
		}
		if !isRetryable(resp.StatusCode) {
			return resp, nil
		}
		if attempt < maxRetries {
			resp.Body.Close()
		}
	}
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}
