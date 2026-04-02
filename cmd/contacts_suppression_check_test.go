package cmd

import (
	"net/http"
	"testing"
)

func TestRunContactsSuppressionCheck(t *testing.T) {
	body := `{"contact":{"id":"cnt_abc123","email":"bob@example.com","userId":"user_123"},"isSuppressed":true,"removalQuota":{"limit":10,"remaining":8}}`

	t.Run("checks by email", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		result, err := runContactsSuppressionCheck(cfg(t), "bob@example.com", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Contact.ID != "cnt_abc123" ||
			result.Contact.Email != "bob@example.com" ||
			deref(result.Contact.UserID) != "user_123" ||
			!result.IsSuppressed ||
			result.RemovalQuota.Limit != 10 ||
			result.RemovalQuota.Remaining != 8 {
			t.Errorf("unexpected result: %+v", result)
		}
	})

	t.Run("checks by user ID", func(t *testing.T) {
		serveJSON(t, http.StatusOK, body)
		result, err := runContactsSuppressionCheck(cfg(t), "", "user_123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsSuppressed {
			t.Errorf("expected suppressed=true, got false")
		}
	})

	t.Run("not suppressed", func(t *testing.T) {
		serveJSON(t, http.StatusOK, `{"contact":{"id":"cnt_abc123","email":"bob@example.com"},"isSuppressed":false,"removalQuota":{"limit":10,"remaining":10}}`)
		result, err := runContactsSuppressionCheck(cfg(t), "bob@example.com", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsSuppressed {
			t.Errorf("expected suppressed=false, got true")
		}
	})

	t.Run("returns error on 404", func(t *testing.T) {
		serveJSON(t, http.StatusNotFound, `{"message":"Contact not found."}`)
		_, err := runContactsSuppressionCheck(cfg(t), "nobody@example.com", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		serveJSON(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)
		_, err := runContactsSuppressionCheck(cfg(t), "bob@example.com", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
