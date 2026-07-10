package charts

import (
	tea "charm.land/bubbletea/v2"
)

// HBarDataMsg sets the percentage of the bar whose ID matches.
type HBarDataMsg struct {
	ID  string
	Pct float64 // 0–100
}

// HBarModel wraps HBar as a tea.Model: the bar stretches to Frame.MaxWidth
// and Used reports the rendered size.
type HBarModel struct {
	ID string
	Frame

	pct float64
}

// NewHBar returns a horizontal bar model consuming messages with the given ID.
func NewHBar(id string) *HBarModel {
	return &HBarModel{ID: id}
}

// Pct returns the current percentage.
func (m *HBarModel) Pct() float64 { return m.pct }

func (m *HBarModel) Init() tea.Cmd { return nil }

func (m *HBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case HBarDataMsg:
		if msg.ID == m.ID {
			m.pct = msg.Pct
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

func (m *HBarModel) View() tea.View {
	width := capOr(m.MaxWidth, defaultHBarWidth)
	return tea.NewView(m.record(HBar(m.pct, width)))
}

const defaultHBarWidth = 20
