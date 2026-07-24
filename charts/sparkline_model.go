package charts

import (
	tea "charm.land/bubbletea/v2"
)

// SparklineDataMsg replaces the history of the sparkline whose ID matches.
type SparklineDataMsg struct {
	ID     string
	Values []float64
}

// SparklinePointMsg appends one sample to the sparkline whose ID matches
// (the history is trimmed to HistoryLen) — the natural shape for live
// metrics streams.
type SparklinePointMsg struct {
	ID    string
	Value float64
}

// SparklineModel wraps Sparkline as a tea.Model: data arrives as ID-routed
// messages, the width stretches to fill Frame.MaxWidth, and Used reports the
// rendered size. See examples/charts for the canonical multi-chart wiring.
type SparklineModel struct {
	// ID selects which data messages this chart consumes. A host with a
	// single sparkline can leave both IDs empty.
	ID string
	Frame
	Opts SparklineOpts

	history []float64
}

// NewSparkline returns a sparkline model consuming messages with the given ID.
func NewSparkline(id string) *SparklineModel {
	return &SparklineModel{ID: id}
}

// History returns the current samples (mainly for tests).
func (m *SparklineModel) History() []float64 { return m.history }

func (m *SparklineModel) Init() tea.Cmd { return nil }

func (m *SparklineModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SparklineDataMsg:
		if msg.ID == m.ID {
			m.history = msg.Values
		}
	case SparklinePointMsg:
		if msg.ID == m.ID {
			m.history = AppendHistory(m.history, msg.Value)
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

func (m *SparklineModel) View() tea.View {
	width := capOr(m.MaxWidth, defaultSparklineWidth)
	return tea.NewView(m.record(Sparkline(m.history, width, m.Opts)))
}

const defaultSparklineWidth = 40
