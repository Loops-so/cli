package cmd

import (
	"fmt"
	"image/color"
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

var isDarkBackground = sync.OnceValue(func() bool {
	if !term.IsTerminal(os.Stdout.Fd()) {
		return true
	}
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
})

var fangColorScheme = sync.OnceValue(func() fang.ColorScheme {
	return fang.DefaultColorScheme(lipgloss.LightDark(isDarkBackground()))
})

func hexColor(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

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
