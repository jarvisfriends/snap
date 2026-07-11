package charts

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

// CellCanvas is a colored whole-cell drawing surface: one rune plus true-color
// foreground and background per terminal cell. Where Canvas draws sub-cell
// braille pixels with a transparent background, CellCanvas owns every cell it
// covers — game boards, heatmaps, block art, anything that paints backgrounds.
// String batches ANSI escapes (colors re-emitted only when they change), so
// full-surface repaints stay cheap on every frame.
//
// Ported from brick-breaker's gameRenderer per the repo sweep; the drawing
// primitives there (bricks, paddle, trails) stay game-side.
type CellCanvas struct {
	width, height int
	cells         []canvasCell
	defFG, defBG  color.Color
}

type canvasCell struct {
	ch rune
	fg color.Color
	bg color.Color
}

// NewCellCanvas creates a canvas of w×h cells filled with spaces in the given
// default colors. A nil default renders as black — pass real colors when the
// surface is meant to be seen.
func NewCellCanvas(w, h int, fg, bg color.Color) *CellCanvas {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	c := &CellCanvas{
		width:  w,
		height: h,
		cells:  make([]canvasCell, w*h),
		defFG:  fg,
		defBG:  bg,
	}
	c.Clear()
	return c
}

// Size returns the canvas dimensions in cells.
func (c *CellCanvas) Size() (w, h int) { return c.width, c.height }

// Clear resets every cell to a space in the default colors.
func (c *CellCanvas) Clear() {
	for i := range c.cells {
		c.cells[i] = canvasCell{ch: ' ', fg: c.defFG, bg: c.defBG}
	}
}

// Set paints one cell: rune, foreground, and background. Out-of-range cells
// are ignored so callers can draw shapes that partially leave the surface.
func (c *CellCanvas) Set(x, y int, ch rune, fg, bg color.Color) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}
	c.cells[y*c.width+x] = canvasCell{ch: ch, fg: fg, bg: bg}
}

// SetFG paints a cell's rune and foreground while keeping its current
// background — glyphs over an already-painted surface (trails, walls, HUD).
func (c *CellCanvas) SetFG(x, y int, ch rune, fg color.Color) {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return
	}
	c.cells[y*c.width+x].ch = ch
	c.cells[y*c.width+x].fg = fg
}

// Rune reports the rune currently at a cell (space when out of range) —
// lets callers avoid overdrawing occupied cells, e.g. glow halos that only
// land on empty background.
func (c *CellCanvas) Rune(x, y int) rune {
	if x < 0 || x >= c.width || y < 0 || y >= c.height {
		return ' '
	}
	return c.cells[y*c.width+x].ch
}

// String renders the whole surface with batched truecolor escapes: fg/bg
// codes are emitted only when they differ from the previous cell's, roughly
// halving the bytes of a naive per-cell render.
func (c *CellCanvas) String() string {
	var b strings.Builder
	b.Grow(c.width * c.height * 15)

	var prevFR, prevFG, prevFB, prevBR, prevBG, prevBB uint8
	fgSet, bgSet := false, false

	for y := range c.height {
		for x := range c.width {
			cell := c.cells[y*c.width+x]
			fr, fg, fb := rgb8(cell.fg)
			br, bg, bb := rgb8(cell.bg)

			if !fgSet || fr != prevFR || fg != prevFG || fb != prevFB {
				fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm", fr, fg, fb)
				prevFR, prevFG, prevFB = fr, fg, fb
				fgSet = true
			}
			if !bgSet || br != prevBR || bg != prevBG || bb != prevBB {
				fmt.Fprintf(&b, "\x1b[48;2;%d;%d;%dm", br, bg, bb)
				prevBR, prevBG, prevBB = br, bg, bb
				bgSet = true
			}

			b.WriteRune(cell.ch)
		}
		if y < c.height-1 {
			// Reset per line so partial redraws and pagers can't smear a
			// row's colors across the terminal.
			b.WriteString("\x1b[0m\n")
			fgSet, bgSet = false, false
		}
	}
	b.WriteString("\x1b[0m")
	return b.String()
}

// rgb8 extracts 8-bit RGB components; nil renders as black.
func rgb8(c color.Color) (red, green, blue uint8) {
	if c == nil {
		return 0, 0, 0
	}
	r, g, b, _ := c.RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)
}

// Gradient returns a smooth HSV blend from one color to the other in the
// given number of steps (ends included). Fewer than two steps yields just the
// start color. Useful wherever a value range maps to color — heatmap scales,
// severity ramps, row tints.
func Gradient(from, to color.Color, steps int) []color.Color {
	start, okFrom := colorful.MakeColor(colorOrBlack(from))
	end, okTo := colorful.MakeColor(colorOrBlack(to))
	if steps < 2 || !okFrom || !okTo {
		return []color.Color{colorOrBlack(from)}
	}
	out := make([]color.Color, steps)
	for i := range steps {
		t := float64(i) / float64(steps-1)
		out[i] = start.BlendHsv(end, t).Clamped()
	}
	return out
}

func colorOrBlack(c color.Color) color.Color {
	if c == nil {
		return color.Black
	}
	return c
}
