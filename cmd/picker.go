package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// pickerHeaderLines is the number of leading lines in the rendered styled
// table that fzf should treat as headers (header text + the BorderHeader
// separator emitted by newStyledTable).
const pickerHeaderLines = 2

func addPickFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("pick", false, "Interactively pick a row with fzf")
}

func isPicking(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("pick")
	return v
}

// reject --pick combined with --output json
func validatePickFlags(cmd *cobra.Command) error {
	if isPicking(cmd) && isJSONOutput() {
		return errors.New("--pick cannot be combined with --output json")
	}
	return nil
}

type pickAction struct {
	OnSelect func(rowIdx int) error
}

// render the styled table for headers/rows and prefixes each data line
// with "<idx>\t" so fzf can identify the original row regardless of
// filtering/reordering. header lines pass through unchanged.
func buildPickerInput(headers []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	t := newStyledTable(&buf, headers...)
	for _, r := range rows {
		t.Row(r...)
	}
	if err := t.Render(); err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) <= pickerHeaderLines {
		return nil, errors.New("no rows to pick")
	}

	var out bytes.Buffer
	for i := range pickerHeaderLines {
		out.WriteString(lines[i])
		out.WriteByte('\n')
	}
	for idx, line := range lines[pickerHeaderLines:] {
		fmt.Fprintf(&out, "%d\t%s\n", idx, line)
	}
	return out.Bytes(), nil
}

// parsePickerSelection parses a single fzf selection line of the form
// "<rowIdx>\t<rendered_row>" and validates the index against numRows.
func parsePickerSelection(s string, numRows int) (int, error) {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return 0, errors.New("empty selection")
	}
	prefix, _, ok := strings.Cut(s, "\t")
	if !ok {
		return 0, errors.New("unexpected selection format")
	}
	idx, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, fmt.Errorf("invalid row index: %w", err)
	}
	if idx < 0 || idx >= numRows {
		return 0, fmt.Errorf("row index %d out of range [0, %d)", idx, numRows)
	}
	return idx, nil
}

func runPicker(headers []string, rows [][]string, action pickAction) error {
	if _, err := exec.LookPath("fzf"); err != nil {
		return errors.New("--pick requires fzf to be installed and on PATH")
	}

	input, err := buildPickerInput(headers, rows)
	if err != nil {
		return err
	}

	fzf := exec.Command("fzf",
		"--ansi",
		"--header-lines", strconv.Itoa(pickerHeaderLines),
		"--delimiter", "\t",
		"--with-nth", "2..",
	)
	fzf.Stdin = bytes.NewReader(input)
	fzf.Stderr = os.Stderr
	selBytes, err := fzf.Output()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			switch exitErr.ExitCode() {
			case 1, 130:
				return nil
			}
		}
		return fmt.Errorf("fzf: %w", err)
	}

	if strings.TrimRight(string(selBytes), "\n") == "" {
		return nil
	}
	rowIdx, err := parsePickerSelection(string(selBytes), len(rows))
	if err != nil {
		return err
	}
	return action.OnSelect(rowIdx)
}

func copyColumnAction(rows [][]string, col int, label string, out io.Writer) pickAction {
	return pickAction{OnSelect: func(rowIdx int) error {
		v := rows[rowIdx][col]
		if err := clipboard.WriteAll(v); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		fmt.Fprintf(out, "Copied %s: %s\n", label, v)
		return nil
	}}
}
