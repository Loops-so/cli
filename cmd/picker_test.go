package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestParsePickerSelection(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		numRows int
		want    int
		wantErr bool
	}{
		{name: "valid first", input: "0\tfoo bar\n", numRows: 3, want: 0},
		{name: "valid mid", input: "2\tcell\n", numRows: 5, want: 2},
		{name: "no trailing newline", input: "1\tx", numRows: 5, want: 1},
		{name: "empty", input: "", numRows: 3, wantErr: true},
		{name: "newline only", input: "\n", numRows: 3, wantErr: true},
		{name: "no tab", input: "0nothing", numRows: 3, wantErr: true},
		{name: "non-numeric", input: "abc\tx", numRows: 3, wantErr: true},
		{name: "negative", input: "-1\tx", numRows: 3, wantErr: true},
		{name: "out of range", input: "5\tx", numRows: 3, wantErr: true},
		{name: "zero numRows", input: "0\tx", numRows: 0, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePickerSelection(tc.input, tc.numRows)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parsePickerSelection(%q, %d) error = %v, wantErr = %v", tc.input, tc.numRows, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Fatalf("parsePickerSelection(%q, %d) = %d, want %d", tc.input, tc.numRows, got, tc.want)
			}
		})
	}
}

func TestBuildPickerInput(t *testing.T) {
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"a", "alpha"},
		{"b", "beta"},
		{"c", "gamma"},
	}

	out, err := buildPickerInput(headers, rows)
	if err != nil {
		t.Fatalf("buildPickerInput: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	wantLines := pickerHeaderLines + len(rows)
	if got := len(lines); got != wantLines {
		t.Fatalf("got %d lines, want %d. output: %q", got, wantLines, out)
	}

	// header lines must NOT round-trip through parsePickerSelection
	for i := range pickerHeaderLines {
		if _, err := parsePickerSelection(lines[i], len(rows)); err == nil {
			t.Errorf("header line %d (%q) unexpectedly parsed as a selection", i, lines[i])
		}
	}

	// data lines must round-trip and the index must match position
	for dataIdx, line := range lines[pickerHeaderLines:] {
		idx, err := parsePickerSelection(line, len(rows))
		if err != nil {
			t.Errorf("data line %d (%q) failed to parse: %v", dataIdx, line, err)
			continue
		}
		if idx != dataIdx {
			t.Errorf("data line %d parsed to idx %d, want %d", dataIdx, idx, dataIdx)
		}
	}
}

func TestBuildPickerInputEmpty(t *testing.T) {
	if _, err := buildPickerInput([]string{"ID"}, nil); err == nil {
		t.Fatalf("expected error for empty rows, got nil")
	}
}

func TestValidatePickFlags(t *testing.T) {
	saved := outputFormat
	t.Cleanup(func() { outputFormat = saved })

	makeCmd := func(t *testing.T, pick bool) *cobra.Command {
		t.Helper()
		c := &cobra.Command{}
		addPickFlag(c)
		if pick {
			if err := c.Flags().Set("pick", "true"); err != nil {
				t.Fatalf("set pick flag: %v", err)
			}
		}
		return c
	}

	t.Run("neither set", func(t *testing.T) {
		outputFormat = "text"
		if err := validatePickFlags(makeCmd(t, false)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("pick only", func(t *testing.T) {
		outputFormat = "text"
		if err := validatePickFlags(makeCmd(t, true)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("json only", func(t *testing.T) {
		outputFormat = "json"
		if err := validatePickFlags(makeCmd(t, false)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("both set", func(t *testing.T) {
		outputFormat = "json"
		if err := validatePickFlags(makeCmd(t, true)); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
