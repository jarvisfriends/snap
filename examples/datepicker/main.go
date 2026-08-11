// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package datepicker implements the `snap_input datepicker` subcommand: a script-usable date prompt built on snap/datepicker:
// pick a day and the ISO date is written to stdout (the TUI itself renders on
// stderr), so a shell can capture it:
//
//	date=$(go run ./examples/snap_input datepicker)
//
// --no-help hides the status bar. Canceling (q) prints nothing, exit 1.
package datepicker

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/datepicker"
	"github.com/jarvisfriends/snap/examples/internal/exui"
)

type demoApp struct {
	dp     *datepicker.DatePickerModel
	chrome *exui.Chrome
	height int
}

func newDemo(start time.Time) demoApp {
	return demoApp{
		dp: datepicker.New(start),
		chrome: exui.NewChrome(
			exui.Bind("↑↓←→", "move"),
			exui.Bind("[]", "month"),
			exui.Bind("{}", "year"),
			exui.Bind("enter", "pick"),
			exui.Bind("q", "quit"),
		),
	}
}

func (a demoApp) Init() tea.Cmd { return a.dp.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.height = msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
		// Fall through to the component so it sizes itself to the window
		// (minus the help bar) — otherwise its natural height can collide
		// with the bar row on short terminals.
		m, cmd := a.dp.Update(tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: max(msg.Height-a.chrome.Height(), 1),
		})
		if dp, ok := m.(*datepicker.DatePickerModel); ok {
			a.dp = dp
		}
		return a, cmd
	case tea.MouseMsg:
		// Mouse events reach the component through the root view's OnMouse
		// (Bubble Tea delivers the raw event to BOTH OnMouse and Update);
		// forwarding them here too would double-process every click.
		return a, nil
	}
	m, cmd := a.dp.Update(msg)
	if dp, ok := m.(*datepicker.DatePickerModel); ok {
		a.dp = dp
	}
	if a.dp.Selected {
		return a, exui.Confirm()
	}
	return a, cmd
}

// View enables mouse reporting on the root view and stacks the shared help
// bar on the terminal's bottom line under the calendar.
func (a demoApp) View() tea.View {
	v := a.dp.View()
	// A click on the already-selected day sets Selected inside the picker's
	// own mouse handler; chain a quit so confirming by mouse ends the demo
	// exactly like enter does.
	prev := v.OnMouse
	v.OnMouse = func(mm tea.MouseMsg) tea.Cmd {
		cmd := prev(mm)
		if a.dp.Selected {
			return exui.Confirm()
		}
		return cmd
	}
	a.chrome.Frame(&v, a.height)
	return v
}

// New builds the datepicker page, opened on today.
func New() exui.Page { return newDemo(time.Now()) }

// Result reports the confirmed date, and nothing until a day is picked.
func (a demoApp) Result() []exui.Field {
	if !a.dp.Selected {
		return nil
	}
	return []exui.Field{exui.F("date", a.dp.Time.Format("2006-01-02"))}
}

// Shell exposes this page's chrome to the tour host.
func (a demoApp) Shell() *exui.Chrome { return a.chrome }
