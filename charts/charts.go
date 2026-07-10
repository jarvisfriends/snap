package charts

import (
	"fmt"
	"image/color"
	"maps"
	"math"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/jarvisfriends/snap/styles"
)

// HistoryLen is the default number of samples kept for a sparkline ring
// buffer (AppendHistory trims to it). Came along from dash's widget config.
const HistoryLen = 120

// SparklineStyle selects the glyph set and rendering mode for a sparkline.
type SparklineStyle int

const (
	// SparklineUserBlocks uses ▁▂▃▄▅▆▇█ — 8-level gradient fill blocks. Default.
	SparklineUserBlocks SparklineStyle = 0
	// SparklineBrailleUp uses directional braille glyphs where rising = good (speed metrics).
	// Glyphs are colored green when rising, red when falling, dim when stable.
	SparklineBrailleUp SparklineStyle = 1
	// SparklineBrailleDown uses directional braille glyphs where rising = bad (latency metrics).
	// Glyphs are colored red when rising, green when falling, dim when stable.
	SparklineBrailleDown SparklineStyle = 2
	// SparklineStdBlocks uses (space)▂▃▄▅▆▇█ — standard blocks with an explicit space for zero.
	SparklineStdBlocks SparklineStyle = 3
)

// SparklineStyleName returns the display name for the given SparklineStyle.
func SparklineStyleName(s SparklineStyle) string {
	return sparklineStyleNames[int(s)%len(sparklineStyleNames)]
}

// sparklineStyleNames is the ordered list of display names matching the SparklineStyle constants.
var sparklineStyleNames = []string{
	"Blocks (gradient)",      // SparklineUserBlocks
	"Braille Up (speed)",     // SparklineBrailleUp
	"Braille Down (latency)", // SparklineBrailleDown
	"Blocks (standard)",      // SparklineStdBlocks
}

// glyphSets holds the rune palette for each SparklineStyle. Braille sets have
// 12 runes arranged as 4 magnitude levels × 3 direction slots
// (slot 0 = rising, 1 = stable, 2 = falling).
var glyphSets = [][]rune{
	[]rune("▁▂▃▄▅▆▇█"),     // SparklineUserBlocks — 8 levels
	[]rune("⢀⣀⡀⣠⣤⣄⣴⣶⣦⣾⣿⣷"), // SparklineBrailleUp — 4 magnitudes × 3 directions
	[]rune("⠈⠉⠁⠙⠛⠋⠻⠿⠟⢿⣿⡿"), // SparklineBrailleDown — 4 magnitudes × 3 directions
	[]rune(" ▂▃▄▅▆▇█"),     // SparklineStdBlocks — 8 levels (space = zero)
}

// SparklineOpts configures a Sparkline call.
type SparklineOpts struct {
	// Style selects the glyph set. Defaults to SparklineUserBlocks when zero.
	Style SparklineStyle
	// Colors, when non-nil, enables per-glyph ANSI coloring for braille styles.
	// Block styles always return plain text regardless of this field.
	Colors *styles.AppStyle
}

// IsBrailleStyle reports whether s is a directional braille style. When true,
// Sparkline returns a pre-ANSI-colored string and callers must not re-apply a
// foreground color (use padRight instead of colorLine).
func IsBrailleStyle(s SparklineStyle) bool {
	return s == SparklineBrailleUp || s == SparklineBrailleDown
}

// Sparkline renders a compact one-row sparkline from a history slice.
// history must be non-empty; width is the number of terminal columns to fill.
// Values are normalised against the local min/max so the full range always
// uses all glyph levels.
//
// For braille styles with non-nil opts.Colors, each glyph is individually
// styled with ANSI color sequences based on the value direction — callers
// must use padRight rather than colorLine to avoid double-coloring.
// Block styles always return plain text.
func Sparkline(history []float64, width int, opts SparklineOpts) string {
	if len(history) == 0 || width <= 0 {
		return strings.Repeat(" ", max(width, 0))
	}

	lo, hi := history[0], history[0]
	for _, v := range history[1:] {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}

	// Sample the most recent `width` values (or pad with lo if shorter).
	samples := make([]float64, width)
	if len(history) >= width {
		copy(samples, history[len(history)-width:])
	} else {
		for i := range samples {
			samples[i] = lo
		}
		offset := width - len(history)
		copy(samples[offset:], history)
	}

	rng := hi - lo
	glyphs := glyphSets[int(opts.Style)%len(glyphSets)]

	if IsBrailleStyle(opts.Style) {
		return renderBraille(samples, rng, lo, glyphs, opts.Style, opts.Colors)
	}
	return renderBlocks(samples, rng, lo, glyphs)
}

// renderBlocks renders a plain-text sparkline using a linear block glyph set.
func renderBlocks(samples []float64, rng, lo float64, glyphs []rune) string {
	n := len(glyphs)
	var sb strings.Builder
	for _, v := range samples {
		var idx int
		if rng > 0 {
			f := (v - lo) / rng
			idx = max(int(f*float64(n-1)+0.5), 0)
			if idx >= n {
				idx = n - 1
			}
		}
		sb.WriteRune(glyphs[idx])
	}
	return sb.String()
}

// renderBraille renders a directional braille sparkline. Each glyph set has
// 12 runes organized as 4 magnitude levels × 3 direction slots:
//
//	glyphIdx = magnitudeLevel*3 + directionSlot
//	directionSlot: 0 = rising, 1 = stable, 2 = falling
//
// When colors is non-nil, each glyph is wrapped with a per-direction style:
//
//	BrailleUp:   rising → Success+bold, stable → Dim, falling → Error
//	BrailleDown: rising → Error,        stable → Dim, falling → Success+bold
func renderBraille(samples []float64, rng, lo float64, glyphs []rune, style SparklineStyle, colors *styles.AppStyle) string {
	const magLevels = 4
	const dirSlots = 3
	const directionalMinMagDelta = 2 // only show rise/fall glyphs on larger visual jumps

	// Stability epsilon: changes ≤2% of the visible range count as stable.
	var eps float64
	if rng > 0 {
		eps = rng * 0.02
	}

	var sb strings.Builder
	prevV := 0.0
	prevMagIdx := 0
	for i, v := range samples {
		// Map value → magnitude level 0–3.
		var magIdx int
		if rng > 0 {
			f := (v - lo) / rng
			magIdx = max(min(int(f*float64(magLevels-1)+0.5), magLevels-1), 0)
		}

		// Determine direction relative to the previous sample.
		// Directional glyphs are only used when the magnitude changed by at
		// least two levels; small one-level moves stay on the stable glyph to
		// reduce visual jitter in braille mode.
		dirIdx := 1 // stable default for the first sample
		if i > 0 {
			magDelta := magIdx - prevMagIdx
			if magDelta < 0 {
				magDelta = -magDelta
			}
			if magDelta >= directionalMinMagDelta {
				delta := v - prevV
				switch {
				case delta > eps:
					dirIdx = 0 // rising
				case delta < -eps:
					dirIdx = 2 // falling
				}
			}
		}

		glyphIdx := magIdx*dirSlots + dirIdx
		if glyphIdx >= len(glyphs) {
			glyphIdx = len(glyphs) - 1
		}
		glyph := string(glyphs[glyphIdx])

		if colors == nil {
			sb.WriteString(glyph)
			prevV = v
			prevMagIdx = magIdx
			continue
		}

		// Select the per-glyph lipgloss style based on style + direction.
		var s lipgloss.Style
		switch {
		case style == SparklineBrailleUp && dirIdx == 0: // rising speed → good
			s = colors.Styles.Success.Bold(true)
		case style == SparklineBrailleUp && dirIdx == 2: // falling speed → bad
			s = colors.Styles.Error
		case style == SparklineBrailleDown && dirIdx == 0: // rising latency → bad
			s = colors.Styles.Error
		case style == SparklineBrailleDown && dirIdx == 2: // falling latency → good
			s = colors.Styles.Success.Bold(true)
		default:
			s = colors.Styles.Dim
		}
		sb.WriteString(s.Render(glyph))
		prevV = v
		prevMagIdx = magIdx
	}
	return sb.String()
}

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

// AppendHistory appends v to history, keeping at most HistoryLen entries.
func AppendHistory(history []float64, v float64) []float64 {
	history = append(history, v)
	if len(history) > HistoryLen {
		history = history[len(history)-HistoryLen:]
	}
	return history
}

// PieSlice represents a single slice in a PieChart.
type PieSlice struct {
	Value float64
	Color color.Color
	Label string // currently unused in rendering the circle, but useful for legends
}

// PieChart renders a text-based pie chart using ANSI background colors.
// radius is the approximate vertical radius in terminal lines.
func PieChart(slices []PieSlice, radius int) string {
	if len(slices) == 0 || radius <= 0 {
		return ""
	}

	total := 0.0
	for _, s := range slices {
		total += s.Value
	}

	if total == 0 {
		return ""
	}

	// Calculate angles (in radians, from 0 to 2pi)
	angles := make([]float64, len(slices))
	currentAngle := 0.0
	for i, s := range slices {
		currentAngle += (s.Value / total) * 2 * 3.1415926535
		angles[i] = currentAngle
	}

	width := radius * 2
	height := radius

	var sb strings.Builder

	for y := range height {
		for x := range width {
			nx := (float64(x) - float64(width)/2.0 + 0.5) / float64(radius)
			ny := (float64(y) - float64(height)/2.0 + 0.5) / float64(height/2)

			// dist logic for a perfect circle
			dist := nx*nx + ny*ny
			if dist > 1.0 {
				sb.WriteString("  ")
				continue
			}

			// atan2 returns -pi to pi.
			// Map to 0 to 2pi, starting from top
			theta := math.Atan2(ny, nx)
			theta += math.Pi / 2.0
			if theta < 0 {
				theta += 2 * math.Pi
			}

			sliceIdx := 0
			for i, a := range angles {
				if theta <= a {
					sliceIdx = i
					break
				}
			}
			if sliceIdx >= len(slices) {
				sliceIdx = len(slices) - 1
			}

			style := lipgloss.NewStyle().Background(slices[sliceIdx].Color)
			sb.WriteString(style.Render("  "))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// BraillePieChart renders a text-based pie chart using Braille characters.
// radius is the approximate vertical radius in terminal lines.
func BraillePieChart(slices []PieSlice, radius int) string {
	if len(slices) == 0 || radius <= 0 {
		return ""
	}

	total := 0.0
	for _, s := range slices {
		total += s.Value
	}

	angles := make([]float64, len(slices))
	currentAngle := 0.0
	for i, s := range slices {
		currentAngle += (s.Value / total) * 2 * math.Pi
		angles[i] = currentAngle
	}

	charW := radius * 2
	charH := radius
	pixelW := charW * 2
	pixelH := charH * 4

	var sb strings.Builder

	for cy := range charH {
		for cx := range charW {
			dotValues := [8]int{-1, -1, -1, -1, -1, -1, -1, -1}
			dotMap := [8][2]int{
				{0, 0},
				{0, 1},
				{0, 2},
				{1, 0},
				{1, 1},
				{1, 2},
				{0, 3},
				{1, 3},
			}

			counts := make([]int, len(slices))
			outsideCount := 0

			for i, offset := range dotMap {
				px := cx*2 + offset[0]
				py := cy*4 + offset[1]

				nx := (float64(px) - float64(pixelW)/2.0 + 0.5) / float64(pixelW/2)
				ny := (float64(py) - float64(pixelH)/2.0 + 0.5) / float64(pixelH/2)

				dist := math.Sqrt(nx*nx + ny*ny)
				if dist > 1.0 {
					outsideCount++
					continue
				}

				theta := math.Atan2(ny, nx)
				theta += math.Pi / 2.0
				if theta < 0 {
					theta += 2 * math.Pi
				}

				sliceIdx := 0
				for j, a := range angles {
					if theta <= a {
						sliceIdx = j
						break
					}
				}
				if sliceIdx >= len(slices) {
					sliceIdx = len(slices) - 1
				}

				dotValues[i] = sliceIdx
				counts[sliceIdx]++
			}

			if outsideCount == 8 {
				sb.WriteString(" ")
				continue
			}

			domSlice := -1
			maxCount := -1
			var domColor color.Color
			for idx := range slices {
				if counts[idx] > maxCount {
					maxCount = counts[idx]
					domSlice = idx
					domColor = slices[idx].Color
				}
			}

			runeVal := 0x2800
			brailleOffsets := []int{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
			for i, v := range dotValues {
				if v == domSlice {
					runeVal |= brailleOffsets[i]
				}
			}

			style := lipgloss.NewStyle()
			if domSlice != -1 {
				style = style.Foreground(domColor)
			}

			sb.WriteString(style.Render(string(rune(runeVal))))
		}
		if cy < charH-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

type SankeyFlow struct {
	Source string
	Target string
	Value  float64
	Color  color.Color
}

func smoothstep(t float64) float64 {
	if t <= 0.0 {
		return 0.0
	}
	if t >= 1.0 {
		return 1.0
	}
	return t * t * (3.0 - 2.0*t)
}

func BrailleSankeyChart(flows []SankeyFlow, charW, charH int) string {
	width := charW * 2
	height := charH * 4

	var srcList []string
	var tgtList []string
	srcTotal := make(map[string]float64)
	tgtTotal := make(map[string]float64)

	for _, f := range flows {
		if _, ok := srcTotal[f.Source]; !ok {
			srcList = append(srcList, f.Source)
		}
		if _, ok := tgtTotal[f.Target]; !ok {
			tgtList = append(tgtList, f.Target)
		}
		srcTotal[f.Source] += f.Value
		tgtTotal[f.Target] += f.Value
	}

	sort.Strings(srcList)
	sort.Strings(tgtList)

	totalSrcVal := 0.0
	for _, v := range srcTotal {
		totalSrcVal += v
	}
	totalTgtVal := 0.0
	for _, v := range tgtTotal {
		totalTgtVal += v
	}

	maxTotalVal := totalSrcVal
	if totalTgtVal > maxTotalVal {
		maxTotalVal = totalTgtVal
	}

	gapPixels := 4.0
	leftGaps := float64(len(srcList) - 1)
	rightGaps := float64(len(tgtList) - 1)
	maxGaps := leftGaps
	if rightGaps > maxGaps {
		maxGaps = rightGaps
	}

	availablePixels := float64(height) - (maxGaps * gapPixels)
	if availablePixels < 1.0 {
		availablePixels = 1.0
	}
	scale := 1.0
	if maxTotalVal > 0 {
		scale = availablePixels / maxTotalVal
	}

	sourceY := make(map[string]float64)
	currentY := 0.0
	for _, src := range srcList {
		sourceY[src] = currentY
		currentY += srcTotal[src]*scale + gapPixels
	}

	targetY := make(map[string]float64)
	currentY = 0.0
	for _, tgt := range tgtList {
		targetY[tgt] = currentY
		currentY += tgtTotal[tgt]*scale + gapPixels
	}

	flowY0Top := make([]float64, len(flows))
	flowY0Bottom := make([]float64, len(flows))
	flowY1Top := make([]float64, len(flows))
	flowY1Bottom := make([]float64, len(flows))

	srcCurrentY := make(map[string]float64)
	tgtCurrentY := make(map[string]float64)
	maps.Copy(srcCurrentY, sourceY)
	maps.Copy(tgtCurrentY, targetY)

	for i, f := range flows {
		thickness := f.Value * scale

		y0 := srcCurrentY[f.Source]
		flowY0Top[i] = y0
		flowY0Bottom[i] = y0 + thickness
		srcCurrentY[f.Source] += thickness

		y1 := tgtCurrentY[f.Target]
		flowY1Top[i] = y1
		flowY1Bottom[i] = y1 + thickness
		tgtCurrentY[f.Target] += thickness
	}

	var sb strings.Builder

	for cy := range charH {
		for cx := range charW {
			dotValues := [8]bool{false, false, false, false, false, false, false, false}
			dotMap := [8][2]int{
				{0, 0},
				{0, 1},
				{0, 2},
				{1, 0},
				{1, 1},
				{1, 2},
				{0, 3},
				{1, 3},
			}

			counts := make([]int, len(flows))

			for i, offset := range dotMap {
				px := cx*2 + offset[0]
				py := cy*4 + offset[1]

				t := float64(px) / float64(width-1)
				if width <= 1 {
					t = 0
				}
				st := smoothstep(t)

				overlaps := false
				for fi := range flows {
					top := flowY0Top[fi] + (flowY1Top[fi]-flowY0Top[fi])*st
					bot := flowY0Bottom[fi] + (flowY1Bottom[fi]-flowY0Bottom[fi])*st
					if float64(py) >= top && float64(py) <= bot {
						counts[fi]++
						overlaps = true
					}
				}

				if overlaps {
					dotValues[i] = true
				}
			}

			runeVal := 0x2800
			brailleOffsets := []int{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
			for i, v := range dotValues {
				if v {
					runeVal |= brailleOffsets[i]
				}
			}

			style := lipgloss.NewStyle()

			// Proportional Color Blending
			var r, g, b float64
			var totalCount int
			for idx := range flows {
				c := counts[idx]
				if c <= 0 {
					continue
				}
				cr, cg, cb, _ := flows[idx].Color.RGBA()
				r += float64(cr) * float64(c)
				g += float64(cg) * float64(c)
				b += float64(cb) * float64(c)
				totalCount += c
			}

			if totalCount > 0 {
				r /= float64(totalCount)
				g /= float64(totalCount)
				b /= float64(totalCount)
				// color.Color RGBA returns 0-65535, we convert back to 0-255
				hexColor := fmt.Sprintf("#%02x%02x%02x", int(r/257.0), int(g/257.0), int(b/257.0))
				style = style.Foreground(lipgloss.Color(hexColor))
			}

			sb.WriteString(style.Render(string(rune(runeVal))))
		}
		if cy < charH-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
