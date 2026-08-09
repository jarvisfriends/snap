// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package timepicker implements the `snap_input timepicker` subcommand: a script-usable time prompt built on snap/timepicker:
// confirm a time and HH:MM:SS is written to stdout (the TUI itself renders on
// stderr), so a shell can capture it:
//
//	when=$(go run ./examples/snap_input timepicker)
//
// --no-help hides the status bar. Canceling (esc) prints nothing, exit 1.
package timepicker

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/styles"
	"github.com/jarvisfriends/snap/timepicker"
)

type demoApp struct {
	tf     *timepicker.TimeFieldModel
	chrome *exui.Chrome
	height int
}

func newDemo(start time.Time) demoApp {
	tf := timepicker.NewTimeField(start)
	// Theme the field from the shared palette instead of its fixed defaults.
	styles.ApplyTimeFieldTheme(tf, exui.Theme())
	tf.ShowSeconds = true
	// The status bar below carries the key hints; the field's own help line
	// would show the same thing twice.
	tf.HideHelp = true
	return demoApp{
		tf: tf,
		chrome: exui.NewChrome(
			exui.Bind("←→↑↓", "move"),
			exui.Bind("0-9", "type"),
			exui.Bind("space", "dropdown"),
			exui.Bind("enter", "confirm"),
			exui.Bind("esc", "cancel"),
		),
	}
}

func (a demoApp) Init() tea.Cmd { return a.tf.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.height = msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
		return a, nil
	case tea.MouseMsg:
		// Mouse events reach the component through the root view's OnMouse
		// (Bubble Tea delivers the raw event to BOTH OnMouse and Update);
		// forwarding them here too would double-process every click.
		return a, nil
	}
	m, cmd := a.tf.Update(msg)
	if tf, ok := m.(*timepicker.TimeFieldModel); ok {
		a.tf = tf
	}
	if a.tf.Done || a.tf.Aborted {
		return a, exui.Confirm()
	}
	return a, cmd
}

// View enables mouse reporting on the root view and stacks the shared help
// bar on the terminal's bottom line under the field.
func (a demoApp) View() tea.View {
	v := a.tf.View()
	// A themed confirm button gives the mouse a way to finish (parity with
	// enter): clicking it marks the field Done and quits.
	t := exui.Theme()
	button := t.Styles.SelectedItem.Padding(0, 1).Render("✓ confirm")
	fieldH := lipgloss.Height(v.Content)
	btnW := lipgloss.Width(button)
	v.SetContent(lipgloss.JoinVertical(lipgloss.Left, v.Content, "", button))
	prev := v.OnMouse
	v.OnMouse = func(mm tea.MouseMsg) tea.Cmd {
		if click, ok := mm.(tea.MouseClickMsg); ok {
			if m := click.Mouse(); m.Y == fieldH+1 && m.X < btnW {
				a.tf.Done = true
				return exui.Confirm()
			}
		}
		if prev != nil {
			return prev(mm)
		}
		return nil
	}
	a.chrome.Frame(&v, a.height)
	return v
}

// New builds the timepicker page on a fixed demo time, so the tape and the
// golden tests see the same starting value on every run.
func New() exui.Page {
	return newDemo(time.Date(2026, 7, 10, 8, 30, 45, 0, time.Local))
}

// Result reports the confirmed time, and nothing until the field is done.
func (a demoApp) Result() []exui.Field {
	if !a.tf.Done {
		return nil
	}
	return []exui.Field{exui.F("time", a.tf.Time().Format("15:04:05"))}
}

// Shell exposes this page's chrome to the tour host.
// The field's type-ahead only takes digits, so it never needs to claim "q".
func (a demoApp) Shell() *exui.Chrome { return a.chrome }
