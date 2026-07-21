package navigation

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

type Page struct {
	ID    string
	Title string
}

// EnsureSettingsLast takes a list of pages and returns a new list where any
// page with ID "settings" is moved to the end.
func EnsureSettingsLast(pages []Page) []Page {
	var normal []Page
	var settings []Page
	for _, p := range pages {
		if strings.EqualFold(p.ID, "settings") {
			settings = append(settings, p)
		} else {
			normal = append(normal, p)
		}
	}
	return append(normal, settings...)
}

// Side indicates where a navigation component docks relative to the page content.
type Side int

const (
	// DockLeft reserves columns on the left (e.g. a sidebar); the page renders
	// to its right via JoinHorizontal.
	DockLeft Side = iota
	// DockTop reserves rows at the top (e.g. a tab bar); the page renders below
	// it via JoinVertical.
	DockTop
)

type Navigator interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() tea.View
	Width() int
	Height() int
	// Dock reports which edge the navigator occupies, letting the router lay it
	// out and route input without type-asserting concrete navigator types.
	Dock() Side
	GetPages() []Page
	SetPages([]Page)
	SetActiveIndex(int)
	GetActiveIndex() int
}

// Focusable is implemented by navigators that support keyboard focus (the
// sidebar). Navigators with no focus concept (tabs) omit it; the router uses a
// capability check rather than a concrete-type assertion to drive focus.
type Focusable interface {
	SetFocused(bool)
}

// horizontalWheelDelta maps a horizontal mouse-wheel event (a tilt wheel, or
// the shift+wheel chord most terminals emit for horizontal scrolling) to a
// page step: -1 for left/previous, +1 for right/next, and 0 for anything that
// is not a horizontal scroll.
func horizontalWheelDelta(mm tea.MouseMsg) int {
	ev, ok := mm.(tea.MouseWheelMsg)
	if !ok {
		return 0
	}
	me := ev.Mouse()
	switch me.Button {
	case tea.MouseWheelLeft:
		return -1
	case tea.MouseWheelRight:
		return 1
	case tea.MouseWheelUp:
		if me.Mod&tea.ModShift != 0 {
			return -1
		}
	case tea.MouseWheelDown:
		if me.Mod&tea.ModShift != 0 {
			return 1
		}
	default:
		return 0
	}
	return 0
}

// verticalWheelDelta is horizontalWheelDelta's twin for vertically laid out
// navigators (the sidebar): -1 for wheel up/previous, +1 for wheel
// down/next, 0 for anything that is not a plain vertical scroll.
func verticalWheelDelta(mm tea.MouseMsg) int {
	ev, ok := mm.(tea.MouseWheelMsg)
	if !ok {
		return 0
	}
	switch ev.Mouse().Button {
	case tea.MouseWheelUp:
		return -1
	case tea.MouseWheelDown:
		return 1
	default:
		return 0
	}
}

// NumberLabeled is implemented by navigators that can optionally show a leading
// per-item number prefix (the minimal top nav). The router applies the user's
// preference via this capability without asserting a concrete type.
type NumberLabeled interface {
	SetShowNumbers(bool)
}

// SelectedMsg is emitted when a navigation item is selected (via click or key).
type SelectedMsg struct {
	PageIndex int
}

// KeyCapturer can be implemented by page models that need exclusive keyboard
// focus. When CapturesKeys returns true the router will bypass its own global
// key shortcuts (quit, page-cycling) and will not forward key events to the
// navigation component, verifying every keystroke reaches the active page.
type KeyCapturer interface {
	CapturesKeys() bool
}

// NavFocusMsg signals that the sidebar has gained or lost keyboard focus.
// The sidebar emits NavFocusMsg{Focused: true} when clicked and
// NavFocusMsg{Focused: false} when Esc is pressed inside it.
// The router emits NavFocusMsg{Focused: false} when the page-content area is
// clicked so the sidebar's visual focus indicator is updated.
type NavFocusMsg struct{ Focused bool }

// CollapseToggleMsg is emitted when the user clicks the collapse/expand handle
// at the top of the sidebar. The router forwards it to the nav's Update so the
// sidebar can toggle its own collapsed state, then triggers a layout resize.
type CollapseToggleMsg struct{}
