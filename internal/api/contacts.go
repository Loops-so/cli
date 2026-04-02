package api

import (
	"bytes"
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

type CreateContactRequest struct {
	Email             string
	FirstName         string
	LastName          string
	Source            string
	Subscribed        *bool
	UserGroup         string
	UserID            string
	MailingLists      map[string]bool
	ContactProperties map[string]any
}

func buildContactBody(
	contactProperties map[string]any,
	firstName, lastName string,
	subscribed *bool,
	userGroup string,
	mailingLists map[string]bool,
) map[string]any {
	body := make(map[string]any)
	for k, v := range contactProperties {
		body[k] = v
	}
	if firstName != "" {
		body["firstName"] = firstName
	}
	if lastName != "" {
		body["lastName"] = lastName
	}
	if subscribed != nil {
		body["subscribed"] = *subscribed
	}
	if userGroup != "" {
		body["userGroup"] = userGroup
	}
	if len(mailingLists) > 0 {
		body["mailingLists"] = mailingLists
	}
	return body
}

func (c *Client) CreateContact(req CreateContactRequest) (string, error) {
	body := buildContactBody(req.ContactProperties, req.FirstName, req.LastName, req.Subscribed, req.UserGroup, req.MailingLists)
	body["email"] = req.Email
	if req.Source != "" {
		body["source"] = req.Source
	}
	if req.UserID != "" {
		body["userId"] = req.UserID
	}

	b, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := c.newRequest(http.MethodPost, "/contacts/create", bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	resp, err := c.do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errorFromResponse(resp)
	}

	var result struct {
		Success bool   `json:"success"`
		ID      string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
}

type UpdateContactRequest struct {
	Email             string
	UserID            string
	FirstName         string
	LastName          string
	Subscribed        *bool
	UserGroup         string
	MailingLists      map[string]bool
	ContactProperties map[string]any
}

func (c *Client) UpdateContact(req UpdateContactRequest) error {
	body := buildContactBody(req.ContactProperties, req.FirstName, req.LastName, req.Subscribed, req.UserGroup, req.MailingLists)
	if req.Email != "" {
		body["email"] = req.Email
	}
	if req.UserID != "" {
		body["userId"] = req.UserID
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := c.newRequest(http.MethodPut, "/contacts/update", bytes.NewReader(b))
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

func (c *Client) DeleteContact(email, userID string) error {
	body := make(map[string]any)
	if email != "" {
		body["email"] = email
	}
	if userID != "" {
		body["userId"] = userID
	}

	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := c.newRequest(http.MethodPost, "/contacts/delete", bytes.NewReader(b))
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

type ContactSuppression struct {
	Contact struct {
		ID     string  `json:"id"`
		Email  string  `json:"email"`
		UserID *string `json:"userId"`
	} `json:"contact"`
	IsSuppressed bool `json:"isSuppressed"`
	RemovalQuota struct {
		Limit     int `json:"limit"`
		Remaining int `json:"remaining"`
	} `json:"removalQuota"`
}

type ContactSuppressionRemoval struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RemovalQuota struct {
		Limit     int `json:"limit"`
		Remaining int `json:"remaining"`
	} `json:"removalQuota"`
}

func (c *Client) CheckContactSuppression(email, userID string) (*ContactSuppression, error) {
	req, err := c.newRequest(http.MethodGet, "/contacts/suppression", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if email != "" {
		q.Set("email", email)
	}
	if userID != "" {
		q.Set("userId", userID)
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

	var result ContactSuppression
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) RemoveContactSuppression(email, userID string) (*ContactSuppressionRemoval, error) {
	req, err := c.newRequest(http.MethodDelete, "/contacts/suppression", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if email != "" {
		q.Set("email", email)
	}
	if userID != "" {
		q.Set("userId", userID)
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

	var result ContactSuppressionRemoval
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
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
