package charts

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// PieDataMsg replaces the slices of the pie whose ID matches.
type PieDataMsg struct {
	ID     string
	Slices []PieSlice
}

// PieModel wraps PieChart / BraillePieChart as a tea.Model. The radius
// stretches to fill the Frame, and slice counts are dynamic: slices too thin
// to be visible at the rendered radius are folded into a single "Other"
// slice — Combined reports what was folded so the host can render a legend
// or drill-down for them.
type PieModel struct {
	ID string
	Frame
	// Braille renders with the higher-resolution braille circle.
	Braille bool
	// MinSliceFrac is the smallest share of the total a slice may hold and
	// still render on its own; thinner slices fold into "Other". Zero means
	// the default (2% of the circle).
	MinSliceFrac float64
	// OtherColor colors the folded "Other" slice (dim gray by default).
	OtherColor color.Color

	slices   []PieSlice
	combined []PieSlice
}

// NewPie returns a pie model consuming messages with the given ID.
func NewPie(id string) *PieModel {
	return &PieModel{ID: id}
}

// Combined returns the slices folded into "Other" during the last View
// (empty when every slice rendered on its own). Hosts use it for legends or
// drill-downs of the long tail.
func (m *PieModel) Combined() []PieSlice { return m.combined }

func (m *PieModel) Init() tea.Cmd { return nil }

func (m *PieModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PieDataMsg:
		if msg.ID == m.ID {
			m.slices = msg.Slices
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

func (m *PieModel) View() tea.View {
	radius := m.radius()
	slices := m.visibleSlices()
	if m.Braille {
		return tea.NewView(m.record(BraillePieChart(slices, radius)))
	}
	return tea.NewView(m.record(PieChart(slices, radius)))
}

// radius fits the circle into the frame: height allows ~2r lines, width
// ~4r cells (terminal cells are half as wide as tall).
func (m *PieModel) radius() int {
	r := min(capOr(m.MaxHeight, 2*defaultPieRadius)/2, capOr(m.MaxWidth, 4*defaultPieRadius)/4)
	return max(r, 1)
}

// visibleSlices folds slices thinner than MinSliceFrac of the total into a
// trailing "Other" slice and records them for Combined.
func (m *PieModel) visibleSlices() []PieSlice {
	m.combined = nil
	minFrac := m.MinSliceFrac
	if minFrac <= 0 {
		minFrac = defaultMinSliceFrac
	}
	total := 0.0
	for _, s := range m.slices {
		total += s.Value
	}
	if total <= 0 {
		return m.slices
	}

	visible := make([]PieSlice, 0, len(m.slices))
	other := 0.0
	for _, s := range m.slices {
		if s.Value/total < minFrac {
			m.combined = append(m.combined, s)
			other += s.Value
			continue
		}
		visible = append(visible, s)
	}
	// A lone straggler stays on its own — folding one slice into an equally
	// thin "Other" would only rename it.
	if len(m.combined) == 1 {
		visible, m.combined = m.slices, nil
		return visible
	}
	if other > 0 {
		c := m.OtherColor
		if c == nil {
			c = lipgloss.Color("240")
		}
		visible = append(visible, PieSlice{Value: other, Color: c, Label: "Other"})
	}
	return visible
}

const (
	defaultPieRadius    = 6
	defaultMinSliceFrac = 0.02
)
