package charts

import (
	tea "charm.land/bubbletea/v2"
)

// SankeyDataMsg replaces the flows of the sankey whose ID matches.
type SankeyDataMsg struct {
	ID    string
	Flows []SankeyFlow
}

// SankeyModel wraps BrailleSankeyChart as a tea.Model: the diagram stretches
// to fill the Frame and Used reports the rendered size.
type SankeyModel struct {
	ID string
	Frame

	flows []SankeyFlow
}

// NewSankey returns a sankey model consuming messages with the given ID.
func NewSankey(id string) *SankeyModel {
	return &SankeyModel{ID: id}
}

func (m *SankeyModel) Init() tea.Cmd { return nil }

func (m *SankeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SankeyDataMsg:
		if msg.ID == m.ID {
			m.flows = msg.Flows
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

func (m *SankeyModel) View() tea.View {
	w := capOr(m.MaxWidth, defaultSankeyWidth)
	h := capOr(m.MaxHeight, defaultSankeyHeight)
	return tea.NewView(m.record(BrailleSankeyChart(m.flows, w, h)))
}

const (
	defaultSankeyWidth  = 40
	defaultSankeyHeight = 12
)
