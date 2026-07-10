package charts

import (
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
// Values are normalized against the local min/max so the full range always
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

// AppendHistory appends v to history, keeping at most HistoryLen entries.
func AppendHistory(history []float64, v float64) []float64 {
	history = append(history, v)
	if len(history) > HistoryLen {
		history = history[len(history)-HistoryLen:]
	}
	return history
}
