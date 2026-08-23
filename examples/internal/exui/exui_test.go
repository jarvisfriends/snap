// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package exui

import (
	"strings"
	"testing"
)

// TestTourExpandedHelpGrowsTourColumn: under the tour, ctrl+h's expanded help
// must answer "how do I move between pages" — page nav (tab/shift+tab,
// alt+arrows), the ctrl+b page strip, and a ctrl+t theme row naming the
// CURRENT tint. Single-page mode must not show tour chords that do nothing.
func TestTourExpandedHelpGrowsTourColumn(t *testing.T) {
	prev := InTour()
	t.Cleanup(func() { SetTourMode(prev) })

	render := func() string {
		c := NewChrome(Bind("enter", "confirm"))
		c.bar.ToggleFullHelpVisible()
		c.SetWidth(220) // wide enough that no column is elided
		return c.bar.View().Content
	}

	SetTourMode(true)
	got := render()
	for _, want := range []string{"shift+tab", "ctrl+b", "ctrl+t", "theme: " + TintID()} {
		if !strings.Contains(got, want) {
			t.Errorf("tour expanded help is missing %q:\n%s", want, got)
		}
	}

	SetTourMode(false)
	if got := render(); strings.Contains(got, "ctrl+t") {
		t.Errorf("single-page expanded help shows the tour's ctrl+t chord:\n%s", got)
	}
}

// TestTourShortHelpUnchanged: the tour chords live only behind ctrl+h — the
// one-line short help stays the page's own bindings either way.
func TestTourShortHelpUnchanged(t *testing.T) {
	prev := InTour()
	t.Cleanup(func() { SetTourMode(prev) })

	SetTourMode(true)
	c := NewChrome(Bind("enter", "confirm"))
	c.SetWidth(220)
	if got := c.bar.View().Content; strings.Contains(got, "ctrl+t") {
		t.Errorf("short help leaked the tour column:\n%s", got)
	}
}
