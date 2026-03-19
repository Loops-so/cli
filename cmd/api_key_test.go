package cmd

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunAPIKey(t *testing.T) {
	t.Run("returns team name", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"teamName":"Acme Corp"}`)
		result, err := runAPIKey(cfg(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := &api.APIKeyResponse{TeamName: "Acme Corp"}
		if !reflect.DeepEqual(result, want) {
			t.Errorf("got %+v, want %+v", result, want)
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"Invalid API key"}`)
		_, err := runAPIKey(cfg(t))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
