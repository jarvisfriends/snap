// Package scrollbar renders minimal scroll indicators for scrolling regions.
// Ported from the tribble console's dashboard scrollbar and decoupled from
// its model: pure geometry in, styled column out.
package scrollbar

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/geom"
)

// Styles selects the track and thumb appearance.
type Styles struct {
	Track lipgloss.Style
	Thumb lipgloss.Style
	// TrackRune / ThumbRune are single-cell glyphs ("░" / "█" by default).
	TrackRune string
	ThumbRune string
}

// DefaultStyles returns a dim track with a solid thumb.
func DefaultStyles() Styles {
	return Styles{
		Track:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Thumb:     lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		TrackRune: "░",
		ThumbRune: "█",
	}
}

// Vertical renders a one-column scrollbar of barHeight cells for content
// that is total lines long with visible lines shown, scrolled to offset.
// Returns "" when the content already fits (no scrollbar needed), so callers
// can join it unconditionally.
func Vertical(total, visible, offset, barHeight int, st Styles) string {
	if total <= visible || barHeight <= 0 || visible <= 0 {
		return ""
	}
	if st.TrackRune == "" || st.ThumbRune == "" {
		d := DefaultStyles()
		if st.TrackRune == "" {
			st.TrackRune = d.TrackRune
		}
		if st.ThumbRune == "" {
			st.ThumbRune = d.ThumbRune
		}
	}

	thumbSize := max(1, barHeight*visible/total)
	maxScroll := total - visible
	thumbPos := 0
	if maxScroll > 0 {
		thumbPos = geom.Clamp(offset, 0, maxScroll) * (barHeight - thumbSize) / maxScroll
	}
	thumbPos = geom.Clamp(thumbPos, 0, barHeight-thumbSize)

	track := st.Track.Render(st.TrackRune)
	thumb := st.Thumb.Render(st.ThumbRune)
	rows := make([]string, barHeight)
	for i := range rows {
		if i >= thumbPos && i < thumbPos+thumbSize {
			rows[i] = thumb
		} else {
			rows[i] = track
		}
	}
	return strings.Join(rows, "\n")
}

// ClampOffset bounds a scroll offset into the valid range for the content:
// [0, max(0, total-visible)]. Use it after wheel or drag adjustments.
func ClampOffset(offset, total, visible int) int {
	return geom.Clamp(offset, 0, max(0, total-visible))
}
