package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const listCampaignsResponse = `{
	"success": true,
	"pagination": {
		"totalResults": 2,
		"returnedResults": 2,
		"perPage": 20,
		"totalPages": 1,
		"nextCursor": "",
		"nextPage": ""
	},
	"data": [
		{
			"campaignId": "cmp_1",
			"emailMessageId": "em_1",
			"name": "Spring Launch",
			"subject": "New arrivals",
			"status": "Draft",
			"createdAt": "2026-04-01T10:00:00Z",
			"updatedAt": "2026-04-02T10:00:00Z"
		},
		{
			"campaignId": "cmp_2",
			"emailMessageId": null,
			"name": "Summer Sale",
			"subject": "50% off",
			"status": "Sent",
			"createdAt": "2026-03-01T10:00:00Z",
			"updatedAt": "2026-03-05T10:00:00Z"
		}
	]
}`

const createCampaignResponse = `{
	"success": true,
	"campaignId": "cmp_new",
	"name": "Spring Launch",
	"status": "Draft",
	"createdAt": "2026-04-20T10:00:00Z",
	"updatedAt": "2026-04-20T10:00:00Z",
	"emailMessageId": "em_new",
	"emailMessageContentRevisionId": "rev_1"
}`

func TestCreateCampaign(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
	}{
		{
			name:       "success",
			statusCode: http.StatusCreated,
			body:       createCampaignResponse,
		},
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			body:       `{"success":false,"message":"name is required"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "name is required"},
		},
		{
			name:       "invalid json",
			statusCode: http.StatusCreated,
			body:       `not json`,
			wantErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			resp, err := client.CreateCampaign(CreateCampaignRequest{Name: "Spring Launch"})

			if tt.wantAPIErr != nil {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.wantAPIErr.StatusCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantAPIErr.StatusCode)
				}
				if apiErr.Message != tt.wantAPIErr.Message {
					t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantAPIErr.Message)
				}
				return
			}

			if tt.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrMsg)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.CampaignID != "cmp_new" {
				t.Errorf("CampaignID = %q, want cmp_new", resp.CampaignID)
			}
			if resp.EmailMessageID == nil || *resp.EmailMessageID != "em_new" {
				t.Errorf("EmailMessageID = %v, want em_new", resp.EmailMessageID)
			}
			if resp.EmailMessageContentRevisionID == nil || *resp.EmailMessageContentRevisionID != "rev_1" {
				t.Errorf("EmailMessageContentRevisionID = %v, want rev_1", resp.EmailMessageContentRevisionID)
			}
		})
	}
}

func TestCreateCampaign_RequestBody(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		json.Unmarshal(b, &body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(createCampaignResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", false)
	if _, err := client.CreateCampaign(CreateCampaignRequest{Name: "Spring"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if body["name"] != "Spring" {
		t.Errorf("name = %v, want Spring", body["name"])
	}
	if _, hasEmail := body["emailMessage"]; hasEmail {
		t.Errorf("emailMessage should not be present in request body, got %v", body["emailMessage"])
	}
}

func TestGetCampaign(t *testing.T) {
	body := `{
		"success": true,
		"campaignId": "cmp_abc123",
		"emailMessageId": "em_abc123",
		"name": "Spring Launch",
		"status": "Draft",
		"createdAt": "2026-04-01T10:00:00Z",
		"updatedAt": "2026-04-02T10:00:00Z"
	}`

	tests := []struct {
		name       string
		id         string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantID     string
	}{
		{
			name:       "success",
			id:         "cmp_abc123",
			statusCode: http.StatusOK,
			body:       body,
			wantID:     "cmp_abc123",
		},
		{
			name:       "not found",
			id:         "cmp_missing",
			statusCode: http.StatusNotFound,
			body:       `{"success":false,"message":"Campaign not found"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusNotFound, Message: "Campaign not found"},
		},
		{
			name:       "invalid id",
			id:         "bad",
			statusCode: http.StatusBadRequest,
			body:       `{"success":false,"message":"Invalid campaignId"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusBadRequest, Message: "Invalid campaignId"},
		},
		{
			name:       "invalid json",
			id:         "cmp_abc123",
			statusCode: http.StatusOK,
			body:       `not json`,
			wantErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			result, err := client.GetCampaign(tt.id)

			if tt.wantAPIErr != nil {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.wantAPIErr.StatusCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantAPIErr.StatusCode)
				}
				if apiErr.Message != tt.wantAPIErr.Message {
					t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantAPIErr.Message)
				}
				return
			}

			if tt.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrMsg)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if want := "/campaigns/" + tt.id; gotPath != want {
				t.Errorf("path = %q, want %q", gotPath, want)
			}
			if result.CampaignID != tt.wantID {
				t.Errorf("CampaignID = %q, want %q", result.CampaignID, tt.wantID)
			}
			if result.EmailMessageID == nil || *result.EmailMessageID != "em_abc123" {
				t.Errorf("EmailMessageID = %v, want em_abc123", result.EmailMessageID)
			}
			if result.Name != "Spring Launch" {
				t.Errorf("Name = %q, want Spring Launch", result.Name)
			}
			if result.Status != "Draft" {
				t.Errorf("Status = %q, want Draft", result.Status)
			}
		})
	}
}

func TestListCampaigns(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantAPIErr *APIError
		wantErrMsg string
		wantCount  int
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			body:       listCampaignsResponse,
			wantCount:  2,
		},
		{
			name:       "empty list",
			statusCode: http.StatusOK,
			body:       `{"success":true,"pagination":{"totalResults":0},"data":[]}`,
			wantCount:  0,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			body:       `{"success":false,"error":"Invalid API key"}`,
			wantAPIErr: &APIError{StatusCode: http.StatusUnauthorized, Message: "Invalid API key"},
		},
		{
			name:       "invalid json",
			statusCode: http.StatusOK,
			body:       `not json`,
			wantErrMsg: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			campaigns, pagination, err := client.ListCampaigns(PaginationParams{})

			if tt.wantAPIErr != nil {
				var apiErr *APIError
				if !errors.As(err, &apiErr) {
					t.Fatalf("expected *APIError, got %T: %v", err, err)
				}
				if apiErr.StatusCode != tt.wantAPIErr.StatusCode {
					t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.wantAPIErr.StatusCode)
				}
				if tt.wantAPIErr.Message != "" && apiErr.Message != tt.wantAPIErr.Message {
					t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantAPIErr.Message)
				}
				return
			}

			if tt.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrMsg)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErrMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(campaigns) != tt.wantCount {
				t.Errorf("len(campaigns) = %d, want %d", len(campaigns), tt.wantCount)
			}
			if pagination == nil {
				t.Fatal("expected pagination, got nil")
			}
		})
	}
}

func TestListCampaigns_ResponseData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(listCampaignsResponse))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", false)
	campaigns, _, err := client.ListCampaigns(PaginationParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if campaigns[0].CampaignID != "cmp_1" {
		t.Errorf("CampaignID = %q, want cmp_1", campaigns[0].CampaignID)
	}
	if campaigns[0].EmailMessageID == nil || *campaigns[0].EmailMessageID != "em_1" {
		t.Errorf("EmailMessageID = %v, want em_1", campaigns[0].EmailMessageID)
	}
	if campaigns[0].Status != "Draft" {
		t.Errorf("Status = %q, want Draft", campaigns[0].Status)
	}
	if campaigns[1].EmailMessageID != nil {
		t.Errorf("expected nil EmailMessageID, got %v", campaigns[1].EmailMessageID)
	}
}

func TestListCampaigns_QueryParams(t *testing.T) {
	tests := []struct {
		name        string
		params      PaginationParams
		wantPerPage string
		wantCursor  string
	}{
		{
			name:   "no params",
			params: PaginationParams{},
		},
		{
			name:        "both params",
			params:      PaginationParams{PerPage: "10", Cursor: "xyz"},
			wantPerPage: "10",
			wantCursor:  "xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPerPage, gotCursor string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPerPage = r.URL.Query().Get("perPage")
				gotCursor = r.URL.Query().Get("cursor")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"pagination":{},"data":[]}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-key", false)
			client.ListCampaigns(tt.params)

			if gotPerPage != tt.wantPerPage {
				t.Errorf("perPage = %q, want %q", gotPerPage, tt.wantPerPage)
			}
			if gotCursor != tt.wantCursor {
				t.Errorf("cursor = %q, want %q", gotCursor, tt.wantCursor)
			}
		})
	}
}
