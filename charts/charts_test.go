package charts

import (
	"strings"
	"testing"
)

func TestSparkline_Empty(t *testing.T) {
	got := Sparkline(nil, 10, SparklineOpts{})
	if len([]rune(got)) != 10 {
		t.Errorf("Sparkline(nil, 10) len = %d, want 10", len([]rune(got)))
	}
}

func TestSparkline_Uniform(t *testing.T) {
	history := []float64{5, 5, 5, 5, 5}
	got := Sparkline(history, 5, SparklineOpts{})
	if len([]rune(got)) != 5 {
		t.Errorf("Sparkline uniform len = %d, want 5", len([]rune(got)))
	}
	// All values equal → all blocks should be the same character.
	runes := []rune(got)
	for _, r := range runes[1:] {
		if r != runes[0] {
			t.Errorf("Sparkline uniform: expected identical chars, got %q", got)
			break
		}
	}
}

func TestSparkline_Ascending(t *testing.T) {
	history := []float64{0, 25, 50, 75, 100}
	got := Sparkline(history, 5, SparklineOpts{}) // default block style
	runes := []rune(got)
	if len(runes) != 5 {
		t.Fatalf("len = %d, want 5", len(runes))
	}
	// Runes should be non-decreasing (ascending values → ascending block height).
	for i := 1; i < len(runes); i++ {
		if runes[i] < runes[i-1] {
			t.Errorf("Sparkline ascending: runes[%d] < runes[%d] in %q", i, i-1, got)
		}
	}
}

func TestSparkline_BrailleUp_SmallStepsUseStableGlyphs(t *testing.T) {
	history := []float64{0, 1, 2, 3}
	got := Sparkline(history, 4, SparklineOpts{Style: SparklineBrailleUp})
	want := "⣀⣤⣶⣿"
	if got != want {
		t.Fatalf("braille small-step output = %q, want %q", got, want)
	}
}

func TestSparkline_BrailleUp_LargeJumpUsesDirectionalGlyph(t *testing.T) {
	history := []float64{0, 3}
	got := Sparkline(history, 2, SparklineOpts{Style: SparklineBrailleUp})
	want := "⣀⣾"
	if got != want {
		t.Fatalf("braille large-jump output = %q, want %q", got, want)
	}
}

func TestSparkline_BrailleDown_LargeDropUsesDirectionalGlyph(t *testing.T) {
	history := []float64{3, 0}
	got := Sparkline(history, 2, SparklineOpts{Style: SparklineBrailleDown})
	want := "⣿⠁"
	if got != want {
		t.Fatalf("braille large-drop output = %q, want %q", got, want)
	}
}

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

func TestAppendHistory_Cap(t *testing.T) {
	var h []float64
	for i := range HistoryLen + 20 {
		h = AppendHistory(h, float64(i))
	}
	if len(h) != HistoryLen {
		t.Errorf("AppendHistory cap = %d, want %d", len(h), HistoryLen)
	}
	// Most recent value must be at the end.
	if h[len(h)-1] != float64(HistoryLen+19) {
		t.Errorf("AppendHistory last = %v, want %v", h[len(h)-1], float64(HistoryLen+19))
	}
}
