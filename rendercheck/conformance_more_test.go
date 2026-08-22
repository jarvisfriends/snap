// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package rendercheck

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// statusModel is a minimal StatusProvider host: a frame with a status line
// that states can retitle or hide.
type statusModel struct {
	text    string
	visible bool
}

type setStatusMsg struct {
	text    string
	visible bool
}

func (m statusModel) Init() tea.Cmd { return nil }

func (m statusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s, ok := msg.(setStatusMsg); ok {
		m.text, m.visible = s.text, s.visible
	}
	return m, nil
}

func (m statusModel) View() tea.View {
	content := "body line\n"
	if m.visible {
		content += m.text
	}
	return tea.NewView(content)
}

func (m statusModel) StatusBarContent() (string, bool) { return m.text, m.visible }

func TestCheckStatusBarVisiblePassesAcrossStates(t *testing.T) {
	CheckStatusBarVisible(t, statusModel{text: "READY · 3 items", visible: true}, []tea.Msg{
		setStatusMsg{text: "SAVING · hold on", visible: true},
		setStatusMsg{text: "hidden by design", visible: false},
		setStatusMsg{text: "BACK · all good", visible: true},
	})
}

func TestCheckStatusBarVisibleSkipsNonProviders(t *testing.T) {
	t.Run("skips", func(t *testing.T) {
		CheckStatusBarVisible(t, boxModel{w: 10, h: 3}, nil)
	})
}

// themedModel emits raw ANSI colors keyed by the last theme message, so the
// theme-responsiveness check sees color changes even where lipgloss detects
// a no-color environment (CI pipes).
type themedModel struct{ sgr string }

type themeMsg string

func (m themedModel) Init() tea.Cmd { return nil }

func (m themedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if c, ok := msg.(themeMsg); ok {
		m.sgr = string(c)
	}
	return m, nil
}

func (m themedModel) View() tea.View {
	return tea.NewView("\x1b[" + m.sgr + "mthemed\x1b[0m")
}

func TestCheckThemeResponsiveSeesColorChange(t *testing.T) {
	CheckThemeResponsive(t, themedModel{sgr: "31"}, themeMsg("31"), themeMsg("34"))
}

// colorlessModel renders no ANSI at all; the check must skip, not fail, in
// no-color environments.
type colorlessModel struct{}

func (colorlessModel) Init() tea.Cmd                         { return nil }
func (m colorlessModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (colorlessModel) View() tea.View                        { return tea.NewView("plain") }

func TestCheckThemeResponsiveSkipsWithoutColors(t *testing.T) {
	t.Run("skips", func(t *testing.T) {
		CheckThemeResponsive(t, colorlessModel{}, themeMsg("31"), themeMsg("34"))
	})
}
