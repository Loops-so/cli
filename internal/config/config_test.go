package config

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestLoad(t *testing.T) {
	t.Run("errors when no credentials are set", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		_, err := Load("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("uses team override when provided", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")
		Save("other-key", "other")
		keyring.Set(keyringService, "key:acme", "acme-key")

		cfg, err := Load("acme")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "acme-key" {
			t.Errorf("got %q, want %q", cfg.APIKey, "acme-key")
		}
	})

	t.Run("errors when team override key not in keyring", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "")

		_, err := Load("nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("errors when activeTeam is set but key not in keyring", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		if err := Save("some-key", "acme"); err != nil {
			t.Fatalf("Save: %v", err)
		}
		keyring.Delete(keyringService, "key:acme")

		_, err := Load("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("reads api key from active team", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		if err := Save("keyring-key", "myteam"); err != nil {
			t.Fatalf("Save: %v", err)
		}

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "keyring-key" {
			t.Errorf("got %q, want %q", cfg.APIKey, "keyring-key")
		}
	})

	t.Run("env var overrides keyring api key", func(t *testing.T) {
		setup(t)
		if err := Save("keyring-key", "myteam"); err != nil {
			t.Fatalf("Save: %v", err)
		}
		t.Setenv("LOOPS_API_KEY", "env-key")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "env-key" {
			t.Errorf("got %q, want %q", cfg.APIKey, "env-key")
		}
	})

	t.Run("defaults endpoint URL", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "some-key")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.EndpointURL != DefaultEndpointURL {
			t.Errorf("got %q, want %q", cfg.EndpointURL, DefaultEndpointURL)
		}
	})

	t.Run("env var overrides endpoint URL", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "some-key")
		t.Setenv("LOOPS_ENDPOINT_URL", "https://custom.example.com/api")

		cfg, err := Load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.EndpointURL != "https://custom.example.com/api" {
			t.Errorf("got %q, want %q", cfg.EndpointURL, "https://custom.example.com/api")
		}
	})
}
