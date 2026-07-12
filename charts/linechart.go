package charts

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// LineSeries is one line in a BrailleLineChart: a rolling history plus the
// color its dots are drawn with.
type LineSeries struct {
	Label string
	Color color.Color
	Data  []float64
}

// brailleDotBit maps a (dx, dy) pixel offset inside one braille cell (2 wide,
// 4 tall) to its bit in the U+2800 block.
var brailleDotBit = [2][4]int{
	{0x01, 0x02, 0x04, 0x40},
	{0x08, 0x10, 0x20, 0x80},
}

// BrailleLineChart renders one or more series as overlaid braille line graphs
// sharing the same scale — ideal for transmit/receive pairs. The newest
// sample is the rightmost column; series shorter than the window leave the
// left edge blank. Vertical gaps between consecutive samples are filled so
// steep changes read as lines, and cells where series overlap blend their
// colors proportionally.
//
// charW/charH are terminal cells. maxVal fixes the top of the scale; pass
// <= 0 to auto-scale to the visible window. Returns the chart and the scale
// actually used, so callers can label it.
func BrailleLineChart(series []LineSeries, charW, charH int, maxVal float64) (chart string, scale float64) {
	if charW <= 0 || charH <= 0 {
		return "", 0
	}
	pixelW, pixelH := charW*2, charH*4

	// Sample each series into the pixel window: last pixelW values,
	// right-aligned, NaN where there's no data yet.
	sampled := make([][]float64, len(series))
	peak := 0.0
	for si, s := range series {
		samples := make([]float64, pixelW)
		for i := range samples {
			samples[i] = math.NaN()
		}
		data := s.Data
		if len(data) > pixelW {
			data = data[len(data)-pixelW:]
		}
		copy(samples[pixelW-len(data):], data)
		for _, v := range data {
			if v > peak {
				peak = v
			}
		}
		sampled[si] = samples
	}
	scale = maxVal
	if scale <= 0 {
		scale = peak
	}
	if scale <= 0 {
		scale = 1
	}

	// Plot into per-cell dot masks and per-cell series hit counts.
	dots := make([][]int, charH)
	counts := make([][][]int, charH)
	for cy := range dots {
		dots[cy] = make([]int, charW)
		counts[cy] = make([][]int, charW)
	}
	plot := func(si, px, py int) {
		cx, cy := px/2, py/4
		dots[cy][cx] |= brailleDotBit[px%2][py%4]
		if counts[cy][cx] == nil {
			counts[cy][cx] = make([]int, len(series))
		}
		counts[cy][cx][si]++
	}

	for si := range sampled {
		prevY := -1
		for px := range pixelW {
			v := sampled[si][px]
			if math.IsNaN(v) {
				prevY = -1
				continue
			}
			f := min(max(v/scale, 0), 1)
			py := pixelH - 1 - int(f*float64(pixelH-1)+0.5)
			// Fill the vertical run to the previous sample so steep moves
			// stay connected instead of raining isolated dots.
			lo, hi := py, py
			if prevY >= 0 {
				lo, hi = min(py, prevY), max(py, prevY)
			}
			for y := lo; y <= hi; y++ {
				plot(si, px, y)
			}
			prevY = py
		}
	}

	var sb strings.Builder
	for cy := range charH {
		for cx := range charW {
			mask := dots[cy][cx]
			if mask == 0 {
				sb.WriteString(" ")
				continue
			}
			style := lipgloss.NewStyle()
			if fg := blendSeriesColors(series, counts[cy][cx]); fg != nil {
				style = style.Foreground(fg)
			}
			sb.WriteString(style.Render(string(rune(0x2800 | mask))))
		}
		if cy < charH-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String(), scale
}

// blendSeriesColors averages the colors of every series present in a cell,
// weighted by how many of the cell's dots each contributed. A cell owned by a
// single series keeps that series' color exactly.
func blendSeriesColors(series []LineSeries, counts []int) color.Color {
	if counts == nil {
		return nil
	}
	present := -1
	total := 0
	for si, c := range counts {
		if c > 0 {
			total += c
			if present == -1 {
				present = si
			} else if present >= 0 && si != present {
				present = -2 // more than one series in this cell
			}
		}
	}
	if total == 0 {
		return nil
	}
	if present >= 0 {
		return series[present].Color
	}

	var r, g, b float64
	for si, c := range counts {
		if c == 0 || series[si].Color == nil {
			continue
		}
		cr, cg, cb, _ := series[si].Color.RGBA()
		r += float64(cr) * float64(c)
		g += float64(cg) * float64(c)
		b += float64(cb) * float64(c)
	}
	r /= float64(total)
	g /= float64(total)
	b /= float64(total)
	return lipgloss.Color(hexRGB(r, g, b))
}

// hexRGB formats blended 0–65535 RGBA components as a "#rrggbb" string
// (scaled back to 0–255). Shared by the line chart's and sankey's cell
// color blending.
func hexRGB(r, g, b float64) string {
	const digits = "0123456789abcdef"
	buf := [7]byte{'#'}
	for i, v := range [3]int{int(r / 257.0), int(g / 257.0), int(b / 257.0)} {
		v = min(max(v, 0), 255)
		buf[1+i*2] = digits[v>>4]
		buf[2+i*2] = digits[v&0xf]
	}
	return string(buf[:])
}
