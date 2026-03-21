package config

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func setup(t *testing.T) {
	t.Helper()
	keyring.MockInit()
	t.Setenv("LOOPS_CONFIG_DIR", t.TempDir())
}

func TestSave(t *testing.T) {
	t.Run("errors when name is empty", func(t *testing.T) {
		setup(t)
		if err := Save("my-key", ""); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("stores api key in keyring under team name", func(t *testing.T) {
		setup(t)

		if err := Save("my-key", "acme"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := keyring.Get(keyringService, "key:acme")
		if err != nil {
			t.Fatalf("could not read keyring: %v", err)
		}
		if got != "my-key" {
			t.Errorf("got %q, want %q", got, "my-key")
		}
	})

	t.Run("writes config file with active team", func(t *testing.T) {
		setup(t)

		if err := Save("my-key", "acme"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pc, err := LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if pc.ActiveTeam != "acme" {
			t.Errorf("activeTeam: got %q, want %q", pc.ActiveTeam, "acme")
		}
		if len(pc.Teams) != 1 || pc.Teams[0] != "acme" {
			t.Errorf("teams: got %v, want [acme]", pc.Teams)
		}
	})

	t.Run("overwrites existing key value in keyring", func(t *testing.T) {
		setup(t)
		Save("old-key", "acme")

		if err := Save("new-key", "acme"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, err := keyring.Get(keyringService, "key:acme")
		if err != nil {
			t.Fatalf("could not read keyring: %v", err)
		}
		if got != "new-key" {
			t.Errorf("got %q, want %q", got, "new-key")
		}
	})

	t.Run("appends new team without duplicates", func(t *testing.T) {
		setup(t)
		Save("key1", "acme")
		Save("key2", "work")
		Save("key3", "acme") // re-save same name

		pc, err := LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if len(pc.Teams) != 2 {
			t.Errorf("teams: got %v, want [acme work]", pc.Teams)
		}
	})

	t.Run("sets active team to the saved team", func(t *testing.T) {
		setup(t)
		Save("key1", "acme")
		Save("key2", "work")

		pc, err := LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if pc.ActiveTeam != "work" {
			t.Errorf("activeTeam: got %q, want %q", pc.ActiveTeam, "work")
		}
	})
}

func TestSetActiveTeam(t *testing.T) {
	t.Run("sets active team", func(t *testing.T) {
		setup(t)
		Save("key1", "acme")
		Save("key2", "work") // work is now active

		if err := SetActiveTeam("acme"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pc, err := LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if pc.ActiveTeam != "acme" {
			t.Errorf("got %q, want %q", pc.ActiveTeam, "acme")
		}
	})

	t.Run("errors when name not in teams list", func(t *testing.T) {
		setup(t)

		if err := SetActiveTeam("nonexistent"); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("clears active team when name is empty", func(t *testing.T) {
		setup(t)
		Save("key1", "acme")

		if err := SetActiveTeam(""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pc, err := LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if pc.ActiveTeam != "" {
			t.Errorf("got %q, want empty", pc.ActiveTeam)
		}
	})
}

func TestListKeys(t *testing.T) {
	t.Run("returns empty slice when no teams configured", func(t *testing.T) {
		setup(t)

		entries, err := ListKeys()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("got %d entries, want 0", len(entries))
		}
	})

	t.Run("returns keys in config order", func(t *testing.T) {
		setup(t)
		Save("key-a", "acme")
		Save("key-b", "work")

		entries, err := ListKeys()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("got %d entries, want 2", len(entries))
		}
		if entries[0].Name != "acme" || entries[0].APIKey != "key-a" {
			t.Errorf("entry 0: got {%q, %q}, want {acme, key-a}", entries[0].Name, entries[0].APIKey)
		}
		if entries[1].Name != "work" || entries[1].APIKey != "key-b" {
			t.Errorf("entry 1: got {%q, %q}, want {work, key-b}", entries[1].Name, entries[1].APIKey)
		}
	})

	t.Run("returns empty api key when keyring entry missing", func(t *testing.T) {
		setup(t)
		Save("key-a", "acme")
		keyring.Delete(keyringService, "key:acme")

		entries, err := ListKeys()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 1 {
			t.Fatalf("got %d entries, want 1", len(entries))
		}
		if entries[0].APIKey != "" {
			t.Errorf("got %q, want empty string", entries[0].APIKey)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("errors when name is empty", func(t *testing.T) {
		setup(t)
		if err := Delete(""); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("removes the named key from keyring and teams list", func(t *testing.T) {
		setup(t)
		Save("key1", "acme")
		Save("key2", "work")

		if err := Delete("acme"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := keyring.Get(keyringService, "key:acme")
		if err != keyring.ErrNotFound {
			t.Errorf("expected ErrNotFound for acme, got %v", err)
		}

		pc, err := LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if len(pc.Teams) != 1 || pc.Teams[0] != "work" {
			t.Errorf("teams: got %v, want [work]", pc.Teams)
		}
	})

	t.Run("clears activeTeam when the active team is deleted", func(t *testing.T) {
		setup(t)
		Save("key1", "acme")
		Save("key2", "work") // work is now active

		if err := Delete("work"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pc, err := LoadPersistentConfig()
		if err != nil {
			t.Fatalf("LoadPersistentConfig: %v", err)
		}
		if pc.ActiveTeam != "" {
			t.Errorf("activeTeam: got %q, want empty", pc.ActiveTeam)
		}
	})

	t.Run("no error when key not in keyring", func(t *testing.T) {
		setup(t)
		if err := Delete("nonexistent"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
