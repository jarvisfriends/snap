// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package datepicker

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// updateBase is a mid-month anchor so week/day arithmetic never crosses a
// month boundary unless a test means it to.
var updateBase = time.Date(2026, time.July, 15, 0, 0, 0, 0, time.UTC)

func TestInitReturnsNil(t *testing.T) {
	t.Parallel()

	if cmd := New(updateBase).Init(); cmd != nil {
		t.Fatal("Init must return a nil command")
	}
}

// TestUpdateArrowsPerFocus pins the arrow-key routing table: what each arrow
// does depends entirely on which component is focused, and FocusNone ignores
// them all.
func TestUpdateArrowsPerFocus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		focus     Focus
		key       tea.KeyPressMsg
		wantTime  time.Time
		wantFocus Focus
	}{
		{"up on calendar goes back a week", FocusCalendar, tea.KeyPressMsg{Code: tea.KeyUp}, updateBase.AddDate(0, 0, -7), FocusCalendar},
		{"up on month header goes back a month", FocusHeaderMonth, tea.KeyPressMsg{Code: tea.KeyUp}, updateBase.AddDate(0, -1, 0), FocusHeaderMonth},
		{"up on year header goes back a year", FocusHeaderYear, tea.KeyPressMsg{Code: tea.KeyUp}, updateBase.AddDate(-1, 0, 0), FocusHeaderYear},
		{"up unfocused is ignored", FocusNone, tea.KeyPressMsg{Code: tea.KeyUp}, updateBase, FocusNone},
		{"down on calendar advances a week", FocusCalendar, tea.KeyPressMsg{Code: tea.KeyDown}, updateBase.AddDate(0, 0, 7), FocusCalendar},
		{"down on month header advances a month", FocusHeaderMonth, tea.KeyPressMsg{Code: tea.KeyDown}, updateBase.AddDate(0, 1, 0), FocusHeaderMonth},
		{"down on year header advances a year", FocusHeaderYear, tea.KeyPressMsg{Code: tea.KeyDown}, updateBase.AddDate(1, 0, 0), FocusHeaderYear},
		{"down unfocused is ignored", FocusNone, tea.KeyPressMsg{Code: tea.KeyDown}, updateBase, FocusNone},
		{"left on calendar goes back a day", FocusCalendar, tea.KeyPressMsg{Code: tea.KeyLeft}, updateBase.AddDate(0, 0, -1), FocusCalendar},
		{"left on year header focuses the month", FocusHeaderYear, tea.KeyPressMsg{Code: tea.KeyLeft}, updateBase, FocusHeaderMonth},
		{"left on month header has nowhere to go", FocusHeaderMonth, tea.KeyPressMsg{Code: tea.KeyLeft}, updateBase, FocusHeaderMonth},
		{"left unfocused is ignored", FocusNone, tea.KeyPressMsg{Code: tea.KeyLeft}, updateBase, FocusNone},
		{"right on calendar advances a day", FocusCalendar, tea.KeyPressMsg{Code: tea.KeyRight}, updateBase.AddDate(0, 0, 1), FocusCalendar},
		{"right on month header focuses the year", FocusHeaderMonth, tea.KeyPressMsg{Code: tea.KeyRight}, updateBase, FocusHeaderYear},
		{"right on year header has nowhere to go", FocusHeaderYear, tea.KeyPressMsg{Code: tea.KeyRight}, updateBase, FocusHeaderYear},
		{"right unfocused is ignored", FocusNone, tea.KeyPressMsg{Code: tea.KeyRight}, updateBase, FocusNone},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			m := New(updateBase)
			m.SetFocus(test.focus)
			_, _ = m.Update(test.key)
			if m.Time != test.wantTime {
				t.Errorf("time = %v; want %v", m.Time, test.wantTime)
			}
			if m.Focused != test.wantFocus {
				t.Errorf("focus = %v; want %v", m.Focused, test.wantFocus)
			}
		})
	}
}

// TestUpdateFocusCycling pins tab/shift+tab: tab walks month → year →
// calendar and stops; shift+tab walks back and stops at the month header.
func TestUpdateFocusCycling(t *testing.T) {
	t.Parallel()

	tab := tea.KeyPressMsg{Code: tea.KeyTab}
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}

	tests := []struct {
		name string
		key  tea.KeyPressMsg
		from Focus
		want Focus
	}{
		{"tab month to year", tab, FocusHeaderMonth, FocusHeaderYear},
		{"tab year to calendar", tab, FocusHeaderYear, FocusCalendar},
		{"tab calendar stays", tab, FocusCalendar, FocusCalendar},
		{"tab none stays", tab, FocusNone, FocusNone},
		{"shift+tab year to month", shiftTab, FocusHeaderYear, FocusHeaderMonth},
		{"shift+tab calendar to year", shiftTab, FocusCalendar, FocusHeaderYear},
		{"shift+tab month stays", shiftTab, FocusHeaderMonth, FocusHeaderMonth},
		{"shift+tab none stays", shiftTab, FocusNone, FocusNone},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			m := New(updateBase)
			m.SetFocus(test.from)
			_, _ = m.Update(test.key)
			if m.Focused != test.want {
				t.Errorf("focus = %v; want %v", m.Focused, test.want)
			}
		})
	}
}

// TestUpdateQuitReturnsQuitCmd: the quit binding must surface tea.Quit so the
// standalone example exits.
func TestUpdateQuitReturnsQuitCmd(t *testing.T) {
	t.Parallel()

	m := New(updateBase)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("quit key returned no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("quit key must return tea.Quit")
	}
}

func TestSelectAndUnselectDate(t *testing.T) {
	t.Parallel()

	m := New(updateBase)
	m.SelectDate()
	if !m.Selected {
		t.Fatal("SelectDate did not set Selected")
	}
	m.UnselectDate()
	if m.Selected {
		t.Fatal("UnselectDate did not clear Selected")
	}
}

// TestFocusString covers the generated stringer, including the out-of-range
// fallback.
func TestFocusString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input Focus
		want  string
	}{
		{FocusNone, "FocusNone"},
		{FocusHeaderMonth, "FocusHeaderMonth"},
		{FocusHeaderYear, "FocusHeaderYear"},
		{FocusCalendar, "FocusCalendar"},
		{Focus(42), "Focus(42)"},
	}
	for i, test := range tests {
		if got := test.input.String(); got != test.want {
			t.Errorf("index %d: String() = %q; want %q", i, got, test.want)
		}
	}
}

// TestDayAtBounds pins the grid hit-test edges: points above the grid, left
// of it, and past the last row all miss (zero time).
func TestDayAtBounds(t *testing.T) {
	t.Parallel()

	m := New(updateBase)
	_ = m.View() // record geometry

	if d := m.dayAt(0, 0); !d.IsZero() {
		t.Errorf("dayAt above the grid = %v; want zero", d)
	}
	if d := m.dayAt(-100, m.gridTopY); !d.IsZero() {
		t.Errorf("dayAt left of the grid = %v; want zero", d)
	}
	if d := m.dayAt(m.gridOffX, m.gridTopY+100*m.cellH); !d.IsZero() {
		t.Errorf("dayAt below the grid = %v; want zero", d)
	}

	unrendered := New(updateBase)
	if d := unrendered.dayAt(1, 1); !d.IsZero() {
		t.Errorf("dayAt before any View = %v; want zero", d)
	}
}
