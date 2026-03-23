package cmd

import (
	"fmt"
	"strconv"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/cmdutil"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

// params common to create and update
type contactFieldParams struct {
	FirstName         string
	LastName          string
	Subscribed        *bool
	UserGroup         string
	MailingLists      map[string]bool
	ContactProperties map[string]any
}

// flags common to create and update
func addContactFieldFlags(cmd *cobra.Command) {
	cmd.Flags().String("first-name", "", "First name")
	cmd.Flags().String("last-name", "", "Last name")
	cmd.Flags().BoolP("subscribed", "s", false, "Subscribed status")
	cmd.Flags().String("user-group", "", "User group")
	cmd.Flags().StringArray("list", nil, "Mailing list subscription as id=true|false (repeatable)")
	cmd.Flags().String("contact-props", "", "Path to a JSON file of contact properties")
}

// read common flags
func contactFieldParamsFromCmd(cmd *cobra.Command) (contactFieldParams, error) {
	firstName, _ := cmd.Flags().GetString("first-name")
	lastName, _ := cmd.Flags().GetString("last-name")
	userGroup, _ := cmd.Flags().GetString("user-group")

	params := contactFieldParams{
		FirstName: firstName,
		LastName:  lastName,
		UserGroup: userGroup,
	}

	if cmd.Flags().Changed("subscribed") {
		sub, _ := cmd.Flags().GetBool("subscribed")
		params.Subscribed = &sub
	}

	listPairs, _ := cmd.Flags().GetStringArray("list")
	mailingLists, err := parseMailingLists(listPairs)
	if err != nil {
		return params, err
	}
	params.MailingLists = mailingLists

	if contactPropsPath, _ := cmd.Flags().GetString("contact-props"); contactPropsPath != "" {
		contactProps, err := cmdutil.ParseJSONFile("contact-props", contactPropsPath)
		if err != nil {
			return params, err
		}
		params.ContactProperties = contactProps
	}

	return params, nil
}

// find

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

// create

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
		source, _ := cmd.Flags().GetString("source")
		userID, _ := cmd.Flags().GetString("user-id")

		fields, err := contactFieldParamsFromCmd(cmd)
		if err != nil {
			return err
		}

		id, err := runContactsCreate(cfg, api.CreateContactRequest{
			Email:             email,
			FirstName:         fields.FirstName,
			LastName:          fields.LastName,
			Source:            source,
			Subscribed:        fields.Subscribed,
			UserGroup:         fields.UserGroup,
			UserID:            userID,
			MailingLists:      fields.MailingLists,
			ContactProperties: fields.ContactProperties,
		})
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

// update

func runContactsUpdate(cfg *config.Config, req api.UpdateContactRequest) error {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey).UpdateContact(req)
}

var contactsUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a contact",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		userID, _ := cmd.Flags().GetString("user-id")

		if email == "" && userID == "" {
			return fmt.Errorf("at least one of --email or --user-id is required")
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		fields, err := contactFieldParamsFromCmd(cmd)
		if err != nil {
			return err
		}

		if err := runContactsUpdate(cfg, api.UpdateContactRequest{
			Email:             email,
			UserID:            userID,
			FirstName:         fields.FirstName,
			LastName:          fields.LastName,
			Subscribed:        fields.Subscribed,
			UserGroup:         fields.UserGroup,
			MailingLists:      fields.MailingLists,
			ContactProperties: fields.ContactProperties,
		}); err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), Result{Success: true})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Updated.")
		return nil
	},
}

// delete

func runContactsDelete(cfg *config.Config, email, userID string) error {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey).DeleteContact(email, userID)
}

var contactsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a contact",
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

		if err := runContactsDelete(cfg, email, userID); err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), Result{Success: true})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Deleted.")
		return nil
	},
}

func init() {
	contactsFindCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsFindCmd.Flags().StringP("user-id", "u", "", "Contact user ID")
	contactsCmd.AddCommand(contactsFindCmd)

	contactsCreateCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsCreateCmd.Flags().String("source", "", "Source")
	contactsCreateCmd.Flags().StringP("user-id", "u", "", "User ID")
	addContactFieldFlags(contactsCreateCmd)
	contactsCreateCmd.MarkFlagRequired("email")
	contactsCmd.AddCommand(contactsCreateCmd)

	contactsUpdateCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsUpdateCmd.Flags().StringP("user-id", "u", "", "User ID")
	addContactFieldFlags(contactsUpdateCmd)
	contactsCmd.AddCommand(contactsUpdateCmd)

	contactsDeleteCmd.Flags().StringP("email", "e", "", "Contact email address")
	contactsDeleteCmd.Flags().StringP("user-id", "u", "", "Contact user ID")
	contactsCmd.AddCommand(contactsDeleteCmd)

	rootCmd.AddCommand(contactsCmd)
}
