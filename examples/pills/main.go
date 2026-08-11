// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package pills implements the `snap_input pills` subcommand: a script-usable PillShape picker built on snap/styles:
// every shape is previewed as pills, a segmented pill, a nav strip, and
// breadcrumbs; Enter writes the selected shape's config value to stdout (the
// TUI itself renders on stderr):
//
//	shape=$(go run ./examples/snap_input pills)
//
// --no-help hides the status bar. Quitting (q/esc) prints nothing, exit 1.
package pills

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
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
	picked bool
	chrome *exui.Chrome
	w, h   int
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		// Arrows and the wheel drive the shape cycle; tab is deliberately not
		// bound here, so the tour can use it to change page.
		switch msg.String() {
		case "left", "up":
			a.sel = (a.sel + len(a.shapes) - 1) % len(a.shapes)
		case "right", "down":
			a.sel = (a.sel + 1) % len(a.shapes)
		case "enter":
			a.picked = true
			return a, exui.Confirm()
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

// Column cell widths. Every cell is centered in its column so the gallery
// reads as a table: name | Go | version | passing | segmented status pill.
// Widths are sized to the widest shape's rendering (Fade's two-cell caps).
// Style.Width/Align pad by terminal cells, so alignment survives any glyphs.
const (
	nameColW    = 12
	goColW      = 10
	versionColW = 12
	passColW    = 13
	segColW     = 28
)

// colCell centers content within a fixed-width column.
func colCell(content string, w int) string {
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, content)
}

func (a *demoApp) shapeRow(shape styles.PillShape, selected bool) string {
	st := styles.PillStyles{Shape: shape}
	label := shape.DisplayName()
	if shape.NeedsNerdFont() {
		label += "*"
	}
	name := label
	if selected {
		name = lipgloss.NewStyle().Bold(true).Foreground(cBlue).Render("▶ " + label)
	}
	segmented := styles.SegmentedPill([]styles.PillSegment{
		{Text: " master ", Bg: cBlue},
		{Text: " +2 ", Bg: cGreen},
		{Text: " ~1 ", Bg: cPeach},
		{Text: " !3 ", Bg: cRed},
	}, st)
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		colCell(name, nameColW),
		colCell(styles.Pill("Go", nil, cBlue, st), goColW),
		colCell(styles.Pill("v0.1.5", nil, cMauve, st), versionColW),
		colCell(styles.Pill("passing", nil, cGreen, st), passColW),
		colCell(segmented, segColW),
	)
}

func (a *demoApp) View() tea.View {
	dim := exui.Dim()

	rows := make([]string, 0, len(a.shapes)+6)
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
	navRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinHorizontal(lipgloss.Top, nav...)),
		lipgloss.NewStyle().PaddingLeft(4).Render(crumbs),
	)
	rows = append(
		rows,
		"",
		dim.Render("nav + breadcrumbs ("+sel.DisplayName()+"):"),
		navRow,
	)

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
	v.OnMouse = a.onMouse
	a.chrome.Frame(&v, a.h)
	return v
}

// New builds the pills page.
func New() exui.Page { return newDemo() }

func newDemo() *demoApp {
	a := &demoApp{
		shapes: styles.PillShapes(),
		chrome: exui.NewChrome(
			exui.Bind("←→ wheel", "shape (* needs Nerd Font)"),
			exui.Bind("enter", "pick"),
			exui.Bind("q", "quit"),
		),
	}
	// Start on the first pure-Unicode shape so the demo (and its rendered
	// gif, whose font has no Powerline glyphs) opens on caps that show
	// everywhere; the Nerd Font shapes are still in the cycle.
	for i, s := range a.shapes {
		if !s.NeedsNerdFont() {
			a.sel = i
			break
		}
	}
	return a
}

// Result reports the chosen shape's config value, and nothing until picked.
func (a *demoApp) Result() []exui.Field {
	if !a.picked {
		return nil
	}
	return []exui.Field{exui.F("shape", string(a.shapes[a.sel]))}
}

// Shell exposes this page's chrome to the tour host.
func (a *demoApp) Shell() *exui.Chrome { return a.chrome }
