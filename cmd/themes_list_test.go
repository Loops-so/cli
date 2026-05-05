package cmd

import (
	"net/http"
	"testing"

	"github.com/loops-so/loops-go"
)

func TestRunThemesList(t *testing.T) {
	t.Run("returns themes", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"pagination":{"nextCursor":""},"data":[{"themeId":"thm_1","name":"Default","isDefault":true,"createdAt":"2026-04-01","updatedAt":"2026-04-02","styles":{}}]}`)
		themes, err := runThemesList(cfg(t), loops.PaginationParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(themes) != 1 {
			t.Fatalf("expected 1 theme, got %d", len(themes))
		}
		if themes[0].ThemeID != "thm_1" {
			t.Errorf("ThemeID = %q, want thm_1", themes[0].ThemeID)
		}
		if !themes[0].IsDefault {
			t.Error("IsDefault = false, want true")
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)
		_, err := runThemesList(cfg(t), loops.PaginationParams{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
