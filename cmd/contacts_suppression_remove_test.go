package cmd

import (
	"net/http"
	"testing"
)

func TestRunContactsSuppressionRemove(t *testing.T) {
	body := `{"success":true,"message":"Email removed from suppression list.","removalQuota":{"limit":10,"remaining":7}}`

	t.Run("removes by email", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		result, err := runContactsSuppressionRemove(cfg(t), "bob@example.com", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success ||
			result.Message != "Email removed from suppression list." ||
			result.RemovalQuota.Limit != 10 ||
			result.RemovalQuota.Remaining != 7 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("removes by user ID", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		result, err := runContactsSuppressionRemove(cfg(t), "", "user_123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success=true")
		}
	})

	t.Run("returns error when not suppressed", func(t *testing.T) {
		serveJSON(t, http.StatusBadRequest, `{"message":"Contact is not suppressed."}`)
		_, err := runContactsSuppressionRemove(cfg(t), "bob@example.com", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on 404", func(t *testing.T) {
		serveJSON(t, http.StatusNotFound, `{"message":"Contact not found."}`)
		_, err := runContactsSuppressionRemove(cfg(t), "nobody@example.com", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)
		_, err := runContactsSuppressionRemove(cfg(t), "bob@example.com", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
