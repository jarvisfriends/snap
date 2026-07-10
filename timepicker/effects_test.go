package timepicker

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/uifx"
)

// TestTimeFieldWheelLeftRightSwitchesSides: the horizontal wheel hops
// between the hour and minute columns (closing any open dropdown).
func TestTimeFieldWheelLeftRightSwitchesSides(t *testing.T) {
	t.Parallel()

	m := NewTimeField(8, 30)
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelRight})
	if m.Focused != SideMinutes {
		t.Fatalf("wheel right focused %v; want minutes", m.Focused)
	}
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if m.Focused != SideHours {
		t.Fatalf("wheel left focused %v; want hours", m.Focused)
	}
}

// TestTimeFieldHoverAndDragTiers: hover tracks columns/rows at High; a drag
// over the open dropdown moves its cursor at Medium.
func TestTimeFieldHoverAndDragTiers(t *testing.T) {
	t.Parallel()

	m := NewTimeField(8, 30)
	m.Effects = uifx.LevelHigh
	_ = m.View()
	_, _ = m.Update(tea.MouseMotionMsg{
		X: m.minuteRect.x + 1, Y: m.minuteRect.y + 1, Button: tea.MouseNone,
	})
	if m.hoverSide != SideMinutes {
		t.Fatalf("hover side = %v; want minutes", m.hoverSide)
	}

	// Open the hours dropdown and drag across a visible row.
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	_ = m.View()
	if len(m.rowRects) == 0 {
		t.Fatal("no dropdown rows recorded")
	}
	r := m.rowRects[len(m.rowRects)-1]
	_, _ = m.Update(tea.MouseMotionMsg{X: r.x, Y: r.y, Button: tea.MouseLeft})
	if m.cursor != m.top+len(m.rowRects)-1 {
		t.Fatalf("drag over last visible row set cursor %d; want %d", m.cursor, m.top+len(m.rowRects)-1)
	}
}

// TestDurationPickerWheelLeftRightMovesFocus: the horizontal wheel moves the
// focused segment across h/m/s without changing the value.
func TestDurationPickerWheelLeftRightMovesFocus(t *testing.T) {
	t.Parallel()

	m := New(time.Hour)
	before := m.Duration
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelRight})
	if m.Focused != FieldMinutes {
		t.Fatalf("wheel right focused %v; want minutes", m.Focused)
	}
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
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
	r := m.segRects[FieldSeconds]
	_, _ = m.Update(tea.MouseMotionMsg{X: r.x + 1, Y: r.y + 1, Button: tea.MouseNone})
	if m.hoverSeg != int(FieldSeconds) {
		t.Fatalf("hover segment = %d; want seconds", m.hoverSeg)
	}

	m2 := New(time.Hour)
	_ = m2.View()
	_, _ = m2.Update(tea.MouseMotionMsg{X: r.x + 1, Y: r.y + 1, Button: tea.MouseNone})
	if m2.hoverSeg != -1 {
		t.Fatalf("hover tracked below High (seg=%d)", m2.hoverSeg)
	}
}
