package charts

import (
	"strings"
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
