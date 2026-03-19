package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

func runAPIKey(cfg *config.Config) (*api.APIKeyResponse, error) {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey).GetAPIKey()
}

var apiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Validate your API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		result, err := runAPIKey(cfg)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), result)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Valid API key for team: %s\n", result.TeamName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(apiKeyCmd)
}
