// Package exui is the shared chrome for snap's example programs, so every
// example reads the same way and doubles as a script-friendly input tool:
//
//   - a uniform status/help bar (the real snap/status bar showing the
//     example's own key bindings) rendered as the bottom line;
//
//   - a --no-help flag that hides that bar, for scripts that only want the
//     component itself;
//
//   - result plumbing: the TUI renders on stderr and Finish writes ONLY the
//     user's choice to stdout, so a shell can capture it directly:
//
//     date=$(go run ./examples/datepicker)
//
// A canceled example (quit without choosing) prints nothing and exits 1.
package exui

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/status"
	"github.com/jarvisfriends/snap/styles"
)

var noHelp = flag.Bool("no-help", false, "hide the status/help bar (script mode)")

// Init parses the shared example flags. Call it first in main.
func Init() { flag.Parse() }

// themeTint is the palette every example renders with: Catppuccin Macchiato,
// whose deep blue base keeps the demos off the terminal-default black and
// gives every component the same injected colors.
const themeTint = "catppuccin_macchiato"

var (
	themeOnce   sync.Once
	sharedTheme *styles.AppStyle
)

// Theme returns the shared example palette. Every example passes it into the
// components it mounts (they are theme-free with injected style hooks) and
// paints its root view background from Theme().Bg, so the whole demo — page,
// components, status bar — agrees on one background.
func Theme() *styles.AppStyle {
	themeOnce.Do(func() {
		// Best-effort: SetCurrentTint initializes the tint registry; an
		// unknown id falls back to styles' default palette.
		_ = styles.SetCurrentTint(themeTint)
		base := styles.Active()
		sharedTheme = base

		// Optional debug mode for GIF audits: force a loud background so any
		// unthemed holes stand out immediately.
		if dbg := strings.TrimSpace(os.Getenv("SNAP_DEMO_DEBUG_BG")); dbg != "" {
			cp := *base
			cp.Bg = lipgloss.Color(dbg)
			cp.StatusBg = lipgloss.Color(dbg)
			cp.Styles = styles.BuildStyles(&cp)
			sharedTheme = &cp
		}
	})
	return sharedTheme
}

// Bindings adapts a flat binding list to the help.KeyMap the status bar
// consumes: everything on one short-help line.
type Bindings []key.Binding

// ShortHelp implements help.KeyMap.
func (b Bindings) ShortHelp() []key.Binding { return b }

// FullHelp implements help.KeyMap.
func (b Bindings) FullHelp() [][]key.Binding { return [][]key.Binding{b} }

var _ help.KeyMap = Bindings(nil)

// Bind builds one help-bar entry: Bind("enter", "confirm").
func Bind(keyName, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(keyName), key.WithHelp(keyName, desc))
}

// Chrome is the example's bottom status/help bar — snap's own status.BarModel
// showing the example's key bindings, or nothing at all under --no-help.
type Chrome struct {
	bar    *status.BarModel
	hidden bool
	width  int
}

// NewChrome builds the bar for the given bindings (shown left-to-right).
// Call after Init so --no-help has been parsed.
func NewChrome(bindings ...key.Binding) *Chrome {
	t := Theme() // select the shared tint before the bar snapshots styles
	c := &Chrome{bar: status.New(), hidden: *noHelp}
	c.bar.SetColors(t)
	c.bar.SetPageBindings(Bindings(bindings))
	return c
}

// SetWidth informs the bar of the terminal width (call on WindowSizeMsg).
// A nil Chrome (tests constructing a demo app bare) is a hidden bar.
func (c *Chrome) SetWidth(w int) {
	if c == nil {
		return
	}
	c.width = w
	c.bar.SetWidth(w)
}

// Height is the number of lines the bar occupies (0 under --no-help), so
// examples can budget the component's height as termH - chrome.Height().
func (c *Chrome) Height() int {
	if c == nil || c.hidden {
		return 0
	}
	return 1
}

// View renders the bar line, or "" under --no-help.
func (c *Chrome) View() string {
	if c == nil || c.hidden {
		return ""
	}
	return c.bar.View().Content
}

// Attach stacks the bar under the example's content: content fills the top,
// the bar sits on the terminal's bottom line (matching how apps mount the
// snap status bar), unless hidden. Content taller than the window is clipped
// so the bar can never be pushed off screen.
func (c *Chrome) Attach(content string, termH int) string {
	if c == nil || c.hidden {
		return content
	}
	if avail := termH - 1; avail > 0 && lipgloss.Height(content) > avail {
		lines := strings.Split(content, "\n")
		content = lipgloss.JoinVertical(lipgloss.Left, lines[:avail]...)
	}
	gap := max(termH-lipgloss.Height(content)-1, 0)
	block := lipgloss.NewStyle().Height(lipgloss.Height(content) + gap).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, block, c.View())
}

// Apply is the one-call frame finisher every example uses: it stacks the
// help bar under v's content and paints the shared theme's background and
// foreground onto the root view, so no demo renders on terminal-default
// black and every unstyled cell agrees with the injected component styles.
func (c *Chrome) Apply(v *tea.View, termH int) {
	v.SetContent(c.Attach(v.Content, termH))
	t := Theme()
	v.BackgroundColor = t.Bg
	v.ForegroundColor = t.Fg
}

// Program builds the example's tea.Program rendering on stderr, keeping
// stdout clean for the Finish value — the split that makes
// value=$(example) work from a shell.
func Program(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
	return tea.NewProgram(m, append([]tea.ProgramOption{tea.WithOutput(os.Stderr)}, opts...)...)
}

// Finish ends a value-producing example and never returns: the chosen values
// go to stdout one per line and the process exits 0; with ok=false
// (canceled) nothing is printed and the exit code is 1 so scripts can tell
// the difference.
func Finish(ok bool, values ...string) {
	if !ok {
		os.Exit(1)
	}
	for _, v := range values {
		fmt.Println(v)
	}
	os.Exit(0)
}

// Fatal reports a program error on stderr and exits.
func Fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
