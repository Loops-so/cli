package cmd

import (
	"fmt"

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

		t := newStyledTable(cmd.OutOrStdout(), "NAME", "ACTIVE", "API KEY")
		for _, e := range entries {
			active := ""
			if e.Name == activeTeam {
				active = "*"
			}
			t.Row(e.Name, active, maskKey(e.APIKey))
		}
		return t.Render()
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
