// Package charts holds terminal chart primitives:
// sparklines (block and directional braille), horizontal bars, pie charts,
// sankey flows, multi-series braille line charts, and a whole-cell CellCanvas
// with color gradients. All take explicit dimensions and render
// lipgloss-styled strings; colors come from the shared snap/styles palette.
//
// The braille line chart plots through ntcharts' canvas/graph primitives
// (github.com/NimbleMarkets/ntcharts). Apps that need axes, tick labels,
// mouse zones, bar charts, heatmaps, or the candlestick/waveline/streamline
// variants should use ntcharts directly — this package only keeps the chart
// shapes ntcharts doesn't have (pie, sankey, directional sparklines, the
// inline HBar, CellCanvas) plus thin ID-routed tea.Model wrappers.
package charts
