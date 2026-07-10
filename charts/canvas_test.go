package charts

import (
	"math/bits"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestCanvasRenderDimensions(t *testing.T) {
	c := NewCanvas(12, 5)
	out := c.Render()
	require.Equal(t, 5, lipgloss.Height(out))
	for _, line := range strings.Split(out, "\n") {
		require.Equal(t, 12, lipgloss.Width(line))
	}
}

func TestCanvasLineLightsCells(t *testing.T) {
	c := NewCanvas(10, 5)
	c.Line(0, 0, 19, 19, lipgloss.Color("2"))

	lines := strings.Split(ansi.Strip(c.Render()), "\n")
	cellAt := func(cx, cy int) rune {
		return []rune(lines[cy])[cx]
	}
	require.NotEqual(t, ' ', cellAt(0, 0), "line start cell should be lit")
	require.NotEqual(t, ' ', cellAt(9, 4), "line end cell should be lit")
	require.NotEqual(t, ' ', cellAt(5, 2), "line midpoint cell should be lit")
	require.Equal(t, ' ', cellAt(9, 0), "cells off the line stay blank")
}

// brailleDotCount sums the lit dots in a rendered canvas.
func brailleDotCount(rendered string) int {
	n := 0
	for _, r := range ansi.Strip(rendered) {
		if r > 0x2800 && r <= 0x28FF {
			n += bits.OnesCount16(uint16(r - 0x2800))
		}
	}
	return n
}

func TestCanvasThickLineWidensStroke(t *testing.T) {
	thin := NewCanvas(10, 5)
	thin.Line(0, 10, 19, 10, nil)
	thick := NewCanvas(10, 5)
	thick.ThickLine(0, 10, 19, 10, 3, nil)

	require.Equal(t, 3*brailleDotCount(thin.Render()), brailleDotCount(thick.Render()),
		"thickness 3 lays three parallel strokes")

	// Thickness 1 is exactly a plain line.
	one := NewCanvas(10, 5)
	one.ThickLine(0, 10, 19, 10, 1, nil)
	require.Equal(t, thin.Render(), one.Render())
}

func TestCanvasTextOverlaysBraille(t *testing.T) {
	c := NewCanvas(10, 3)
	c.Line(0, 4, 19, 4, nil) // braille row across cy=1
	c.Text(2, 1, "AB", lipgloss.NewStyle())

	lines := strings.Split(ansi.Strip(c.Render()), "\n")
	require.Contains(t, lines[1], "AB", "overlay text renders on top of braille")

	// Clipping and out-of-range writes must not panic.
	c.Text(9, 1, "overflow", lipgloss.NewStyle())
	c.Text(-3, 99, "gone", lipgloss.NewStyle())
	c.SetPixel(-1, -1, nil)
	c.SetPixel(999, 999, nil)
	require.Equal(t, 3, lipgloss.Height(c.Render()))
}
