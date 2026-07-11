package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jarvisfriends/snap/notifications"
)

func asDemo(t *testing.T, m tea.Model) *demoApp {
	t.Helper()
	a, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	return a
}

// TestSeverityKeysAddNotifications: i/w/e add one notification each; the
// count line reflects the history and the frame still renders one screen.
func TestSeverityKeysAddNotifications(t *testing.T) {
	a := newDemo()
	var m tea.Model = a
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = asDemo(t, m)

	for _, k := range []string{"i", "w", "e"} {
		m, _ = a.Update(tea.KeyPressMsg{Code: rune(k[0]), Text: k})
		a = asDemo(t, m)
	}
	if got := a.mgr.Count(); got != 3 {
		t.Fatalf("after i/w/e, notification count = %d, want 3", got)
	}
	frame := ansi.Strip(a.View().Content)
	if !strings.Contains(frame, "notifications in history: 3") {
		t.Fatalf("frame missing the history count:\n%s", frame)
	}
}

// TestProgressRunsToCompletion: p starts a progress notification and the
// tick loop fills it to 100, ending with a completion toast.
func TestProgressRunsToCompletion(t *testing.T) {
	a := newDemo()
	var m tea.Model = a
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = asDemo(t, m)

	m, _ = a.Update(tea.KeyPressMsg{Code: 'p', Text: "p"})
	a = asDemo(t, m)
	if a.dlID < 0 {
		t.Fatal("p should start the download notification")
	}
	for range 20 {
		m, _ = a.Update(progressTickMsg{})
		a = asDemo(t, m)
		if a.dlID < 0 {
			break
		}
	}
	if a.dlID >= 0 {
		t.Fatal("progress never completed")
	}
	var done bool
	for _, n := range a.mgr.All() {
		if n.Content == "download complete" && n.Severity == notifications.SeverityInfo {
			done = true
		}
	}
	if !done {
		t.Fatal("completion notification missing from history")
	}
}

// TestHistoryToggle: ctrl+n opens the history panel overlay and a second
// ctrl+n closes it.
func TestHistoryToggle(t *testing.T) {
	a := newDemo()
	var m tea.Model = a
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	a = asDemo(t, m)
	m, _ = a.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	a = asDemo(t, m)

	m, _ = a.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	a = asDemo(t, m)
	if !a.bar.IsHistoryVisible() {
		t.Fatal("ctrl+n should open the history panel")
	}
	m, _ = a.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	a = asDemo(t, m)
	if a.bar.IsHistoryVisible() {
		t.Fatal("second ctrl+n should close the history panel")
	}
}
