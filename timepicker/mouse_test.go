package timepicker

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// TestAllTimepickerViewsSetOnMouse: both models must expose mouse handling to
// hosts that honor tea.View.OnMouse.
func TestAllTimepickerViewsSetOnMouse(t *testing.T) {
	t.Parallel()

	if New(time.Hour).View().OnMouse == nil {
		t.Error("TimePickerModel View must set OnMouse")
	}
	if NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC)).View().OnMouse == nil {
		t.Error("TimeFieldModel View must set OnMouse")
	}
}

// TestSegmentClickFocuses: clicking the minutes cell of the duration picker
// focuses it, so the wheel then adjusts minutes.
func TestSegmentClickFocuses(t *testing.T) {
	t.Parallel()

	m := New(time.Hour)
	_ = m.View() // record segment hit zones
	r, ok := m.zones.Bounds(fmt.Sprintf("%s%d", zoneRow, int(FieldMinutes)))
	if !ok || r.W == 0 {
		t.Fatal("minutes segment zone not recorded during View")
	}
	_ = m.View().OnMouse(tea.MouseClickMsg{X: r.X + r.W/2, Y: r.Y + r.H/2, Button: tea.MouseLeft})
	if m.Focused != FieldMinutes {
		t.Fatalf("segment click focused %v; want minutes", m.Focused)
	}

	before := m.Duration
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.Duration != before+time.Minute {
		t.Fatalf("wheel after focusing minutes changed %v -> %v; want +1m", before, m.Duration)
	}
}

// TestOnMouseOpensDropdown: OnMouse is the pointer's only door — a click
// dispatched through it drives the component directly (no Update involved).
func TestOnMouseOpensDropdown(t *testing.T) {
	t.Parallel()

	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))
	v := m.View()
	hours, ok := m.zones.Bounds(zoneHours)
	if !ok {
		t.Fatal("hours zone not recorded")
	}
	_ = v.OnMouse(tea.MouseClickMsg{
		X: hours.X + 1, Y: hours.Y + 1, Button: tea.MouseLeft,
	})
	if s, ok := m.DropdownOpen(); !ok || s != SideHours {
		t.Fatalf("OnMouse click did not open the hours dropdown (side=%v ok=%v)", s, ok)
	}
}
