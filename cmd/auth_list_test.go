package cmd

import (
	"testing"

	"github.com/loops-so/cli/internal/config"
)

func TestRunAuthList(t *testing.T) {
	t.Run("returns empty list when no keys stored", func(t *testing.T) {
		mockKeyring(t)
		entries, err := runAuthList()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("got %d entries, want 0", len(entries))
		}
	})

	t.Run("returns stored keys", func(t *testing.T) {
		mockKeyring(t)
		config.Save("key-abc1234", "acme")
		config.Save("key-xyz5678", "work")

		entries, err := runAuthList()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("got %d entries, want 2", len(entries))
		}
		if entries[0].Name != "acme" || entries[0].APIKey != "key-abc1234" {
			t.Errorf("entry 0: got {%q, %q}, want {acme, key-abc1234}", entries[0].Name, entries[0].APIKey)
		}
		if entries[1].Name != "work" || entries[1].APIKey != "key-xyz5678" {
			t.Errorf("entry 1: got {%q, %q}, want {work, key-xyz5678}", entries[1].Name, entries[1].APIKey)
		}
	})
}
