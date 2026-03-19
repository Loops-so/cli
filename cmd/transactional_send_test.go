package cmd

import (
	"net/http"
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
