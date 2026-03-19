package cmd

import (
	"testing"
)

func TestRunAuthLogout(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		mockKeyring(t)
		if err := runAuthLogout(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
