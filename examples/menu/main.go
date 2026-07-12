// Command menu demos snap/menu: right-click anywhere in the pane (or press
// m) to pop a context menu at the pointer; hover, wheel, and arrow keys move
// the cursor; click or Enter chooses; clicking outside or Esc dismisses.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/menu"
)

type demoApp struct {
	menu   menu.Menu
	w, h   int
	status string
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
	case tea.KeyPressMsg:
		switch {
		case a.menu.IsOpen():
			// HandleKey mirrors HandleMouse: the open menu owns the keyboard.
			if chosen, _ := a.menu.HandleKey(msg); chosen != nil {
				a.status = "chose " + chosen.ID
			}
		case msg.String() == "m":
			a.menu.Open(a.w/2, a.h/2, items(), "keyboard")
		case msg.String() == "q" || msg.String() == "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

// onMouse owns pointer input per the snap contract: while the menu is open
// it consumes events; a right-click opens it at the pointer.
func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	if a.menu.IsOpen() {
		if chosen, handled := a.menu.HandleMouse(mm, a.w, a.h); chosen != nil {
			a.status = "chose " + chosen.ID
		} else if handled {
			return nil
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
	rows := make([]string, 0, max(a.h-1, 2))
	for range max(a.h-2, 1) {
		rows = append(rows, line)
	}
	rows = append(rows, "right-click (or m) opens the menu — q quits   "+a.status)
	base := lipgloss.JoinVertical(lipgloss.Left, rows...)
	v := tea.NewView(a.menu.Composite(base, a.w, a.h))
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	v.OnMouse = a.onMouse
	return v
}

func main() {
	if _, err := tea.NewProgram(&demoApp{status: "ready"}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
