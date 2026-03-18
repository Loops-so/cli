package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var apiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Validate your API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.EndpointURL, cfg.APIKey)
		result, err := client.GetAPIKey()
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(result)
		}
		fmt.Printf("Valid API key for team: %s\n", result.TeamName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(apiKeyCmd)
}
