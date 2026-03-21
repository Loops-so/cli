package cmd

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunContactPropertiesCreate(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"success":true}`)
		err := runContactPropertiesCreate(cfg(t), "age", "number")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error on failure", func(t *testing.T) {
		serveJSON(t, http.StatusBadRequest, `{"message":"Property already exists"}`)
		err := runContactPropertiesCreate(cfg(t), "age", "number")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestRunContactPropertiesList(t *testing.T) {
	t.Run("returns properties", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `[{"key":"firstName","label":"First name","type":"string"},{"key":"score","label":"Score","type":"number"}]`)
		props, err := runContactPropertiesList(cfg(t), false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []api.ContactProperty{
			{Key: "firstName", Label: "First name", Type: "string"},
			{Key: "score", Label: "Score", Type: "number"},
		}
		if !reflect.DeepEqual(props, want) {
			t.Errorf("got %+v, want %+v", props, want)
		}
	})

	t.Run("handles empty array", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `[]`)
		props, err := runContactPropertiesList(cfg(t), false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(props) != 0 {
			t.Errorf("expected empty slice, got %+v", props)
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)
		_, err := runContactPropertiesList(cfg(t), false)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
