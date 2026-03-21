package cmd

import (
	"testing"

	"github.com/loops-so/cli/internal/config"
)

func TestRunAuthUse(t *testing.T) {
	t.Run("sets active team", func(t *testing.T) {
		mockKeyring(t)
		config.Save("key1", "acme")
		config.Save("key2", "work")

		if err := runAuthUse("acme"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pc, err := config.LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if pc.ActiveTeam != "acme" {
			t.Errorf("got %q, want %q", pc.ActiveTeam, "acme")
		}
	})

	t.Run("returns error when name not in teams list", func(t *testing.T) {
		mockKeyring(t)

		if err := runAuthUse("nonexistent"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("clears active team", func(t *testing.T) {
		mockKeyring(t)
		config.Save("key1", "acme")

		if err := runAuthUse(""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pc, err := config.LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if pc.ActiveTeam != "" {
			t.Errorf("got %q, want empty", pc.ActiveTeam)
		}
	})
}
