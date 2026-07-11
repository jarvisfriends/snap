package charts

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	ccRed  = lipgloss.Color("#aa0000")
	ccBlue = lipgloss.Color("#0000aa")
)

// TestCellCanvasSetAndClear: cells paint, clipping ignores out-of-range
// writes, Clear restores the defaults.
func TestCellCanvasSetAndClear(t *testing.T) {
	c := NewCellCanvas(4, 2, ccRed, ccBlue)
	c.Set(1, 0, 'x', ccBlue, ccRed)
	c.Set(-1, 0, '!', ccBlue, ccRed) // clipped
	c.Set(4, 2, '!', ccBlue, ccRed)  // clipped

	if got := ansi.Strip(c.String()); got != " x  \n    " {
		t.Fatalf("got %q", got)
	}
	if c.Rune(1, 0) != 'x' || c.Rune(0, 0) != ' ' || c.Rune(9, 9) != ' ' {
		t.Fatal("Rune accessor wrong")
	}

	c.Clear()
	if got := ansi.Strip(c.String()); got != "    \n    " {
		t.Fatalf("after Clear: %q", got)
	}
}

// TestCellCanvasSetFGKeepsBackground: SetFG swaps rune+fg and leaves the
// cell's background untouched.
func TestCellCanvasSetFGKeepsBackground(t *testing.T) {
	c := NewCellCanvas(2, 1, ccRed, ccBlue)
	c.Set(0, 0, '#', ccRed, ccRed)
	c.SetFG(0, 0, '@', ccBlue)
	cell := c.cells[0]
	if cell.ch != '@' {
		t.Fatalf("rune = %q", cell.ch)
	}
	if r, g, b := rgb8(cell.bg); r != 0xaa || g != 0 || b != 0 {
		t.Fatalf("background changed: %d %d %d", r, g, b)
	}
	if r, g, b := rgb8(cell.fg); r != 0 || g != 0 || b != 0xaa {
		t.Fatalf("foreground not updated: %d %d %d", r, g, b)
	}
}

// TestCellCanvasBatchedEscapes: a run of same-colored cells emits its colors
// once, not per cell — the whole point of the batched writer.
func TestCellCanvasBatchedEscapes(t *testing.T) {
	c := NewCellCanvas(8, 1, ccRed, ccBlue)
	out := c.String()
	if got := strings.Count(out, "\x1b[38;2;"); got != 1 {
		t.Errorf("fg escapes = %d; want 1 for a uniform row\n%q", got, out)
	}
	if got := strings.Count(out, "\x1b[48;2;"); got != 1 {
		t.Errorf("bg escapes = %d; want 1 for a uniform row\n%q", got, out)
	}
	if !strings.HasSuffix(out, "\x1b[0m") {
		t.Error("render should end with a reset")
	}

	// A color change mid-row forces exactly one more fg emit.
	c.SetFG(4, 0, 'x', ccBlue)
	out = c.String()
	if got := strings.Count(out, "\x1b[38;2;"); got != 3 {
		// red ×4, blue ×1, red ×3 → three fg runs
		t.Errorf("fg escapes after one change = %d; want 3\n%q", got, out)
	}
}

// TestGradientEndsAndSteps: the ramp starts and ends on the given colors,
// yields the requested count, and degenerate steps fall back to the start.
func TestGradientEndsAndSteps(t *testing.T) {
	g := Gradient(ccRed, ccBlue, 5)
	if len(g) != 5 {
		t.Fatalf("len = %d", len(g))
	}
	if r, _, b := rgb8(g[0]); r != 0xaa || b != 0 {
		t.Fatalf("start = %v", g[0])
	}
	if r, _, b := rgb8(g[4]); r != 0 || b != 0xaa {
		t.Fatalf("end = %v", g[4])
	}
	for i, c := range g {
		if c == nil {
			t.Fatalf("nil color at %d", i)
		}
	}

	if g := Gradient(ccRed, ccBlue, 1); len(g) != 1 {
		t.Fatalf("steps=1 should yield the start color only, got %d", len(g))
	}
	if g := Gradient(nil, nil, 3); len(g) != 3 {
		t.Fatalf("nil colors should still ramp (as black), got %d", len(g))
	}
	var _ color.Color = Gradient(ccRed, ccBlue, 2)[0]
}
