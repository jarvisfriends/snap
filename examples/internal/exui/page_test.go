// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package exui

import "testing"

// TestConfirmSemanticsFollowHost is the crux of the two hosting modes: a
// one-shot prompt exits on confirm (keeping `date=$(snap_input datepicker)`
// a single action), while the tour records the value and keeps running.
func TestConfirmSemanticsFollowHost(t *testing.T) {
	prev := InTour()
	t.Cleanup(func() { SetTourMode(prev) })

	SetTourMode(false)
	if Confirm() == nil {
		t.Fatal("single-page confirm returned no command; want tea.Quit")
	}
	SetTourMode(true)
	if cmd := Confirm(); cmd != nil {
		t.Fatal("tour confirm returned a command; want nil so the page stays up")
	}
}

func TestNamespacePrefixesEveryKey(t *testing.T) {
	t.Parallel()

	got := Namespace("datepicker", []Field{F("date", "2026-07-12"), F("tz", "UTC")})
	want := []Field{{Key: "datepicker.date", Value: "2026-07-12"}, {Key: "datepicker.tz", Value: "UTC"}}
	if len(got) != len(want) {
		t.Fatalf("Namespace returned %v; want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Namespace[%d] = %v; want %v", i, got[i], want[i])
		}
	}
}

// TestNamespaceLeavesInputAlone: the caller's slice is still needed by the
// page it came from, so namespacing must copy rather than rewrite in place.
func TestNamespaceLeavesInputAlone(t *testing.T) {
	t.Parallel()

	in := []Field{F("date", "2026-07-12")}
	_ = Namespace("datepicker", in)
	if in[0].Key != "date" {
		t.Fatalf("Namespace mutated its input: key = %q; want %q", in[0].Key, "date")
	}
}

// TestCycleThemeAdvancesAndWraps: theme cycling walks the built-in tints and
// returns to the start, and the shared palette follows it.
func TestCycleThemeAdvancesAndWraps(t *testing.T) {
	ring := cycleTints()
	if len(ring) < 2 {
		t.Skip("need at least two tints to cycle")
	}
	if ring[0] != defaultTint {
		t.Fatalf("cycle ring starts on %q; want the demo default %q", ring[0], defaultTint)
	}

	themeMu.Lock()
	prevID, prevTheme := activeTint, sharedTheme
	themeMu.Unlock()
	t.Cleanup(func() {
		themeMu.Lock()
		activeTint, sharedTheme = prevID, prevTheme
		themeMu.Unlock()
	})

	start := TintID()
	if got := CycleTheme(); got == start {
		t.Fatalf("CycleTheme stayed on %q", got)
	}
	if TintID() == start {
		t.Fatal("TintID did not follow CycleTheme")
	}
	if Theme() == nil {
		t.Fatal("Theme is nil after cycling")
	}

	// Walking the whole ring returns to where it started.
	for range len(ring) - 1 {
		CycleTheme()
	}
	if got := TintID(); got != start {
		t.Fatalf("cycling every tint landed on %q; want %q", got, start)
	}
}

// TestCapturingGuardsPlainChords: a nil predicate never captures, and a page
// that reports capturing keeps the shell off plain-letter chords like "q".
func TestCapturingGuardsPlainChords(t *testing.T) {
	t.Parallel()

	var c *Chrome
	if c.Capturing() || c.OverlayOpen() {
		t.Fatal("a nil Chrome must report neither capturing nor an open overlay")
	}

	c = &Chrome{}
	if c.Capturing() {
		t.Fatal("a Chrome with no capture predicate must not report capturing")
	}
	capturing := true
	c.SetCapture(func() bool { return capturing })
	if !c.Capturing() {
		t.Fatal("SetCapture predicate was not consulted")
	}
	capturing = false
	if c.Capturing() {
		t.Fatal("capture predicate is read once instead of per call")
	}
}
