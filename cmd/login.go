package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your Loops API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(os.Stderr, "Enter your API key: ")
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return fmt.Errorf("failed to read API key: %w", err)
		}

		apiKey := strings.TrimSpace(string(raw))
		if apiKey == "" {
			return fmt.Errorf("API key cannot be empty")
		}

		client := api.NewClient(config.EndpointURL(), apiKey)
		result, err := client.GetAPIKey()
		if err != nil {
			return fmt.Errorf("API key verification failed: %w", err)
		}

		if err := config.Save(apiKey); err != nil {
			return err
		}

		fmt.Printf("API key saved. Authenticated as team: %s\n", result.TeamName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
