package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/loops-so/loops-go"
)

func TestRunContactsFind(t *testing.T) {
	body := `[{"id":"cnt_abc123","email":"bob@example.com","firstName":"Bob","lastName":"Smith","source":"api","subscribed":true,"userGroup":"default","userId":"user_123","mailingLists":{},"optInStatus":"accepted","company":"Loops","plan":"pro"}]`

	assertContact := func(t *testing.T, got loops.Contact) {
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
		if got.Custom["company"] != "Loops" || got.Custom["plan"] != "pro" {
			t.Errorf("unexpected custom properties: %v", got.Custom)
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

	t.Run("marshal preserves custom properties", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		contacts, err := runContactsFind(cfg(t), "bob@example.com", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		b, err := json.Marshal(contacts[0])
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		for _, key := range []string{"company", "plan"} {
			if _, ok := raw[key]; !ok {
				t.Errorf("expected custom property %q in marshaled JSON", key)
			}
		}
	})
}

func TestFormatMailingLists(t *testing.T) {
	tests := []struct {
		name string
		in   map[string]bool
		want string
	}{
		{"nil", nil, ""},
		{"empty", map[string]bool{}, ""},
		{"one subscribed", map[string]bool{"list_a": true}, "list_a"},
		{"skips unsubscribed", map[string]bool{"list_a": true, "list_b": false}, "list_a"},
		{"sorted", map[string]bool{"list_c": true, "list_a": true, "list_b": true}, "list_a, list_b, list_c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMailingLists(tt.in)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
