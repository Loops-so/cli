/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var outputFormat outputFlag = "text"
var teamFlag string

func loadConfig() (*config.Config, error) {
	return config.Load(teamFlag)
}

var rootCmd = &cobra.Command{
	Use:           "loops",
	Short:         "The official CLI for Loops (https://loops.so)",
	Long:          "The official CLI for Loops (https://loops.so)",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func fixHelpFlags(cmd *cobra.Command) {
	cmd.InitDefaultHelpFlag()
	if f := cmd.Flags().Lookup("help"); f != nil {
		name := cmd.Name()
		if name == "" {
			name = "this command"
		}
		f.Usage = "Help for " + name
	}
	for _, sub := range cmd.Commands() {
		fixHelpFlags(sub)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	fixHelpFlags(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		if isJSONOutput() {
			printJSON(os.Stderr, Result{Success: false, Message: err.Error()})
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().VarP(&outputFormat, "output", "o", "Output format (text, json)")
	rootCmd.PersistentFlags().StringVarP(&teamFlag, "team", "t", "", "Team key name to use")
}
