// Command datepicker demos snap/datepicker standalone (used by the VHS tape).
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/datepicker"
)

type demoApp struct{ dp *datepicker.DatePickerModel }

func (a demoApp) Init() tea.Cmd { return a.dp.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		if s := k.String(); s == "q" || s == "ctrl+c" {
			return a, tea.Quit
		}
	}
	m, cmd := a.dp.Update(msg)
	if dp, ok := m.(*datepicker.DatePickerModel); ok {
		a.dp = dp
	}
	if a.dp.Selected {
		return a, tea.Quit
	}
	return a, cmd
}


func (a demoApp) View() tea.View { return a.dp.View() }

func main() {
	app := demoApp{dp: datepicker.New(time.Now())}
	final, err := tea.NewProgram(app).Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Print the confirmed date after the alt-screen restores, so the choice
	// stays visible in the console (and the VHS tape captures a clean exit).
	if a, ok := final.(demoApp); ok && a.dp.Selected {
		fmt.Println("Selected:", a.dp.Time.Format("Monday, January 2, 2006"))
	}
}
