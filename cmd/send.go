package cmd

import "github.com/spf13/cobra"

var sendCmd = &cobra.Command{
	Use:   "send <id>",
	Short: "Send a transactional email",
	Args:  cobra.ExactArgs(1),
	RunE:  transactionalSendRunE,
}

func init() {
	addTransactionalSendFlags(sendCmd)
	rootCmd.AddCommand(sendCmd)
}
