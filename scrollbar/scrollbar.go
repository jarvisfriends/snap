// Package scrollbar renders minimal scroll indicators for scrolling regions.
// Ported from the tribble console's dashboard scrollbar, decoupled from its
// model, and restyled: the default is a thin line track with a heavy thumb
// (the fzf/yazi look), and PresetSmooth draws the thumb with eighth-block
// glyphs for 8x positional resolution (the btop look) so it glides instead
// of jumping cell by cell.
package scrollbar

import (
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/geom"
)

// Preset selects the scrollbar's look.
type Preset int

const (
	// PresetLine (default): a thin "│" track with a heavy "┃" thumb.
	PresetLine Preset = iota
	// PresetSmooth: a floating block thumb with sub-cell (1/8) resolution —
	// partial cells render as eighth blocks, so slow scrolls glide.
	PresetSmooth
	// PresetClassic: the retro "░" track with a "█" thumb.
	PresetClassic
)

// Styles selects the scrollbar's appearance. The zero value uses PresetLine
// with the default glyphs and colors.
type Styles struct {
	Preset Preset
	// Track / Thumb carry the colors. For PresetSmooth the thumb style must
	// be color-only (no underline etc.): partial cells invert it to paint
	// the sub-cell boundary.
	Track lipgloss.Style
	Thumb lipgloss.Style
	// TrackRune / ThumbRune override the single-cell glyphs for PresetLine
	// and PresetClassic.
	TrackRune string
	ThumbRune string
}

// DefaultStyles returns the thin-line look: dim track, bright heavy thumb.
func DefaultStyles() Styles {
	return Styles{
		Track: lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
		Thumb: lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
	}
}

// glyphs returns the effective track and thumb runes for the preset.
func (st Styles) glyphs() (track, thumb string) {
	track, thumb = st.TrackRune, st.ThumbRune
	if track == "" {
		if st.Preset == PresetClassic {
			track = "░"
		} else {
			track = "│"
		}
	}
	if thumb == "" {
		if st.Preset == PresetClassic {
			thumb = "█"
		} else {
			thumb = "┃"
		}
	}
	return track, thumb
}

// Vertical renders a one-column scrollbar of barHeight cells for content
// that is total lines long with visible lines shown, scrolled to offset.
// Returns "" when the content already fits (no scrollbar needed), so callers
// can join it unconditionally.
func Vertical(total, visible, offset, barHeight int, st Styles) string {
	if total <= visible || barHeight <= 0 || visible <= 0 {
		return ""
	}
	if st.Preset == PresetSmooth {
		return verticalSmooth(total, visible, offset, barHeight, st)
	}

	trackRune, thumbRune := st.glyphs()
	thumbSize := max(1, barHeight*visible/total)
	maxScroll := total - visible
	thumbPos := 0
	if maxScroll > 0 {
		thumbPos = geom.Clamp(offset, 0, maxScroll) * (barHeight - thumbSize) / maxScroll
	}
	thumbPos = geom.Clamp(thumbPos, 0, barHeight-thumbSize)

	track := st.Track.Render(trackRune)
	thumb := st.Thumb.Render(thumbRune)
	rows := make([]string, barHeight)
	for i := range rows {
		if i >= thumbPos && i < thumbPos+thumbSize {
			rows[i] = thumb
		} else {
			rows[i] = track
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// lowerBlocks[k] is the "lower k eighths" block (index 0 unused).
var lowerBlocks = []string{"", "▁", "▂", "▃", "▄", "▅", "▆", "▇"}

// verticalSmooth draws the thumb in eighth-cell units: the boundary cells
// render as partial blocks — the thumb's bottom edge as a regular lower
// block, its top edge as an inverted one — so the thumb glides with 8x the
// positional resolution.
func verticalSmooth(total, visible, offset, barHeight int, st Styles) string {
	barE := barHeight * 8
	thumbE := max(4, barE*visible/total) // at least half a cell of thumb
	maxScroll := total - visible
	posE := 0
	if maxScroll > 0 {
		posE = geom.Clamp(offset, 0, maxScroll) * (barE - thumbE) / maxScroll
	}
	posE = geom.Clamp(posE, 0, barE-thumbE)
	endE := posE + thumbE

	rows := make([]string, barHeight)
	for i := range rows {
		cellStart, cellEnd := i*8, i*8+8
		k := min(endE, cellEnd) - max(posE, cellStart) // eighths of thumb here
		switch {
		case k <= 0:
			rows[i] = " "
		case k >= 8:
			rows[i] = st.Thumb.Render("█")
		case endE >= cellEnd:
			// Thumb enters from below: the bottom k eighths are thumb.
			rows[i] = st.Thumb.Render(lowerBlocks[k])
		default:
			// Thumb leaves through the top: the top k eighths are thumb.
			// There are no "upper k eighths" glyphs, so render the
			// complement as a lower block with the style reversed — the
			// glyph area shows the terminal background and the remainder
			// shows the thumb color.
			rows[i] = st.Thumb.Reverse(true).Render(lowerBlocks[8-k])
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// ClampOffset bounds a scroll offset into the valid range for the content:
// [0, max(0, total-visible)]. Use it after wheel or drag adjustments.
func ClampOffset(offset, total, visible int) int {
	return geom.Clamp(offset, 0, max(0, total-visible))
}

// OffsetAt maps a pointer row on the scrollbar to the scroll offset that
// puts the thumb's center there — the standard click-the-track /
// drag-the-thumb behavior, and the inverse of Vertical's thumb placement.
// y is the row within the bar (0-based: subtract the bar's top screen row
// from the event's Y). Feed it both clicks and drag motion while the button
// is held; the result is already clamped.
func OffsetAt(y, barHeight, total, visible int) int {
	if total <= visible || barHeight <= 0 || visible <= 0 {
		return 0
	}
	thumbSize := max(1, barHeight*visible/total)
	track := barHeight - thumbSize
	if track <= 0 {
		// The thumb fills the bar (content barely overflows): every row maps
		// to the same degenerate position.
		return 0
	}
	maxScroll := total - visible
	pos := geom.Clamp(y-thumbSize/2, 0, track)
	// Round to nearest so mid-track clicks don't all bias toward the top.
	return geom.Clamp((pos*maxScroll+track/2)/track, 0, maxScroll)
}
