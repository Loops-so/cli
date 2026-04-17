package cmd

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
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
		}{version, commit})
	}
	fmt.Fprintf(w, "loops %s (commit: %s)\n", version, commit)
	return nil
}

func init() {
	if info, ok := debug.ReadBuildInfo(); ok && version == "dev" {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = strings.TrimPrefix(info.Main.Version, "v")
		}
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				commit = s.Value[:7]
				break
			}
		}
	}
	rootCmd.AddCommand(versionCmd)
}
