package cmd

import (
	"fmt"
	"strconv"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/cmdutil"
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

type contactCreateResult struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

func runContactsCreate(cfg *config.Config, req api.CreateContactRequest) (string, error) {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey).CreateContact(req)
}

var contactsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("email")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		source, _ := cmd.Flags().GetString("source")
		userGroup, _ := cmd.Flags().GetString("user-group")
		userID, _ := cmd.Flags().GetString("user-id")

		req := api.CreateContactRequest{
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			Source:    source,
			UserGroup: userGroup,
			UserID:    userID,
		}

		if cmd.Flags().Changed("subscribed") {
			sub, _ := cmd.Flags().GetBool("subscribed")
			req.Subscribed = &sub
		}

		listPairs, _ := cmd.Flags().GetStringArray("list")
		mailingLists, err := parseMailingLists(listPairs)
		if err != nil {
			return err
		}
		req.MailingLists = mailingLists

		if contactPropsPath, _ := cmd.Flags().GetString("contact-props"); contactPropsPath != "" {
			contactProps, err := cmdutil.ParseJSONFile("contact-props", contactPropsPath)
			if err != nil {
				return err
			}
			req.ContactProperties = contactProps
		}

		id, err := runContactsCreate(cfg, req)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), contactCreateResult{Success: true, ID: id})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created. (id: %s)\n", id)
		return nil
	},
}

func init() {
	contactsFindCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsFindCmd.Flags().StringP("user-id", "u", "", "Contact user ID")
	contactsCmd.AddCommand(contactsFindCmd)

	contactsCreateCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsCreateCmd.Flags().String("first-name", "", "First name")
	contactsCreateCmd.Flags().String("last-name", "", "Last name")
	contactsCreateCmd.Flags().String("source", "", "Source")
	contactsCreateCmd.Flags().BoolP("subscribed", "s", false, "Subscribed status")
	contactsCreateCmd.Flags().String("user-group", "", "User group")
	contactsCreateCmd.Flags().StringP("user-id", "u", "", "User ID")
	contactsCreateCmd.Flags().StringArray("list", nil, "Mailing list subscription as id=true|false (repeatable)")
	contactsCreateCmd.Flags().String("contact-props", "", "Path to a JSON file of contact properties")
	contactsCreateCmd.MarkFlagRequired("email")
	contactsCmd.AddCommand(contactsCreateCmd)

	rootCmd.AddCommand(contactsCmd)
}
