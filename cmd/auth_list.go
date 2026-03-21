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
		entries, err := runAuthList()
		if err != nil {
			return err
		}

		if isJSONOutput() {
			type jsonEntry struct {
				Name   string `json:"name"`
				APIKey string `json:"apiKey"`
			}
			out := make([]jsonEntry, len(entries))
			for i, e := range entries {
				out[i] = jsonEntry{e.Name, maskKey(e.APIKey)}
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
			fmt.Fprintf(w, "%s\t%s\n", e.Name, maskKey(e.APIKey))
		}
		return w.Flush()
	},
}

func runAuthList() ([]config.KeyEntry, error) {
	return config.ListKeys()
}

func init() {
	authCmd.AddCommand(authListCmd)
}
