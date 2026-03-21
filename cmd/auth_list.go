package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stored API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, activeTeam, err := runAuthList()
		if err != nil {
			return err
		}

		if isJSONOutput() {
			type jsonEntry struct {
				Name   string `json:"name"`
				APIKey string `json:"apiKey"`
				Active bool   `json:"active"`
			}
			out := make([]jsonEntry, len(entries))
			for i, e := range entries {
				out[i] = jsonEntry{e.Name, maskKey(e.APIKey), e.Name == activeTeam}
			}
			return printJSON(cmd.OutOrStdout(), out)
		}

		if len(entries) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No keys stored.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tAPI KEY")
		for _, e := range entries {
			name := e.Name
			if e.Name == activeTeam {
				name += " (active)"
			}
			fmt.Fprintf(w, "%s\t%s\n", name, maskKey(e.APIKey))
		}
		return w.Flush()
	},
}

func runAuthList() ([]config.KeyEntry, string, error) {
	entries, err := config.ListKeys()
	if err != nil {
		return nil, "", err
	}
	pc, err := config.LoadPersistentConfig()
	if err != nil {
		return nil, "", err
	}
	return entries, pc.ActiveTeam, nil
}

func init() {
	authCmd.AddCommand(authListCmd)
}
