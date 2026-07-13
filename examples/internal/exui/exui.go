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

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/status"
)

var noHelp = flag.Bool("no-help", false, "hide the status/help bar (script mode)")

// Init parses the shared example flags. Call it first in main.
func Init() { flag.Parse() }

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
	c := &Chrome{bar: status.New(), hidden: *noHelp}
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
// snap status bar), unless hidden.
func (c *Chrome) Attach(content string, termH int) string {
	if c == nil || c.hidden {
		return content
	}
	gap := max(termH-lipgloss.Height(content)-1, 0)
	block := lipgloss.NewStyle().Height(lipgloss.Height(content) + gap).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, block, c.View())
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
