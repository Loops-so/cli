package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/loops-so/cli/internal/api"
	"github.com/spf13/cobra"
)

func newFieldCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	addContactFieldFlags(cmd)
	return cmd
}

func TestContactFieldParamsFromCmd(t *testing.T) {
	t.Run("reads string fields", func(t *testing.T) {
		cmd := newFieldCmd(t)
		cmd.Flags().Set("first-name", "Bob")
		cmd.Flags().Set("last-name", "Smith")
		cmd.Flags().Set("user-group", "vip")
		params, err := contactFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.FirstName != "Bob" || params.LastName != "Smith" || params.UserGroup != "vip" {
			t.Errorf("unexpected params: %+v", params)
		}
	})

	t.Run("subscribed not set is nil", func(t *testing.T) {
		params, err := contactFieldParamsFromCmd(newFieldCmd(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.Subscribed != nil {
			t.Errorf("expected Subscribed nil, got %v", *params.Subscribed)
		}
	})

	t.Run("subscribed=true", func(t *testing.T) {
		cmd := newFieldCmd(t)
		cmd.Flags().Set("subscribed", "true")
		params, err := contactFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.Subscribed == nil || *params.Subscribed != true {
			t.Errorf("expected Subscribed=true, got %v", params.Subscribed)
		}
	})

	t.Run("subscribed=false", func(t *testing.T) {
		cmd := newFieldCmd(t)
		cmd.Flags().Set("subscribed", "false")
		params, err := contactFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.Subscribed == nil || *params.Subscribed != false {
			t.Errorf("expected Subscribed=false, got %v", params.Subscribed)
		}
	})

	t.Run("valid --list", func(t *testing.T) {
		cmd := newFieldCmd(t)
		cmd.Flags().Set("list", "abc123=true")
		params, err := contactFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v, ok := params.MailingLists["abc123"]; !ok || !v {
			t.Errorf("expected MailingLists[abc123]=true, got %v", params.MailingLists)
		}
	})

	t.Run("invalid --list returns error", func(t *testing.T) {
		cmd := newFieldCmd(t)
		cmd.Flags().Set("list", "badvalue")
		_, err := contactFieldParamsFromCmd(cmd)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("valid --contact-props", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "props.json")
		os.WriteFile(f, []byte(`{"plan":"pro","score":42}`), 0644)
		cmd := newFieldCmd(t)
		cmd.Flags().Set("contact-props", f)
		params, err := contactFieldParamsFromCmd(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if params.ContactProperties["plan"] != "pro" {
			t.Errorf("expected plan=pro, got %v", params.ContactProperties["plan"])
		}
	})

	t.Run("nonexistent --contact-props returns error", func(t *testing.T) {
		cmd := newFieldCmd(t)
		cmd.Flags().Set("contact-props", "/nonexistent/path.json")
		_, err := contactFieldParamsFromCmd(cmd)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestRunContactsCreate(t *testing.T) {
	t.Run("creates contact and returns ID", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"success":true,"id":"cnt_abc123"}`)
		id, err := runContactsCreate(cfg(t), api.CreateContactRequest{Email: "bob@example.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != "cnt_abc123" {
			t.Errorf("id = %q, want %q", id, "cnt_abc123")
		}
	})

	t.Run("returns error on API failure", func(t *testing.T) {
		serveJSON(t, http.StatusConflict, `{"success":false,"message":"Contact already exists"}`)
		_, err := runContactsCreate(cfg(t), api.CreateContactRequest{Email: "existing@example.com"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
