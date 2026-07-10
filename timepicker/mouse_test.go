package timepicker

import (
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
	if NewTimeField(8, 30).View().OnMouse == nil {
		t.Error("TimeFieldModel View must set OnMouse")
	}
}

// TestSegmentClickFocuses: clicking the minutes cell of the duration picker
// focuses it, so the wheel then adjusts minutes.
func TestSegmentClickFocuses(t *testing.T) {
	t.Parallel()

	m := New(time.Hour)
	_ = m.View() // record segment hit zones
	r := m.segRects[FieldMinutes]
	if r.W == 0 {
		t.Fatal("minutes segment rect not recorded during View")
	}
	_, _ = m.Update(tea.MouseClickMsg{X: r.X + r.W/2, Y: r.Y + r.H/2, Button: tea.MouseLeft})
	if m.Focused != FieldMinutes {
		t.Fatalf("segment click focused %v; want minutes", m.Focused)
	}

	before := m.Duration
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.Duration != before+time.Minute {
		t.Fatalf("wheel after focusing minutes changed %v -> %v; want +1m", before, m.Duration)
	}
}

// TestOnMouseRoutesToUpdate: driving the OnMouse closure must behave exactly
// like sending the message to Update (the TimeField dropdown opens).
func TestOnMouseRoutesToUpdate(t *testing.T) {
	t.Parallel()

	m := NewTimeField(8, 30)
	v := m.View()
	_ = v.OnMouse(tea.MouseClickMsg{
		X: m.hourRect.X + 1, Y: m.hourRect.Y + 1, Button: tea.MouseLeft,
	})
	if s, ok := m.DropdownOpen(); !ok || s != SideHours {
		t.Fatalf("OnMouse click did not open the hours dropdown (side=%v ok=%v)", s, ok)
	}
}
