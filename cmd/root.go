package cmd

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"charm.land/fang/v2"
	"github.com/loops-so/cli/internal/api"
	"github.com/loops-so/cli/internal/config"
	"github.com/spf13/cobra"
)

var outputFormat outputFlag = "text"
var teamFlag string
var debugFlag bool
var colorFlag = true

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

func jsonAwareErrorHandler(w io.Writer, styles fang.Styles, err error) {
	if isJSONOutput() {
		_ = printJSON(w, Result{Success: false, Message: err.Error()})
		return
	}
	fang.DefaultErrorHandler(w, styles, err)
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

	applyColorArg(os.Args[1:])

	err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version),
		fang.WithCommit(commit),
		fang.WithErrorHandler(jsonAwareErrorHandler),
	)

	checkForUpdate(os.Stderr)

	if err != nil {
		os.Exit(1)
	}
}

// applyColorArg scans args for --color=<bool> and sets NO_COLOR=1 when the user
// passes a false value. fang/lipgloss capture the color profile before cobra
// parses persistent flags, so a flag-parse hook would miss cases like unknown
// command errors — hence the early scan.
func applyColorArg(args []string) {
	for _, a := range args {
		if !strings.HasPrefix(a, "--color=") {
			continue
		}
		v := strings.TrimPrefix(a, "--color=")
		if b, err := strconv.ParseBool(v); err == nil && !b {
			os.Setenv("NO_COLOR", "1")
		}
		return
	}
}

func init() {
	rootCmd.PersistentFlags().VarP(&outputFormat, "output", "o", "Output format (text, json)")
	rootCmd.PersistentFlags().StringVarP(&teamFlag, "team", "t", "", "Team key name to use")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Print API request details before sending")
	rootCmd.PersistentFlags().BoolVar(&colorFlag, "color", true, "Enable colored output (--color=false to disable)")
}
