package cmd

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunContactsFind(t *testing.T) {
	body := `[{"id":"cnt_abc123","email":"bob@example.com","firstName":"Bob","lastName":"Smith","source":"api","subscribed":true,"userGroup":"default","userId":"user_123","mailingLists":{},"optInStatus":"accepted"}]`

	assertContact := func(t *testing.T, got api.Contact) {
		t.Helper()
		if got.ID != "cnt_abc123" ||
			got.Email != "bob@example.com" ||
			deref(got.FirstName) != "Bob" ||
			deref(got.LastName) != "Smith" ||
			got.Source != "api" ||
			!got.Subscribed ||
			got.UserGroup != "default" ||
			deref(got.UserID) != "user_123" ||
			deref(got.OptInStatus) != "accepted" {
			t.Errorf("unexpected contact: %+v", got)
		}
	}

	t.Run("finds by email", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		contacts, err := runContactsFind(cfg(t), "bob@example.com", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(contacts) != 1 {
			t.Fatalf("expected 1 contact, got %d", len(contacts))
		}
		assertContact(t, contacts[0])
	})

	t.Run("finds by user ID", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		contacts, err := runContactsFind(cfg(t), "", "user_123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(contacts) != 1 {
			t.Fatalf("expected 1 contact, got %d", len(contacts))
		}
		assertContact(t, contacts[0])
	})

	t.Run("handles empty result", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `[]`)
		contacts, err := runContactsFind(cfg(t), "notfound@example.com", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(contacts) != 0 {
			t.Errorf("expected empty slice, got %+v", contacts)
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)
		_, err := runContactsFind(cfg(t), "bob@example.com", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestContactPropertyKeys(t *testing.T) {
	var contacts []api.Contact
	body := `[{"id":"cnt_abc123","email":"bob@example.com","subscribed":true,"mailingLists":{},"thisKey":"thisValue","plan":"pro","score":42,"isActive":false}]`
	if err := json.Unmarshal([]byte(body), &contacts); err != nil {
		t.Fatalf("failed to decode contact test fixture: %v", err)
	}

	got := contacts[0].PropertyKeys()
	want := []string{
		"id",
		"email",
		"subscribed",
		"mailingLists",
		"thisKey",
		"plan",
		"score",
		"isActive",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PropertyKeys() = %v, want %v", got, want)
	}

	props := contacts[0].Properties()
	if props["thisKey"] != "thisValue" {
		t.Fatalf("thisKey = %v, want thisValue", props["thisKey"])
	}
	if props["plan"] != "pro" {
		t.Fatalf("plan = %v, want pro", props["plan"])
	}
	if props["score"] != float64(42) {
		t.Fatalf("score = %v, want 42", props["score"])
	}
	if props["isActive"] != false {
		t.Fatalf("isActive = %v, want false", props["isActive"])
	}
}

func TestFormatContactTableValue(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: "null"},
		{name: "string", value: "abc", want: "abc"},
		{name: "bool", value: true, want: "true"},
		{name: "whole number", value: float64(42), want: "42"},
		{name: "decimal", value: float64(42.5), want: "42.5"},
		{name: "object", value: map[string]any{"a": "b"}, want: `{"a":"b"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatContactTableValue(tt.value); got != tt.want {
				t.Fatalf("formatContactTableValue(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestAddSpacesToCamelCase(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		lowerCase bool
		want      string
	}{
		{name: "camel case", input: "firstName", want: "First Name"},
		{name: "single word", input: "email", want: "Email"},
		{name: "lowercase words", input: "optInStatus", lowerCase: true, want: "opt in status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addSpacesToCamelCase(tt.input, tt.lowerCase); got != tt.want {
				t.Fatalf("addSpacesToCamelCase(%q, %v) = %q, want %q", tt.input, tt.lowerCase, got, tt.want)
			}
		})
	}
}
