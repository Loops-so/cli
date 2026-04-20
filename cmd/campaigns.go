package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func fromEmailUsername(s string) string {
	before, _, _ := strings.Cut(s, "@")
	return before
}

func runCampaignsGet(cfg *config.Config, id string) (*api.Campaign, error) {
	return newAPIClient(cfg).GetCampaign(id)
}

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

func runCampaignsCreate(cfg *config.Config, req api.CreateCampaignRequest) (*api.CampaignCreateResponse, error) {
	return newAPIClient(cfg).CreateCampaign(req)
}

var campaignsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a draft campaign",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		subject, _ := cmd.Flags().GetString("subject")
		previewText, _ := cmd.Flags().GetString("preview-text")
		fromName, _ := cmd.Flags().GetString("from-name")
		fromEmail, _ := cmd.Flags().GetString("from-email")
		replyTo, _ := cmd.Flags().GetString("reply-to")
		lmx, _ := cmd.Flags().GetString("lmx")
		lmxFile, _ := cmd.Flags().GetString("lmx-file")

		if lmxFile != "" {
			data, err := os.ReadFile(lmxFile)
			if err != nil {
				return fmt.Errorf("read --lmx-file: %w", err)
			}
			lmx = string(data)
		}

		req := api.CreateCampaignRequest{Name: name}
		if subject != "" || previewText != "" || fromName != "" || fromEmail != "" || replyTo != "" || lmx != "" {
			req.EmailMessage = &api.CampaignEmailMessageFields{
				Subject:      subject,
				PreviewText:  previewText,
				FromName:     fromName,
				FromEmail:    fromEmailUsername(fromEmail),
				ReplyToEmail: replyTo,
				LMX:          lmx,
			}
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		resp, err := runCampaignsCreate(cfg, req)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), resp)
		}

		emailMessageID := ""
		if resp.EmailMessage != nil {
			emailMessageID = resp.EmailMessage.EmailMessageID
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created. (id: %s, emailMessageId: %s)\n", resp.CampaignID, emailMessageID)

		if len(resp.Warnings) > 0 {
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Warnings:")
			for _, warn := range resp.Warnings {
				if warn.Path != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s (%s)\n", warn.Rule, warn.Message, warn.Path)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", warn.Rule, warn.Message)
				}
			}
		}

		return nil
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

		w := newTableWriter(cmd.OutOrStdout())
		fmt.Fprintln(w, "FIELD\tVALUE")
		row := func(field, value string) {
			fmt.Fprintf(w, "%s\t%s\n", field, value)
		}
		row("campaignId", c.CampaignID)
		row("emailMessageId", deref(c.EmailMessageID))
		row("name", c.Name)
		row("status", c.Status)
		row("createdAt", c.CreatedAt)
		row("updatedAt", c.UpdatedAt)
		w.Flush()

		return nil
	},
}

func init() {
	addPaginationFlags(campaignsListCmd)
	campaignsCmd.AddCommand(campaignsListCmd)
	campaignsCmd.AddCommand(campaignsGetCmd)

	campaignsCreateCmd.Flags().StringP("name", "n", "", "Campaign name (required)")
	campaignsCreateCmd.Flags().String("subject", "", "Email subject")
	campaignsCreateCmd.Flags().String("preview-text", "", "Email preview text")
	campaignsCreateCmd.Flags().String("from-name", "", "Sender name")
	campaignsCreateCmd.Flags().String("from-email", "", "Username only: a@example.com -> a")
	campaignsCreateCmd.Flags().String("reply-to", "", "Reply-to email address")
	campaignsCreateCmd.Flags().String("lmx", "", "LMX markup (inline)")
	campaignsCreateCmd.Flags().String("lmx-file", "", "Path to a file containing LMX markup")
	campaignsCreateCmd.MarkFlagRequired("name")
	campaignsCreateCmd.MarkFlagsMutuallyExclusive("lmx", "lmx-file")
	campaignsCmd.AddCommand(campaignsCreateCmd)

	rootCmd.AddCommand(campaignsCmd)
}
