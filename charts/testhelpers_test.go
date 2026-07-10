package charts

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// gridSize returns the rendered width (widest line, ANSI stripped) and height
// (line count) of a chart string.
func gridSize(s string) (w, h int) {
	for line := range strings.SplitSeq(s, "\n") {
		h++
		if lw := lipgloss.Width(line); lw > w {
			w = lw
		}
	}
	return w, h
}

// brailleDots counts set dots across all braille runes in s.
func brailleDots(s string) int {
	n := 0
	for _, r := range ansi.Strip(s) {
		if r >= 0x2800 && r <= 0x28FF {
			for mask := 0x01; mask <= 0x80; mask <<= 1 {
				if int(r-0x2800)&mask != 0 {
					n++
				}
			}
		}
	}
	return n
}
