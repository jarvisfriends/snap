// Command pills demos snap/styles' pill shapes: single-color pills,
// color-divided segmented pills, a nav strip, and breadcrumb separators in
// every PillShape. Left/right (or the wheel) select the shape; q quits.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/styles"
)

// Categorical fills (Catppuccin Mocha accents) — identifiers like syntax
// highlighting, not theme colors, so they are hardcoded app-side.
var (
	cBlue  = lipgloss.Color("#89b4fa")
	cPeach = lipgloss.Color("#fab387")
	cGreen = lipgloss.Color("#a6e3a1")
	cMauve = lipgloss.Color("#cba6f7")
	cRed   = lipgloss.Color("#f38ba8")
	cGray  = lipgloss.Color("#45475a")
)

type demoApp struct {
	shapes []styles.PillShape
	sel    int
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "left", "shift+tab":
			a.sel = (a.sel + len(a.shapes) - 1) % len(a.shapes)
		case "right", "tab":
			a.sel = (a.sel + 1) % len(a.shapes)
		default:
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	if wheel, ok := mm.(tea.MouseWheelMsg); ok {
		switch wheel.Button {
		case tea.MouseWheelUp:
			a.sel = (a.sel + len(a.shapes) - 1) % len(a.shapes)
		case tea.MouseWheelDown:
			a.sel = (a.sel + 1) % len(a.shapes)
		}
	}
	return nil
}

func (a *demoApp) shapeRow(shape styles.PillShape, selected bool) string {
	st := styles.PillStyles{Shape: shape}
	name := fmt.Sprintf("  %-6s", shape.DisplayName())
	if selected {
		name = lipgloss.NewStyle().Bold(true).Foreground(cBlue).Render("▶ " + name[2:])
	}
	badges := strings.Join([]string{
		styles.Pill("Go", nil, cBlue, st),
		styles.Pill("v0.1.5", nil, cMauve, st),
		styles.Pill("passing", nil, cGreen, st),
	}, " ")
	segmented := styles.SegmentedPill([]styles.PillSegment{
		{Text: " master ", Bg: cBlue},
		{Text: " +2 ", Bg: cGreen},
		{Text: " ~1 ", Bg: cPeach},
		{Text: " !3 ", Bg: cRed},
	}, st)
	return name + "  " + badges + "   " + segmented
}

func (a *demoApp) View() tea.View {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	rows := make([]string, 0, len(a.shapes)+6)
	rows = append(rows,
		dim.Render("←/→ or wheel select shape — q quits"),
		"")
	for i, shape := range a.shapes {
		rows = append(rows, a.shapeRow(shape, i == a.sel))
	}

	sel := a.shapes[a.sel]
	st := styles.PillStyles{Shape: sel}
	nav := make([]string, 0, 3)
	for i, label := range []string{"Home", "Settings", "About"} {
		fill := cGray
		if i == 0 {
			fill = cMauve
		}
		nav = append(nav, styles.Pill(label, nil, fill, st))
	}
	crumbs := styles.Breadcrumbs([]string{"home", "projects", "snap"}, dim, st)
	rows = append(rows,
		"",
		dim.Render(fmt.Sprintf("nav + breadcrumbs (%s):", sel.DisplayName())),
		"  "+strings.Join(nav, " ")+"    "+crumbs,
	)

	v := tea.NewView(strings.Join(rows, "\n"))
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	v.OnMouse = a.onMouse
	return v
}

func main() {
	app := &demoApp{shapes: styles.PillShapes()}
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
