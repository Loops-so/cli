package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runCampaignsGet(cfg *config.Config, id string) (*api.Campaign, error) {
	return newAPIClient(cfg).GetCampaign(id)
}

func runCampaignsList(cfg *config.Config, params api.PaginationParams) ([]api.CampaignListItem, error) {
	client := newAPIClient(cfg)
	if params.Cursor != "" {
		campaigns, _, err := client.ListCampaigns(params)
		return campaigns, err
	}
	return api.Paginate(func(cursor string) ([]api.CampaignListItem, *api.Pagination, error) {
		return client.ListCampaigns(api.PaginationParams{
			PerPage: params.PerPage,
			Cursor:  cursor,
		})
	})
}

var campaignsCmd = &cobra.Command{
	Use:    "campaigns",
	Short:  "Manage campaigns",
	Hidden: true,
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
				campaigns = []api.CampaignListItem{}
			}
			return printJSON(cmd.OutOrStdout(), campaigns)
		}

		if len(campaigns) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No campaigns found.")
			return nil
		}

		t := newStyledTable(cmd.OutOrStdout(), "ID", "MESSAGE ID", "NAME", "STATUS", "SUBJECT", "UPDATED")
		for _, c := range campaigns {
			t.Row(
				c.CampaignID,
				deref(c.EmailMessageID),
				c.Name,
				c.Status,
				c.Subject,
				c.UpdatedAt,
			)
		}
		return t.Render()
	},
}

func runCampaignsCreate(cfg *config.Config, req api.CreateCampaignRequest) (*api.CampaignCreateResponse, error) {
	return newAPIClient(cfg).CreateCampaign(req)
}

var campaignsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a draft campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		resp, err := runCampaignsCreate(cfg, api.CreateCampaignRequest{Name: name})
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), resp)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created. (id: %s, emailMessageId: %s, contentRevisionId: %s)\n", resp.CampaignID, deref(resp.EmailMessageID), deref(resp.EmailMessageContentRevisionID))
		return nil
	},
}

func runCampaignsUpdate(cfg *config.Config, id string, req api.UpdateCampaignRequest) (*api.Campaign, error) {
	return newAPIClient(cfg).UpdateCampaign(id, req)
}

var campaignsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a draft campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		c, err := runCampaignsUpdate(cfg, args[0], api.UpdateCampaignRequest{Name: name})
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), c)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Updated. (id: %s)\n\n", c.CampaignID)

		t := newStyledTable(cmd.OutOrStdout(), "FIELD", "VALUE")
		t.Row("campaignId", c.CampaignID)
		t.Row("emailMessageId", deref(c.EmailMessageID))
		t.Row("name", c.Name)
		t.Row("status", c.Status)
		t.Row("createdAt", c.CreatedAt)
		t.Row("updatedAt", c.UpdatedAt)
		return t.Render()
	},
}

var campaignsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a campaign",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		c, err := runCampaignsGet(cfg, args[0])
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), c)
		}

		t := newStyledTable(cmd.OutOrStdout(), "FIELD", "VALUE")
		t.Row("campaignId", c.CampaignID)
		t.Row("emailMessageId", deref(c.EmailMessageID))
		t.Row("name", c.Name)
		t.Row("status", c.Status)
		t.Row("createdAt", c.CreatedAt)
		t.Row("updatedAt", c.UpdatedAt)
		return t.Render()
	},
}

func init() {
	addPaginationFlags(campaignsListCmd)
	campaignsCmd.AddCommand(campaignsListCmd)
	campaignsCmd.AddCommand(campaignsGetCmd)

	campaignsCreateCmd.Flags().StringP("name", "n", "", "Campaign name (required)")
	campaignsCreateCmd.MarkFlagRequired("name")
	campaignsCmd.AddCommand(campaignsCreateCmd)

	campaignsUpdateCmd.Flags().StringP("name", "n", "", "Campaign name (required)")
	campaignsUpdateCmd.MarkFlagRequired("name")
	campaignsCmd.AddCommand(campaignsUpdateCmd)

	rootCmd.AddCommand(campaignsCmd)
}
