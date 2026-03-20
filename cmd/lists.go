package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runListsList(cfg *config.Config) ([]api.MailingList, error) {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey).ListMailingLists()
}

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage mailing lists",
}

var listsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List mailing lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		lists, err := runListsList(cfg)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			if lists == nil {
				lists = []api.MailingList{}
			}
			return printJSON(cmd.OutOrStdout(), lists)
		}

		if len(lists) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No mailing lists found.")
			return nil
		}

		w := newTableWriter(cmd.OutOrStdout())
		fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tPUBLIC")
		for _, l := range lists {
			fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", l.ID, l.Name, l.Description, l.IsPublic)
		}
		w.Flush()

		return nil
	},
}

func init() {
	listsCmd.AddCommand(listsListCmd)
	rootCmd.AddCommand(listsCmd)
}
