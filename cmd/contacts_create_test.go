package cmd

import (
	"net/http"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunContactsCreate(t *testing.T) {
	t.Run("creates contact and returns ID", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"success":true,"id":"cnt_abc123"}`)
		id, err := runContactsCreate(cfg(t), api.CreateContactRequest{Email: "bob@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "cnt_abc123" {
			t.Errorf("id = %q, want %q", id, "cnt_abc123")
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		serveJSON(t, http.StatusConflict, `{"success":false,"message":"Contact already exists"}`)
		_, err := runContactsCreate(cfg(t), api.CreateContactRequest{Email: "existing@example.com"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
