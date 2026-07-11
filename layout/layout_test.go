package layout

import (
	"testing"

	"charm.land/lipgloss/v2"
)

func borderedPadded() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
}

func TestContentOrigin(t *testing.T) {
	t.Parallel()

	x, y := ContentOrigin(borderedPadded())
	if x != 3 || y != 2 { // 1 border + 2 padding, 1 border + 1 padding
		t.Fatalf("ContentOrigin = (%d, %d), want (3, 2)", x, y)
	}
	x, y = ContentOrigin(lipgloss.NewStyle())
	if x != 0 || y != 0 {
		t.Fatalf("ContentOrigin(plain) = (%d, %d), want (0, 0)", x, y)
	}
}

func TestInnerSize(t *testing.T) {
	t.Parallel()

	// Frame: 2 border + 4 padding wide, 2 border + 2 padding tall.
	w, h := InnerSize(borderedPadded(), 40, 10)
	if w != 34 || h != 6 {
		t.Fatalf("InnerSize = (%d, %d), want (34, 6)", w, h)
	}
	// Outer smaller than the frame floors at 1x1 rather than going negative.
	w, h = InnerSize(borderedPadded(), 3, 2)
	if w != 1 || h != 1 {
		t.Fatalf("InnerSize tiny = (%d, %d), want (1, 1)", w, h)
	}
}

func TestRenderInBox(t *testing.T) {
	t.Parallel()

	st := borderedPadded()
	out := RenderInBox(st, 20, 7, "hi")
	if got := lipgloss.Width(out); got != 20 {
		t.Errorf("RenderInBox width = %d, want 20", got)
	}
	if got := lipgloss.Height(out); got != 7 {
		t.Errorf("RenderInBox height = %d, want 7", got)
	}

	// Non-positive dimensions derive from content + frame: "hi" is 2x1,
	// frame adds 6 wide and 4 tall.
	out = RenderInBox(st, 0, 0, "hi")
	if got := lipgloss.Width(out); got != 8 {
		t.Errorf("RenderInBox auto width = %d, want 8", got)
	}
	if got := lipgloss.Height(out); got != 5 {
		t.Errorf("RenderInBox auto height = %d, want 5", got)
	}
}
