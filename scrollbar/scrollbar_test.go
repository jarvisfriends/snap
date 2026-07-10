package scrollbar

import (
	"strings"
	"testing"
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
