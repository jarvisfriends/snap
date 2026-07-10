package scrollbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// column renders with plain styles so glyphs are countable.
func column(total, visible, offset, barHeight int) []string {
	s := Vertical(total, visible, offset, barHeight, Styles{TrackRune: ".", ThumbRune: "#"})
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func TestVerticalHiddenWhenContentFits(t *testing.T) {
	t.Parallel()

	if got := Vertical(10, 10, 0, 10, DefaultStyles()); got != "" {
		t.Fatalf("scrollbar rendered for content that fits: %q", got)
	}
	if got := Vertical(5, 10, 0, 10, DefaultStyles()); got != "" {
		t.Fatalf("scrollbar rendered for short content: %q", got)
	}
}

func TestVerticalThumbTracksOffset(t *testing.T) {
	t.Parallel()

	// 100 lines, 10 visible, bar of 10: thumb is 1 cell.
	top := column(100, 10, 0, 10)
	if len(top) != 10 {
		t.Fatalf("bar height = %d; want 10", len(top))
	}
	if top[0] != "#" {
		t.Fatalf("thumb at offset 0 not at the top: %v", top)
	}
	bottom := column(100, 10, 90, 10)
	if bottom[len(bottom)-1] != "#" {
		t.Fatalf("thumb at max offset not at the bottom: %v", bottom)
	}
	mid := column(100, 10, 45, 10)
	if mid[0] == "#" || mid[len(mid)-1] == "#" {
		t.Fatalf("thumb at half offset pinned to an edge: %v", mid)
	}
	if !strings.Contains(strings.Join(mid[3:7], ""), "#") {
		t.Fatalf("thumb at half offset not near the middle: %v", mid)
	}

	// Out-of-range offsets clamp rather than panic or vanish.
	over := column(100, 10, 9999, 10)
	if over[len(over)-1] != "#" {
		t.Fatalf("overscrolled thumb not clamped to the bottom: %v", over)
	}
}

func TestVerticalThumbSizeProportional(t *testing.T) {
	t.Parallel()

	// Half the content visible: thumb fills half the bar.
	bar := column(20, 10, 0, 10)
	thumbs := 0
	for _, c := range bar {
		if c == "#" {
			thumbs++
		}
	}
	if thumbs != 5 {
		t.Fatalf("thumb size = %d cells of 10; want 5 (visible/total = 1/2)", thumbs)
	}
}

func TestClampOffset(t *testing.T) {
	t.Parallel()

	if got := ClampOffset(-5, 100, 10); got != 0 {
		t.Fatalf("ClampOffset(-5) = %d; want 0", got)
	}
	if got := ClampOffset(500, 100, 10); got != 90 {
		t.Fatalf("ClampOffset(500) = %d; want 90", got)
	}
	if got := ClampOffset(3, 5, 10); got != 0 {
		t.Fatalf("ClampOffset with fitting content = %d; want 0", got)
	}
}

// TestPresetGlyphDefaults: each preset picks its signature glyphs when no
// overrides are given.
func TestPresetGlyphDefaults(t *testing.T) {
	t.Parallel()

	line := ansi.Strip(Vertical(100, 10, 0, 10, Styles{}))
	if !strings.Contains(line, "┃") || !strings.Contains(line, "│") {
		t.Fatalf("line preset missing thin/heavy glyphs: %q", line)
	}
	classic := ansi.Strip(Vertical(100, 10, 0, 10, Styles{Preset: PresetClassic}))
	if !strings.Contains(classic, "█") || !strings.Contains(classic, "░") {
		t.Fatalf("classic preset missing retro glyphs: %q", classic)
	}
}

// TestSmoothPresetSubCellGlide pins the eighth-block behavior: a mid-way
// offset produces partial-block boundary cells, offsets one line apart
// produce distinct frames (sub-cell motion), and the ends are exact.
func TestSmoothPresetSubCellGlide(t *testing.T) {
	t.Parallel()

	st := Styles{Preset: PresetSmooth}
	render := func(offset int) []string {
		return strings.Split(ansi.Strip(Vertical(200, 20, offset, 10, st)), "\n")
	}

	top := render(0)
	if len(top) != 10 || top[0] != "█" {
		t.Fatalf("smooth thumb at offset 0 not flush with the top: %v", top)
	}
	bottom := render(180)
	if bottom[len(bottom)-1] != "█" {
		t.Fatalf("smooth thumb at max offset not flush with the bottom: %v", bottom)
	}

	// A mid-way offset must land the thumb edges inside cells: at least one
	// partial (eighth-block) boundary glyph.
	partials := "▁▂▃▄▅▆▇"
	mid := strings.Join(render(37), "")
	if !strings.ContainsAny(mid, partials) {
		t.Fatalf("smooth thumb at odd offset has no sub-cell boundary: %q", mid)
	}

	// Nearby offsets render differently — the point of sub-cell resolution.
	// (Exact adjacency can share an eighth: 180 scroll positions map onto 72
	// eighth-steps here, so compare offsets ≥ one eighth-step apart. A
	// cell-based thumb would need ~20 lines of scrolling to move at all.)
	if a, b := render(40), render(45); strings.Join(a, "") == strings.Join(b, "") {
		t.Fatal("nearby offsets rendered identically; sub-cell glide lost")
	}
}
