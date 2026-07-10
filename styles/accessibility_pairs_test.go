package styles

import (
	"image/color"
	"testing"

	tint "github.com/lrstanley/bubbletint/v2"
)

// TestColorPairsFromTintBothBackgrounds pins colorPairsFromTint's output
// shape before the two palette blocks were collapsed into one helper: a tint
// with both Bg and SelectionBg yields the 16 base pairs followed by the same
// 16 with the "Select " prefix and the selection background.
func TestColorPairsFromTintBothBackgrounds(t *testing.T) {
	t.Parallel()

	bg := &tint.Color{R: 10, G: 10, B: 10, A: 255}
	selBg := &tint.Color{R: 40, G: 40, B: 80, A: 255}
	red := &tint.Color{R: 200, G: 30, B: 30, A: 255}
	tt := &tint.Tint{Bg: bg, SelectionBg: selBg, Red: red}

	pairs := colorPairsFromTint(tt, false)
	if len(pairs) != 32 {
		t.Fatalf("pairs = %d; want 32 (16 base + 16 selection)", len(pairs))
	}
	if pairs[0].Name != "Black" || pairs[16].Name != "Select Black" {
		t.Fatalf("order changed: [0]=%q [16]=%q", pairs[0].Name, pairs[16].Name)
	}
	if pairs[1].Name != "Red" || pairs[1].Fg != color.Color(red) || pairs[1].Bg != color.Color(bg) {
		t.Fatalf("base Red pair = %+v; want tint red on Bg", pairs[1])
	}
	if pairs[17].Name != "Select Red" || pairs[17].Fg != color.Color(red) || pairs[17].Bg != color.Color(selBg) {
		t.Fatalf("selection Red pair = %+v; want tint red on SelectionBg", pairs[17])
	}

	// Missing tint slots fall back to the numbered terminal color.
	if pairs[2].Name != "Green" || pairs[2].Fg == nil {
		t.Fatalf("Green fallback missing: %+v", pairs[2])
	}

	// Only Bg set: just the 16 base pairs.
	if got := colorPairsFromTint(&tint.Tint{Bg: bg}, false); len(got) != 16 {
		t.Fatalf("Bg-only pairs = %d; want 16", len(got))
	}
	// Neither set: no pairs.
	if got := colorPairsFromTint(&tint.Tint{}, false); len(got) != 0 {
		t.Fatalf("empty tint pairs = %d; want 0", len(got))
	}
}
