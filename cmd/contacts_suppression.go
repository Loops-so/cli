package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var contactsSuppressionCmd = &cobra.Command{
	Use:   "suppression",
	Short: "Manage contact suppression",
}

// check

func runContactsSuppressionCheck(cfg *config.Config, email, userID string) (*api.ContactSuppression, error) {
	return newAPIClient(cfg).CheckContactSuppression(email, userID)
}

var contactsSuppressionCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check suppression status for a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		userID, _ := cmd.Flags().GetString("user-id")

		if (email == "") == (userID == "") {
			return fmt.Errorf("exactly one of --email or --user-id is required")
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		result, err := runContactsSuppressionCheck(cfg, email, userID)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), result)
		}

		suppressed := "no"
		if result.IsSuppressed {
			suppressed = "yes"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Suppressed: %s\n", suppressed)
		fmt.Fprintf(cmd.OutOrStdout(), "Removal quota: %d/%d remaining\n", result.RemovalQuota.Remaining, result.RemovalQuota.Limit)
		return nil
	},
}

func init() {
	contactsSuppressionCheckCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsSuppressionCheckCmd.Flags().StringP("user-id", "u", "", "Contact user ID")
	contactsSuppressionCmd.AddCommand(contactsSuppressionCheckCmd)

	contactsCmd.AddCommand(contactsSuppressionCmd)
}
