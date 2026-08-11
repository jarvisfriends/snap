// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package styles

import (
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/pickers"
)

// PickerStyles returns pickers styles themed to the active palette, replacing
// pickers.DefaultStyles' fixed ANSI colors so the picker follows tint changes
// like every other component: the title takes the accent, the browsed path and
// dimmed rows recede with the muted color, and the highlighted entry uses the
// shared selection pair.
func PickerStyles(c *AppStyle) pickers.Styles {
	st := pickers.DefaultStyles()
	st.Title = c.Styles.Title
	st.Path = lipgloss.NewStyle().Foreground(c.Muted)
	st.Selected = c.Styles.SelectedItem
	st.Normal = c.Styles.TextOnBg
	st.Dim = c.Styles.Dim
	return st
}
