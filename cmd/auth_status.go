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
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cfg)
		}

		fmt.Printf("API Key:  %s\n", cfg.APIKey)
		fmt.Printf("Endpoint: %s\n", cfg.EndpointURL)
		return nil
	},
}

func init() {
	authCmd.AddCommand(statusCmd)
}
