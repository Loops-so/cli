package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/loops-so/cli/internal/config"
	"github.com/zalando/go-keyring"
)

func cfg(t *testing.T) *config.Config {
	t.Helper()
	c, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	return c
}

func mockKeyring(t *testing.T) {
	t.Helper()
	keyring.MockInit()
}

func serveJSON(t *testing.T, status int, body string) {
	t.Helper()
	keyring.MockInit()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("LOOPS_API_KEY", "test-key")
	t.Setenv("LOOPS_ENDPOINT_URL", srv.URL)
}
