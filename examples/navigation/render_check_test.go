package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func asDemo(t *testing.T, m tea.Model) *demoApp {
	t.Helper()
	a, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	return a
}

// TestStyleSwapKeepsActivePage: n cycles through all three navigator styles,
// the active index carries over, and every style renders the page titles.
func TestStyleSwapKeepsActivePage(t *testing.T) {
	t.Parallel()

	a := newDemo()
	var m tea.Model = a
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = asDemo(t, m)

	// Move to the second page on the sidebar.
	m, _ = a.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	a = asDemo(t, m)
	if got := a.nav().GetActiveIndex(); got != 1 {
		t.Fatalf("down should move to page index 1, got %d", got)
	}

	for cycle := range 3 {
		frame := ansi.Strip(a.View().Content)
		for _, title := range []string{"Home", "Metrics", "Logs"} {
			if !strings.Contains(frame, title) {
				t.Fatalf("style %d (%s): frame missing page %q:\n%s",
					cycle, styleNames[a.style], title, frame)
			}
		}
		m, _ = a.Update(tea.KeyPressMsg{Code: 'n', Text: "n"})
		a = asDemo(t, m)
		if got := a.nav().GetActiveIndex(); got != 1 {
			t.Fatalf("style swap %d lost the active index: got %d want 1", cycle, got)
		}
	}
	if a.style != 0 {
		t.Fatalf("three swaps should land back on the first style, got %d", a.style)
	}
}

// TestEnterPicksActivePage: Enter records the active page's ID (the value
// main writes to stdout) and quits.
func TestEnterPicksActivePage(t *testing.T) {
	t.Parallel()

	a := newDemo()
	var m tea.Model = a
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = asDemo(t, m)
	m, _ = a.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	a = asDemo(t, m)
	m, cmd := a.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	a = asDemo(t, m)
	if a.picked != "metrics" {
		t.Fatalf("picked = %q, want %q", a.picked, "metrics")
	}
	if cmd == nil {
		t.Fatal("enter should quit (nil cmd)")
	}
}
