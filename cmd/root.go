package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var outputFormat outputFlag = "text"
var teamFlag string
var debugFlag bool

func newAPIClient(cfg *config.Config) *api.Client {
	return api.NewClient(cfg.EndpointURL, cfg.APIKey, cfg.Debug).
		WithUserAgent("loops-cli/" + version)
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(teamFlag)
	if err != nil {
		return nil, err
	}
	cfg.Debug = debugFlag
	return cfg, nil
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

func Execute() {
	defer func() {
		if updateCheckDone != nil {
			select {
			case <-updateCheckDone:
			case <-time.After(500 * time.Millisecond):
			}
		}
		if updateCheckCancel != nil {
			updateCheckCancel()
		}
	}()

	fixHelpFlags(rootCmd)
	err := rootCmd.Execute()

	checkForUpdate(os.Stderr)

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
	rootCmd.PersistentFlags().VarP(&outputFormat, "output", "o", "Output format (text, json)")
	rootCmd.PersistentFlags().StringVarP(&teamFlag, "team", "t", "", "Team key name to use")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Print API request details before sending")
}
