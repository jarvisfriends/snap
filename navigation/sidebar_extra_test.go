// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package navigation

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// newSizedSidebar builds a sidebar with the standard four-page fixture at the
// given terminal size, rendered once so mouse geometry is live.
func newSizedSidebar(t *testing.T, w, h int) *Sidebar {
	t.Helper()
	m := New()
	m.SetPages([]Page{
		{ID: pageIDHome, Title: pageHome},
		{ID: "p1", Title: "Placeholder 1"},
		{ID: "p2", Title: "Placeholder 2"},
		{ID: pageIDSettings, Title: pageSettings},
	})
	_, _ = m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	_ = m.View()
	return m
}

// TestSidebarKeySelectAndDismiss: Select re-emits the active page and esc
// releases keyboard focus back to the page (with a NavFocusMsg).
func TestSidebarKeySelectAndDismiss(t *testing.T) {
	t.Parallel()
	m := newSizedSidebar(t, 40, 40)
	_, _ = m.Update(NavFocusMsg{Focused: true})

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if sel, ok := selectedFrom(cmd); !ok || sel.PageIndex != 0 {
		t.Fatalf("select did not emit the active page (cmd=%v)", cmd)
	}

	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("dismiss returned no command")
	}
	if msg, ok := cmd().(NavFocusMsg); !ok || msg.Focused {
		t.Fatalf("dismiss emitted %v; want NavFocusMsg{Focused: false}", cmd())
	}
	if m.focused {
		t.Fatal("dismiss must blur the sidebar")
	}
}

// TestSidebarCollapseCycle: CollapseToggleMsg shrinks the sidebar to the
// 3-column strip (which window resizes preserve), the strip's mouse release
// re-expands it, and other mouse events are ignored.
func TestSidebarCollapseCycle(t *testing.T) {
	t.Parallel()
	m := newSizedSidebar(t, 40, 40)
	expandedW := m.Width()

	_, _ = m.Update(CollapseToggleMsg{})
	if m.Width() != sidebarCollapsedWidth {
		t.Fatalf("collapsed width = %d; want %d", m.Width(), sidebarCollapsedWidth)
	}
	_, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 30})
	if m.Width() != sidebarCollapsedWidth {
		t.Fatal("a resize must not re-expand a collapsed sidebar")
	}

	v := m.View()
	if !strings.Contains(v.Content, "≡") {
		t.Fatal("collapsed view must show the expand affordance")
	}
	if cmd := v.OnMouse(tea.MouseMotionMsg{X: 1, Y: 1}); cmd != nil {
		t.Fatal("collapsed strip must ignore non-release mouse events")
	}
	cmd := v.OnMouse(tea.MouseReleaseMsg{X: 1, Y: 1, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("clicking the collapsed strip returned no command")
	}
	if _, ok := cmd().(CollapseToggleMsg); !ok {
		t.Fatalf("collapsed strip click emitted %T; want CollapseToggleMsg", cmd())
	}

	_, _ = m.Update(CollapseToggleMsg{})
	if m.Width() != expandedW {
		t.Fatalf("re-expanded width = %d; want %d", m.Width(), expandedW)
	}
}

// TestSidebarMouseZones covers the click regions around the main list: the
// title row collapses, the pinned Settings row selects Settings, and clicks
// that hit no item (spacing rows, rows past the list, non-left buttons,
// non-release events) only focus the sidebar or do nothing.
func TestSidebarMouseZones(t *testing.T) {
	t.Parallel()
	m := newSizedSidebar(t, 40, 40)
	v := m.View()

	// Title row (y=0): collapse.
	cmd := v.OnMouse(tea.MouseReleaseMsg{X: 1, Y: 0, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("title click returned no command")
	}
	if _, ok := cmd().(CollapseToggleMsg); !ok {
		t.Fatalf("title click emitted %T; want CollapseToggleMsg", cmd())
	}

	// Pinned Settings row (sidebarFooterRows from the bottom).
	cmd = v.OnMouse(tea.MouseReleaseMsg{X: 1, Y: 40 - sidebarFooterRows, Button: tea.MouseLeft})
	if sel, ok := selectedFrom(cmd); !ok || sel.PageIndex != m.settingsIdx {
		t.Fatalf("settings click selected %v; want the pinned settings index %d", cmd, m.settingsIdx)
	}

	// A spacing row between items focuses without selecting.
	m.focused = false
	cmd = v.OnMouse(tea.MouseReleaseMsg{X: 1, Y: sidebarHeaderRows + 1, Button: tea.MouseLeft})
	if _, ok := selectedFrom(cmd); ok {
		t.Fatal("a spacing-row click must not select a page")
	}
	if msg, ok := cmd().(NavFocusMsg); !ok || !msg.Focused {
		t.Fatalf("spacing-row click emitted %v; want NavFocusMsg{Focused: true}", cmd())
	}
	if !m.focused {
		t.Fatal("spacing-row click must focus the sidebar")
	}

	// An item-aligned row past the last main item also falls through to focus.
	y := sidebarHeaderRows + (m.numMainItems()+1)*navItemStride
	cmd = v.OnMouse(tea.MouseReleaseMsg{X: 1, Y: y, Button: tea.MouseLeft})
	if _, ok := selectedFrom(cmd); ok {
		t.Fatal("a click past the last item must not select a page")
	}

	// Non-left buttons and non-release events are ignored.
	if cmd := v.OnMouse(tea.MouseReleaseMsg{X: 1, Y: 0, Button: tea.MouseRight}); cmd != nil {
		t.Fatal("right-button release must be ignored")
	}
	if cmd := v.OnMouse(tea.MouseMotionMsg{X: 1, Y: 0}); cmd != nil {
		t.Fatal("mouse motion must be ignored")
	}
}

// TestSidebarActiveMarkerFollowsFocus: the active item renders the focused
// cursor (▶) while the sidebar owns the keyboard and the passive marker (●)
// otherwise.
func TestSidebarActiveMarkerFollowsFocus(t *testing.T) {
	t.Parallel()
	m := newSizedSidebar(t, 40, 40)

	_, _ = m.Update(NavFocusMsg{Focused: true})
	if v := m.View().Content; !strings.Contains(v, "▶") {
		t.Fatal("focused sidebar must mark the active item with ▶")
	}
	_, _ = m.Update(NavFocusMsg{Focused: false})
	v := m.View().Content
	if !strings.Contains(v, "●") {
		t.Fatal("blurred sidebar must mark the active item with ●")
	}
	if strings.Contains(v, "▶") {
		t.Fatal("blurred sidebar must not show the focused cursor")
	}
}
