// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package charts

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestModelInitsReturnNil pins the shared lifecycle contract: every chart
// model is message-driven, so Init never schedules work.
func TestModelInitsReturnNil(t *testing.T) {
	t.Parallel()

	inits := map[string]tea.Cmd{
		"sparkline": NewSparkline("s").Init(),
		"hbar":      NewHBar("h").Init(),
		"linechart": NewLineChart("l").Init(),
		"pie":       NewPie("p").Init(),
		"sankey":    NewSankey("k").Init(),
	}
	for name, cmd := range inits {
		if cmd != nil {
			t.Errorf("%s: Init returned a non-nil command", name)
		}
	}
}

// TestHBarPctAccessor: Pct reflects the last routed data message.
func TestHBarPctAccessor(t *testing.T) {
	t.Parallel()

	h := NewHBar("h")
	if h.Pct() != 0 {
		t.Fatalf("initial Pct = %v; want 0", h.Pct())
	}
	_, _ = h.Update(HBarDataMsg{ID: "h", Pct: 62.5})
	if h.Pct() != 62.5 {
		t.Fatalf("Pct = %v; want 62.5", h.Pct())
	}
}

// TestCellCanvasSizeAndClamp: dimensions are reported by Size and clamped to
// at least 1×1 so degenerate frames still render.
func TestCellCanvasSizeAndClamp(t *testing.T) {
	t.Parallel()

	c := NewCellCanvas(3, 2, nil, nil)
	if w, h := c.Size(); w != 3 || h != 2 {
		t.Fatalf("Size() = %dx%d; want 3x2", w, h)
	}
	c = NewCellCanvas(0, -1, nil, nil)
	if w, h := c.Size(); w != 1 || h != 1 {
		t.Fatalf("clamped Size() = %dx%d; want 1x1", w, h)
	}
}

// TestCellCanvasSetFGBounds: SetFG keeps the cell background and ignores
// out-of-range coordinates like Set does.
func TestCellCanvasSetFGBounds(t *testing.T) {
	t.Parallel()

	c := NewCellCanvas(2, 2, nil, nil)
	c.SetFG(1, 1, '@', nil)
	if got := c.Rune(1, 1); got != '@' {
		t.Fatalf("Rune(1,1) = %q; want '@'", got)
	}
	for _, p := range [][2]int{{-1, 0}, {2, 0}, {0, -1}, {0, 2}} {
		c.SetFG(p[0], p[1], 'X', nil)
	}
	for y := range 2 {
		for x := range 2 {
			if c.Rune(x, y) == 'X' {
				t.Fatalf("out-of-range SetFG painted cell (%d,%d)", x, y)
			}
		}
	}
}

// TestRGB8NilIsBlack: a nil color renders as black.
func TestRGB8NilIsBlack(t *testing.T) {
	t.Parallel()

	if r, g, b := rgb8(nil); r != 0 || g != 0 || b != 0 {
		t.Fatalf("rgb8(nil) = %d,%d,%d; want black", r, g, b)
	}
}

// TestCapOr: a positive limit wins, otherwise the default applies.
func TestCapOr(t *testing.T) {
	t.Parallel()

	if got := capOr(7, 20); got != 7 {
		t.Fatalf("capOr(7, 20) = %d; want 7", got)
	}
	if got := capOr(0, 20); got != 20 {
		t.Fatalf("capOr(0, 20) = %d; want 20", got)
	}
}
