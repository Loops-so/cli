package cmd

import "github.com/spf13/cobra"

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a transactional email",
	RunE:  transactionalSendRunE,
}

func init() {
	addTransactionalSendFlags(sendCmd)
	rootCmd.AddCommand(sendCmd)
}
