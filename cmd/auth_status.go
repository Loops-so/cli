package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the resolved configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := runAuthStatus()
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), cfg)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "API Key:  %s\n", cfg.APIKey)
		fmt.Fprintf(cmd.OutOrStdout(), "Endpoint: %s\n", cfg.EndpointURL)
		return nil
	},
}

func runAuthStatus() (*config.Config, error) {
	return config.Load()
}

func init() {
	authCmd.AddCommand(statusCmd)
}
