package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersion(cmd.OutOrStdout())
	},
}

func runVersion(w io.Writer) error {
	if isJSONOutput() {
		return printJSON(w, struct {
			Version string `json:"version"`
			Commit  string `json:"commit"`
			Date    string `json:"date"`
		}{version, commit, date})
	}
	fmt.Fprintf(w, "loops version %s (commit: %s, built: %s)\n", version, commit, date)
	return nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
