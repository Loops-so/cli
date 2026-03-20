package cmd

import (
	"net/http"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestRunAuthStatus(t *testing.T) {
	t.Run("returns config and team name", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"teamName":"Acme"}`)
		cfg, keyResp, err := runAuthStatus()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey == "" {
			t.Error("expected APIKey to be set")
		}
		if cfg.EndpointURL == "" {
			t.Error("expected EndpointURL to be set")
		}
		if keyResp.TeamName != "Acme" {
			t.Errorf("got team %q, want %q", keyResp.TeamName, "Acme")
		}
	})

	t.Run("returns error when no key set", func(t *testing.T) {
		keyring.MockInit()
		_, _, err := runAuthStatus()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"Invalid API key"}`)
		_, _, err := runAuthStatus()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
