// Command timefield demos snap/timepicker's TimeFieldModel (used by the VHS tape).
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/timepicker"
)

type demoApp struct{ tf *timepicker.TimeFieldModel }

func (a demoApp) Init() tea.Cmd { return a.tf.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Mouse events reach the component through the root view's OnMouse
	// (Bubble Tea delivers the raw event to BOTH OnMouse and Update);
	// forwarding them here too would double-process every click — the first
	// click would highlight AND select.
	if _, isMouse := msg.(tea.MouseMsg); isMouse {
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

// View enables mouse reporting on the root view — in Bubble Tea v2 the
// terminal only sends mouse events when the root view asks for them, so
// without this the component's OnMouse never fires.
func (a demoApp) View() tea.View {
	v := a.tf.View()
	v.MouseMode = tea.MouseModeCellMotion
	// AltScreen gives the demo the whole window: rendered inline (the
	// default), the content is pinned to the prompt line and the tall VHS
	// window stays empty — the "Height not showing up" symptom.
	v.AltScreen = true
	return v
}

func main() {
	tf := timepicker.NewTimeField(time.Date(2026, 7, 10, 8, 30, 45, 0, time.Local))
	tf.ShowSeconds = true
	app := demoApp{tf: tf}
	final, err := tea.NewProgram(app).Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Print the confirmed time after the alt-screen restores, so the choice
	// stays visible in the console (and the VHS tape captures a clean exit).
	if a, ok := final.(demoApp); ok && a.tf.Done {
		fmt.Printf("Selected: %s\n", a.tf.Time().Format("15:04:05"))
	}
}
