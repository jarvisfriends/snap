package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// pump advances the demo through a resize and a few animation ticks.
func pump(t *testing.T, w, h, ticks int) *demoApp {
	t.Helper()
	var m tea.Model = &demoApp{colors: palette()}
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

// TestCellCanvasDemoFitsWindow: the plasma frame fills exactly the window
// (header + canvas rows, no overflow) at a large and a small size.
func TestCellCanvasDemoFitsWindow(t *testing.T) {
	t.Parallel()

	for _, size := range []struct{ w, h int }{{90, 28}, {40, 10}} {
		a := pump(t, size.w, size.h, 5)
		frame := ansi.Strip(a.View().Content)
		lines := strings.Split(frame, "\n")
		if len(lines) != size.h {
			t.Fatalf("%dx%d: frame is %d lines tall; want %d", size.w, size.h, len(lines), size.h)
		}
		for i, line := range lines[1:] { // row 0 is the header
			if lw := lipgloss.Width(line); lw != size.w {
				t.Fatalf("%dx%d: canvas line %d is %d cells wide; want %d",
					size.w, size.h, i+1, lw, size.w)
			}
		}
	}
}

// TestCellCanvasDemoAnimates: successive ticks actually change the frame —
// the point of the plasma (a static canvas means the tick wiring broke).
func TestCellCanvasDemoAnimates(t *testing.T) {
	t.Parallel()

	a := pump(t, 60, 16, 1)
	before := a.View().Content
	m, _ := a.Update(tickMsg{})
	next, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	if before == next.View().Content {
		t.Fatal("tick did not change the rendered plasma frame")
	}
}
