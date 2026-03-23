package cmd

import "testing"

func TestDeref(t *testing.T) {
	s := "hello"
	if got := deref(&s); got != "hello" {
		t.Errorf("deref(&s) = %q, want %q", got, "hello")
	}
	if got := deref(nil); got != "" {
		t.Errorf("deref(nil) = %q, want %q", got, "")
	}
}

func TestOutputFlagSet(t *testing.T) {
	for _, valid := range []string{"text", "json"} {
		var f outputFlag
		if err := f.Set(valid); err != nil {
			t.Errorf("Set(%q) unexpected error: %v", valid, err)
		}
		if string(f) != valid {
			t.Errorf("Set(%q) got %q", valid, f)
		}
	}

	for _, invalid := range []string{"yaml", "csv", ""} {
		var f outputFlag
		if err := f.Set(invalid); err == nil {
			t.Errorf("Set(%q) expected error, got nil", invalid)
		}
	}
}
