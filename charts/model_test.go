package charts

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TestSparklineModelRoutesByID pins the ID contract every chart model
// shares: a message with a different ID is ignored, a matching one lands.
func TestSparklineModelRoutesByID(t *testing.T) {
	t.Parallel()

	cpu := NewSparkline("cpu")
	_, _ = cpu.Update(SparklineDataMsg{ID: "mem", Values: []float64{9, 9}})
	if len(cpu.History()) != 0 {
		t.Fatal("sparkline consumed another chart's data")
	}
	_, _ = cpu.Update(SparklineDataMsg{ID: "cpu", Values: []float64{1, 2, 3}})
	if len(cpu.History()) != 3 {
		t.Fatalf("history = %v; want the 3 routed samples", cpu.History())
	}
	_, _ = cpu.Update(SparklinePointMsg{ID: "cpu", Value: 4})
	if len(cpu.History()) != 4 {
		t.Fatal("point message did not append")
	}
	_, _ = cpu.Update(SparklinePointMsg{ID: "mem", Value: 5})
	if len(cpu.History()) != 4 {
		t.Fatal("point message for another ID appended")
	}
}

// TestModelsStretchAndReportUsed: the rendered frame fills the given size
// caps (stretch-to-fill) and Used() reports the actual footprint.
func TestModelsStretchAndReportUsed(t *testing.T) {
	t.Parallel()

	spark := NewSparkline("s")
	_, _ = spark.Update(SparklineDataMsg{ID: "s", Values: []float64{1, 5, 3}})
	spark.SetSize(25, 1)
	frame := ansi.Strip(spark.View().Content)
	if lipgloss.Width(frame) != 25 {
		t.Fatalf("sparkline width = %d; want the 25-cell cap", lipgloss.Width(frame))
	}
	if w, h := spark.Used(); w != 25 || h != 1 {
		t.Fatalf("sparkline Used() = %dx%d; want 25x1", w, h)
	}

	bar := NewHBar("b")
	_, _ = bar.Update(HBarDataMsg{ID: "b", Pct: 50})
	bar.SetSize(30, 1)
	if got := lipgloss.Width(bar.View().Content); got != 30 {
		t.Fatalf("hbar width = %d; want 30", got)
	}

	sankey := NewSankey("k")
	_, _ = sankey.Update(SankeyDataMsg{ID: "k", Flows: []SankeyFlow{
		{Source: "a", Target: "x", Value: 3, Color: lipgloss.Color("1")},
	}})
	sankey.SetSize(30, 8)
	v := sankey.View().Content
	if h := lipgloss.Height(v); h != 8 {
		t.Fatalf("sankey height = %d; want 8", h)
	}
	if w, h := sankey.Used(); w > 30 || h != 8 {
		t.Fatalf("sankey Used() = %dx%d; want within 30x8", w, h)
	}

	line := NewLineChart("l")
	_, _ = line.Update(LineDataMsg{ID: "l", Series: []LineSeries{
		{Data: []float64{1, 2, 3, 2, 5}, Color: lipgloss.Color("2")},
	}})
	line.SetSize(24, 6)
	_ = line.View()
	if w, h := line.Used(); w > 24 || h != 6 {
		t.Fatalf("linechart Used() = %dx%d; want within 24x6", w, h)
	}
	if line.Scale() <= 0 {
		t.Fatal("linechart Scale() not reported after View")
	}
}

// TestModelsHandleWindowSize: a tea.WindowSizeMsg re-caps the frame — the
// chart shrinks with the terminal instead of overflowing it.
func TestModelsHandleWindowSize(t *testing.T) {
	t.Parallel()

	spark := NewSparkline("s")
	_, _ = spark.Update(SparklineDataMsg{ID: "s", Values: []float64{1, 2}})
	_, _ = spark.Update(tea.WindowSizeMsg{Width: 10, Height: 4})
	_ = spark.View()
	if w, _ := spark.Used(); w != 10 {
		t.Fatalf("sparkline width after resize = %d; want 10", w)
	}
}

// TestPieModelFoldsThinSlices: slices under MinSliceFrac fold into a single
// "Other" slice and Combined() reports them for the host's legend.
func TestPieModelFoldsThinSlices(t *testing.T) {
	t.Parallel()

	pie := NewPie("p")
	pie.SetSize(24, 12)
	slices := []PieSlice{
		{Value: 50, Color: lipgloss.Color("1"), Label: "big"},
		{Value: 48, Color: lipgloss.Color("2"), Label: "second"},
		{Value: 0.5, Color: lipgloss.Color("3"), Label: "tiny-a"},
		{Value: 0.5, Color: lipgloss.Color("4"), Label: "tiny-b"},
		{Value: 1, Color: lipgloss.Color("5"), Label: "tiny-c"},
	}
	_, _ = pie.Update(PieDataMsg{ID: "p", Slices: slices})
	if out := pie.View().Content; out == "" {
		t.Fatal("pie rendered empty")
	}
	combined := pie.Combined()
	if len(combined) != 3 {
		t.Fatalf("Combined() = %d slices; want the 3 thin ones", len(combined))
	}
	for _, c := range combined {
		if !strings.HasPrefix(c.Label, "tiny-") {
			t.Fatalf("folded the wrong slice: %+v", c)
		}
	}

	// A single thin slice stays on its own (folding would only rename it).
	_, _ = pie.Update(PieDataMsg{ID: "p", Slices: slices[:3]})
	_ = pie.View()
	if len(pie.Combined()) != 0 {
		t.Fatalf("lone thin slice was folded: %+v", pie.Combined())
	}
}
