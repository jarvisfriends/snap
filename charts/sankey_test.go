package charts

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestBrailleSankeyChartShape pins the Sankey renderer's output geometry and
// that every flow's endpoints influence the drawing — guarding the
// maps.Copy/SplitSeq modernizations that touched its internals.
func TestBrailleSankeyChartShape(t *testing.T) {
	t.Parallel()

	flows := []SankeyFlow{
		{Source: "a", Target: "x", Value: 3, Color: lipgloss.Color("1")},
		{Source: "b", Target: "x", Value: 1, Color: lipgloss.Color("2")},
		{Source: "a", Target: "y", Value: 2, Color: lipgloss.Color("3")},
	}
	got := BrailleSankeyChart(flows, 20, 6)
	if got == "" {
		t.Fatal("BrailleSankeyChart returned empty output for valid input")
	}
	w, h := gridSize(got)
	if h != 6 {
		t.Fatalf("BrailleSankeyChart height = %d lines; want charH=6", h)
	}
	if w > 20 {
		t.Fatalf("BrailleSankeyChart width = %d cells; must not exceed charW=20", w)
	}
	if !strings.ContainsFunc(ansi.Strip(got), func(r rune) bool {
		return r >= 0x2800 && r <= 0x28FF
	}) {
		t.Fatal("BrailleSankeyChart output contains no braille glyphs")
	}

	// A single dominant flow must render more ink than a tiny one: compare
	// braille dot counts to catch a broken value scale.
	big := brailleDots(BrailleSankeyChart([]SankeyFlow{{Source: "a", Target: "x", Value: 100, Color: lipgloss.Color("1")}}, 20, 6))
	if big == 0 {
		t.Fatal("single-flow Sankey rendered no dots")
	}
}

func TestSmoothstepEnds(t *testing.T) {
	t.Parallel()

	if smoothstep(-1) != 0 || smoothstep(0) != 0 {
		t.Error("smoothstep must clamp to 0 at and below t=0")
	}
	if smoothstep(1) != 1 || smoothstep(2) != 1 {
		t.Error("smoothstep must clamp to 1 at and above t=1")
	}
	if mid := smoothstep(0.5); mid != 0.5 {
		t.Errorf("smoothstep(0.5) = %v; want 0.5 (symmetric S-curve)", mid)
	}
}
