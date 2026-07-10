package charts

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Canvas is a general-purpose braille drawing surface: a pixel layer (2×4
// dots per terminal cell) for lines and points, plus a text overlay layer for
// glyphs and labels drawn on top. The game prototype drives it, but it is
// deliberately generic — any chart or diagram can use it.
type Canvas struct {
	wCells, hCells int
	dots           []int         // braille bitmask per cell
	dotColor       []color.Color // per-cell dot color; the last write wins
	overlay        []string      // pre-styled single-cell strings; "" = none
}

// NewCanvas creates an empty canvas of the given size in terminal cells.
func NewCanvas(wCells, hCells int) *Canvas {
	if wCells < 1 {
		wCells = 1
	}
	if hCells < 1 {
		hCells = 1
	}
	n := wCells * hCells
	return &Canvas{
		wCells:   wCells,
		hCells:   hCells,
		dots:     make([]int, n),
		dotColor: make([]color.Color, n),
		overlay:  make([]string, n),
	}
}

// PixelSize returns the drawable resolution: 2 dots per cell across, 4 down.
func (c *Canvas) PixelSize() (w, h int) { return c.wCells * 2, c.hCells * 4 }

// SetPixel lights one braille dot. Out-of-range pixels are ignored, so
// callers can draw shapes that partially leave the field.
func (c *Canvas) SetPixel(px, py int, col color.Color) {
	if px < 0 || py < 0 || px >= c.wCells*2 || py >= c.hCells*4 {
		return
	}
	idx := (py/4)*c.wCells + px/2
	c.dots[idx] |= brailleDotBit[px%2][py%4]
	if col != nil {
		c.dotColor[idx] = col
	}
}

// Line draws a straight pixel line between two points (Bresenham).
func (c *Canvas) Line(x0, y0, x1, y1 int, col color.Color) {
	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		c.SetPixel(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			return
		}
		if e2 := 2 * err; e2 >= dy {
			err += dy
			x0 += sx
		} else {
			err += dx
			y0 += sy
		}
	}
}

// ThickLine draws a line of the given pixel thickness by laying parallel
// lines offset perpendicular to the direction of travel. Thickness 1 is a
// plain Line; 2–3 read as progressively heavier strokes.
func (c *Canvas) ThickLine(x0, y0, x1, y1, thickness int, col color.Color) {
	c.Line(x0, y0, x1, y1, col)
	if thickness <= 1 {
		return
	}
	// Perpendicular step: offset along the minor axis of the line so the
	// extra strokes hug the original.
	dx, dy := x1-x0, y1-y0
	ox, oy := 0, 1 // horizontal-ish lines widen vertically
	if absInt(dy) > absInt(dx) {
		ox, oy = 1, 0 // vertical-ish lines widen horizontally
	}
	for t := 1; t < thickness; t++ {
		// Alternate sides: +1, -1, +2, -2, …
		off := (t + 1) / 2
		if t%2 == 0 {
			off = -off
		}
		c.Line(x0+ox*off, y0+oy*off, x1+ox*off, y1+oy*off, col)
	}
}

// Text lays a styled string over the cells starting at (cx, cy), one rune per
// cell, clipping at the canvas edge. Overlay cells hide the braille beneath.
func (c *Canvas) Text(cx, cy int, text string, style lipgloss.Style) {
	if cy < 0 || cy >= c.hCells {
		return
	}
	for _, r := range text {
		if cx >= c.wCells {
			return
		}
		if cx >= 0 {
			c.overlay[cy*c.wCells+cx] = style.Render(string(r))
		}
		cx++
	}
}

// Render composes the canvas into terminal lines.
func (c *Canvas) Render() string {
	var sb strings.Builder
	for cy := range c.hCells {
		for cx := range c.wCells {
			idx := cy*c.wCells + cx
			if c.overlay[idx] != "" {
				sb.WriteString(c.overlay[idx])
				continue
			}
			if c.dots[idx] == 0 {
				sb.WriteString(" ")
				continue
			}
			style := lipgloss.NewStyle()
			if c.dotColor[idx] != nil {
				style = style.Foreground(c.dotColor[idx])
			}
			sb.WriteString(style.Render(string(rune(0x2800 | c.dots[idx]))))
		}
		if cy < c.hCells-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
