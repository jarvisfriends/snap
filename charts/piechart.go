package charts

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

// PieSlice represents a single slice in a PieChart.
type PieSlice struct {
	Value float64
	Color color.Color
	Label string // currently unused in rendering the circle, but useful for legends
}

// PieChart renders a text-based pie chart using ANSI background colors.
// radius is the approximate vertical radius in terminal lines.
func PieChart(slices []PieSlice, radius int) string {
	if len(slices) == 0 || radius <= 0 {
		return ""
	}

	total := 0.0
	for _, s := range slices {
		total += s.Value
	}

	if total == 0 {
		return ""
	}

	// Calculate angles (in radians, from 0 to 2pi)
	angles := make([]float64, len(slices))
	currentAngle := 0.0
	for i, s := range slices {
		currentAngle += (s.Value / total) * 2 * 3.1415926535
		angles[i] = currentAngle
	}

	width := radius * 2
	height := radius

	var sb strings.Builder

	for y := range height {
		for x := range width {
			nx := (float64(x) - float64(width)/2.0 + 0.5) / float64(radius)
			ny := (float64(y) - float64(height)/2.0 + 0.5) / float64(height/2)

			// dist logic for a perfect circle
			dist := nx*nx + ny*ny
			if dist > 1.0 {
				sb.WriteString("  ")
				continue
			}

			// atan2 returns -pi to pi.
			// Map to 0 to 2pi, starting from top
			theta := math.Atan2(ny, nx)
			theta += math.Pi / 2.0
			if theta < 0 {
				theta += 2 * math.Pi
			}

			sliceIdx := 0
			for i, a := range angles {
				if theta <= a {
					sliceIdx = i
					break
				}
			}
			if sliceIdx >= len(slices) {
				sliceIdx = len(slices) - 1
			}

			style := lipgloss.NewStyle().Background(slices[sliceIdx].Color)
			sb.WriteString(style.Render("  "))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// BraillePieChart renders a text-based pie chart using Braille characters.
// radius is the approximate vertical radius in terminal lines.
func BraillePieChart(slices []PieSlice, radius int) string {
	if len(slices) == 0 || radius <= 0 {
		return ""
	}

	total := 0.0
	for _, s := range slices {
		total += s.Value
	}

	angles := make([]float64, len(slices))
	currentAngle := 0.0
	for i, s := range slices {
		currentAngle += (s.Value / total) * 2 * math.Pi
		angles[i] = currentAngle
	}

	charW := radius * 2
	charH := radius
	pixelW := charW * 2
	pixelH := charH * 4

	var sb strings.Builder

	for cy := range charH {
		for cx := range charW {
			dotValues := [8]int{-1, -1, -1, -1, -1, -1, -1, -1}
			dotMap := [8][2]int{
				{0, 0},
				{0, 1},
				{0, 2},
				{1, 0},
				{1, 1},
				{1, 2},
				{0, 3},
				{1, 3},
			}

			counts := make([]int, len(slices))
			outsideCount := 0

			for i, offset := range dotMap {
				px := cx*2 + offset[0]
				py := cy*4 + offset[1]

				nx := (float64(px) - float64(pixelW)/2.0 + 0.5) / float64(pixelW/2)
				ny := (float64(py) - float64(pixelH)/2.0 + 0.5) / float64(pixelH/2)

				dist := math.Sqrt(nx*nx + ny*ny)
				if dist > 1.0 {
					outsideCount++
					continue
				}

				theta := math.Atan2(ny, nx)
				theta += math.Pi / 2.0
				if theta < 0 {
					theta += 2 * math.Pi
				}

				sliceIdx := 0
				for j, a := range angles {
					if theta <= a {
						sliceIdx = j
						break
					}
				}
				if sliceIdx >= len(slices) {
					sliceIdx = len(slices) - 1
				}

				dotValues[i] = sliceIdx
				counts[sliceIdx]++
			}

			if outsideCount == 8 {
				sb.WriteString(" ")
				continue
			}

			domSlice := -1
			maxCount := -1
			var domColor color.Color
			for idx := range slices {
				if counts[idx] > maxCount {
					maxCount = counts[idx]
					domSlice = idx
					domColor = slices[idx].Color
				}
			}

			runeVal := 0x2800
			brailleOffsets := []int{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
			for i, v := range dotValues {
				if v == domSlice {
					runeVal |= brailleOffsets[i]
				}
			}

			style := lipgloss.NewStyle()
			if domSlice != -1 {
				style = style.Foreground(domColor)
			}

			sb.WriteString(style.Render(string(rune(runeVal))))
		}
		if cy < charH-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
