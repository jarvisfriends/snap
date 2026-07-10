package uifx

import (
	"testing"

	"charm.land/lipgloss/v2"
)

// TestZonesHitAndBounds pins the zone contract: hits resolve by position and
// measured content size, higher z wins on overlap, unnamed layers and misses
// return "", and a nil Zones is safe.
func TestZonesHitAndBounds(t *testing.T) {
	t.Parallel()

	z := NewZones(
		lipgloss.NewLayer("ab\ncd").ID("left"),
		lipgloss.NewLayer("bb").ID("right").X(4).Y(1),
		lipgloss.NewLayer("!").ID("badge").X(4).Y(1).Z(1), // overlaps right
		lipgloss.NewLayer("=-"),                           // no ID: never matches
	)

	if got := z.Hit(1, 1); got != "left" {
		t.Fatalf("Hit(1,1) = %q; want left", got)
	}
	if got := z.Hit(5, 1); got != "right" {
		t.Fatalf("Hit(5,1) = %q; want right", got)
	}
	if got := z.Hit(4, 1); got != "badge" {
		t.Fatalf("Hit(4,1) = %q; want badge (higher z wins)", got)
	}
	if got := z.Hit(40, 40); got != "" {
		t.Fatalf("Hit outside = %q; want empty", got)
	}

	r, ok := z.Bounds("right")
	if !ok || r.X != 4 || r.Y != 1 || r.W != 2 || r.H != 1 {
		t.Fatalf("Bounds(right) = %+v %v; want {4 1 2 1} true", r, ok)
	}
	if _, ok := z.Bounds("ghost"); ok {
		t.Fatal("Bounds of unknown zone must miss")
	}

	var nilz *Zones
	if nilz.Hit(0, 0) != "" {
		t.Fatal("nil Zones must miss, not panic")
	}
}
