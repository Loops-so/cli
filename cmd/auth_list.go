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
		if err := validatePickFlags(cmd); err != nil {
			return err
		}

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

		headers := []string{"NAME", "ACTIVE", "API KEY"}
		rows := make([][]string, 0, len(entries))
		for _, e := range entries {
			active := ""
			if e.Name == activeTeam {
				active = "*"
			}
			rows = append(rows, []string{e.Name, active, maskKey(e.APIKey)})
		}

		if isPicking(cmd) {
			out := cmd.OutOrStdout()
			return runPicker(headers, rows, pickAction{OnSelect: func(rowIdx int) error {
				name := entries[rowIdx].Name
				if err := runAuthUse(name); err != nil {
					return err
				}
				fmt.Fprintf(out, "Active team set to %q.\n", name)
				return nil
			}})
		}

		t := newStyledTable(cmd.OutOrStdout(), headers...)
		for _, r := range rows {
			t.Row(r...)
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
	addPickFlag(authListCmd)
	authCmd.AddCommand(authListCmd)
}
