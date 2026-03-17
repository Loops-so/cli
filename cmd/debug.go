package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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

		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout, string(out))
		return nil
	},
}

func init() {
	authCmd.AddCommand(statusCmd)
}
