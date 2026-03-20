package cmd

import (
	"fmt"
	"strings"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Print the resolved configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, keyResp, err := runAuthStatus()
		if err != nil {
			return err
		}

		masked := maskKey(cfg.APIKey)

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), struct {
				APIKey      string `json:"apiKey"`
				EndpointURL string `json:"endpointUrl"`
				TeamName    string `json:"teamName"`
			}{masked, cfg.EndpointURL, keyResp.TeamName})
		}

		fmt.Fprintf(cmd.OutOrStdout(), "API Key:  %s\n", masked)
		fmt.Fprintf(cmd.OutOrStdout(), "Endpoint: %s\n", cfg.EndpointURL)
		fmt.Fprintf(cmd.OutOrStdout(), "Team:     %s\n", keyResp.TeamName)
		return nil
	},
}

func runAuthStatus() (*config.Config, *api.APIKeyResponse, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	keyResp, err := api.NewClient(cfg.EndpointURL, cfg.APIKey).GetAPIKey()
	if err != nil {
		return nil, nil, fmt.Errorf("API key verification failed: %w", err)
	}
	return cfg, keyResp, nil
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return fmt.Sprintf("%s%s", strings.Repeat("*", len(key)-4), key[len(key)-4:])
}

func init() {
	authCmd.AddCommand(statusCmd)
}
