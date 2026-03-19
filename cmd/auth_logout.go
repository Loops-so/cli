package cmd

import (
	"fmt"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored Loops credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runAuthLogout(); err != nil {
			return err
		}
		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), Result{Success: true})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Logged out.")
		return nil
	},
}

func runAuthLogout() error {
	return config.Delete()
}

func init() {
	authCmd.AddCommand(logoutCmd)
}
