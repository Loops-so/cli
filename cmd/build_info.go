package cmd

import (
	"runtime/debug"
	"strings"
)

var (
	version = "dev"
	commit  = "none"
)

// includes a leading newline to make it easy to read the ascii art here in the source
const versionHeader = `
    __    ____  ____  ____  _____
   / /   / __ \/ __ \/ __ \/ ___/
  / /   / / / / / / / /_/ /\__ \
 / /___/ /_/ / /_/ / ____/___/ /
/_____/\____/\____/_/    /____/
`

func init() {
	if info, ok := debug.ReadBuildInfo(); ok && version == "dev" {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = strings.TrimPrefix(info.Main.Version, "v")
		}
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				commit = s.Value[:7]
				break
			}
		}
	}
	header := strings.TrimPrefix(versionHeader, "\n")
	rootCmd.SetVersionTemplate(header + "\n{{with .Name}}{{printf \"%s \" .}}{{end}}{{printf \"version %s\" .Version}}\n")
}
