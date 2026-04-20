package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runEmailMessagesGet(cfg *config.Config, id string) (*api.EmailMessage, error) {
	return newAPIClient(cfg).GetEmailMessage(id)
}

var emailMessagesCmd = &cobra.Command{
	Use:   "email-messages",
	Short: "Manage email messages",
}

var emailMessagesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an email message",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		msg, err := runEmailMessagesGet(cfg, args[0])
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), msg)
		}

		w := newTableWriter(cmd.OutOrStdout())
		fmt.Fprintln(w, "FIELD\tVALUE")
		row := func(field, value string) {
			fmt.Fprintf(w, "%s\t%s\n", field, value)
		}
		row("emailMessageId", msg.EmailMessageID)
		row("campaignId", deref(msg.CampaignID))
		row("subject", msg.Subject)
		row("previewText", msg.PreviewText)
		row("fromName", msg.FromName)
		row("fromEmail", msg.FromEmail)
		row("replyToEmail", msg.ReplyToEmail)
		row("contentRevisionId", deref(msg.ContentRevisionID))
		row("updatedAt", msg.UpdatedAt)
		w.Flush()

		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "LMX:")
		fmt.Fprintln(cmd.OutOrStdout(), msg.LMX)

		return nil
	},
}

func init() {
	emailMessagesCmd.AddCommand(emailMessagesGetCmd)
	rootCmd.AddCommand(emailMessagesCmd)
}
