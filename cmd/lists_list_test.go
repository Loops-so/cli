package cmd

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunListsList(t *testing.T) {
	t.Run("returns lists", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `[{"id":"list_1","name":"Newsletter","description":"Weekly updates","isPublic":true}]`)
		lists, err := runListsList(cfg(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []api.MailingList{
			{ID: "list_1", Name: "Newsletter", Description: "Weekly updates", IsPublic: true},
		}
		if !reflect.DeepEqual(lists, want) {
			t.Errorf("got %+v, want %+v", lists, want)
		}
	})

	t.Run("handles empty array", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `[]`)
		lists, err := runListsList(cfg(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(lists) != 0 {
			t.Errorf("expected empty slice, got %+v", lists)
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)
		_, err := runListsList(cfg(t))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
