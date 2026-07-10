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
func TestDemoFlowSelectsAndQuits(t *testing.T) {
	t.Parallel()

	a := demoApp{dp: datepicker.New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))}
	view := a.View().Content
	if !strings.Contains(view, "09") {
		t.Fatalf("calendar missing the initial day:\n%s", view)
	}

	m, _ := a.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	a = m.(demoApp)
	m, cmd := a.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	a = m.(demoApp)
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
