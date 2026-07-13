// Command timepicker is a script-usable time prompt built on snap/timepicker:
// confirm a time and HH:MM:SS is written to stdout (the TUI itself renders on
// stderr), so a shell can capture it:
//
//	when=$(go run ./examples/timepicker)
//
// --no-help hides the status bar. Canceling (esc) prints nothing, exit 1.
package main

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/timepicker"
)

type demoApp struct {
	tf     *timepicker.TimeFieldModel
	chrome *exui.Chrome
	height int
}

func newDemo(start time.Time) demoApp {
	tf := timepicker.NewTimeField(start)
	tf.ShowSeconds = true
	return demoApp{
		tf: tf,
		chrome: exui.NewChrome(
			exui.Bind("←/→", "column"),
			exui.Bind("↑/↓", "spin"),
			exui.Bind("0-9", "type"),
			exui.Bind("space", "dropdown"),
			exui.Bind("enter", "confirm"),
			exui.Bind("esc", "cancel"),
		),
	}
}

func (a demoApp) Init() tea.Cmd { return a.tf.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.height = msg.Height
		a.chrome.SetWidth(msg.Width)
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
		return a, tea.Quit
	}
	return a, cmd
}

// View enables mouse reporting on the root view and stacks the shared help
// bar on the terminal's bottom line under the field.
func (a demoApp) View() tea.View {
	v := a.tf.View()
	v.SetContent(a.chrome.Attach(v.Content, a.height))
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	return v
}

func main() {
	exui.Init()
	final, err := exui.Program(newDemo(time.Date(2026, 7, 10, 8, 30, 45, 0, time.Local))).Run()
	if err != nil {
		exui.Fatal(err)
	}
	if a, ok := final.(demoApp); ok && a.tf.Done {
		exui.Finish(true, a.tf.Time().Format("15:04:05"))
	}
	exui.Finish(false)
}
