package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// typeText feeds text into the demo one key at a time.
func typeText(t *testing.T, m tea.Model, text string) tea.Model {
	t.Helper()
	for _, r := range text {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

// frameOf renders the demo's current frame with styling stripped.
func frameOf(t *testing.T, m tea.Model) string {
	t.Helper()
	a, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	return ansi.Strip(a.View().Content)
}

// TestFormsDemoValidatesLive drives the form through the same flow as the
// tape: errors show for empty/partial input and flip to parsed values.
func TestFormsDemoValidatesLive(t *testing.T) {
	t.Parallel()

	var m tea.Model = newDemo()
	frame := frameOf(t, m)
	if !strings.Contains(frame, "task is required") {
		t.Fatalf("empty form does not show the required error:\n%s", frame)
	}

	m = typeText(t, m, "ship it")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // -> duration
	m = typeText(t, m, "7h30m")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // -> date
	m = typeText(t, m, "2026-07-14")
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab}) // -> tags
	m = typeText(t, m, "go,  tui ,, release")

	frame = frameOf(t, m)
	for _, want := range []string{"✓ ship it", "✓ 7h30m", "2026-07-14", "3 tags", "go,tui,release"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("frame missing %q:\n%s", want, frame)
		}
	}
	if strings.Contains(frame, "✗") {
		t.Fatalf("fully valid form still shows an error:\n%s", frame)
	}
}

// TestFormsDemoFitsWidth: rendered lines stay within a narrow terminal.
func TestFormsDemoFitsWidth(t *testing.T) {
	t.Parallel()

	var m tea.Model = newDemo()
	m = typeText(t, m, strings.Repeat("x", 60)) // overfill the first input
	frame := frameOf(t, m)
	for i, line := range strings.Split(frame, "\n") {
		if lw := lipgloss.Width(line); lw > 80 {
			t.Fatalf("line %d is %d cells wide (input should scroll, not grow): %q", i, lw, line)
		}
	}
}
