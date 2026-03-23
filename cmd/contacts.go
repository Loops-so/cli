package cmd

import (
	"fmt"
	"strconv"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runContactsFind(cfg *config.Config, email, userID string) ([]api.Contact, error) {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey).FindContacts(api.FindContactParams{
		Email:  email,
		UserID: userID,
	})
}

var contactsCmd = &cobra.Command{
	Use:   "contacts",
	Short: "Manage contacts",
}

var contactsFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find a contact by email or user ID",
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

		contacts, err := runContactsFind(cfg, email, userID)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			if contacts == nil {
				contacts = []api.Contact{}
			}
			return printJSON(cmd.OutOrStdout(), contacts)
		}

		if len(contacts) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No contacts found.")
			return nil
		}

		w := newTableWriter(cmd.OutOrStdout())
		fmt.Fprintln(w, "USER ID\tEMAIL\tFIRST NAME\tLAST NAME\tSUBSCRIBED\tSOURCE\tUSER GROUP\tOPT-IN STATUS")
		for _, c := range contacts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				deref(c.UserID),
				c.Email,
				deref(c.FirstName),
				deref(c.LastName),
				strconv.FormatBool(c.Subscribed),
				c.Source,
				c.UserGroup,
				deref(c.OptInStatus),
			)
		}
		w.Flush()

		return nil
	},
}

func init() {
	contactsFindCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsFindCmd.Flags().StringP("user-id", "u", "", "Contact user ID")
	contactsCmd.AddCommand(contactsFindCmd)
	rootCmd.AddCommand(contactsCmd)
}
