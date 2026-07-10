package charts

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestPieChartRendersAllSlices(t *testing.T) {
	t.Parallel()

	slices := []PieSlice{
		{Value: 3, Color: lipgloss.Color("1"), Label: "a"},
		{Value: 1, Color: lipgloss.Color("2"), Label: "b"},
	}
	got := PieChart(slices, 4)
	if got == "" {
		t.Fatal("PieChart returned empty output for valid input")
	}
	if _, h := gridSize(got); h < 4 {
		t.Fatalf("PieChart height = %d lines; want at least the radius (4)", h)
	}

	// Degenerate inputs render nothing rather than panicking.
	if PieChart(nil, 4) != "" || PieChart(slices, 0) != "" {
		t.Fatal("PieChart must return empty output for no slices / no radius")
	}
	if PieChart([]PieSlice{{Value: 0}}, 3) != "" {
		t.Fatal("PieChart with zero total must return empty output")
	}
}

func TestBraillePieChartRendersCircle(t *testing.T) {
	t.Parallel()

	slices := []PieSlice{
		{Value: 2, Color: lipgloss.Color("3")},
		{Value: 2, Color: lipgloss.Color("4")},
	}
	got := BraillePieChart(slices, 4)
	if got == "" {
		t.Fatal("BraillePieChart returned empty output for valid input")
	}
	if !strings.ContainsFunc(ansi.Strip(got), func(r rune) bool {
		return r >= 0x2800 && r <= 0x28FF
	}) {
		t.Fatal("BraillePieChart output contains no braille glyphs")
	}
	if BraillePieChart(nil, 4) != "" || BraillePieChart(slices, 0) != "" {
		t.Fatal("BraillePieChart must return empty output for degenerate input")
	}
}
