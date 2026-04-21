package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

var fangColorScheme = sync.OnceValue(func() fang.ColorScheme {
	isDark := true
	if term.IsTerminal(os.Stdout.Fd()) {
		isDark = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	}
	return fang.DefaultColorScheme(lipgloss.LightDark(isDark))
})

type styledTable struct {
	out io.Writer
	t   *table.Table
}

func newStyledTable(out io.Writer, headers ...string) *styledTable {
	cs := fangColorScheme()
	t := table.New().
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		BorderRow(false).
		BorderHeader(true).
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			base := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return base.
					Bold(true).
					Foreground(cs.Title).
					Transform(strings.ToUpper)
			}
			return base
		})
	return &styledTable{out: out, t: t}
}

func (s *styledTable) Row(cells ...string) {
	s.t.Row(cells...)
}

func (s *styledTable) Render() error {
	cw := colorprofile.NewWriter(s.out, os.Environ())
	_, err := fmt.Fprintln(cw, s.t.Render())
	return err
}
