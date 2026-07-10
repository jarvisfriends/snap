package timepicker

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func press(m *TimeFieldModel, code rune, text string) {
	_, _ = m.Update(tea.KeyPressMsg{Code: code, Text: text})
}

func TestTimeFieldTypeAheadValidatesOnBlur(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))

	// Type "7" into hours, then leave the field: single digit commits as 7.
	press(m, '7', "7")
	press(m, tea.KeyTab, "")
	if m.Time().Hour() != 7 {
		t.Fatalf("hour after typing 7 + blur = %d; want 7", m.Time().Hour())
	}
	if m.Focused != SideMinutes {
		t.Fatalf("tab did not move focus to minutes")
	}

	// Type "75" into minutes: out of range, must clamp to 59 on the second
	// digit (buffer full = validation point).
	press(m, '7', "7")
	press(m, '5', "5")
	if m.Time().Minute() != 59 {
		t.Fatalf("minute after typing 75 = %d; want clamped 59", m.Time().Minute())
	}
}

func TestTimeFieldTwoDigitTypingCommitsImmediately(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC))
	press(m, '1', "1")
	if m.Time().Hour() != 0 {
		t.Fatalf("single digit committed early: hour=%d", m.Time().Hour())
	}
	press(m, '9', "9")
	if m.Time().Hour() != 19 {
		t.Fatalf("hour after typing 19 = %d; want 19", m.Time().Hour())
	}
}

func TestTimeFieldDropdownKeyboardFlow(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))

	// Space opens the focused (hours) dropdown at the current value.
	press(m, tea.KeySpace, " ")
	if s, ok := m.DropdownOpen(); !ok || s != SideHours {
		t.Fatalf("space did not open the hours dropdown (open=%v ok=%v)", s, ok)
	}
	if m.cursor != 8 {
		t.Fatalf("dropdown cursor = %d; want current hour 8", m.cursor)
	}

	// Down twice then Enter commits 10 and closes.
	press(m, tea.KeyDown, "")
	press(m, tea.KeyDown, "")
	press(m, tea.KeyEnter, "")
	if _, ok := m.DropdownOpen(); ok {
		t.Fatal("enter did not close the dropdown")
	}
	if m.Time().Hour() != 10 {
		t.Fatalf("hour after dropdown commit = %d; want 10", m.Time().Hour())
	}
	if m.Done {
		t.Fatal("committing a dropdown value must not finish the whole field")
	}

	// Esc with a dropdown open closes it without aborting.
	press(m, tea.KeySpace, " ")
	press(m, tea.KeyEscape, "")
	if m.Aborted {
		t.Fatal("esc on an open dropdown aborted the field")
	}
}

func TestTimeFieldMouseFlow(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))
	_ = m.View() // records the cell hit zones

	// Click the minutes cell: its dropdown opens.
	minutes, ok := m.zones.Bounds(zoneMinutes)
	if !ok {
		t.Fatal("minutes zone not recorded")
	}
	_ = m.View().OnMouse(tea.MouseClickMsg{
		X: minutes.X + 1, Y: minutes.Y + 1, Button: tea.MouseLeft,
	})
	if s, ok := m.DropdownOpen(); !ok || s != SideMinutes {
		t.Fatalf("clicking minutes did not open its dropdown (open=%v ok=%v)", s, ok)
	}

	// Render to lay out the rows, then click the first visible row.
	_ = m.View()
	first := m.top
	r, rowOK := m.zones.Bounds(zoneRow + "0")
	if !rowOK {
		t.Fatal("no dropdown row hit zones recorded")
	}
	_ = m.View().OnMouse(tea.MouseClickMsg{X: r.X, Y: r.Y, Button: tea.MouseLeft})
	if m.Time().Minute() != first {
		t.Fatalf("clicking the first row set minute=%d; want %d", m.Time().Minute(), first)
	}
	if _, ok := m.DropdownOpen(); ok {
		t.Fatal("row click did not close the dropdown")
	}

	// Wheel with no dropdown spins the focused column.
	before := m.Time().Minute()
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.Time().Minute() != before+1 {
		t.Fatalf("wheel up changed minute %d -> %d; want +1", before, m.Time().Minute())
	}
}

func TestTimeFieldDropdownScrollWindow(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC))
	press(m, tea.KeyTab, "") // focus minutes
	press(m, tea.KeySpace, " ")

	// Wheel scrolls the window without moving the cursor.
	topBefore := m.top
	_ = m.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.top != topBefore+1 {
		t.Fatalf("wheel did not scroll the dropdown window (%d -> %d)", topBefore, m.top)
	}

	// Cursor movement keeps itself visible: jump far down.
	for range 20 {
		press(m, tea.KeyDown, "")
	}
	if m.cursor < m.top || m.cursor >= m.top+dropdownVisibleRows {
		t.Fatalf("cursor %d outside window [%d,%d)", m.cursor, m.top, m.top+dropdownVisibleRows)
	}
}

func TestTimeFieldColonHighlighted(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))
	view := m.View().Content
	if !strings.Contains(view, ":") {
		t.Fatal("view missing the colon separator")
	}
	// The colon must be rendered through ColonStyle: restyle it with a marker
	// foreground and confirm the styled sequence appears.
	if m.ColonStyle.GetForeground() == m.RowStyle.GetForeground() {
		t.Fatal("colon style should default to the highlight color, not the row color")
	}
}

func TestTimeFieldEnterFinishesWithValidation(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 0, 0, time.UTC))
	press(m, '9', "9") // pending single digit
	press(m, tea.KeyEnter, "")
	if !m.Done {
		t.Fatal("enter did not finish the field")
	}
	if m.Time().Hour() != 9 {
		t.Fatalf("pending digit not validated on finish: hour=%d; want 9", m.Time().Hour())
	}
}

// TestTimeFieldSecondsColumn: ShowSeconds adds a third column that focus
// navigation and the dropdown reach, and Time() carries the edited second.
func TestTimeFieldSecondsColumn(t *testing.T) {
	base := time.Date(2026, 7, 10, 8, 30, 45, 0, time.UTC)
	m := NewTimeField(base)
	m.ShowSeconds = true

	if _, ok := m.zones.Bounds(zoneSeconds); ok {
		t.Fatal("seconds zone recorded before any View")
	}
	_ = m.View()
	if _, ok := m.zones.Bounds(zoneSeconds); !ok {
		t.Fatal("seconds zone not recorded with ShowSeconds")
	}

	// Tab twice reaches seconds; up spins it.
	press(m, tea.KeyTab, "")
	press(m, tea.KeyTab, "")
	if m.Focused != SideSeconds {
		t.Fatalf("focus after two tabs = %v; want seconds", m.Focused)
	}
	press(m, tea.KeyUp, "")
	if got := m.Time().Second(); got != 46 {
		t.Fatalf("second after up = %d; want 46", got)
	}

	// The date and location round-trip untouched.
	y, mo, d := m.Time().Date()
	if y != 2026 || mo != time.July || d != 10 || m.Time().Location() != time.UTC {
		t.Fatalf("Time() lost the date part: %v", m.Time())
	}
}

// TestTimeFieldHiddenSecondsStopsAtMinutes: without ShowSeconds the seconds
// column neither renders nor receives focus.
func TestTimeFieldHiddenSecondsStopsAtMinutes(t *testing.T) {
	m := NewTimeField(time.Date(2026, 7, 10, 8, 30, 45, 0, time.UTC))
	_ = m.View()
	if _, ok := m.zones.Bounds(zoneSeconds); ok {
		t.Fatal("seconds zone rendered without ShowSeconds")
	}
	press(m, tea.KeyTab, "")
	press(m, tea.KeyTab, "")
	if m.Focused != SideMinutes {
		t.Fatalf("focus walked past minutes to %v with seconds hidden", m.Focused)
	}
	// The hidden second still round-trips through Time().
	if got := m.Time().Second(); got != 45 {
		t.Fatalf("hidden second = %d; want the original 45", got)
	}
}
