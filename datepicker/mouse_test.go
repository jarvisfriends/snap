package datepicker

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// clickAt sends a content-relative left click.
func clickAt(m *DatePickerModel, x, y int) {
	_ = m.View().OnMouse(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
}

// cellCenter returns the content coordinates of the cell showing date d,
// using the geometry recorded by the last View.
func cellCenter(t *testing.T, m *DatePickerModel, d time.Time) (x, y int) {
	t.Helper()
	for row, days := range m.dayGrid {
		for col, day := range days {
			if !day.IsZero() && day.Day() == d.Day() && day.Month() == d.Month() {
				return m.gridOffX + col*m.cellW + m.cellW/2,
					m.gridTopY + row*m.cellH
			}
		}
	}
	t.Fatalf("date %v not present in the rendered grid", d)
	return 0, 0
}

// TestViewSetsOnMouse: the component must expose mouse handling to hosts that
// honor tea.View.OnMouse.
func TestViewSetsOnMouse(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	if m.View().OnMouse == nil {
		t.Fatal("datepicker View must set OnMouse")
	}
}

// TestClickMovesHighlightAndClickAgainSelects covers the primary mouse flow:
// clicking a day highlights it; clicking the highlighted day confirms it.
func TestClickMovesHighlightAndClickAgainSelects(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	_ = m.View() // record geometry

	target := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	x, y := cellCenter(t, m, target)
	clickAt(m, x, y)
	if m.Time.Day() != 15 {
		t.Fatalf("click on day 15 moved highlight to %d", m.Time.Day())
	}
	if m.Selected {
		t.Fatal("first click must highlight, not select")
	}

	_ = m.View() // re-render; same month so geometry holds
	x, y = cellCenter(t, m, target)
	clickAt(m, x, y)
	if !m.Selected {
		t.Fatal("clicking the highlighted day must select it")
	}
}

// TestClickOnBlankCellIsNoOp: leading blanks (days outside the month) must
// not move the highlight.
func TestClickOnBlankCellIsNoOp(t *testing.T) {
	t.Parallel()

	// July 2026 starts on Wednesday: row 0, col 0 (Sunday) is blank.
	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	_ = m.View()
	before := m.Time
	clickAt(m, m.gridOffX+m.cellW/2, m.gridTopY)
	if !m.Time.Equal(before) || m.Selected {
		t.Fatalf("blank-cell click changed state (time %v -> %v)", before, m.Time)
	}
}

// TestWheelPagesWeeks: the wheel mirrors up/down week navigation.
func TestWheelPagesWeeks(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.Time.Day() != 16 {
		t.Fatalf("wheel down = day %d; want 16 (one week later)", m.Time.Day())
	}
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.Time.Day() != 9 {
		t.Fatalf("wheel up = day %d; want 9", m.Time.Day())
	}
}

// TestTitleClickFocusesHeaders: the left half of the title focuses the month,
// the right half the year.
func TestTitleClickFocusesHeaders(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	_ = m.View()
	clickAt(m, 1, m.titleH-1)
	if m.Focused != FocusHeaderMonth {
		t.Fatalf("left title click focused %v; want month", m.Focused)
	}
	clickAt(m, m.totalW-2, m.titleH-1)
	if m.Focused != FocusHeaderYear {
		t.Fatalf("right title click focused %v; want year", m.Focused)
	}
}

// TestOnMouseHandlesClick drives a click through View().OnMouse — the
// pointer's only door (component Updates carry no mouse cases) — and checks
// it lands on the intended day.
func TestOnMouseHandlesClick(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC))
	m.Focused = FocusCalendar
	v := m.View()
	target := time.Date(2026, 7, 21, 0, 0, 0, 0, time.UTC)
	x, y := cellCenter(t, m, target)
	_ = v.OnMouse(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	if !m.Time.Equal(target) {
		t.Fatalf("click via OnMouse landed on %v; want %v", m.Time, target)
	}
}
