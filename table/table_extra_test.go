// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package table

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/styles"
)

// TestWithPageSizeIsFixed: WithPageSize pins the page size so SetSize's
// height-derived value never overrides it.
func TestWithPageSizeIsFixed(t *testing.T) {
	m := New(sampleCols(), WithPageSize(2))
	m.SetRows(sampleRows())
	m.SetSize(60, 40)
	if m.pageSize != 2 {
		t.Fatalf("pageSize = %d; want the fixed 2", m.pageSize)
	}
	if start, end := m.bt.VisibleIndices(); end-start+1 != 2 {
		t.Fatalf("visible rows = %d..%d; want a 2-row page", start, end)
	}
}

// TestHelpBindings: ShortHelp and FullHelp expose the live bindings (the
// help.KeyMap contract), and bindings self-heals from a nil keymap.
func TestHelpBindings(t *testing.T) {
	m := New(sampleCols())
	if got := len(m.ShortHelp()); got != 5 {
		t.Fatalf("ShortHelp returned %d bindings; want 5", got)
	}
	groups := m.FullHelp()
	if len(groups) != 2 || len(groups[0]) != 6 || len(groups[1]) != 4 {
		t.Fatalf("FullHelp shape = %v; want groups of 6 and 4", groups)
	}

	m.keyMap = nil
	if m.bindings() == nil {
		t.Fatal("bindings must restore the defaults from a nil keymap")
	}
}

// TestSelectRowAt mirrors HandleClick's row selection for right-clicks: a
// valid y highlights and returns the row, the header row and out-of-range
// points miss.
func TestSelectRowAt(t *testing.T) {
	m := New(sampleCols(), WithSort(1, false))
	m.SetRows(sampleRows())
	m.SetSize(60, 20)
	_ = m.View(styles.Active(), 0)

	r, ok := m.SelectRowAt(m.dataStartY + 1)
	if !ok || r.Key != "c" {
		t.Fatalf("SelectRowAt second data row = %q ok=%v; want c", r.Key, ok)
	}
	if r, ok := m.SelectedRow(); !ok || r.Key != "c" {
		t.Fatalf("SelectRowAt must move the highlight (got %q ok=%v)", r.Key, ok)
	}
	if _, ok := m.SelectRowAt(m.headerY); ok {
		t.Fatal("the header row must not select")
	}
	if _, ok := m.SelectRowAt(m.dataStartY + 40); ok {
		t.Fatal("a y past the visible rows must not select")
	}
}

// TestCycleSortWalksColumns pins the keyboard sort cycle: asc → desc on the
// current column, then on to the next column, wrapping at the end.
func TestCycleSortWalksColumns(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())

	m.cycleSort() // inactive → column 0 asc
	if !m.sortActive || m.sortCol != 0 || !m.sortAsc {
		t.Fatalf("first cycle: col=%d asc=%v active=%v; want 0/asc", m.sortCol, m.sortAsc, m.sortActive)
	}
	m.cycleSort() // asc → desc
	if m.sortCol != 0 || m.sortAsc {
		t.Fatalf("second cycle: col=%d asc=%v; want 0/desc", m.sortCol, m.sortAsc)
	}
	m.cycleSort() // desc → next column asc
	if m.sortCol != 1 || !m.sortAsc || !m.sortActive {
		t.Fatalf("third cycle: col=%d asc=%v active=%v; want 1/asc", m.sortCol, m.sortAsc, m.sortActive)
	}
	m.cycleSort()
	m.cycleSort() // wraps back to column 0
	if m.sortCol != 0 || !m.sortAsc {
		t.Fatalf("wrap cycle: col=%d asc=%v; want 0/asc", m.sortCol, m.sortAsc)
	}

	empty := New(nil)
	empty.cycleSort() // no columns: must not panic or activate
	if empty.sortActive {
		t.Fatal("cycleSort on an empty table must stay inactive")
	}
}

// TestUpdateIgnoresNonKeyMsgs: only key presses are handled today.
func TestUpdateIgnoresNonKeyMsgs(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())
	if cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24}); cmd != nil {
		t.Fatal("non-key messages must be ignored")
	}
}

// TestFooterFilterReadouts: the footer shows the live filter input while
// typing and the applied-filter hint after blurring.
func TestFooterFilterReadouts(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())
	m.SetSize(60, 20)

	m.Update(keyText("/"))
	m.Update(keyText("a"))
	if v := m.View(styles.Active(), 0); !strings.Contains(v, "/a▏") {
		t.Fatal("footer must echo the focused filter input")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if v := m.View(styles.Active(), 0); !strings.Contains(v, "esc clears") {
		t.Fatal("footer must show the applied filter hint")
	}
}
