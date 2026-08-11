// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package charts

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// HBar renders a horizontal proportional bar of the given width.
// pct is 0–100. Filled cells use '█', empty cells use '░'.
func HBar(pct float64, width int) string {
	if width <= 0 {
		return ""
	}
	pct = min(max(pct, 0), 100)
	filled := min(width, int(pct/100.0*float64(width)+0.5))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// HBarGradient renders HBar with the filled cells colored along a two-color
// gradient across the bar's full width (position-based, so the bar's tip
// shows how far into the gradient the value reaches).
func HBarGradient(pct float64, width int, from, to color.Color) string {
	if width <= 0 {
		return ""
	}
	if from == nil || to == nil {
		return HBar(pct, width)
	}
	pct = min(max(pct, 0), 100)
	filled := min(width, int(pct/100.0*float64(width)+0.5))
	ramp := Gradient(from, to, width)
	var sb strings.Builder
	for i := range filled {
		sb.WriteString(lipgloss.NewStyle().Foreground(ramp[i]).Render("█"))
	}
	sb.WriteString(strings.Repeat("░", width-filled))
	return sb.String()
}
