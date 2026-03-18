package cmd

import "testing"

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
