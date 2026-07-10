// Command timefield demos snap/timepicker's TimeFieldModel (used by the VHS tape).
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/timepicker"
)

type demoApp struct{ tf *timepicker.TimeFieldModel }

func (a demoApp) Init() tea.Cmd { return a.tf.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	return v
}

func main() {
	app := demoApp{tf: timepicker.NewTimeField(8, 30)}
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
