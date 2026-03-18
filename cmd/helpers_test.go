package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zalando/go-keyring"
)

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	outputFormat = "text"
	t.Cleanup(func() { outputFormat = "text" })

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(&bytes.Buffer{})
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
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
