// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package forms implements the `snap_input forms` subcommand: a script-usable task form proving snap/forms extends huh
// rather than replacing it: a plain huh.Form whose fields validate through
// forms.HuhValidate(ParseRequired/ParseDuration/ParseISODate), with
// SplitAndClean cleaning the tags on submit. One form submit yields several
// typed values at once, so completing the form writes them to stdout as a
// small YAML document (still machine-readable; the TUI itself renders on
// stderr):
//
//	go run ./examples/snap_input forms | yq .duration
//
// --no-help hides the status bar (huh's own help line is off — the bar shows
// the keys instead). Ctrl+C cancels: nothing printed, exit 1.
package forms

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/forms"
	"github.com/jarvisfriends/snap/styles"
)

// newTaskForm builds the huh form: standard huh fields, snap/forms parsers as
// their validators, themed by the shared example palette.
func newTaskForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("task").
				Title("Task").
				Placeholder("what needs doing (required)").
				Validate(forms.HuhValidate(forms.ParseRequired, "task")),
			huh.NewInput().
				Key("duration").
				Title("Duration").
				Placeholder("5m, 1h, 7h30m").
				Validate(forms.HuhValidate(forms.ParseDuration, "duration")),
			huh.NewInput().
				Key("due").
				Title("Due date").
				Placeholder("YYYY-MM-DD").
				Validate(forms.HuhValidate(forms.ParseISODate, "due date")),
			huh.NewInput().
				Key("tags").
				Title("Tags").
				Placeholder("comma, separated , list,, of tags"),
		),
	).
		WithTheme(styles.HuhThemeFunc()).
		WithShowHelp(false) // the status bar below carries the keys
}

type demoApp struct {
	form   *huh.Form
	chrome *exui.Chrome
	height int
}

func newDemo() *demoApp {
	a := &demoApp{
		form: newTaskForm(),
		chrome: exui.NewChrome(
			exui.Bind("tab shift+tab", "move"),
			exui.Bind("enter", "next"),
			exui.Bind("ctrl+c", "cancel"),
		),
	}
	// Every keystroke belongs to the form's inputs while it is running, so
	// the shell must not claim "q" and the tour must not claim tab.
	a.chrome.SetCapture(a.Capturing)
	return a
}

func (a *demoApp) Init() tea.Cmd { return a.form.Init() }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		a.height = msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
	}
	model, cmd := a.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		a.form = f
	}
	if a.form.State != huh.StateNormal {
		return a, exui.Confirm()
	}
	return a, cmd
}

func (a *demoApp) View() tea.View {
	v := tea.NewView(a.form.View())
	a.chrome.Frame(&v, a.height)
	return v
}

// New builds the forms page.
func New() exui.Page { return newDemo() }

// Result reports the submitted task, and nothing until the form completes.
// One submit yields several values at once.
func (a *demoApp) Result() []exui.Field {
	if a.form.State != huh.StateCompleted {
		return nil
	}
	// The same parsers that validated the fields now produce the typed
	// values, so output can never disagree with what validation accepted.
	d, _ := forms.ParseDuration(a.form.GetString("duration"), "duration")
	due, _ := forms.ParseISODate(a.form.GetString("due"), "due date")
	tags := forms.SplitAndClean(a.form.GetString("tags"), ",")
	return []exui.Field{
		exui.F("task", strings.TrimSpace(a.form.GetString("task"))),
		exui.F("duration", d.String()),
		exui.F("due", due.Format(time.DateOnly)),
		exui.F("tags", strings.Join(tags, ", ")),
	}
}

// Shell exposes this page's chrome to the tour host.
func (a *demoApp) Shell() *exui.Chrome { return a.chrome }

// Capturing reports whether the form is still taking input.
func (a *demoApp) Capturing() bool { return a.form.State == huh.StateNormal }
