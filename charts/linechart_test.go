package charts

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func TestBrailleLineChartDimensionsAndScale(t *testing.T) {
	rx := LineSeries{Label: "rx", Color: lipgloss.Color("2"), Data: []float64{0, 10, 20, 40, 80}}
	tx := LineSeries{Label: "tx", Color: lipgloss.Color("5"), Data: []float64{5, 5, 5, 5, 5}}

	chart, scale := BrailleLineChart([]LineSeries{rx, tx}, 20, 6, 0)
	require.Equal(t, 80.0, scale, "auto-scale should track the window peak")
	require.Equal(t, 6, lipgloss.Height(chart))
	for line := range strings.SplitSeq(chart, "\n") {
		require.Equal(t, 20, lipgloss.Width(line))
	}
	hasBraille := false
	for _, r := range chart {
		if r > 0x2800 && r <= 0x28FF {
			hasBraille = true
			break
		}
	}
	require.True(t, hasBraille, "braille cells expected")
}

func TestBrailleLineChartFixedScale(t *testing.T) {
	s := LineSeries{Data: []float64{50}}
	_, scale := BrailleLineChart([]LineSeries{s}, 10, 4, 200)
	require.Equal(t, 200.0, scale, "explicit maxVal must win over the data peak")
}

func TestBrailleLineChartEmptyAndDegenerate(t *testing.T) {
	chart, scale := BrailleLineChart(nil, 10, 3, 0)
	require.Equal(t, 1.0, scale)
	require.Equal(t, 3, lipgloss.Height(chart))

	chart, _ = BrailleLineChart([]LineSeries{{Data: nil}}, 10, 3, 0)
	require.Equal(t, 3, lipgloss.Height(chart))

	chart, scale = BrailleLineChart([]LineSeries{{Data: []float64{1}}}, 0, 0, 0)
	require.Empty(t, chart)
	require.Equal(t, 0.0, scale)
}

func TestBrailleLineChartOverlapBlendsWithoutPanic(t *testing.T) {
	a := LineSeries{Color: lipgloss.Color("#ff0000"), Data: make([]float64, 100)}
	b := LineSeries{Color: lipgloss.Color("#00ff00"), Data: make([]float64, 100)}
	for i := range a.Data {
		a.Data[i] = 50
		b.Data[i] = 50 // fully overlapping lines share every cell
	}
	chart, _ := BrailleLineChart([]LineSeries{a, b}, 30, 5, 100)
	require.NotEmpty(t, chart)
	require.Equal(t, 5, lipgloss.Height(chart))
}
