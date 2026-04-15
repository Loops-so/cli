package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

type outputFlag string

func (o *outputFlag) Set(s string) error {
	switch s {
	case "text", "json":
		*o = outputFlag(s)
		return nil
	default:
		return fmt.Errorf("must be \"text\" or \"json\"")
	}
}

func (o *outputFlag) String() string { return string(*o) }
func (o *outputFlag) Type() string   { return "format" }

type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

func isJSONOutput() bool {
	return outputFormat == "json"
}

func newTableWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatMailingLists(m map[string]bool) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if v {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

func formatCustomPropLines(m map[string]any) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, len(keys))
	for i, k := range keys {
		lines[i] = fmt.Sprintf("%s=%v", k, m[k])
	}
	return lines
}
