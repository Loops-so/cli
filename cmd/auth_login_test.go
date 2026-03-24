package cmd

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunAuthLogin(t *testing.T) {
	t.Run("saves key and returns team name", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"teamName":"Acme"}`)
		result, err := runAuthLogin("test-key", "acme", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := &api.APIKeyResponse{TeamName: "Acme"}
		if !reflect.DeepEqual(result, want) {
			t.Errorf("got %+v, want %+v", result, want)
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"Invalid API key"}`)
		_, err := runAuthLogin("bad-key", "acme", false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error when name is empty", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"teamName":"Acme"}`)
		_, err := runAuthLogin("test-key", "", false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("skip-verify saves key without calling api", func(t *testing.T) {
		result, err := runAuthLogin("test-key", "acme", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %+v", result)
		}
	})
}
