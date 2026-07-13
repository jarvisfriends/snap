// Command linechart demos snap/charts' braille line chart model: two live
// series (a sine sweep and its noisy echo) stream through ID-routed
// LineDataMsgs into a LineChartModel that stretches to fill the window —
// braille dots give 2x4 sub-cell resolution, and overlapping series blend
// their colors per cell. q quits.
package main

import (
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/charts"
	"github.com/jarvisfriends/snap/examples/internal/exui"
)

const (
	chartID = "waves"
	window  = 120 // rolling points kept per series
)

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

type demoApp struct {
	chart  *charts.LineChartModel
	chrome *exui.Chrome
	sine   []float64
	echo   []float64
	t      float64
	w, h   int
}

func newDemo() *demoApp {
	c := charts.NewLineChart(chartID)
	c.MaxVal = 2.2 // fixed scale so the waves breathe inside a stable frame
	return &demoApp{chart: c, chrome: exui.NewChrome(exui.Bind("any key", "quit"))}
}

func (a *demoApp) Init() tea.Cmd { return tick() }

// step advances the generators one sample and returns the refreshed series.
func (a *demoApp) step() []charts.LineSeries {
	a.t += 0.15
	sine := 1.1 + math.Sin(a.t)
	// Deterministic pseudo-noise (fast incommensurate sines) keeps the echo
	// jittery without pulling in a random source.
	noise := 0.08*math.Sin(a.t*7.3) + 0.05*math.Sin(a.t*11.7)
	echo := 1.1 + 0.8*math.Sin(a.t-0.9) + noise
	a.sine = append(a.sine, sine)
	a.echo = append(a.echo, echo)
	if len(a.sine) > window {
		a.sine = a.sine[len(a.sine)-window:]
		a.echo = a.echo[len(a.echo)-window:]
	}
	return []charts.LineSeries{
		{Label: "sine", Color: lipgloss.Color("#89b4fa"), Data: a.sine},
		{Label: "echo", Color: lipgloss.Color("#fab387"), Data: a.echo},
	}
}

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Reserve one line for the header (and one for the help bar); the
		// chart fills the rest.
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetWidth(msg.Width)
		a.chart.SetSize(msg.Width, max(msg.Height-1-a.chrome.Height(), 4))
		return a, nil
	case tickMsg:
		// The canonical wiring: producers tag data with the chart's ID and
		// hosts forward everything; the chart ignores other IDs.
		m, _ := a.chart.Update(charts.LineDataMsg{ID: chartID, Series: a.step()})
		if c, ok := m.(*charts.LineChartModel); ok {
			a.chart = c
		}
		return a, tick()
	case tea.KeyPressMsg:
		return a, tea.Quit
	}
	return a, nil
}

func (a *demoApp) View() tea.View {
	// MaxWidth truncates the header on narrow windows so it can never pad
	// the joined frame wider than the chart.
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).MaxWidth(max(a.w, 1))
	header := dim.Render("braille line chart — 2x4 dots per cell, blended overlaps")
	v := tea.NewView(a.chrome.Attach(
		lipgloss.JoinVertical(lipgloss.Left, header, a.chart.View().Content), a.h))
	v.AltScreen = true
	return v
}

func main() {
	exui.Init()
	if _, err := exui.Program(newDemo()).Run(); err != nil {
		exui.Fatal(err)
	}
}
