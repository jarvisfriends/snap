package charts

import (
	"charm.land/lipgloss/v2"
)

// Frame is the sizing contract shared by the chart models: the host caps the
// space a chart may use (via SetSize or a tea.WindowSizeMsg), the chart
// stretches its drawing to fill that space, and View records what was
// actually used — so flexible layouts can pack charts, rolling ones that
// don't fit onto the next line as the terminal resizes.
type Frame struct {
	// MaxWidth / MaxHeight cap the rendered size in cells. Zero means "use
	// the chart's natural default" for that axis.
	MaxWidth  int
	MaxHeight int

	usedW, usedH int
}

// SetSize sets both caps at once — the natural call for hosts splitting a
// window between several charts.
func (f *Frame) SetSize(w, h int) { f.MaxWidth, f.MaxHeight = w, h }

// Used reports the size of the most recently rendered frame. Zero until the
// first View.
func (f *Frame) Used() (w, h int) { return f.usedW, f.usedH }

// record measures a rendered frame into Used and passes it through.
func (f *Frame) record(frame string) string {
	f.usedW = lipgloss.Width(frame)
	f.usedH = lipgloss.Height(frame)
	return frame
}

// capOr returns the limit when set, def otherwise.
func capOr(limit, def int) int {
	if limit > 0 {
		return limit
	}
	return def
}
