package charts

import (
	"fmt"
	"image/color"
	"maps"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
)

type SankeyFlow struct {
	Source string
	Target string
	Value  float64
	Color  color.Color
}

func smoothstep(t float64) float64 {
	if t <= 0.0 {
		return 0.0
	}
	if t >= 1.0 {
		return 1.0
	}
	return t * t * (3.0 - 2.0*t)
}

func BrailleSankeyChart(flows []SankeyFlow, charW, charH int) string {
	width := charW * 2
	height := charH * 4

	var srcList []string
	var tgtList []string
	srcTotal := make(map[string]float64)
	tgtTotal := make(map[string]float64)

	for _, f := range flows {
		if _, ok := srcTotal[f.Source]; !ok {
			srcList = append(srcList, f.Source)
		}
		if _, ok := tgtTotal[f.Target]; !ok {
			tgtList = append(tgtList, f.Target)
		}
		srcTotal[f.Source] += f.Value
		tgtTotal[f.Target] += f.Value
	}

	sort.Strings(srcList)
	sort.Strings(tgtList)

	totalSrcVal := 0.0
	for _, v := range srcTotal {
		totalSrcVal += v
	}
	totalTgtVal := 0.0
	for _, v := range tgtTotal {
		totalTgtVal += v
	}

	maxTotalVal := totalSrcVal
	if totalTgtVal > maxTotalVal {
		maxTotalVal = totalTgtVal
	}

	gapPixels := 4.0
	leftGaps := float64(len(srcList) - 1)
	rightGaps := float64(len(tgtList) - 1)
	maxGaps := leftGaps
	if rightGaps > maxGaps {
		maxGaps = rightGaps
	}

	availablePixels := float64(height) - (maxGaps * gapPixels)
	if availablePixels < 1.0 {
		availablePixels = 1.0
	}
	scale := 1.0
	if maxTotalVal > 0 {
		scale = availablePixels / maxTotalVal
	}

	sourceY := make(map[string]float64)
	currentY := 0.0
	for _, src := range srcList {
		sourceY[src] = currentY
		currentY += srcTotal[src]*scale + gapPixels
	}

	targetY := make(map[string]float64)
	currentY = 0.0
	for _, tgt := range tgtList {
		targetY[tgt] = currentY
		currentY += tgtTotal[tgt]*scale + gapPixels
	}

	flowY0Top := make([]float64, len(flows))
	flowY0Bottom := make([]float64, len(flows))
	flowY1Top := make([]float64, len(flows))
	flowY1Bottom := make([]float64, len(flows))

	srcCurrentY := make(map[string]float64)
	tgtCurrentY := make(map[string]float64)
	maps.Copy(srcCurrentY, sourceY)
	maps.Copy(tgtCurrentY, targetY)

	for i, f := range flows {
		thickness := f.Value * scale

		y0 := srcCurrentY[f.Source]
		flowY0Top[i] = y0
		flowY0Bottom[i] = y0 + thickness
		srcCurrentY[f.Source] += thickness

		y1 := tgtCurrentY[f.Target]
		flowY1Top[i] = y1
		flowY1Bottom[i] = y1 + thickness
		tgtCurrentY[f.Target] += thickness
	}

	var sb strings.Builder

	for cy := range charH {
		for cx := range charW {
			dotValues := [8]bool{false, false, false, false, false, false, false, false}
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

			counts := make([]int, len(flows))

			for i, offset := range dotMap {
				px := cx*2 + offset[0]
				py := cy*4 + offset[1]

				t := float64(px) / float64(width-1)
				if width <= 1 {
					t = 0
				}
				st := smoothstep(t)

				overlaps := false
				for fi := range flows {
					top := flowY0Top[fi] + (flowY1Top[fi]-flowY0Top[fi])*st
					bot := flowY0Bottom[fi] + (flowY1Bottom[fi]-flowY0Bottom[fi])*st
					if float64(py) >= top && float64(py) <= bot {
						counts[fi]++
						overlaps = true
					}
				}

				if overlaps {
					dotValues[i] = true
				}
			}

			runeVal := 0x2800
			brailleOffsets := []int{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
			for i, v := range dotValues {
				if v {
					runeVal |= brailleOffsets[i]
				}
			}

			style := lipgloss.NewStyle()

			// Proportional Color Blending
			var r, g, b float64
			var totalCount int
			for idx := range flows {
				c := counts[idx]
				if c <= 0 {
					continue
				}
				cr, cg, cb, _ := flows[idx].Color.RGBA()
				r += float64(cr) * float64(c)
				g += float64(cg) * float64(c)
				b += float64(cb) * float64(c)
				totalCount += c
			}

			if totalCount > 0 {
				r /= float64(totalCount)
				g /= float64(totalCount)
				b /= float64(totalCount)
				// color.Color RGBA returns 0-65535, we convert back to 0-255
				hexColor := fmt.Sprintf("#%02x%02x%02x", int(r/257.0), int(g/257.0), int(b/257.0))
				style = style.Foreground(lipgloss.Color(hexColor))
			}

			sb.WriteString(style.Render(string(rune(runeVal))))
		}
		if cy < charH-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
