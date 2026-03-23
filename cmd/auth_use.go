package cmd

import (
	"errors"
	"fmt"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var clearActive bool

var authUseCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Set or clear the active API key",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if clearActive && len(args) > 0 {
			return errors.New("cannot use --clear with a key name")
		}
		if !clearActive && len(args) == 0 {
			return errors.New("provide a key name or --clear")
		}

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		if err := runAuthUse(name); err != nil {
			return err
		}

		if isJSONOutput() {
			msg := fmt.Sprintf("Active team set to: %s", name)
			if name == "" {
				msg = "Active team cleared"
			}
			return printJSON(cmd.OutOrStdout(), Result{Success: true, Message: msg})
		}

		if name == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Active team cleared.")
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Active team set to %q.\n", name)
		}
		return nil
	},
}

func runAuthUse(name string) error {
	return config.SetActiveTeam(name)
}

func init() {
	authUseCmd.Flags().BoolVar(&clearActive, "clear", false, "Clear the active team")
	authCmd.AddCommand(authUseCmd)
}
