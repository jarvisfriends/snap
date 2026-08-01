// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package styles

import (
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/timepicker"
)

// ApplyTimeFieldTheme replaces the time field's fixed ANSI defaults with the
// active palette (same intent as PickerStyles/ScrollbarStyles): the focused
// column takes the accent, unfocused columns recede with muted/border tones,
// and dropdown selection uses the shared selection pair — so the field
// follows tint changes like every other component.
func ApplyTimeFieldTheme(m *timepicker.TimeFieldModel, c *AppStyle) {
	rounded := lipgloss.RoundedBorder()
	m.ActiveStyle = lipgloss.NewStyle().
		Foreground(c.Accent).Bold(true).Padding(0, 1).
		Border(rounded).BorderForeground(c.Accent)
	m.InactiveStyle = lipgloss.NewStyle().
		Foreground(c.Muted).Padding(0, 1).
		Border(rounded).BorderForeground(c.Border)
	m.ColonStyle = lipgloss.NewStyle().Foreground(c.Accent).Bold(true)
	m.ListStyle = lipgloss.NewStyle().Border(rounded).BorderForeground(c.Border)
	m.SelectedStyle = c.Styles.SelectedItem.Padding(0, 1)
	m.RowStyle = lipgloss.NewStyle().Foreground(c.Fg).Padding(0, 1)
	m.HelpStyle = c.Styles.Dim
}
