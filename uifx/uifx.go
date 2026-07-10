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

// levelNameMedium is the default tier's name — also what unknown Level
// values read as, mirroring how the zero value behaves everywhere else.
const levelNameMedium = "medium"

// String returns the tier name.
func (l Level) String() string {
	switch l {
	case LevelMinimal:
		return "minimal"
	case LevelHigh:
		return "high"
	case LevelMedium:
		return levelNameMedium
	default:
		return levelNameMedium
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

// MouseHandlers dispatches View.OnMouse events to per-kind handlers, so a
// component's mouse logic lives in OnMouse (or functions it calls) and stays
// out of Update entirely. Keeping the two paths separate means hosts deliver
// mouse through exactly one door — Bubble Tea v2 hands the root view's
// OnMouse the message AND delivers it to Update, so a component that reacts
// in both places double-processes every event — and leaves room to run
// pointer handling independently of state updates later. Nil handlers ignore
// that event kind.
type MouseHandlers struct {
	Click   func(tea.Mouse) tea.Cmd
	Release func(tea.Mouse) tea.Cmd
	Wheel   func(tea.Mouse) tea.Cmd
	Motion  func(tea.Mouse) tea.Cmd
}

// OnMouse routes one mouse message; assign it to View.OnMouse.
func (h MouseHandlers) OnMouse(msg tea.MouseMsg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if h.Click != nil {
			return h.Click(msg.Mouse())
		}
	case tea.MouseReleaseMsg:
		if h.Release != nil {
			return h.Release(msg.Mouse())
		}
	case tea.MouseWheelMsg:
		if h.Wheel != nil {
			return h.Wheel(msg.Mouse())
		}
	case tea.MouseMotionMsg:
		if h.Motion != nil {
			return h.Motion(msg.Mouse())
		}
	}
	return nil
}
