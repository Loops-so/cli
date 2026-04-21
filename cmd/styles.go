package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

var headingStyle = sync.OnceValue(func() lipgloss.Style {
	isDark := true
	if term.IsTerminal(os.Stdout.Fd()) {
		isDark = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	}
	cs := fang.DefaultColorScheme(lipgloss.LightDark(isDark))
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(cs.Title).
		Transform(strings.ToUpper)
})

// styledTable wraps tabwriter so the first written row (the header) is rendered
// with the fang heading style on Flush. Buffering until Flush lets tabwriter
// compute column widths from the plain text, then styling is applied after
// alignment so ANSI escapes don't throw off the layout. Color is stripped
// automatically when the destination is not a TTY.
type styledTable struct {
	out io.Writer
	buf bytes.Buffer
	tw  *tabwriter.Writer
}

func newStyledTable(out io.Writer) *styledTable {
	s := &styledTable{out: out}
	s.tw = tabwriter.NewWriter(&s.buf, 0, 0, 2, ' ', 0)
	return s
}

func (s *styledTable) Write(p []byte) (int, error) { return s.tw.Write(p) }

func (s *styledTable) Flush() error {
	if err := s.tw.Flush(); err != nil {
		return err
	}
	cw := colorprofile.NewWriter(s.out, os.Environ())
	header, body, _ := strings.Cut(s.buf.String(), "\n")
	if _, err := fmt.Fprintln(cw, headingStyle().Render(header)); err != nil {
		return err
	}
	if body != "" {
		if _, err := fmt.Fprint(cw, body); err != nil {
			return err
		}
	}
	return nil
}
