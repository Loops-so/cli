package cmd

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunTransactionalList(t *testing.T) {
	t.Run("returns emails", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"pagination":{"nextCursor":""},"data":[{"id":"tx_1","name":"Welcome","lastUpdated":"2024-01-01","dataVariables":[]}]}`)
		emails, err := runTransactionalList(cfg(t), api.PaginationParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []api.TransactionalEmail{
			{ID: "tx_1", Name: "Welcome", LastUpdated: "2024-01-01", DataVariables: []string{}},
		}
		if !reflect.DeepEqual(emails, want) {
			t.Errorf("got %+v, want %+v", emails, want)
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)
		_, err := runTransactionalList(cfg(t), api.PaginationParams{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
