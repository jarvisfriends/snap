// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

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
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/keys"
	"github.com/jarvisfriends/snap/notifications"
	"github.com/jarvisfriends/snap/status"
	"github.com/jarvisfriends/snap/styles"
)

var (
	noHelp      = flag.Bool("no-help", false, "hide the status/help bar (script mode)")
	showVersion = flag.Bool("version", false, "print the version and exit")
)

// Version is injected at build time via -ldflags for release binaries:
//
//	-X github.com/jarvisfriends/snap/examples/internal/exui.Version=v1.2.3
//
// A local `go run` build keeps the "dev" default.
var Version = "dev"

// Init parses the shared example flags. Call it first in main. Exits 0 after
// printing Version when --version is passed, so a released archive's binary
// is self-identifying without needing to consult the archive's file name.
func Init() {
	flag.Parse()
	if *showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}
}

// defaultTint is the palette every example opens with: Catppuccin Macchiato,
// whose deep blue base keeps the demos off the terminal-default black and
// gives every component the same injected colors.
const defaultTint = "catppuccin_macchiato"

// The shared palette is rebuildable rather than a sync.Once value because the
// tour cycles themes at runtime. Components snapshot styles at construction,
// so a cycle is always followed by rebuilding the chrome and the live page —
// mutating this alone would leave already-built widgets on the old palette.
var (
	themeMu     sync.Mutex
	sharedTheme *styles.AppStyle
	activeTint  = defaultTint
)

// Theme returns the shared example palette. Every example passes it into the
// components it mounts (they are theme-free with injected style hooks) and
// paints its root view background from Theme().Bg, so the whole demo — page,
// components, status bar — agrees on one background.
func Theme() *styles.AppStyle {
	themeMu.Lock()
	defer themeMu.Unlock()
	if sharedTheme == nil {
		buildThemeLocked()
	}
	return sharedTheme
}

// TintID is the active tint's ID, shown in the debug overlay and the tour's
// expanded help so a theme cycle names what it landed on.
func TintID() string {
	themeMu.Lock()
	defer themeMu.Unlock()
	return activeTint
}

// cycleTints is the ring ctrl+t walks: the demo default followed by snap's
// built-in themes. It is deliberately not every registered tint — bubbletint
// ships hundreds, which is a theme picker's job, not a cycle key's — and the
// default leads so cycling always comes back to the palette the gifs were
// recorded in.
func cycleTints() []string {
	ids := styles.BuiltinTintIDs()
	ring := make([]string, 0, len(ids)+1)
	ring = append(ring, defaultTint)
	for _, id := range ids {
		if id != defaultTint {
			ring = append(ring, id)
		}
	}
	return ring
}

// CycleTheme advances to the next tint in the ring and rebuilds the shared
// palette, returning the new tint's ID. Callers must rebuild any chrome and
// page already constructed — see the note on the vars above.
func CycleTheme() string {
	themeMu.Lock()
	defer themeMu.Unlock()
	ring := cycleTints()
	next := 0
	for i, id := range ring {
		if id == activeTint {
			next = (i + 1) % len(ring)
			break
		}
	}
	activeTint = ring[next]
	buildThemeLocked()
	return activeTint
}

// buildThemeLocked selects activeTint and snapshots the palette. Callers must
// hold themeMu.
func buildThemeLocked() {
	// Best-effort: SetCurrentTint initializes the tint registry; an
	// unknown id falls back to styles' default palette.
	_ = styles.SetCurrentTint(activeTint)
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

	// shared shell surfaces (see shell.go): notifications history behind
	// the bar's bell, the info modal behind ⓘ, and the ctrl+d debug overlay.
	height int
	mgr    *notifications.Manager
	km     *keys.AppKeyMap
	modal  *status.InfoModal
	debug  bool
	start  time.Time

	// capture reports whether the hosted page is eating keystrokes; see
	// SetCapture. nil means the page never captures text.
	capture func() bool
}

// NewChrome builds the bar for the given bindings (shown left-to-right).
// Call after Init so --no-help has been parsed.
func NewChrome(bindings ...key.Binding) *Chrome {
	t := Theme() // select the shared tint before the bar snapshots styles
	c := &Chrome{bar: status.New(), hidden: *noHelp, start: time.Now()}
	c.mgr, c.km, c.modal = newShellParts(c.bar)
	c.bar.SetColors(t)
	// Example bindings first, then the shell's shared surface chords.
	c.bar.SetPageBindings(Bindings(append(bindings,
		Bind("ctrl+n", "notifs"), Bind(infoKey, "info"), Bind("ctrl+d", "debug"))))
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

// Dim is the shared caption/filler style examples use for de-emphasized
// text (headers, rulers, line numbers), so every demo dims the same way.
func Dim() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
}

// Frame is the one-call View finisher: it stacks the help bar and paints the
// theme (Apply), composites any open shell surface (history, info, debug),
// switches to the alt screen with mouse reporting on, and chains the shell's
// pointer handling in front of any v.OnMouse the example set before calling.
func (c *Chrome) Frame(v *tea.View, termH int) {
	if c != nil {
		c.height = termH
	}
	prev := v.OnMouse
	c.Apply(v, termH)
	if c != nil {
		v.SetContent(c.overlays(v.Content))
	}
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.OnMouse = c.wrapMouse(prev)
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
