package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runCampaignsList(cfg *config.Config, params api.PaginationParams) ([]api.Campaign, error) {
	client := newAPIClient(cfg)
	if params.Cursor != "" {
		campaigns, _, err := client.ListCampaigns(params)
		return campaigns, err
	}
	return api.Paginate(func(cursor string) ([]api.Campaign, *api.Pagination, error) {
		return client.ListCampaigns(api.PaginationParams{
			PerPage: params.PerPage,
			Cursor:  cursor,
		})
	})
}

var campaignsCmd = &cobra.Command{
	Use:   "campaigns",
	Short: "Manage campaigns",
}

var campaignsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		campaigns, err := runCampaignsList(cfg, paginationParams(cmd))
		if err != nil {
			return err
		}

		if isJSONOutput() {
			if campaigns == nil {
				campaigns = []api.Campaign{}
			}
			return printJSON(cmd.OutOrStdout(), campaigns)
		}

		if len(campaigns) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No campaigns found.")
			return nil
		}

		w := newTableWriter(cmd.OutOrStdout())
		fmt.Fprintln(w, "ID\tMESSAGE ID\tNAME\tSTATUS\tSUBJECT\tUPDATED")
		for _, c := range campaigns {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				c.CampaignID,
				deref(c.EmailMessageID),
				c.Name,
				c.Status,
				c.Subject,
				c.UpdatedAt,
			)
		}
		w.Flush()

		return nil
	},
}

func init() {
	addPaginationFlags(campaignsListCmd)
	campaignsCmd.AddCommand(campaignsListCmd)
	rootCmd.AddCommand(campaignsCmd)
}
