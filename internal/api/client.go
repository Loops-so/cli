package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
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
	debug      bool
	userAgent  string
}

func NewClient(baseURL, apiKey string, debug bool) *Client {
	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		debug:      debug,
		userAgent:  "loops-go/dev",
	}
}

func (c *Client) WithUserAgent(ua string) *Client {
	c.userAgent = ua
	return c
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

func (c *Client) logResponse(resp *http.Response) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[debug] Response: %s (body read failed: %v)\n", resp.Status, err)
		resp.Body = io.NopCloser(bytes.NewReader(nil))
		return
	}
	resp.Body = io.NopCloser(bytes.NewReader(raw))
	fmt.Fprintf(os.Stderr, "[debug] Response: %s (%d bytes)\n", resp.Status, len(raw))
	if len(raw) == 0 {
		return
	}
	var pretty bytes.Buffer
	if json.Indent(&pretty, raw, "", "  ") == nil {
		fmt.Fprintf(os.Stderr, "[debug] Body:\n%s\n", pretty.String())
	} else {
		fmt.Fprintf(os.Stderr, "[debug] Body: %s\n", raw)
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to reset request body: %w", err)
				}
				req.Body = body
			}
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
			if c.debug {
				c.logResponse(resp)
			}
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

	var bodyBytes []byte
	if body != nil && c.debug {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.debug {
		fmt.Fprintf(os.Stderr, "[debug] %s %s\n", method, url)
		fmt.Fprintf(os.Stderr, "[debug] Authorization: Bearer [REDACTED]\n")
		if req.Header.Get("Content-Type") != "" {
			fmt.Fprintf(os.Stderr, "[debug] Content-Type: %s\n", req.Header.Get("Content-Type"))
		}
		if len(bodyBytes) > 0 {
			var pretty bytes.Buffer
			if json.Indent(&pretty, bodyBytes, "", "  ") == nil {
				fmt.Fprintf(os.Stderr, "[debug] Body:\n%s\n", pretty.String())
			} else {
				fmt.Fprintf(os.Stderr, "[debug] Body: %s\n", bodyBytes)
			}
		}
	}

	return req, nil
}
