// Package uifx defines the shared interaction-effect tiers for snap
// components. The tiers map directly onto Bubble Tea v2 mouse reporting:
//
//   - LevelMinimal — interactions only. Clicks, wheel, and keys work, but
//     purely cosmetic feedback (hover highlights, drag previews) is
//     suppressed and components avoid changing rendered output unless the
//     change is meaningful — the "extreme transport saving" option for
//     remote/serial links.
//   - LevelMedium — the default. Adds feedback that costs nothing extra to
//     receive: wheel scrolling everywhere, and drag tracking while a mouse
//     button is held (cell-motion reporting already delivers those).
//   - LevelHigh — adds hover: components highlight the element under the
//     pointer as it moves. Hover requires the ROOT view to request
//     tea.MouseModeAllMotion (see MouseMode), which is a firehose of motion
//     events — fine on local machines, expensive over thin links.
//
// Components expose an Effects field of this type; hosts pick the tier once
// and set the matching mouse mode on their root view.
package uifx

import tea "charm.land/bubbletea/v2"

// Level selects how much interactive feedback a component renders.
type Level int

const (
	// LevelMedium is the default (zero value): click/wheel/drag feedback.
	LevelMedium Level = iota
	// LevelMinimal suppresses cosmetic-only feedback and extra redraws.
	LevelMinimal
	// LevelHigh adds hover highlighting (needs MouseModeAllMotion at the root).
	LevelHigh
)

// String returns the tier name.
func (l Level) String() string {
	switch l {
	case LevelMinimal:
		return "minimal"
	case LevelHigh:
		return "high"
	default:
		return "medium"
	}
}

// MouseMode returns the root-view mouse mode this tier needs: AllMotion for
// hover, CellMotion otherwise (clicks/wheel/drag all arrive under cell
// motion; even LevelMinimal keeps interactions working).
func (l Level) MouseMode() tea.MouseMode {
	if l == LevelHigh {
		return tea.MouseModeAllMotion
	}
	return tea.MouseModeCellMotion
}

// Hover reports whether hover feedback should render at this tier.
func (l Level) Hover() bool { return l == LevelHigh }

// Drag reports whether drag feedback should render at this tier.
func (l Level) Drag() bool { return l != LevelMinimal }
