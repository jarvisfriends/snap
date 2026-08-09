// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package menu implements the `snap_input menu` subcommand: a script-usable context-menu picker built on snap/menu:
// right-click anywhere (or press m) to pop the menu, choose an item, and the
// chosen item's ID is written to stdout (the TUI itself renders on stderr):
//
//	action=$(go run ./examples/snap_input menu)
//
// --no-help hides the status bar. Quitting (q/esc) prints nothing, exit 1.
package menu

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/menu"
)

type demoApp struct {
	menu   menu.Menu
	chrome *exui.Chrome
	picked string
	w, h   int
}

func newDemo() *demoApp {
	a := &demoApp{
		chrome: exui.NewChrome(
			exui.Bind("right-click m", "open"),
			exui.Bind("↑↓", "move"),
			exui.Bind("enter click", "select"),
			exui.Bind("esc", "dismiss"),
			exui.Bind("q", "quit"),
		),
	}
	// An open menu owns the keyboard, the same way the shell's own overlays
	// do — esc dismisses it, and only then does "q" quit again.
	a.chrome.SetCapture(a.Capturing)
	return a
}

func items() []menu.Item {
	return []menu.Item{
		{ID: "open", Label: "Open"},
		{ID: "rename", Label: "Rename", Disabled: true},
		{ID: "copy", Label: "Copy path"},
		{ID: "delete", Label: "Delete"},
	}
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		switch {
		case a.menu.IsOpen():
			// HandleKey mirrors HandleMouse: the open menu owns the keyboard.
			if chosen, _ := a.menu.HandleKey(msg); chosen != nil {
				a.picked = chosen.ID
				return a, exui.Confirm()
			}
		case msg.String() == "m":
			a.menu.Open(a.w/2, a.h/2, items(), "keyboard")
		}
	}
	return a, nil
}

// onMouse owns pointer input per the snap contract: while the menu is open
// it consumes events; a right-click opens it at the pointer.
func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	if a.menu.IsOpen() {
		if chosen, _ := a.menu.HandleMouse(mm, a.w, a.h); chosen != nil {
			a.picked = chosen.ID
			return exui.Confirm()
		}
		return nil
	}
	if click, ok := mm.(tea.MouseClickMsg); ok && click.Button == tea.MouseRight {
		me := click.Mouse()
		a.menu.Open(me.X, me.Y, items(), "cell "+strconv.Itoa(me.X)+","+strconv.Itoa(me.Y))
	}
	return nil
}

func (a *demoApp) View() tea.View {
	dim := exui.Dim()
	line := dim.Render(strings.Repeat("·", max(a.w, 1)))
	paneH := max(a.h-a.chrome.Height(), 1)
	rows := make([]string, 0, paneH)
	for range paneH {
		rows = append(rows, line)
	}
	base := lipgloss.JoinVertical(lipgloss.Left, rows...)
	v := tea.NewView(a.menu.Composite(base, a.w, paneH))
	v.OnMouse = a.onMouse
	a.chrome.Frame(&v, a.h)
	return v
}

// New builds the menu page.
func New() exui.Page { return newDemo() }

// Result reports the chosen menu item, and nothing until one is selected.
func (a *demoApp) Result() []exui.Field {
	if a.picked == "" {
		return nil
	}
	return []exui.Field{exui.F("action", a.picked)}
}

// Shell exposes this page's chrome to the tour host.
func (a *demoApp) Shell() *exui.Chrome { return a.chrome }

// Capturing reports whether the context menu is open and owning the keyboard.
func (a *demoApp) Capturing() bool { return a.menu.IsOpen() }
