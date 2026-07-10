package main

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/datepicker"
)

// TestDemoFlowSelectsAndQuits drives the demo app the way the VHS tape does:
// the highlighted day is visible, Enter selects, and the app quits.
// asDemo asserts the returned model is still the demoApp wrapper.
func asDemo(t *testing.T, m tea.Model) demoApp {
	t.Helper()
	a, ok := m.(demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want demoApp", m)
	}
	return a
}

func TestDemoFlowSelectsAndQuits(t *testing.T) {
	t.Parallel()

	a := demoApp{dp: datepicker.New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))}
	view := a.View().Content
	if !strings.Contains(view, "09") {
		t.Fatalf("calendar missing the initial day:\n%s", view)
	}

	m, _ := a.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	a = asDemo(t, m)
	m, cmd := a.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	a = asDemo(t, m)
	if !a.dp.Selected || a.dp.Time.Day() != 10 {
		t.Fatalf(
			"enter did not select the expected day (selected=%v day=%d)",
			a.dp.Selected, a.dp.Time.Day(),
		)
	}
	if cmd == nil {
		t.Fatal("enter did not produce a command (expected tea.Quit)")
	}
}

// TestDemoEnablesMouse: the demo's root view must request mouse reporting —
// without it the terminal never sends mouse events and OnMouse is dead code.
func TestDemoEnablesMouse(t *testing.T) {
	t.Parallel()

	a := demoApp{dp: datepicker.New(time.Now())}
	v := a.View()
	if v.MouseMode != tea.MouseModeCellMotion {
		t.Fatalf("demo root view MouseMode = %v; want cell motion", v.MouseMode)
	}
	if v.OnMouse == nil {
		t.Fatal("demo root view lost the component's OnMouse handler")
	}
}

// TestDemoDoesNotDoubleProcessMouse is the regression test for the
// double-delivery bug: Bubble Tea sends every mouse event to BOTH the root
// view's OnMouse and Update, so the demo's Update must ignore mouse messages
// (OnMouse owns them). Before this, one click highlighted AND selected.
func TestDemoDoesNotDoubleProcessMouse(t *testing.T) {
	t.Parallel()

	a := demoApp{dp: datepicker.New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))}
	_ = a.View()
	before := a.dp.Time
	m, cmd := a.Update(tea.MouseClickMsg{X: 5, Y: 5, Button: tea.MouseLeft})
	a = asDemo(t, m)
	if !a.dp.Time.Equal(before) || a.dp.Selected || cmd != nil {
		t.Fatalf("demo Update processed a mouse event (time %v->%v selected=%v)",
			before, a.dp.Time, a.dp.Selected)
	}
}
