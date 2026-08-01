// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package charts implements the `snap_input charts` subcommand: the canonical multi-chart wiring example: several chart
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
package charts

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/charts"
	"github.com/jarvisfriends/snap/examples/internal/exui"
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

	chrome *exui.Chrome
	height int
	third  int // pie column width (left third of the window)
	bodyH  int // rows available to the pie/sankey body
	t      float64
}

func newDemo() demoApp {
	t := exui.Theme()
	pie := charts.NewPie("share")
	pie.Braille = true
	// Dual-color gradients from the shared theme: load-style metrics ramp
	// success→error as they climb; memory ramps accent→warning.
	cpu := charts.NewSparkline("cpu")
	cpu.Opts.GradientFrom, cpu.Opts.GradientTo = t.Success, t.Error
	mem := charts.NewSparkline("mem")
	mem.Opts.GradientFrom, mem.Opts.GradientTo = t.Accent, t.Warning
	disk := charts.NewHBar("disk")
	disk.GradientFrom, disk.GradientTo = t.Success, t.Error
	return demoApp{
		cpu:    cpu,
		mem:    mem,
		pie:    pie,
		sankey: charts.NewSankey("traffic"),
		disk:   disk,
		chrome: exui.NewChrome(exui.Bind("any key", "quit")),
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
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return a, tea.Quit

	case tea.WindowSizeMsg:
		// Split the window: two sparkline rows up top, pie beside sankey
		// below, one bar across the bottom. Each chart stretches to the
		// space it is given; the pie/sankey body takes every row the fixed
		// lines (2 sparklines, 1 legend, 1 hbar, the help bar) don't.
		a.height = msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
		// Pie owns the left third (centered vertically); sankey stretches
		// over the remaining two-thirds at full body height.
		a.third = max(msg.Width/3, 12)
		a.bodyH = max(msg.Height-4-a.chrome.Height(), 8)
		a.cpu.SetSize(msg.Width-8, 1)
		a.mem.SetSize(msg.Width-8, 1)
		a.pie.SetSize(a.third-2, a.bodyH-1) // -1: the always-present legend line
		a.sankey.SetSize(msg.Width-a.third-2, a.bodyH)
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

	// Legend for pie slices folded into "Other". The line is ALWAYS there
	// (an em dash when nothing folded) at a fixed width, so the pie column's
	// height and width never change as thin slices fold or unfold — which
	// keeps the vertically-centered pie from jumping and the sankey from
	// shifting sideways.
	pieFrame := a.pie.View().Content
	legend := "Other: —"
	if combined := a.pie.Combined(); len(combined) > 0 {
		labels := make([]string, len(combined))
		for i, s := range combined {
			labels[i] = s.Label
		}
		legend = "Other: " + strings.Join(labels, ", ")
	}
	pieW := max(a.third-2, 1)
	legendLine := exui.Dim().
		Width(pieW).MaxWidth(pieW).Render(legend)

	// Left third: pie + legend centered vertically. Right two-thirds: the
	// sankey at full body height.
	pieCol := lipgloss.Place(a.third, max(a.bodyH, 1), lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Left, pieFrame, legendLine))
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		pieCol,
		"  ",
		a.sankey.View().Content,
	)

	v := tea.NewView(lipgloss.JoinVertical(
		lipgloss.Left,
		row("cpu", a.cpu.View().Content),
		row("mem", a.mem.View().Content),
		body,
		row("disk", a.disk.View().Content),
	))
	a.chrome.Frame(&v, a.height)
	return v
}

// Run is the snap_input subcommand entry point.
func Run() {
	exui.Init()
	if _, err := exui.Program(newDemo()).Run(); err != nil {
		exui.Fatal(err)
	}
}
