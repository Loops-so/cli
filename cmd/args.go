package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var requiredArgRe = regexp.MustCompile(`<(\w+)>`)

func wrapArgsWithNames(cmd *cobra.Command) {
	if cmd.Args != nil {
		v := cmd.Args
		cmd.Args = func(cmd *cobra.Command, args []string) error {
			err := v(cmd, args)
			if err == nil {
				return nil
			}
			required := requiredArgNames(cmd.Use)
			if len(args) < len(required) {
				missing := required[len(args):]
				return fmt.Errorf(`required argument(s) "%s" not set`, strings.Join(missing, `", "`))
			}
			return err
		}
	}
	for _, c := range cmd.Commands() {
		wrapArgsWithNames(c)
	}
}

func requiredArgNames(use string) []string {
	matches := requiredArgRe.FindAllStringSubmatch(use, -1)
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = m[1]
	}
	return names
}
