package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Print the resolved configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, string(out))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
