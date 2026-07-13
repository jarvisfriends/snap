// Command cellcanvas demos snap/charts' whole-cell canvas and color
// gradients: a classic plasma field animates over a truecolor palette built
// from chained charts.Gradient blends. Each cell renders '▀' with an
// independent foreground (top pixel) and background (bottom pixel), doubling
// the vertical resolution, and CellCanvas.String() batches the escapes so
// colors are re-emitted only when they change. q (any key) quits.
package main

import (
	"image/color"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/charts"
	"github.com/jarvisfriends/snap/examples/internal/exui"
)

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(66*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

// palette chains two gradients (blue → mauve → peach) into one smooth ramp.
func palette() []color.Color {
	blue := lipgloss.Color("#89b4fa")
	mauve := lipgloss.Color("#cba6f7")
	peach := lipgloss.Color("#fab387")
	return append(charts.Gradient(blue, mauve, 32), charts.Gradient(mauve, peach, 32)...)
}

type demoApp struct {
	canvas *charts.CellCanvas
	colors []color.Color
	chrome *exui.Chrome
	w, h   int
	termH  int
	t      float64
}

func (a *demoApp) Init() tea.Cmd { return tick() }

// plasma evaluates the field at pixel (x, y) — three drifting sine waves —
// and maps it onto a palette index.
func (a *demoApp) plasma(x, y int) color.Color {
	fx, fy := float64(x), float64(y)
	v := math.Sin(fx/9+a.t) +
		math.Sin((fx+fy)/13) +
		math.Sin(math.Sqrt(fx*fx+fy*fy)/8+a.t/2)
	// v spans [-3, 3]; normalize into palette bounds.
	idx := int((v + 3) / 6 * float64(len(a.colors)-1))
	return a.colors[min(max(idx, 0), len(a.colors)-1)]
}

func (a *demoApp) redraw() {
	for y := range a.h {
		for x := range a.w {
			// '▀' shows fg on the top half and bg on the bottom half, so one
			// cell carries two vertically stacked plasma pixels.
			a.canvas.Set(x, y, '▀', a.plasma(x, 2*y), a.plasma(x, 2*y+1))
		}
	}
}

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.termH = msg.Height
		a.chrome.SetWidth(msg.Width)
		a.w, a.h = max(msg.Width, 1), max(msg.Height-1-a.chrome.Height(), 1)
		a.canvas = charts.NewCellCanvas(a.w, a.h, nil, nil)
		a.redraw()
		return a, nil
	case tickMsg:
		if a.canvas != nil {
			a.t += 0.18
			a.redraw()
		}
		return a, tick()
	case tea.KeyPressMsg:
		return a, tea.Quit
	}
	return a, nil
}

func (a *demoApp) View() tea.View {
	// MaxWidth truncates the header on narrow windows so it can never pad
	// the joined frame wider than the canvas.
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MaxWidth(max(a.w, 1))
	header := dim.Render("charts.CellCanvas + charts.Gradient — truecolor plasma, 2 pixels per cell")
	body := ""
	if a.canvas != nil {
		body = a.canvas.String()
	}
	v := tea.NewView(a.chrome.Attach(lipgloss.JoinVertical(lipgloss.Left, header, body), a.termH))
	v.AltScreen = true
	return v
}

func main() {
	exui.Init()
	app := &demoApp{colors: palette(), chrome: exui.NewChrome(exui.Bind("any key", "quit"))}
	if _, err := exui.Program(app).Run(); err != nil {
		exui.Fatal(err)
	}
}
