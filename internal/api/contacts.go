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
	properties   map[string]any
	propertyKeys []string
}

func (c *Contact) UnmarshalJSON(data []byte) error {
	type contactAlias Contact
	var alias contactAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	var props map[string]any
	if err := json.Unmarshal(data, &props); err != nil {
		return err
	}
	keys, err := jsonObjectKeysInOrder(data)
	if err != nil {
		return err
	}

	*c = Contact(alias)
	c.properties = props
	c.propertyKeys = keys
	return nil
}

func (c Contact) MarshalJSON() ([]byte, error) {
	if len(c.properties) > 0 {
		return json.Marshal(c.properties)
	}

	props := c.Properties()
	return json.Marshal(props)
}

func (c Contact) Properties() map[string]any {
	if len(c.properties) > 0 {
		clone := make(map[string]any, len(c.properties))
		for k, v := range c.properties {
			clone[k] = v
		}
		return clone
	}

	props := map[string]any{
		"id":           c.ID,
		"email":        c.Email,
		"source":       c.Source,
		"subscribed":   c.Subscribed,
		"userGroup":    c.UserGroup,
		"mailingLists": c.MailingLists,
	}
	if c.FirstName != nil {
		props["firstName"] = *c.FirstName
	} else {
		props["firstName"] = nil
	}
	if c.LastName != nil {
		props["lastName"] = *c.LastName
	} else {
		props["lastName"] = nil
	}
	if c.UserID != nil {
		props["userId"] = *c.UserID
	} else {
		props["userId"] = nil
	}
	if c.OptInStatus != nil {
		props["optInStatus"] = *c.OptInStatus
	} else {
		props["optInStatus"] = nil
	}

	return props
}

func (c Contact) PropertyKeys() []string {
	if len(c.propertyKeys) > 0 {
		keys := make([]string, len(c.propertyKeys))
		copy(keys, c.propertyKeys)
		return keys
	}
	keys := make([]string, 0, len(c.properties))
	for key := range c.properties {
		keys = append(keys, key)
	}
	return keys
}

func jsonObjectKeysInOrder(data []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, fmt.Errorf("expected JSON object")
	}

	keys := make([]string, 0)
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected JSON object key")
		}
		keys = append(keys, key)

		var ignored json.RawMessage
		if err := dec.Decode(&ignored); err != nil {
			return nil, err
		}
	}
	if _, err := dec.Token(); err != nil {
		return nil, err
	}
	return keys, nil
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
