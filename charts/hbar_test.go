package charts

import (
	"strings"
	"testing"
)

func TestHBar_Zero(t *testing.T) {
	got := HBar(0, 10)
	if !strings.HasPrefix(got, "░") {
		t.Errorf("HBar(0, 10) = %q, expected all empty", got)
	}
	if len([]rune(got)) != 10 {
		t.Errorf("HBar(0, 10) len = %d, want 10", len([]rune(got)))
	}
}

func TestHBar_Full(t *testing.T) {
	got := HBar(100, 8)
	if !strings.HasPrefix(got, "█") {
		t.Errorf("HBar(100, 8) = %q, expected all filled", got)
	}
	if strings.Contains(got, "░") {
		t.Errorf("HBar(100, 8) = %q, should have no empty cells", got)
	}
}

func TestHBar_Half(t *testing.T) {
	got := HBar(50, 10)
	runes := []rune(got)
	if len(runes) != 10 {
		t.Fatalf("HBar(50, 10) len = %d, want 10", len(runes))
	}
	filled := strings.Count(got, "█")
	// Rounding: 50% of 10 = 5 ± 1
	if filled < 4 || filled > 6 {
		t.Errorf("HBar(50, 10) filled = %d, want ~5", filled)
	}
}
