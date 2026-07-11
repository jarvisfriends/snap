package styles

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	testRed  = lipgloss.Color("#aa0000")
	testBlue = lipgloss.Color("#0000aa")
)

// TestPillCapGlyphsPerShape pins each shape's signature caps around the body.
func TestPillCapGlyphsPerShape(t *testing.T) {
	cases := []struct {
		shape       PillShape
		left, right string
	}{
		{PillRound, "", ""},
		{PillArrow, "", ""},
		{PillSlant, "", ""},
		{PillBlock, "▐", "▌"},
		{PillBracket, "[", "]"},
		{PillChevron, "❮", "❯"},
		{PillDiagonal, "◢", "◤"},
		{PillFade, "░▒", "▒░"},
	}
	for _, tc := range cases {
		got := ansi.Strip(Pill("hi", nil, testRed, PillStyles{Shape: tc.shape}))
		want := tc.left + "hi" + tc.right
		if got != want {
			t.Errorf("%s: got %q want %q", tc.shape, got, want)
		}
	}
}

// TestPillPlainPadsInsteadOfCaps: the plain shape has no caps and pads the
// body by one cell each side, matching the classic BadgeStyle look.
func TestPillPlainPadsInsteadOfCaps(t *testing.T) {
	got := ansi.Strip(Pill("hi", nil, testRed, PillStyles{Shape: PillPlain}))
	if got != " hi " {
		t.Errorf("got %q want %q", got, " hi ")
	}
}

// TestSegmentedPillDividerCarriesColors: between two different-bg segments
// the solid divider paints prev's bg as foreground over next's bg, so the
// glyph reads as the color boundary.
func TestSegmentedPillDividerCarriesColors(t *testing.T) {
	segs := []PillSegment{
		{Text: "a", Bg: testRed},
		{Text: "b", Bg: testBlue},
	}
	out := SegmentedPill(segs, PillStyles{})

	if got := ansi.Strip(out); got != "ab" {
		t.Fatalf("structure: got %q", got)
	}
	divider := lipgloss.NewStyle().
		Foreground(testRed).
		Background(testBlue).
		Render("")
	if !strings.Contains(out, divider) {
		t.Errorf("divider not painted prev-fg/next-bg:\nout %q\nwant substring %q", out, divider)
	}
}

// TestSegmentedPillSameBgUsesThinDivider: adjacent segments sharing a bg are
// separated by the thin outline glyph, not the solid one.
func TestSegmentedPillSameBgUsesThinDivider(t *testing.T) {
	segs := []PillSegment{
		{Text: "a", Bg: testRed},
		{Text: "b", Bg: testRed},
	}
	got := ansi.Strip(SegmentedPill(segs, PillStyles{}))
	if got != "ab" {
		t.Errorf("got %q", got)
	}
}

// TestSegmentedPillEmpty: no segments render nothing.
func TestSegmentedPillEmpty(t *testing.T) {
	if got := SegmentedPill(nil, PillStyles{}); got != "" {
		t.Errorf("got %q want empty", got)
	}
}

// TestPillBaseBackgroundOnCaps: a Base color paints the concave side of the
// caps so pills embed cleanly in status bars.
func TestPillBaseBackgroundOnCaps(t *testing.T) {
	out := Pill("x", nil, testRed, PillStyles{Base: testBlue})
	leftCap := lipgloss.NewStyle().Foreground(testRed).Background(testBlue).Render("")
	if !strings.HasPrefix(out, leftCap) {
		t.Errorf("left cap missing base bg:\nout %q\nwant prefix %q", out, leftCap)
	}
}

// TestPillFgAutoContrast: nil Fg picks black on light fills, white on dark.
func TestPillFgAutoContrast(t *testing.T) {
	light := pillFg(PillSegment{Bg: lipgloss.Color("#eeeeee")})
	if !sameColor(light, lipgloss.Color("#000000")) {
		t.Errorf("light bg: got %v want black", light)
	}
	dark := pillFg(PillSegment{Bg: lipgloss.Color("#222222")})
	if !sameColor(dark, lipgloss.Color("#ffffff")) {
		t.Errorf("dark bg: got %v want white", dark)
	}
	explicit := pillFg(PillSegment{Fg: testBlue, Bg: lipgloss.Color("#eeeeee")})
	if !sameColor(explicit, testBlue) {
		t.Errorf("explicit fg overridden: got %v", explicit)
	}
}

// TestSameColorAcrossTypes: equal colors expressed via different color.Color
// implementations compare equal, so thin-vs-solid divider choice is stable.
func TestSameColorAcrossTypes(t *testing.T) {
	if !sameColor(lipgloss.Color("#aa0000"), color.RGBA{R: 0xaa, A: 0xff}) {
		t.Error("hex vs RGBA should match")
	}
	if sameColor(testRed, testBlue) {
		t.Error("distinct colors should differ")
	}
	if sameColor(testRed, nil) {
		t.Error("nil vs color should differ")
	}
	if !sameColor(nil, nil) {
		t.Error("nil vs nil should match")
	}
}

// TestBreadcrumbsJoinsWithThinGlyph: items join on the shape's thin divider
// with one space each side; plain/unknown shapes fall back to "│".
func TestBreadcrumbsJoinsWithThinGlyph(t *testing.T) {
	sep := lipgloss.NewStyle()
	got := ansi.Strip(Breadcrumbs([]string{"a", "b", "c"}, sep, PillStyles{Shape: PillArrow}))
	if got != "a  b  c" {
		t.Errorf("arrow: got %q", got)
	}
	got = ansi.Strip(Breadcrumbs([]string{"a", "b"}, sep, PillStyles{Shape: PillPlain}))
	if got != "a │ b" {
		t.Errorf("plain: got %q", got)
	}
}

// TestPillShapeNormalization: unknown or cased input falls back sanely and
// every listed shape has a display name.
func TestPillShapeNormalization(t *testing.T) {
	if got := NormalizePillShape("  Arrow "); got != PillArrow {
		t.Errorf("got %q want %q", got, PillArrow)
	}
	if got := NormalizePillShape("bogus"); got != DefaultPillShape {
		t.Errorf("got %q want default", got)
	}
	if got := NormalizePillShape(""); got != DefaultPillShape {
		t.Errorf("empty: got %q want default", got)
	}
	shapes := PillShapes()
	if len(shapes) != len(pillGlyphSets) {
		t.Fatalf("PillShapes lists %d of %d shapes", len(shapes), len(pillGlyphSets))
	}
	for _, s := range shapes {
		if s.DisplayName() == "" {
			t.Errorf("%q has no display name", s)
		}
	}
	if PillShape("bogus").DisplayName() != PillRound.DisplayName() {
		t.Error("unknown shape should borrow the default display name")
	}
}

// TestPillNerdFontFlag: powerline shapes need a patched font, the Unicode
// fallbacks do not.
func TestPillNerdFontFlag(t *testing.T) {
	for _, s := range []PillShape{PillRound, PillArrow, PillSlant} {
		if !s.NeedsNerdFont() {
			t.Errorf("%q should need a Nerd Font", s)
		}
	}
	for _, s := range []PillShape{
		PillBlock, PillPlain, PillBracket, PillChevron, PillDiagonal, PillFade,
	} {
		if s.NeedsNerdFont() {
			t.Errorf("%q should not need a Nerd Font", s)
		}
	}
}
