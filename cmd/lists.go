package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runListsList(cfg *config.Config) ([]api.MailingList, error) {
	return newAPIClient(cfg).ListMailingLists()
}

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage mailing lists",
}

var listsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List mailing lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validatePickFlags(cmd); err != nil {
			return err
		}

		cfg, err := loadConfig()
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

		headers := []string{"ID", "NAME", "DESCRIPTION", "PUBLIC"}
		rows := make([][]string, 0, len(lists))
		for _, l := range lists {
			rows = append(rows, []string{l.ID, l.Name, l.Description, fmt.Sprintf("%v", l.IsPublic)})
		}

		if isPicking(cmd) {
			return runPicker(headers, rows, copyColumnAction(rows, 0, "list ID", cmd.OutOrStdout()))
		}

		t := newStyledTable(cmd.OutOrStdout(), headers...)
		for _, r := range rows {
			t.Row(r...)
		}
		return t.Render()
	},
}

func init() {
	addPickFlag(listsListCmd)
	listsCmd.AddCommand(listsListCmd)
	rootCmd.AddCommand(listsCmd)
}
