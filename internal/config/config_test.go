package config

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func setup(t *testing.T) {
	t.Helper()
	keyring.MockInit()
}

func TestLoad(t *testing.T) {
	t.Run("errors when no api key is set", func(t *testing.T) {
		setup(t)
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("reads api key from keyring", func(t *testing.T) {
		setup(t)
		keyring.Set(keyringService, keyringUser, "keyring-key")
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "keyring-key" {
			t.Errorf("got %q, want %q", cfg.APIKey, "keyring-key")
		}
	})

	t.Run("env var overrides keyring api key", func(t *testing.T) {
		setup(t)
		keyring.Set(keyringService, keyringUser, "keyring-key")
		t.Setenv("LOOPS_API_KEY", "env-key")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		cfg, err := Load()
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

		cfg, err := Load()
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

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.EndpointURL != "https://custom.example.com/api" {
			t.Errorf("got %q, want %q", cfg.EndpointURL, "https://custom.example.com/api")
		}
	})
}

func TestSave(t *testing.T) {
	t.Run("stores api key in keyring", func(t *testing.T) {
		setup(t)

		if err := Save("my-key"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := keyring.Get(keyringService, keyringUser)
		if err != nil {
			t.Fatalf("could not read keyring: %v", err)
		}
		if got != "my-key" {
			t.Errorf("got %q, want %q", got, "my-key")
		}
	})

	t.Run("overwrites existing api key", func(t *testing.T) {
		setup(t)
		keyring.Set(keyringService, keyringUser, "old-key")

		if err := Save("new-key"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := keyring.Get(keyringService, keyringUser)
		if err != nil {
			t.Fatalf("could not read keyring: %v", err)
		}
		if got != "new-key" {
			t.Errorf("got %q, want %q", got, "new-key")
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("removes api key from keyring", func(t *testing.T) {
		setup(t)
		keyring.Set(keyringService, keyringUser, "my-key")

		if err := Delete(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := keyring.Get(keyringService, keyringUser)
		if err != keyring.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("no error when no credentials stored", func(t *testing.T) {
		setup(t)

		if err := Delete(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
