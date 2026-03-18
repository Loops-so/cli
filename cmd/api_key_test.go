package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/loops-so/cli/internal/api"
)

func TestAPIKeyCmd(t *testing.T) {
	t.Run("text: shows team name", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"teamName":"Acme Corp"}`)
		out, err := runCmd(t, "api-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "Acme Corp") {
			t.Errorf("output %q should contain team name", out)
		}
	})

	t.Run("json: valid json with team name", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"teamName":"Acme Corp"}`)
		out, err := runCmd(t, "--output=json", "api-key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result api.APIKeyResponse
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("output is not valid JSON: %v\n%s", err, out)
		}
		if result.TeamName != "Acme Corp" {
			t.Errorf("TeamName = %q, want %q", result.TeamName, "Acme Corp")
		}
	})

	t.Run("api error propagates", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"Invalid API key"}`)
		_, err := runCmd(t, "api-key")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
