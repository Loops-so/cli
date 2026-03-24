package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginName string

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your Loops API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if loginName == "" {
			return errors.New("use --name to give this key a name (e.g. loops auth login --name my-team)")
		}

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

		result, err := runAuthLogin(apiKey, loginName)
		if err != nil {
			return err
		}

		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), Result{Success: true, Message: fmt.Sprintf("Authenticated as team: %s", result.TeamName)})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "API key saved as %q. Authenticated as team: %s\n", loginName, result.TeamName)
		return nil
	},
}

func runAuthLogin(apiKey, name string) (*api.APIKeyResponse, error) {
	if name == "" {
		return nil, errors.New("use --name to give this key a name")
	}
	result, err := api.NewClient(config.EndpointURL(), apiKey, debugFlag).GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("API key verification failed: %w", err)
	}
	if err := config.Save(apiKey, name); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	loginCmd.Flags().StringVarP(&loginName, "name", "n", "", "Name for this API key (e.g. my-team)")
	authCmd.AddCommand(loginCmd)
}
