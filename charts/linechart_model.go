package charts

import (
	tea "charm.land/bubbletea/v2"
)

// LineDataMsg replaces the series of the line chart whose ID matches.
type LineDataMsg struct {
	ID     string
	Series []LineSeries
}

// LineChartModel wraps BrailleLineChart as a tea.Model: the plot stretches
// to fill the Frame, Used reports the rendered size, and Scale exposes the
// vertical scale actually used so hosts can label the axis.
type LineChartModel struct {
	ID string
	Frame
	// MaxVal fixes the vertical scale; <= 0 auto-scales to the data.
	MaxVal float64

	series []LineSeries
	scale  float64
}

// NewLineChart returns a line-chart model consuming messages with the given ID.
func NewLineChart(id string) *LineChartModel {
	return &LineChartModel{ID: id}
}

// Scale returns the vertical scale used by the last View (0 until rendered).
func (m *LineChartModel) Scale() float64 { return m.scale }

func (m *LineChartModel) Init() tea.Cmd { return nil }

func (m *LineChartModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LineDataMsg:
		if msg.ID == m.ID {
			m.series = msg.Series
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

func (m *LineChartModel) View() tea.View {
	w := capOr(m.MaxWidth, defaultLineChartWidth)
	h := capOr(m.MaxHeight, defaultLineChartHeight)
	chart, scale := BrailleLineChart(m.series, w, h, m.MaxVal)
	m.scale = scale
	return tea.NewView(m.record(chart))
}

const (
	defaultLineChartWidth  = 40
	defaultLineChartHeight = 10
)
