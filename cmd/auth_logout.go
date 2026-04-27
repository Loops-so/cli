package cmd

import (
	"errors"
	"fmt"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout <name>",
	Short: "Remove stored Loops credentials",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := runAuthLogout(name); err != nil {
			return err
		}
		if isJSONOutput() {
			return printJSON(cmd.OutOrStdout(), Result{Success: true})
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Logged out of %q.\n", name)
		return nil
	},
}

func runAuthLogout(name string) error {
	if name == "" {
		return errors.New("a key name is required")
	}
	return config.Delete(name)
}

func init() {
	authCmd.AddCommand(logoutCmd)
}
