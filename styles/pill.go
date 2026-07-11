package styles

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// PillShape selects the geometry of pill caps and dividers. Shapes are
// user-selectable (settings pickers, YAML config) the same way StylePreset
// is: string-valued, normalized, with display names. Four shapes use
// Powerline-extras glyphs (private use area) and need a patched Nerd Font;
// the rest — Circle, Triangle, Diagonal, Fade, Block, Plain — are pure
// Unicode and render everywhere. NeedsNerdFont reports which is which.
type PillShape string

const (
	// PillRound (default): half-circle caps — the classic pill.
	PillRound PillShape = "round"
	// PillArrow: solid triangle caps, the original Powerline look.
	PillArrow PillShape = "arrow"
	// PillSlant: diagonal caps forming a parallelogram.
	PillSlant PillShape = "slant"
	// PillFlame: flame-edge caps for the ornate moods.
	PillFlame PillShape = "flame"
	// PillBlock: half-block caps — pure Unicode, no Nerd Font needed.
	PillBlock PillShape = "block"
	// PillPlain: no caps, one cell of padding inside the pill body.
	PillPlain PillShape = "plain"
	// PillCircle: geometric half-circle caps (◖ ◗) — the Round look without
	// a Nerd Font. Most fonts leave a hairline of cell background around the
	// glyph, so the caps read slightly detached compared to PillRound.
	PillCircle PillShape = "circle"
	// PillTriangle: solid pointer caps (◀ ▶) — the Arrow look without a
	// Nerd Font.
	PillTriangle PillShape = "triangle"
	// PillDiagonal: corner-triangle caps (◢ ◤) forming a parallelogram —
	// the Slant look without a Nerd Font.
	PillDiagonal PillShape = "diagonal"
	// PillFade: shade-block caps (░▒ … ▒░) that dissolve the pill into the
	// background — two cells per cap, pure Unicode.
	PillFade PillShape = "fade"
)

// DefaultPillShape is used when no shape has been chosen or an unknown value
// is supplied.
const DefaultPillShape = PillRound

// pillGlyphs holds the four glyphs a shape needs. Caps render with the pill
// background as their foreground over a transparent (or Base) background;
// dividers render with the previous segment's background as foreground and
// the next segment's as background, so the shape reads as a color boundary.
type pillGlyphs struct {
	left     string // outer left cap
	right    string // outer right cap
	divider  string // solid boundary between segments with different bg
	thin     string // outline boundary between segments sharing a bg
	nerdFont bool
	display  string
}

// The Nerd Font rows hold Powerline-extras private-use literals that render
// as tofu (or nothing) in unpatched fonts. Codepoints, in field order
// left/right/divider/thin: Round E0B6/E0B4/E0B4/E0B5, Arrow E0B2/E0B0/E0B0/
// E0B1, Slant E0BA/E0BC/E0BC/E0BD, Flame E0C2/E0C0/E0C0/E0C1.
var pillGlyphSets = map[PillShape]pillGlyphs{
	PillRound:    {left: "", right: "", divider: "", thin: "", nerdFont: true, display: "Round"},
	PillArrow:    {left: "", right: "", divider: "", thin: "", nerdFont: true, display: "Arrow"},
	PillSlant:    {left: "", right: "", divider: "", thin: "", nerdFont: true, display: "Slant"},
	PillFlame:    {left: "", right: "", divider: "", thin: "", nerdFont: true, display: "Flame"},
	PillBlock:    {left: "▐", right: "▌", divider: "▌", thin: "│", display: "Block"},
	PillPlain:    {thin: "│", display: "Plain"},
	PillCircle:   {left: "◖", right: "◗", divider: "◗", thin: "│", display: "Circle"},
	PillTriangle: {left: "◀", right: "▶", divider: "▶", thin: "›", display: "Triangle"},
	PillDiagonal: {left: "◢", right: "◤", divider: "◤", thin: "╱", display: "Diagonal"},
	PillFade:     {left: "░▒", right: "▒░", divider: "▒", thin: "░", display: "Fade"},
}

var orderedPillShapes = []PillShape{
	PillRound, PillArrow, PillSlant, PillFlame,
	PillCircle, PillTriangle, PillDiagonal, PillFade, PillBlock, PillPlain,
}

// PillShapes returns all shapes in presentation order, for settings pickers.
func PillShapes() []PillShape {
	out := make([]PillShape, len(orderedPillShapes))
	copy(out, orderedPillShapes)
	return out
}

// NormalizePillShape maps a stored string to a known shape, falling back to
// DefaultPillShape for unknown or empty input.
func NormalizePillShape(s string) PillShape {
	shape := PillShape(strings.ToLower(strings.TrimSpace(s)))
	if _, ok := pillGlyphSets[shape]; ok {
		return shape
	}
	return DefaultPillShape
}

// DisplayName returns the human-facing name for pickers.
func (p PillShape) DisplayName() string {
	if g, ok := pillGlyphSets[p]; ok {
		return g.display
	}
	return pillGlyphSets[DefaultPillShape].display
}

// NeedsNerdFont reports whether the shape's glyphs are Powerline-extras
// private-use codepoints that require a patched (Nerd) font.
func (p PillShape) NeedsNerdFont() bool {
	return pillGlyphSets[normalizeShape(p)].nerdFont
}

func normalizeShape(p PillShape) PillShape {
	if _, ok := pillGlyphSets[p]; ok {
		return p
	}
	return DefaultPillShape
}

// PillStyles selects a pill's appearance. The zero value uses PillRound with
// transparent cap backgrounds (the terminal background shows through the
// concave side of the caps).
type PillStyles struct {
	Shape PillShape
	// Base, when set, is painted behind the concave side of the caps —
	// use the bar's background when embedding pills in a status bar or
	// nav strip so resets don't expose the terminal default.
	Base color.Color
}

// PillSegment is one colored run inside a segmented pill.
type PillSegment struct {
	Text string
	// Fg is the text color; nil picks black or white by Bg luminance.
	Fg color.Color
	// Bg is the segment fill and the color the caps/dividers take.
	Bg color.Color
}

// Pill renders a single-color pill: cap, colored body, cap.
func Pill(text string, fg, bg color.Color, st PillStyles) string {
	return SegmentedPill([]PillSegment{{Text: text, Fg: fg, Bg: bg}}, st)
}

// SegmentedPill renders one pill whose interior is divided by color: each
// segment's background meets the next at the shape's divider glyph, and the
// outer ends get the shape's caps. Single-line only.
func SegmentedPill(segs []PillSegment, st PillStyles) string {
	if len(segs) == 0 {
		return ""
	}
	g := pillGlyphSets[normalizeShape(st.Shape)]

	var b strings.Builder
	b.WriteString(capStyle(segs[0].Bg, st.Base).Render(g.left))
	for i, seg := range segs {
		b.WriteString(bodyStyle(seg).Render(pillBody(seg.Text, g)))
		if i+1 < len(segs) {
			b.WriteString(pillDivider(g, seg, segs[i+1]))
		}
	}
	b.WriteString(capStyle(segs[len(segs)-1].Bg, st.Base).Render(g.right))
	return b.String()
}

// Breadcrumbs joins pre-rendered items with the shape's thin divider glyph,
// one space either side, rendered in sep (typically Styles.Dim). Items keep
// whatever styling they already carry, so this suits paths, nav trails, and
// mixed pill/text rows alike.
func Breadcrumbs(items []string, sep lipgloss.Style, st PillStyles) string {
	g := pillGlyphSets[normalizeShape(st.Shape)]
	thin := g.thin
	if thin == "" {
		thin = "│"
	}
	return strings.Join(items, sep.Render(" "+thin+" "))
}

// pillBody pads plain-shape pills by one cell; capped shapes already get
// their breathing room from the cap glyphs.
func pillBody(text string, g pillGlyphs) string {
	if g.left == "" && g.right == "" {
		return " " + text + " "
	}
	return text
}

// pillDivider renders the boundary between two adjacent segments: a solid
// glyph carrying prev's bg over next's bg, or the thin outline variant when
// both share a background.
func pillDivider(g pillGlyphs, prev, next PillSegment) string {
	if sameColor(prev.Bg, next.Bg) {
		div := lipgloss.NewStyle().Foreground(pillFg(prev)).Background(prev.Bg)
		return div.Render(g.thin)
	}
	div := lipgloss.NewStyle().Foreground(prev.Bg).Background(next.Bg)
	if g.divider == "" {
		return div.Render("")
	}
	return div.Render(g.divider)
}

func capStyle(bg, base color.Color) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(bg)
	if base != nil {
		s = s.Background(base)
	}
	return s
}

func bodyStyle(seg PillSegment) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(pillFg(seg)).Background(seg.Bg)
}

// pillFg resolves a segment's text color: explicit Fg wins, otherwise black
// or white by background luminance so labels stay readable on any fill.
func pillFg(seg PillSegment) color.Color {
	if seg.Fg != nil {
		return seg.Fg
	}
	// colorLuminance is on a 0–255 scale; 128 splits light from dark fills.
	if seg.Bg != nil && colorLuminance(seg.Bg) > 128 {
		return lipgloss.Color("#000000")
	}
	return lipgloss.Color("#ffffff")
}

// sameColor compares colors by their RGBA values, so equal colors expressed
// through different color.Color types still match.
func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	aR, aG, aB, aA := a.RGBA()
	bR, bG, bB, bA := b.RGBA()
	return aR == bR && aG == bG && aB == bB && aA == bA
}
