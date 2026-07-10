package timepicker

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/uifx"
)

// TestTimeFieldWheelLeftRightSwitchesSides: the horizontal wheel hops
// between the hour and minute columns (closing any open dropdown).
func TestTimeFieldWheelLeftRightSwitchesSides(t *testing.T) {
	t.Parallel()

	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelRight})
	if m.Focused != SideMinutes {
		t.Fatalf("wheel right focused %v; want minutes", m.Focused)
	}
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if m.Focused != SideHours {
		t.Fatalf("wheel left focused %v; want hours", m.Focused)
	}
}

// TestTimeFieldHoverAndDragTiers: hover tracks columns/rows at High; a drag
// over the open dropdown moves its cursor at Medium.
func TestTimeFieldHoverAndDragTiers(t *testing.T) {
	t.Parallel()

	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))
	m.Effects = uifx.LevelHigh
	_ = m.View()
	minutes, ok := m.zones.Bounds(zoneMinutes)
	if !ok {
		t.Fatal("minutes zone not recorded")
	}
	_ = m.View().OnMouse(tea.MouseMotionMsg{
		X: minutes.X + 1, Y: minutes.Y + 1, Button: tea.MouseNone,
	})
	if m.hoverSide != SideMinutes {
		t.Fatalf("hover side = %v; want minutes", m.hoverSide)
	}

	// Open the hours dropdown and drag across a visible row.
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	_ = m.View()
	last := dropdownVisibleRows - 1
	r, ok := m.zones.Bounds(fmt.Sprintf("%s%d", zoneRow, last))
	if !ok {
		t.Fatal("no dropdown row zones recorded")
	}
	_ = m.View().OnMouse(tea.MouseMotionMsg{X: r.X, Y: r.Y, Button: tea.MouseLeft})
	if m.cursor != m.top+last {
		t.Fatalf("drag over last visible row set cursor %d; want %d", m.cursor, m.top+last)
	}
}

// TestDurationPickerWheelLeftRightMovesFocus: the horizontal wheel moves the
// focused segment across h/m/s without changing the value.
func TestDurationPickerWheelLeftRightMovesFocus(t *testing.T) {
	t.Parallel()

	m := New(time.Hour)
	before := m.Duration
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelRight})
	if m.Focused != FieldMinutes {
		t.Fatalf("wheel right focused %v; want minutes", m.Focused)
	}
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if m.Focused != FieldHours {
		t.Fatalf("wheel left focused %v; want hours", m.Focused)
	}
	if m.Duration != before {
		t.Fatal("horizontal wheel must not change the duration")
	}
}

// TestDurationPickerHoverSegmentAtHigh: hovering a segment tracks it only at
// LevelHigh.
func TestDurationPickerHoverSegmentAtHigh(t *testing.T) {
	t.Parallel()

	m := New(time.Hour)
	m.Effects = uifx.LevelHigh
	_ = m.View()
	r, ok := m.zones.Bounds(fmt.Sprintf("%s%d", zoneRow, int(FieldSeconds)))
	if !ok {
		t.Fatal("seconds segment zone not recorded")
	}
	_ = m.View().OnMouse(tea.MouseMotionMsg{X: r.X + 1, Y: r.Y + 1, Button: tea.MouseNone})
	if m.hoverSeg != int(FieldSeconds) {
		t.Fatalf("hover segment = %d; want seconds", m.hoverSeg)
	}

	m2 := New(time.Hour)
	_ = m2.View()
	_ = m2.View().OnMouse(tea.MouseMotionMsg{X: r.X + 1, Y: r.Y + 1, Button: tea.MouseNone})
	if m2.hoverSeg != -1 {
		t.Fatalf("hover tracked below High (seg=%d)", m2.hoverSeg)
	}
}
