// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package navigation

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSidebar_InterfaceAccessors(t *testing.T) {
	t.Parallel()

	s := New()
	if cmd := s.Init(); cmd != nil {
		t.Errorf("Init() = %v; want nil", cmd)
	}
	if s.Dock() != DockLeft {
		t.Errorf("Dock() = %v; want DockLeft", s.Dock())
	}
	if len(s.GetPages()) != 2 {
		t.Errorf("GetPages() = %d pages; want 2 (Home, Settings)", len(s.GetPages()))
	}

	// Active index round-trips through the Navigator setters.
	s.SetActiveIndex(1)
	if s.GetActiveIndex() != 1 {
		t.Errorf("GetActiveIndex() = %d; want 1", s.GetActiveIndex())
	}

	// Focus toggle is a pure visual-state setter.
	s.SetFocused(true)
	if !s.focused {
		t.Error("SetFocused(true) should mark the sidebar focused")
	}

	// A window size updates the reported width/height.
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if s.Height() != 40 {
		t.Errorf("Height() = %d; want 40", s.Height())
	}
	if s.Width() <= 0 {
		t.Errorf("Width() = %d; want a positive expanded width", s.Width())
	}
}

func TestSidebar_CollapsedView(t *testing.T) {
	t.Parallel()

	s := New()
	s.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	// Collapse via the toggle message; width shrinks to the collapsed strip.
	s.Update(CollapseToggleMsg{})
	if got := s.Width(); got != sidebarCollapsedWidth {
		t.Errorf("collapsed Width() = %d; want %d", got, sidebarCollapsedWidth)
	}
	v := s.View() // exercises collapsedView
	if !strings.Contains(v.Content, "≡") {
		t.Errorf("collapsed view should show the expand glyph ≡; got %q", v.Content)
	}

	// Toggling again expands it back to the computed width.
	s.Update(CollapseToggleMsg{})
	if got := s.Width(); got != s.expandedWidth {
		t.Errorf("re-expanded Width() = %d; want %d", got, s.expandedWidth)
	}
}

func TestPageItem_FilterValue(t *testing.T) {
	t.Parallel()

	pi := pageItem{id: "home", title: "Home"}
	if pi.FilterValue() != "Home" {
		t.Errorf("FilterValue() = %q; want %q", pi.FilterValue(), "Home")
	}
}

func TestNavDelegate_Update(t *testing.T) {
	t.Parallel()

	// The delegate holds no per-message state, so Update is always a nil cmd.
	if cmd := (navDelegate{}).Update(nil, nil); cmd != nil {
		t.Errorf("navDelegate.Update = %v; want nil", cmd)
	}
}

func TestMinimalTopNav_InterfaceAccessors(t *testing.T) {
	t.Parallel()

	m := NewMinimalTopNav()
	if cmd := m.Init(); cmd != nil {
		t.Errorf("Init() = %v; want nil", cmd)
	}
	if m.Dock() != DockTop {
		t.Errorf("Dock() = %v; want DockTop", m.Dock())
	}
	if m.Width() != 0 {
		t.Errorf("Width() = %d; want 0 (top nav spans the full width)", m.Width())
	}
	if len(m.GetPages()) != 2 {
		t.Errorf("GetPages() = %d; want 2", len(m.GetPages()))
	}

	m.SetActiveIndex(1)
	if m.GetActiveIndex() != 1 {
		t.Errorf("GetActiveIndex() = %d; want 1", m.GetActiveIndex())
	}

	// Height is derived from the rendered view and is always at least one row.
	if m.Height() < 1 {
		t.Errorf("Height() = %d; want >= 1", m.Height())
	}
}
