package cmd

import (
	"net/http"
	"testing"
)

func TestRunAuthLogout(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{}`)
		runAuthLogin("test-key", "acme", true)
		if err := runAuthLogout("acme"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error when name is empty", func(t *testing.T) {
		mockKeyring(t)
		if err := runAuthLogout(""); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
