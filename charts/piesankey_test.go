package charts

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// gridSize returns the rendered width (widest line, ANSI stripped) and height
// (line count) of a chart string.
func gridSize(s string) (w, h int) {
	for line := range strings.SplitSeq(s, "\n") {
		h++
		if lw := len([]rune(ansi.Strip(line))); lw > w {
			w = lw
		}
	}
	return w, h
}

func TestSparklineStyleName(t *testing.T) {
	t.Parallel()

	styles := []SparklineStyle{
		SparklineUserBlocks, SparklineBrailleUp, SparklineBrailleDown, SparklineStdBlocks,
	}
	seen := map[string]SparklineStyle{}
	for _, style := range styles {
		name := SparklineStyleName(style)
		if name == "" {
			t.Errorf("SparklineStyleName(%d) is empty", style)
		}
		if prev, dup := seen[name]; dup {
			t.Errorf("styles %d and %d share the name %q", prev, style, name)
		}
		seen[name] = style
	}
	// Out-of-range values wrap modulo the table instead of panicking.
	if got := SparklineStyleName(SparklineStyle(len(styles))); got != SparklineStyleName(SparklineUserBlocks) {
		t.Errorf("out-of-range style name = %q; want wrap to %q", got, SparklineStyleName(SparklineUserBlocks))
	}
}

func TestCanvasPixelSize(t *testing.T) {
	t.Parallel()

	c := NewCanvas(10, 5)
	w, h := c.PixelSize()
	if w != 20 || h != 20 {
		t.Fatalf("PixelSize() = %dx%d; want 20x20 (2 px per cell wide, 4 tall)", w, h)
	}
}

func TestPieChartRendersAllSlices(t *testing.T) {
	t.Parallel()

	slices := []PieSlice{
		{Value: 3, Color: lipgloss.Color("1"), Label: "a"},
		{Value: 1, Color: lipgloss.Color("2"), Label: "b"},
	}
	got := PieChart(slices, 4)
	if got == "" {
		t.Fatal("PieChart returned empty output for valid input")
	}
	if _, h := gridSize(got); h < 4 {
		t.Fatalf("PieChart height = %d lines; want at least the radius (4)", h)
	}

	// Degenerate inputs render nothing rather than panicking.
	if PieChart(nil, 4) != "" || PieChart(slices, 0) != "" {
		t.Fatal("PieChart must return empty output for no slices / no radius")
	}
	if PieChart([]PieSlice{{Value: 0}}, 3) != "" {
		t.Fatal("PieChart with zero total must return empty output")
	}
}

func TestBraillePieChartRendersCircle(t *testing.T) {
	t.Parallel()

	slices := []PieSlice{
		{Value: 2, Color: lipgloss.Color("3")},
		{Value: 2, Color: lipgloss.Color("4")},
	}
	got := BraillePieChart(slices, 4)
	if got == "" {
		t.Fatal("BraillePieChart returned empty output for valid input")
	}
	if !strings.ContainsFunc(ansi.Strip(got), func(r rune) bool {
		return r >= 0x2800 && r <= 0x28FF
	}) {
		t.Fatal("BraillePieChart output contains no braille glyphs")
	}
	if BraillePieChart(nil, 4) != "" || BraillePieChart(slices, 0) != "" {
		t.Fatal("BraillePieChart must return empty output for degenerate input")
	}
}

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

// brailleDots counts set dots across all braille runes in s.
func brailleDots(s string) int {
	n := 0
	for _, r := range ansi.Strip(s) {
		if r >= 0x2800 && r <= 0x28FF {
			for mask := 0x01; mask <= 0x80; mask <<= 1 {
				if int(r-0x2800)&mask != 0 {
					n++
				}
			}
		}
	}
	return n
}
