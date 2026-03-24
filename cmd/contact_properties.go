package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runContactPropertiesList(cfg *config.Config, customOnly bool) ([]api.ContactProperty, error) {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey, cfg.Debug).ListContactProperties(customOnly)
}

func runContactPropertiesCreate(cfg *config.Config, name, propType string) error {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey, cfg.Debug).CreateContactProperty(name, propType)
}

var contactPropertiesCmd = &cobra.Command{
	Use:   "contact-properties",
	Short: "Manage contact properties",
}

var contactPropertiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List contact properties",
	RunE: func(cmd *cobra.Command, args []string) error {
		customOnly, _ := cmd.Flags().GetBool("custom")

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		props, err := runContactPropertiesList(cfg, customOnly)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			if props == nil {
				props = []api.ContactProperty{}
			}
			return printJSON(cmd.OutOrStdout(), props)
		}

		if len(props) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No contact properties found.")
			return nil
		}

		w := newTableWriter(cmd.OutOrStdout())
		fmt.Fprintln(w, "KEY\tLABEL\tTYPE")
		for _, p := range props {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Key, p.Label, p.Type)
		}
		w.Flush()

		return nil
	},
}

var contactPropertiesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a contact property",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		propType, _ := cmd.Flags().GetString("type")

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if err := runContactPropertiesCreate(cfg, name, propType); err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), Result{Success: true})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Created.")
		return nil
	},
}

func init() {
	contactPropertiesListCmd.Flags().Bool("custom", false, "Only list custom properties")
	contactPropertiesCmd.AddCommand(contactPropertiesListCmd)

	contactPropertiesCreateCmd.Flags().String("name", "", "Property name (camelCase, e.g. planName)")
	contactPropertiesCreateCmd.Flags().String("type", "", "Property type")
	contactPropertiesCreateCmd.MarkFlagRequired("name")
	contactPropertiesCreateCmd.MarkFlagRequired("type")
	contactPropertiesCmd.AddCommand(contactPropertiesCreateCmd)

	rootCmd.AddCommand(contactPropertiesCmd)
}
