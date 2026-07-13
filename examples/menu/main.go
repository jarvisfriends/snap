// Command menu is a script-usable context-menu picker built on snap/menu:
// right-click anywhere (or press m) to pop the menu, choose an item, and the
// chosen item's ID is written to stdout (the TUI itself renders on stderr):
//
//	action=$(go run ./examples/menu)
//
// --no-help hides the status bar. Quitting (q/esc) prints nothing, exit 1.
package main

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/menu"
)

type demoApp struct {
	menu   menu.Menu
	chrome *exui.Chrome
	picked string
	w, h   int
}

func newDemo() *demoApp {
	return &demoApp{
		chrome: exui.NewChrome(
			exui.Bind("right-click/m", "open menu"),
			exui.Bind("↑/↓", "move"),
			exui.Bind("enter/click", "choose"),
			exui.Bind("esc", "dismiss"),
			exui.Bind("q", "quit"),
		),
	}
}

func items() []menu.Item {
	return []menu.Item{
		{ID: "open", Label: "Open"},
		{ID: "rename", Label: "Rename", Disabled: true},
		{ID: "copy", Label: "Copy path"},
		{ID: "delete", Label: "Delete"},
	}
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetWidth(msg.Width)
	case tea.KeyPressMsg:
		switch {
		case a.menu.IsOpen():
			// HandleKey mirrors HandleMouse: the open menu owns the keyboard.
			if chosen, _ := a.menu.HandleKey(msg); chosen != nil {
				a.picked = chosen.ID
				return a, tea.Quit
			}
		case msg.String() == "m":
			a.menu.Open(a.w/2, a.h/2, items(), "keyboard")
		case msg.String() == "q" || msg.String() == "esc" || msg.String() == "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

// onMouse owns pointer input per the snap contract: while the menu is open
// it consumes events; a right-click opens it at the pointer.
func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	if a.menu.IsOpen() {
		if chosen, _ := a.menu.HandleMouse(mm, a.w, a.h); chosen != nil {
			a.picked = chosen.ID
			return tea.Quit
		}
		return nil
	}
	if click, ok := mm.(tea.MouseClickMsg); ok && click.Button == tea.MouseRight {
		me := click.Mouse()
		a.menu.Open(me.X, me.Y, items(), "cell "+strconv.Itoa(me.X)+","+strconv.Itoa(me.Y))
	}
	return nil
}

func (a *demoApp) View() tea.View {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	line := dim.Render(strings.Repeat("·", max(a.w, 1)))
	paneH := max(a.h-a.chrome.Height(), 1)
	rows := make([]string, 0, paneH)
	for range paneH {
		rows = append(rows, line)
	}
	base := lipgloss.JoinVertical(lipgloss.Left, rows...)
	v := tea.NewView(a.menu.Composite(base, a.w, paneH))
	a.chrome.Apply(&v, a.h)
	v.MouseMode = tea.MouseModeCellMotion
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
