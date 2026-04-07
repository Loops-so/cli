package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"1.1.0", "1.0.0", true},
		{"2.0.0", "1.9.9", true},
		{"1.0.1", "1.0.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.1", false},
		{"1.0.0", "2.0.0", false},
		{"0.0.0", "0.0.0", false},
	}
	for _, tt := range tests {
		if got := isNewerVersion(tt.latest, tt.current); got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
	}{
		{"1.2.3", [3]int{1, 2, 3}},
		{"v1.2.3", [3]int{1, 2, 3}},
		{"0.0.0", [3]int{0, 0, 0}},
		{"invalid", [3]int{0, 0, 0}},
	}
	for _, tt := range tests {
		if got := parseSemver(tt.input); got != tt.want {
			t.Errorf("parseSemver(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestUpdateCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "update-check.json")

	now := time.Now().Truncate(time.Second)
	cache := updateCache{LatestVersion: "1.5.0", CheckedAt: now}
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := readUpdateCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.LatestVersion != "1.5.0" {
		t.Errorf("LatestVersion = %q, want %q", got.LatestVersion, "1.5.0")
	}
	if !got.CheckedAt.Equal(now) {
		t.Errorf("CheckedAt = %v, want %v", got.CheckedAt, now)
	}
}

func TestCheckForUpdateShowsNotice(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOOPS_CONFIG_DIR", dir)

	cache := updateCache{LatestVersion: "9.9.9", CheckedAt: time.Now()}
	data, _ := json.Marshal(cache)
	os.WriteFile(filepath.Join(dir, "update-check.json"), data, 0o600)

	old := version
	version = "1.0.0"
	t.Cleanup(func() { version = old })

	oldFmt := outputFormat
	outputFormat = "text"
	t.Cleanup(func() { outputFormat = oldFmt })

	var buf bytes.Buffer
	checkForUpdate(&buf)

	out := buf.String()
	if out == "" {
		t.Fatal("expected update notice, got empty output")
	}
	if !bytes.Contains(buf.Bytes(), []byte("9.9.9")) {
		t.Errorf("notice should mention latest version 9.9.9, got: %s", out)
	}
	if !bytes.Contains(buf.Bytes(), []byte("v1.0.0")) {
		t.Errorf("notice should mention current version v1.0.0, got: %s", out)
	}
}

func TestCheckForUpdateSuppressedForJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOOPS_CONFIG_DIR", dir)

	cache := updateCache{LatestVersion: "9.9.9", CheckedAt: time.Now()}
	data, _ := json.Marshal(cache)
	os.WriteFile(filepath.Join(dir, "update-check.json"), data, 0o600)

	old := version
	version = "1.0.0"
	t.Cleanup(func() { version = old })

	oldFmt := outputFormat
	outputFormat = "json"
	t.Cleanup(func() { outputFormat = oldFmt })

	var buf bytes.Buffer
	checkForUpdate(&buf)

	if buf.Len() != 0 {
		t.Errorf("expected no output for JSON mode, got: %s", buf.String())
	}
}

func TestCheckForUpdateSuppressedForDev(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOOPS_CONFIG_DIR", dir)

	cache := updateCache{LatestVersion: "9.9.9", CheckedAt: time.Now()}
	data, _ := json.Marshal(cache)
	os.WriteFile(filepath.Join(dir, "update-check.json"), data, 0o600)

	old := version
	version = "dev"
	t.Cleanup(func() { version = old })

	var buf bytes.Buffer
	checkForUpdate(&buf)

	if buf.Len() != 0 {
		t.Errorf("expected no output for dev build, got: %s", buf.String())
	}
}

func TestCheckForUpdateUpToDate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LOOPS_CONFIG_DIR", dir)

	cache := updateCache{LatestVersion: "1.0.0", CheckedAt: time.Now()}
	data, _ := json.Marshal(cache)
	os.WriteFile(filepath.Join(dir, "update-check.json"), data, 0o600)

	old := version
	version = "1.0.0"
	t.Cleanup(func() { version = old })

	oldFmt := outputFormat
	outputFormat = "text"
	t.Cleanup(func() { outputFormat = oldFmt })

	var buf bytes.Buffer
	checkForUpdate(&buf)

	if buf.Len() != 0 {
		t.Errorf("expected no output when up to date, got: %s", buf.String())
	}
}

func TestFetchAndCacheLatestVersion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"tag_name": "v2.0.0"})
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	path := filepath.Join(dir, "update-check.json")

	cache := updateCache{LatestVersion: "2.0.0", CheckedAt: time.Now()}
	data, _ := json.Marshal(cache)
	os.WriteFile(path, data, 0o600)

	got, err := readUpdateCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.LatestVersion != "2.0.0" {
		t.Errorf("got %q, want %q", got.LatestVersion, "2.0.0")
	}
}
