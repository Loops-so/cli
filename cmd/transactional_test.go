package cmd

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttachmentFromPath(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "test-*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		content := []byte("hello attachment")
		f.Write(content)
		f.Close()

		a, err := attachmentFromPath(f.Name())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Filename != filepath.Base(f.Name()) {
			t.Errorf("Filename = %q, want %q", a.Filename, filepath.Base(f.Name()))
		}
		if !strings.HasPrefix(a.ContentType, "text/plain") {
			t.Errorf("ContentType = %q, want text/plain prefix", a.ContentType)
		}
		if a.Data != base64.StdEncoding.EncodeToString(content) {
			t.Errorf("Data = %q, want base64 of content", a.Data)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := attachmentFromPath("/nonexistent/file.pdf")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "/nonexistent/file.pdf") {
			t.Errorf("error %q should mention the path", err.Error())
		}
	})

	t.Run("directory", func(t *testing.T) {
		_, err := attachmentFromPath(t.TempDir())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "is a directory") {
			t.Errorf("error %q should mention 'is a directory'", err.Error())
		}
	})
}
