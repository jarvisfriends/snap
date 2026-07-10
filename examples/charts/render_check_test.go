package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// pump advances the demo through a resize and a few ticks.
func pump(t *testing.T, w, h, ticks int) demoApp {
	t.Helper()
	var m tea.Model = newDemo()
	m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	for range ticks {
		m, _ = m.Update(tickMsg{})
	}
	a, ok := m.(demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want demoApp", m)
	}
	return a
}

// TestChartsDemoFitsWindow: after a resize every chart stretches into (and
// stays within) the split it was given, at both a large and a small size.
func TestChartsDemoFitsWindow(t *testing.T) {
	t.Parallel()

	for _, size := range []struct{ w, h int }{{100, 32}, {60, 20}} {
		a := pump(t, size.w, size.h, 12)
		frame := ansi.Strip(a.View().Content)
		for i, line := range strings.Split(frame, "\n") {
			if lw := lipgloss.Width(line); lw > size.w {
				t.Fatalf("%dx%d: line %d is %d cells wide: %q", size.w, size.h, i, lw, line)
			}
		}
		// The stream actually landed: cpu sparkline consumed its points.
		if len(a.cpu.History()) == 0 {
			t.Fatal("cpu sparkline never received routed data")
		}
		if uw, _ := a.cpu.Used(); uw != size.w-8 {
			t.Fatalf("cpu sparkline used %d cells; want the %d it was given", uw, size.w-8)
		}
	}
}

// TestChartsDemoRoutesPerID: same-type charts consume only their own
// stream — the core of the multi-chart pattern this example demonstrates.
func TestChartsDemoRoutesPerID(t *testing.T) {
	t.Parallel()

	a := pump(t, 100, 32, 3)
	if len(a.cpu.History()) != 3 || len(a.mem.History()) != 3 {
		t.Fatalf("history lengths cpu=%d mem=%d; want 3 each",
			len(a.cpu.History()), len(a.mem.History()))
	}
	if a.cpu.History()[0] == a.mem.History()[0] {
		t.Fatal("cpu and mem received identical values — IDs not routing")
	}
}
