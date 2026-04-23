package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/loops-so/cli/internal/api"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

func TestRunEmailMessagesUpdate(t *testing.T) {
	body := `{
		"success": true,
		"emailMessageId": "em_abc123",
		"campaignId": "cmp_xyz789",
		"subject": "Updated",
		"previewText": "",
		"fromName": "",
		"fromEmail": "",
		"replyToEmail": "",
		"lmx": "",
		"contentRevisionId": "rev_2",
		"updatedAt": "2026-04-20T11:00:00Z"
	}`

	t.Run("returns updated message", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		msg, err := runEmailMessagesUpdate(cfg(t), "em_abc123", api.UpdateEmailMessageRequest{
			EmailMessageFields: api.EmailMessageFields{Subject: "Updated"},
			Set:                map[string]bool{"subject": true},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if msg.EmailMessageID != "em_abc123" {
			t.Errorf("EmailMessageID = %q, want em_abc123", msg.EmailMessageID)
		}
		if deref(msg.ContentRevisionID) != "rev_2" {
			t.Errorf("ContentRevisionID = %q, want rev_2", deref(msg.ContentRevisionID))
		}
	})

	t.Run("returns error on 409 revision mismatch", func(t *testing.T) {
		serveJSON(t, http.StatusConflict, `{"success":false,"message":"Revision mismatch"}`)
		_, err := runEmailMessagesUpdate(cfg(t), "em_abc123", api.UpdateEmailMessageRequest{
			EmailMessageFields: api.EmailMessageFields{Subject: "Updated"},
			Set:                map[string]bool{"subject": true},
			ExpectedRevisionID: "rev_stale",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestEmailMessageFieldParamsFromCmd(t *testing.T) {
	t.Run("unset flags are absent from Set", func(t *testing.T) {
		cmd := &cobra.Command{}
		addEmailMessageFieldFlags(cmd)
		cmd.ParseFlags([]string{"--subject", "Hello"})

		params, err := emailMessageFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !params.Set["subject"] {
			t.Errorf(`Set["subject"] = false, want true`)
		}
		for _, k := range []string{"previewText", "fromName", "fromEmail", "replyToEmail", "lmx"} {
			if params.Set[k] {
				t.Errorf(`Set[%q] = true, want absent`, k)
			}
		}
		if params.Subject != "Hello" {
			t.Errorf("Subject = %q, want Hello", params.Subject)
		}
	})

	t.Run("empty-string flag still marks field as set", func(t *testing.T) {
		cmd := &cobra.Command{}
		addEmailMessageFieldFlags(cmd)
		cmd.ParseFlags([]string{"--preview-text", ""})

		params, err := emailMessageFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !params.Set["previewText"] {
			t.Error("expected previewText in Set even when empty")
		}
		if params.PreviewText != "" {
			t.Errorf("PreviewText = %q, want empty", params.PreviewText)
		}
	})

	t.Run("from-email strips @domain", func(t *testing.T) {
		cmd := &cobra.Command{}
		addEmailMessageFieldFlags(cmd)
		cmd.ParseFlags([]string{"--from-email", "hello@acme.com"})

		params, err := emailMessageFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.FromEmail != "hello" {
			t.Errorf("FromEmail = %q, want hello", params.FromEmail)
		}
		if !params.Set["fromEmail"] {
			t.Error("expected fromEmail in Set")
		}
	})

	t.Run("lmx-file reads file into LMX and sets lmx", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "msg.lmx")
		if err := os.WriteFile(path, []byte("<Paragraph>From file</Paragraph>"), 0o600); err != nil {
			t.Fatalf("write temp file: %v", err)
		}

		cmd := &cobra.Command{}
		addEmailMessageFieldFlags(cmd)
		cmd.ParseFlags([]string{"--lmx-file", path})

		params, err := emailMessageFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.LMX != "<Paragraph>From file</Paragraph>" {
			t.Errorf("LMX = %q", params.LMX)
		}
		if !params.Set["lmx"] {
			t.Error("expected lmx in Set when --lmx-file is used")
		}
	})

	t.Run("missing lmx-file returns error", func(t *testing.T) {
		cmd := &cobra.Command{}
		addEmailMessageFieldFlags(cmd)
		cmd.ParseFlags([]string{"--lmx-file", "/does/not/exist.lmx"})

		if _, err := emailMessageFieldParamsFromCmd(cmd); err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})
}

func TestResolveExpectedRevisionID(t *testing.T) {
	t.Run("supplied value is returned without any HTTP call", func(t *testing.T) {
		var called bool
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)
		keyring.MockInit()
		t.Setenv("LOOPS_CONFIG_DIR", t.TempDir())
		t.Setenv("LOOPS_API_KEY", "test-key")
		t.Setenv("LOOPS_ENDPOINT_URL", srv.URL)

		got, err := resolveExpectedRevisionID(cfg(t), "em_abc123", "rev_supplied")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "rev_supplied" {
			t.Errorf("got %q, want rev_supplied", got)
		}
		if called {
			t.Error("expected no HTTP call when revision id supplied")
		}
	})

	t.Run("empty supplied triggers GET and returns current contentRevisionId", func(t *testing.T) {
		body := `{
			"success": true,
			"emailMessageId": "em_abc123",
			"campaignId": "cmp_xyz789",
			"subject": "Hello",
			"previewText": "",
			"fromName": "",
			"fromEmail": "",
			"replyToEmail": "",
			"lmx": "",
			"contentRevisionId": "rev_current",
			"updatedAt": "2026-04-20T10:00:00Z"
		}`
		serveJSON(t, http.StatusOK, body)

		got, err := resolveExpectedRevisionID(cfg(t), "em_abc123", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "rev_current" {
			t.Errorf("got %q, want rev_current", got)
		}
	})

	t.Run("GET failure surfaces a wrapped error", func(t *testing.T) {
		serveJSON(t, http.StatusNotFound, `{"success":false,"message":"Email message not found"}`)

		_, err := resolveExpectedRevisionID(cfg(t), "em_missing", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
