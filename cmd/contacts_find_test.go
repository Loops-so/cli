package cmd

import (
	"net/http"
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
