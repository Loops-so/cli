package cmd

import (
	"net/http"
	"testing"
)

func TestRunContactsDelete(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"success":true,"message":"Contact deleted."}`)
		err := runContactsDelete(cfg(t), "bob@example.com", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		serveJSON(t, http.StatusNotFound, `{"success":false,"message":"Contact not found."}`)
		err := runContactsDelete(cfg(t), "nobody@example.com", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
