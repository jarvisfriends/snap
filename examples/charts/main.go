// Command charts is the canonical multi-chart wiring example: several chart
// models of the same and different types live in one app, every data message
// carries the ID of the chart it belongs to, and the window is split between
// charts on resize via SetSize with layout driven by Used().
//
// The pattern to copy:
//  1. Give every chart a unique ID at construction (NewSparkline("cpu")).
//  2. Producers tag data messages with that ID (SparklinePointMsg{ID: "cpu"}).
//  3. Forward every message to every chart — each consumes only its own ID,
//     so the host never demultiplexes by hand.
//  4. On tea.WindowSizeMsg, divide the space and SetSize each chart; charts
//     stretch to fill and report their actual footprint via Used().
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/charts"
)

// tickMsg drives the fake metrics stream.
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type demoApp struct {
	cpu    *charts.SparklineModel
	mem    *charts.SparklineModel
	pie    *charts.PieModel
	sankey *charts.SankeyModel
	disk   *charts.HBarModel

	t float64
}

func newDemo() demoApp {
	pie := charts.NewPie("share")
	pie.Braille = true
	return demoApp{
		cpu:    charts.NewSparkline("cpu"),
		mem:    charts.NewSparkline("mem"),
		pie:    pie,
		sankey: charts.NewSankey("traffic"),
		disk:   charts.NewHBar("disk"),
	}
}

func (a demoApp) Init() tea.Cmd { return tick() }

// forward hands msg to every chart — each consumes only its own ID.
// (Value receiver: the chart fields are pointers, so updates stick.)
func (a demoApp) forward(msg tea.Msg) {
	_, _ = a.cpu.Update(msg)
	_, _ = a.mem.Update(msg)
	_, _ = a.pie.Update(msg)
	_, _ = a.sankey.Update(msg)
	_, _ = a.disk.Update(msg)
}

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return a, tea.Quit

	case tea.WindowSizeMsg:
		// Split the window: two sparkline rows up top, pie beside sankey
		// below, one bar across the bottom. Each chart stretches to the
		// space it is given.
		half := max(msg.Width/2, 10)
		a.cpu.SetSize(msg.Width-8, 1)
		a.mem.SetSize(msg.Width-8, 1)
		bodyH := max(msg.Height-8, 6)
		a.pie.SetSize(half-2, bodyH)
		a.sankey.SetSize(msg.Width-half-2, bodyH)
		a.disk.SetSize(msg.Width-8, 1)
		return a, nil

	case tickMsg:
		a.t += 0.3
		const apiSvc, webSvc = "api", "web"
		// Producers tag each message with the target chart's ID.
		a.forward(charts.SparklinePointMsg{ID: "cpu", Value: 50 + 40*math.Sin(a.t)})
		a.forward(charts.SparklinePointMsg{ID: "mem", Value: 60 + 25*math.Cos(a.t/2)})
		a.forward(charts.HBarDataMsg{ID: "disk", Pct: 35 + 30*math.Sin(a.t/4)})
		a.forward(charts.PieDataMsg{ID: "share", Slices: []charts.PieSlice{
			{Value: 45, Color: lipgloss.Color("4"), Label: apiSvc},
			{Value: 30 + 10*math.Sin(a.t), Color: lipgloss.Color("2"), Label: webSvc},
			{Value: 15, Color: lipgloss.Color("5"), Label: "batch"},
			{Value: 1 + math.Sin(a.t*3), Color: lipgloss.Color("3"), Label: "cron"},
			{Value: 0.5 + 0.5*math.Cos(a.t*5), Color: lipgloss.Color("6"), Label: "misc"},
		}})
		a.forward(charts.SankeyDataMsg{ID: "traffic", Flows: []charts.SankeyFlow{
			{Source: "lb", Target: apiSvc, Value: 6 + 3*math.Sin(a.t), Color: lipgloss.Color("4")},
			{Source: "lb", Target: webSvc, Value: 4, Color: lipgloss.Color("2")},
			{Source: "cdn", Target: webSvc, Value: 3 + 2*math.Cos(a.t), Color: lipgloss.Color("6")},
		}})
		return a, tick()
	}
	return a, nil
}

func (a demoApp) View() tea.View {
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(6)
	row := func(name, frame string) string {
		return lipgloss.JoinHorizontal(lipgloss.Center, label.Render(name), frame)
	}

	// Legend for pie slices folded into "Other" (rendered after pie.View()).
	pieFrame := a.pie.View().Content
	legend := ""
	if combined := a.pie.Combined(); len(combined) > 0 {
		labels := make([]string, len(combined))
		for i, s := range combined {
			labels[i] = s.Label
		}
		legend = "Other: " + strings.Join(labels, ", ")
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, pieFrame,
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(legend)),
		"  ",
		a.sankey.View().Content,
	)

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
		row("cpu", a.cpu.View().Content),
		row("mem", a.mem.View().Content),
		body,
		row("disk", a.disk.View().Content),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("any key quits"),
	))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(newDemo()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
