package styles

import (
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/scrollbar"
)

// ScrollbarStyles returns scrollbar styles themed to the active palette: the
// thumb takes the same color as the app's main panel/nav borders (c.Border) so
// a scroll indicator reads as part of the surrounding frame, and the track uses
// the muted secondary color so it recedes behind the thumb. Pass the result to
// scrollbar.Vertical in place of scrollbar.DefaultStyles for a theme-aware bar.
func ScrollbarStyles(c *AppStyle) scrollbar.Styles {
	st := scrollbar.DefaultStyles()
	st.Track = lipgloss.NewStyle().Foreground(c.Muted)
	st.Thumb = lipgloss.NewStyle().Foreground(c.Border)
	return st
}
