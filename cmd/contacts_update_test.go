package cmd

import (
	"net/http"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunContactsUpdate(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"success":true,"id":"cnt_abc123"}`)
		err := runContactsUpdate(cfg(t), api.UpdateContactRequest{
			Email:     "bob@example.com",
			FirstName: "Bob",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		serveJSON(t, http.StatusBadRequest, `{"success":false,"message":"Invalid email address"}`)
		err := runContactsUpdate(cfg(t), api.UpdateContactRequest{Email: "notanemail"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
