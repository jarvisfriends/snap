// Command table demos snap/table: arrow keys move the selection, clicking a
// header sorts that column, `/` filters, Enter (or double-click) opens the
// row, q quits.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/styles"
	"github.com/jarvisfriends/snap/table"
	"github.com/jarvisfriends/snap/uifx"
)

type demoApp struct {
	tbl    *table.TableModel
	status string
	w, h   int
}

func newDemo() *demoApp {
	t := table.New([]table.Column{
		{Title: "Service", Filter: true},
		{Title: "Region", Filter: true},
		{Title: "P99 ms"},
		{Title: "Errors"},
	})
	t.SetRows([]table.Row{
		{Key: "api", Cells: []table.Cell{table.Text("api"), table.Text("us-east"), table.Num("41", 41), table.Num("3", 3)}},
		{Key: "web", Cells: []table.Cell{table.Text("web"), table.Text("us-east"), table.Num("120", 120), table.Num("0", 0)}},
		{Key: "auth", Cells: []table.Cell{table.Text("auth"), table.Text("eu-west"), table.Num("9", 9), table.Num("12", 12)}},
		{Key: "batch", Cells: []table.Cell{table.Text("batch"), table.Text("us-west"), table.Num("310", 310), table.Num("1", 1)}},
		{Key: "cdn", Cells: []table.Cell{table.Text("cdn"), table.Text("global"), table.Num("18", 18), table.Num("0", 0)}},
	})
	return &demoApp{tbl: t, status: "↑↓ move · click header sorts · / filters · q quits"}
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.tbl.SetSize(msg.Width, msg.Height-2)
		return a, nil
	case table.OpenDetailMsg:
		a.status = "opened row: " + msg.Key
		return a, nil
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case tea.MouseMsg:
		// Pointer input arrives via the root view's OnMouse below.
		return a, nil
	}
	return a, a.tbl.Update(msg)
}

// onMouse routes clicks and wheels to the table's handlers (page-relative
// coordinates; the table sits one line below the status header).
func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	me := mm.Mouse()
	switch mm.(type) {
	case tea.MouseClickMsg:
		if me.Button == tea.MouseLeft {
			return a.tbl.HandleClick(me.X, me.Y-1)
		}
	case tea.MouseWheelMsg:
		switch me.Button {
		case tea.MouseWheelUp:
			a.tbl.HandleWheel(true)
		case tea.MouseWheelDown:
			a.tbl.HandleWheel(false)
		}
	}
	return nil
}

func (a *demoApp) View() tea.View {
	frame := a.tbl.View(styles.Active(), 1) + "\n\n" + a.status
	v := tea.NewView(frame)
	v.MouseMode = uifx.LevelMedium.MouseMode()
	v.AltScreen = true
	v.OnMouse = a.onMouse
	return v
}

func main() {
	if _, err := tea.NewProgram(newDemo()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
