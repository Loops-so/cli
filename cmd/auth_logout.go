package cmd

import (
	"errors"
	"fmt"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutName string

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored Loops credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if logoutName == "" {
			return errors.New("use --name to specify which key to remove (e.g. loops auth logout --name loops-prod)")
		}
		if err := runAuthLogout(logoutName); err != nil {
			return err
		}
		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), Result{Success: true})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Logged out of %q.\n", logoutName)
		return nil
	},
}

func runAuthLogout(name string) error {
	return config.Delete(name)
}

func init() {
	logoutCmd.Flags().StringVarP(&logoutName, "name", "n", "", "Name of the API key to remove")
	authCmd.AddCommand(logoutCmd)
}
