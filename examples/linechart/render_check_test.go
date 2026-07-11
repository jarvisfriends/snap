package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// pump advances the demo through a resize and a few ticks.
func pump(t *testing.T, w, h, ticks int) *demoApp {
	t.Helper()
	var m tea.Model = newDemo()
	m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	for range ticks {
		m, _ = m.Update(tickMsg{})
	}
	a, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	return a
}

// TestLinechartDemoFitsWindow: the streamed chart stretches into (and stays
// within) the window at a large and a small size, and the rolling window
// actually received points.
func TestLinechartDemoFitsWindow(t *testing.T) {
	t.Parallel()

	for _, size := range []struct{ w, h int }{{100, 30}, {50, 12}} {
		a := pump(t, size.w, size.h, 20)
		frame := ansi.Strip(a.View().Content)
		lines := strings.Split(frame, "\n")
		if len(lines) > size.h {
			t.Fatalf("%dx%d: frame is %d lines tall", size.w, size.h, len(lines))
		}
		for i, line := range lines {
			if lw := lipgloss.Width(line); lw > size.w {
				t.Fatalf("%dx%d: line %d is %d cells wide: %q", size.w, size.h, i, lw, line)
			}
		}
		if len(a.sine) != 20 || len(a.echo) != 20 {
			t.Fatalf("series lengths sine=%d echo=%d; want 20 each", len(a.sine), len(a.echo))
		}
	}
}

// TestLinechartDemoRollsWindow: the series cap at the rolling window size.
func TestLinechartDemoRollsWindow(t *testing.T) {
	t.Parallel()

	a := pump(t, 80, 24, window+25)
	if len(a.sine) != window || len(a.echo) != window {
		t.Fatalf("series lengths sine=%d echo=%d; want %d (rolling window)",
			len(a.sine), len(a.echo), window)
	}
}
