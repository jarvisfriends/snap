// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package dependencies implements the `snap_input dependencies` subcommand,
// demoing snap/dependencies rendered by snap/status's InfoModal — the same
// modal every example opens from the bar's ⓘ icon (or ctrl+e); this demo
// just starts with it open. Build info (Go version, OS, VCS revision) sits
// above a scrollable dependency list read via dependencies.ExpandedBuildInfo.
// Wheel or ↑/↓/PgUp/PgDn scroll, Esc or an outside click closes, q quits.
package dependencies

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
)

type demoApp struct {
	chrome *exui.Chrome
	w, h   int
	opened bool
}

func newDemo() *demoApp {
	return &demoApp{chrome: exui.NewChrome(
		exui.Bind("wheel ↑↓", "move"),
		exui.Bind("esc outside-click", "close"),
		exui.Bind("q", "quit"),
	)}
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		if !a.opened { // this demo IS the info modal: open on first size
			a.opened = true
			a.chrome.OpenInfo()
		}
	case tea.KeyPressMsg:
		if msg.String() == "i" {
			a.chrome.OpenInfo()
		}
	}
	return a, nil
}

func (a *demoApp) View() tea.View {
	line := exui.Dim().Render(strings.Repeat("·", max(a.w, 1)))
	paneH := max(a.h-a.chrome.Height(), 1)
	rows := make([]string, 0, paneH)
	for range paneH {
		rows = append(rows, line)
	}
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
	a.chrome.Frame(&v, a.h)
	return v
}

// New builds the dependencies page.
func New() exui.Page { return newDemo() }

// Result is always empty: this page is a reader, not a prompt.
func (a *demoApp) Result() []exui.Field { return nil }

// Shell exposes this page's chrome to the tour host.
func (a *demoApp) Shell() *exui.Chrome { return a.chrome }
