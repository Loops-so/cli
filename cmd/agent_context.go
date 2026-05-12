package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// cobra stores flag-group constraints in cmd.Annotations under these keys.
// the keys arent exported in cobra/flag_groups.go, so we hardcode the
// literal strings. each value is a []string where every element is one
// group with member flag names joined by a single space.
const (
	annotOneRequired       = "cobra_annotation_one_required"
	annotMutuallyExclusive = "cobra_annotation_mutually_exclusive"
	annotRequiredTogether  = "cobra_annotation_required_if_others_set"
)

type AgentContext struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Commit      string    `json:"commit"`
	Short       string    `json:"short"`
	Long        string    `json:"long"`
	GlobalFlags []Flag    `json:"globalFlags"`
	Commands    []Command `json:"commands"`
}

type Command struct {
	Name       string       `json:"name"`
	Path       []string     `json:"path"`
	Use        string       `json:"use"`
	Short      string       `json:"short"`
	Long       string       `json:"long"`
	Hidden     bool         `json:"hidden"`
	Runnable   bool         `json:"runnable"`
	Args       []Positional `json:"args"`
	Flags      []Flag       `json:"flags"`
	FlagGroups FlagGroups   `json:"flagGroups"`
	Commands   []Command    `json:"commands"`
}

type Flag struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Persistent  bool   `json:"persistent"`
}

type FlagGroups struct {
	OneRequired       [][]string `json:"oneRequired"`
	MutuallyExclusive [][]string `json:"mutuallyExclusive"`
	RequiredTogether  [][]string `json:"requiredTogether"`
}

func buildAgentContext(root *cobra.Command) AgentContext {
	rootPersistent := persistentFlagNames(root)
	return AgentContext{
		Name:        root.Name(),
		Version:     version,
		Commit:      commit,
		Short:       root.Short,
		Long:        root.Long,
		GlobalFlags: collectRootPersistentFlags(root),
		Commands:    walkChildren(root, []string{}, rootPersistent),
	}
}

func persistentFlagNames(cmd *cobra.Command) map[string]struct{} {
	out := map[string]struct{}{}
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { out[f.Name] = struct{}{} })
	return out
}

func collectRootPersistentFlags(root *cobra.Command) []Flag {
	out := []Flag{}
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		out = append(out, flagFromPFlag(f, true))
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func walkChildren(parent *cobra.Command, parentPath []string, rootPersistent map[string]struct{}) []Command {
	out := []Command{}
	for _, c := range parent.Commands() {
		// filter out commands that arent useful for agents
		switch c.Name() {
		case "help", "completion", "man", "spam":
			continue
		}
		path := append(append([]string{}, parentPath...), c.Name())
		out = append(out, Command{
			Name:       c.Name(),
			Path:       path,
			Use:        c.Use,
			Short:      c.Short,
			Long:       c.Long,
			Hidden:     c.Hidden,
			Runnable:   c.RunE != nil || c.Run != nil,
			Args:       parsePositionals(c.Use),
			Flags:      walkLocalFlags(c, rootPersistent),
			FlagGroups: extractFlagGroups(c),
			Commands:   walkChildren(c, path, rootPersistent),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func walkLocalFlags(cmd *cobra.Command, rootPersistent map[string]struct{}) []Flag {
	out := []Flag{}
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		if _, skip := rootPersistent[f.Name]; skip {
			return
		}
		persistent := cmd.PersistentFlags().Lookup(f.Name) != nil
		out = append(out, flagFromPFlag(f, persistent))
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func flagFromPFlag(f *pflag.Flag, persistent bool) Flag {
	_, required := f.Annotations[cobra.BashCompOneRequiredFlag]
	return Flag{
		Name:        f.Name,
		Shorthand:   f.Shorthand,
		Type:        f.Value.Type(),
		Default:     f.DefValue,
		Description: f.Usage,
		Required:    required,
		Persistent:  persistent,
	}
}

func extractFlagGroups(cmd *cobra.Command) FlagGroups {
	groups := FlagGroups{
		OneRequired:       [][]string{},
		MutuallyExclusive: [][]string{},
		RequiredTogether:  [][]string{},
	}
	// each flag in a group carries the same annotation value. dedupe by raw
	// joined string before splitting back into a name slice.
	seen := map[string]map[string]struct{}{
		annotOneRequired:       {},
		annotMutuallyExclusive: {},
		annotRequiredTogether:  {},
	}
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		for _, key := range []string{annotOneRequired, annotMutuallyExclusive, annotRequiredTogether} {
			for _, raw := range f.Annotations[key] {
				if _, dup := seen[key][raw]; dup {
					continue
				}
				seen[key][raw] = struct{}{}
				names := strings.Split(raw, " ")
				switch key {
				case annotOneRequired:
					groups.OneRequired = append(groups.OneRequired, names)
				case annotMutuallyExclusive:
					groups.MutuallyExclusive = append(groups.MutuallyExclusive, names)
				case annotRequiredTogether:
					groups.RequiredTogether = append(groups.RequiredTogether, names)
				}
			}
		}
	})
	for _, g := range []*[][]string{&groups.OneRequired, &groups.MutuallyExclusive, &groups.RequiredTogether} {
		sort.Slice(*g, func(i, j int) bool {
			return strings.Join((*g)[i], " ") < strings.Join((*g)[j], " ")
		})
	}
	return groups
}

var agentContextCmd = &cobra.Command{
	Use:   "agent-context",
	Short: "Print a JSON description of all CLI commands and flags",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return printJSON(cmd.OutOrStdout(), buildAgentContext(rootCmd))
	},
}

func init() {
	rootCmd.AddCommand(agentContextCmd)
}
