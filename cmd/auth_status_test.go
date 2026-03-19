package cmd

import (
	"net/http"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestRunAuthStatus(t *testing.T) {
	t.Run("returns config", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{}`)
		cfg, err := runAuthStatus()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey == "" {
			t.Error("expected APIKey to be set")
		}
		if cfg.EndpointURL == "" {
			t.Error("expected EndpointURL to be set")
		}
	})

	t.Run("returns error when no key set", func(t *testing.T) {
		keyring.MockInit()
		_, err := runAuthStatus()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
