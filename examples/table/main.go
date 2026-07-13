// Command table is a script-usable row picker built on snap/table: browse,
// sort, and filter, then Enter (or double-click) writes the chosen row's key
// to stdout (the TUI itself renders on stderr), so a shell can capture it:
//
//	service=$(go run ./examples/table)
//
// --no-help hides the status bar. Quitting (q) prints nothing, exit 1.
package main

import (
	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/table"
	"github.com/jarvisfriends/snap/uifx"
)

type demoApp struct {
	tbl    *table.TableModel
	chrome *exui.Chrome
	picked string
	w, h   int
}

func newDemo() *demoApp {
	t := table.New([]table.Column{
		{Title: "Service", Filter: true},
		{Title: "Region", Filter: true},
		{Title: "P99 ms"},
		{Title: "Errors"},
	})
	// The status bar below carries the key hints; the table footer's own
	// hint text would show the same thing twice.
	t.HideFooterHint = true
	t.SetRows([]table.Row{
		{Key: "api", Cells: []table.Cell{table.Text("api"), table.Text("us-east"), table.Num("41", 41), table.Num("3", 3)}},
		{Key: "web", Cells: []table.Cell{table.Text("web"), table.Text("us-east"), table.Num("120", 120), table.Num("0", 0)}},
		{Key: "auth", Cells: []table.Cell{table.Text("auth"), table.Text("eu-west"), table.Num("9", 9), table.Num("12", 12)}},
		{Key: "batch", Cells: []table.Cell{table.Text("batch"), table.Text("us-west"), table.Num("310", 310), table.Num("1", 1)}},
		{Key: "cdn", Cells: []table.Cell{table.Text("cdn"), table.Text("global"), table.Num("18", 18), table.Num("0", 0)}},
	})
	return &demoApp{
		tbl: t,
		chrome: exui.NewChrome(
			exui.Bind("↑/↓", "move"),
			exui.Bind("s/click header", "sort"),
			exui.Bind("/", "filter"),
			exui.Bind("enter", "pick"),
			exui.Bind("q", "quit"),
		),
	}
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetWidth(msg.Width)
		a.tbl.SetSize(msg.Width, msg.Height-a.chrome.Height())
		return a, nil
	case table.OpenDetailMsg:
		a.picked = msg.Key
		return a, tea.Quit
	case tea.KeyPressMsg:
		if s := msg.String(); !a.tbl.Filtering() && (s == "q" || s == "ctrl+c") {
			return a, tea.Quit
		}
	case tea.MouseMsg:
		// Pointer input arrives via the root view's OnMouse below.
		return a, nil
	}
	return a, a.tbl.Update(msg)
}

// onMouse routes clicks and wheels to the table's handlers (the table starts
// at the top of the frame, so screen coordinates are page coordinates).
func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	me := mm.Mouse()
	switch mm.(type) {
	case tea.MouseClickMsg:
		if me.Button == tea.MouseLeft {
			return a.tbl.HandleClick(me.X, me.Y)
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
	v := tea.NewView(a.tbl.View(exui.Theme(), 0))
	a.chrome.Apply(&v, a.h)
	v.MouseMode = uifx.LevelMedium.MouseMode()
	v.AltScreen = true
	v.OnMouse = a.onMouse
	return v
}

func main() {
	exui.Init()
	final, err := exui.Program(newDemo()).Run()
	if err != nil {
		exui.Fatal(err)
	}
	if a, ok := final.(*demoApp); ok && a.picked != "" {
		exui.Finish(true, a.picked)
	}
	exui.Finish(false)
}
