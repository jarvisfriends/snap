package charts

import (
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
