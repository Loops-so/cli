package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestRunTransactionalSend(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{}`)
		err := runTransactionalSend(cfg(t), api.SendTransactionalRequest{
			Email:           "user@example.com",
			TransactionalID: "tx_1",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error on api failure", func(t *testing.T) {
		serveJSON(t, http.StatusBadRequest, `{"error":"invalid request"}`)
		err := runTransactionalSend(cfg(t), api.SendTransactionalRequest{
			Email:           "user@example.com",
			TransactionalID: "tx_1",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParseDataVars(t *testing.T) {
	t.Run("var only", func(t *testing.T) {
		m, err := parseDataVars([]string{"name=Alice", "city=NYC"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["name"] != "Alice" || m["city"] != "NYC" {
			t.Errorf("unexpected map: %v", m)
		}
	})

	t.Run("json-vars only", func(t *testing.T) {
		f, _ := os.CreateTemp(t.TempDir(), "vars-*.json")
		json.NewEncoder(f).Encode(map[string]any{"items": []string{"a", "b"}, "count": "2"})
		f.Close()

		m, err := parseDataVars(nil, f.Name())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		items, ok := m["items"].([]any)
		if !ok || len(items) != 2 {
			t.Errorf("expected items slice, got %v", m["items"])
		}
		if m["count"] != "2" {
			t.Errorf("expected count=2, got %v", m["count"])
		}
	})

	t.Run("var overrides json-vars", func(t *testing.T) {
		f, _ := os.CreateTemp(t.TempDir(), "vars-*.json")
		json.NewEncoder(f).Encode(map[string]any{"name": "Bob", "role": "admin"})
		f.Close()

		m, err := parseDataVars([]string{"name=Alice"}, f.Name())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["name"] != "Alice" {
			t.Errorf("expected name=Alice, got %v", m["name"])
		}
		if m["role"] != "admin" {
			t.Errorf("expected role=admin, got %v", m["role"])
		}
	})

	t.Run("var missing equals", func(t *testing.T) {
		_, err := parseDataVars([]string{"badvalue"}, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "KEY=value") {
			t.Errorf("error %q should mention KEY=value", err.Error())
		}
	})

	t.Run("json-vars file not found", func(t *testing.T) {
		_, err := parseDataVars(nil, "/nonexistent/vars.json")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "--json-vars") {
			t.Errorf("error %q should mention --json-vars", err.Error())
		}
	})

	t.Run("json-vars invalid JSON", func(t *testing.T) {
		f, _ := os.CreateTemp(t.TempDir(), "vars-*.json")
		f.WriteString("not json")
		f.Close()

		_, err := parseDataVars(nil, f.Name())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "--json-vars must be a valid JSON object") {
			t.Errorf("error %q should mention --json-vars must be a valid JSON object", err.Error())
		}
	})

	t.Run("both empty returns nil", func(t *testing.T) {
		m, err := parseDataVars(nil, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m != nil {
			t.Errorf("expected nil map, got %v", m)
		}
	})
}
