package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunVersion_Text(t *testing.T) {
	outputFormat = "text"
	var buf bytes.Buffer
	if err := runVersion(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.HasPrefix(got, "loops ") {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestRunVersion_JSON(t *testing.T) {
	outputFormat = "json"
	t.Cleanup(func() { outputFormat = "text" })

	var buf bytes.Buffer
	if err := runVersion(&buf); err != nil {
		t.Fatal(err)
	}

	var got struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Version == "" {
		t.Error("version field is empty")
	}
}
