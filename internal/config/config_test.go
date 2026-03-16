package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "loops.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	t.Run("errors when no api key is set", func(t *testing.T) {
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		_, err := load("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("reads api key from config file", func(t *testing.T) {
		path := writeConfigFile(t, "[config]\napi-key = \"file-key\"\n")
		t.Setenv("LOOPS_API_KEY", "")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		cfg, err := load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "file-key" {
			t.Errorf("got %q, want %q", cfg.APIKey, "file-key")
		}
	})

	t.Run("env var overrides config file api key", func(t *testing.T) {
		path := writeConfigFile(t, "[config]\napi-key = \"file-key\"\n")
		t.Setenv("LOOPS_API_KEY", "env-key")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		cfg, err := load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.APIKey != "env-key" {
			t.Errorf("got %q, want %q", cfg.APIKey, "env-key")
		}
	})

	t.Run("defaults endpoint URL", func(t *testing.T) {
		t.Setenv("LOOPS_API_KEY", "some-key")
		t.Setenv("LOOPS_ENDPOINT_URL", "")

		cfg, err := load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.EndpointURL != DefaultEndpointURL {
			t.Errorf("got %q, want %q", cfg.EndpointURL, DefaultEndpointURL)
		}
	})

	t.Run("env var overrides endpoint URL", func(t *testing.T) {
		t.Setenv("LOOPS_API_KEY", "some-key")
		t.Setenv("LOOPS_ENDPOINT_URL", "https://custom.example.com/api")

		cfg, err := load("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.EndpointURL != "https://custom.example.com/api" {
			t.Errorf("got %q, want %q", cfg.EndpointURL, "https://custom.example.com/api")
		}
	})

	t.Run("errors on malformed config file", func(t *testing.T) {
		path := writeConfigFile(t, "not valid toml ][[[")
		t.Setenv("LOOPS_API_KEY", "")

		_, err := load(path)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
