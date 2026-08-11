// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package charts

import (
	"image/color"
	"math"

	"charm.land/lipgloss/v2"
	ntcanvas "github.com/NimbleMarkets/ntcharts/v2/canvas"
	"github.com/NimbleMarkets/ntcharts/v2/canvas/graph"
)

// LineSeries is one line in a BrailleLineChart: a rolling history plus the
// color its dots are drawn with.
type LineSeries struct {
	Label string
	Color color.Color
	Data  []float64
}

// BrailleLineChart renders one or more series as overlaid braille line graphs
// sharing the same scale — ideal for transmit/receive pairs. The newest
// sample is the rightmost column; series shorter than the window leave the
// left edge blank. Consecutive samples are connected with interpolated
// braille line segments, and dots from overlapping series merge within a
// cell (the later series' color wins the cell).
//
// The plotting is delegated to ntcharts' braille grid + canvas primitives
// (github.com/NimbleMarkets/ntcharts) — this wrapper keeps snap's rolling
// right-aligned window, NaN gaps, and scale reporting. Apps that want axes,
// tick labels, mouse zones, or the candlestick/waveline/streamline variants
// use ntcharts' linechart packages directly.
//
// charW/charH are terminal cells. maxVal fixes the top of the scale; pass
// <= 0 to auto-scale to the visible window. Returns the chart and the scale
// actually used, so callers can label it.
func BrailleLineChart(series []LineSeries, charW, charH int, maxVal float64) (chart string, scale float64) {
	if charW <= 0 || charH <= 0 {
		return "", 0
	}
	pixelW := charW * 2 // braille cells are 2 dots wide

	// Sample each series into the pixel window: the last pixelW values, or
	// a shorter history stretched across the full width so the chart always
	// spans the frame instead of huddling right while the buffer fills.
	sampled := make([][]float64, len(series))
	peak := 0.0
	for si, s := range series {
		samples := make([]float64, pixelW)
		data := s.Data
		switch {
		case len(data) >= pixelW:
			copy(samples, data[len(data)-pixelW:])
		case len(data) > 0:
			for i := range samples {
				samples[i] = data[i*len(data)/pixelW]
			}
		default:
			for i := range samples {
				samples[i] = math.NaN()
			}
		}
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

	c := ntcanvas.New(charW, charH)
	for si, s := range series {
		// One braille grid per series: dots accumulate per grid, then merge
		// onto the shared canvas (existing braille runes are combined, blank
		// cells skipped) so overlapping series keep both dot patterns.
		grid := graph.NewBrailleGrid(charW, charH, 0, float64(pixelW-1), 0, scale)
		prev := ntcanvas.Point{X: -1, Y: -1}
		havePrev := false
		for px := range pixelW {
			v := sampled[si][px]
			if math.IsNaN(v) {
				havePrev = false
				continue
			}
			p := grid.GridPoint(ntcanvas.Float64Point{
				X: float64(px),
				Y: min(max(v, 0), scale),
			})
			if havePrev {
				for _, lp := range graph.GetLinePoints(prev, p) {
					grid.Set(lp)
				}
			} else {
				grid.Set(p)
			}
			prev, havePrev = p, true
		}
		style := lipgloss.NewStyle()
		if s.Color != nil {
			style = style.Foreground(s.Color)
		}
		graph.DrawBraillePatterns(&c, ntcanvas.Point{X: 0, Y: 0}, grid.BraillePatterns(), style)
	}
	return c.View(), scale
}
