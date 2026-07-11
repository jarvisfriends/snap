package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// resize opens the demo at w x h (the first WindowSizeMsg opens the modal).
func resize(t *testing.T, w, h int) *demoApp {
	t.Helper()
	var m tea.Model = newDemo()
	m, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	a, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	return a
}

// TestDependenciesDemoShowsModal: the modal opens on the first resize,
// composites over the backdrop, and lists this test binary's dependencies.
func TestDependenciesDemoShowsModal(t *testing.T) {
	t.Parallel()

	a := resize(t, 100, 32)
	if !a.modal.IsVisible() {
		t.Fatal("modal not visible after the first WindowSizeMsg")
	}
	frame := ansi.Strip(a.View().Content)
	if !strings.Contains(frame, "Dependencies") {
		t.Fatalf("frame missing the dependencies header:\n%s", frame)
	}
	if !strings.Contains(frame, "charm.land/bubbletea") {
		t.Fatal("frame does not list the bubbletea dependency")
	}
	for i, line := range strings.Split(frame, "\n") {
		if lw := lipgloss.Width(line); lw > 100 {
			t.Fatalf("line %d is %d cells wide: %q", i, lw, line)
		}
	}
}

// TestDependenciesDemoWheelAndDismiss: the wheel scrolls via HandleMouse,
// Esc closes, i reopens — the tape's exact flow.
func TestDependenciesDemoWheelAndDismiss(t *testing.T) {
	t.Parallel()

	a := resize(t, 90, 24)
	before := ansi.Strip(a.View().Content)
	if cmd := a.onMouse(tea.MouseWheelMsg{X: 45, Y: 12, Button: tea.MouseWheelDown}); cmd != nil {
		t.Fatalf("wheel inside the modal produced a cmd: %v", cmd())
	}
	if after := ansi.Strip(a.View().Content); after == before {
		t.Fatal("wheel down did not scroll the dependency list")
	}

	m, _ := a.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	a, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	if a.modal.IsVisible() {
		t.Fatal("Esc did not close the modal")
	}
	m, _ = a.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	a, ok = m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	if !a.modal.IsVisible() {
		t.Fatal("i did not reopen the modal")
	}
}
