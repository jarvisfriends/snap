package datepicker

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/uifx"
)

// TestWheelLeftRightPagesMonths: the horizontal wheel pages whole months.
func TestWheelLeftRightPagesMonths(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelRight})
	if m.Time.Month() != time.August {
		t.Fatalf("wheel right month = %v; want August", m.Time.Month())
	}
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelLeft})
	if m.Time.Month() != time.July {
		t.Fatalf("wheel left month = %v; want July", m.Time.Month())
	}
}

// TestDragMovesHighlightByTier: a held-button drag over a day cell moves the
// highlight at LevelMedium, and is suppressed at LevelMinimal.
func TestDragMovesHighlightByTier(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	_ = m.View()
	x, y := cellCenter(t, m, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	_, _ = m.Update(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseLeft})
	if m.Time.Day() != 15 {
		t.Fatalf("drag at Medium: day = %d; want 15", m.Time.Day())
	}

	m2 := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	m2.Effects = uifx.LevelMinimal
	_ = m2.View()
	_, _ = m2.Update(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseLeft})
	if m2.Time.Day() != 9 {
		t.Fatalf("drag at Minimal moved the highlight (day=%d)", m2.Time.Day())
	}
}

// TestHoverTracksOnlyAtHigh: hover motion records the day under the pointer
// at LevelHigh and stays untracked below it.
func TestHoverTracksOnlyAtHigh(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	m.Effects = uifx.LevelHigh
	_ = m.View()
	x, y := cellCenter(t, m, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))
	_, _ = m.Update(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseNone})
	if m.hoverDay.Day() != 15 {
		t.Fatalf("hover at High tracked day %v; want 15", m.hoverDay)
	}
	if m.Time.Day() != 9 {
		t.Fatal("hover must not move the highlight")
	}

	m2 := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	_ = m2.View()
	_, _ = m2.Update(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseNone})
	if !m2.hoverDay.IsZero() {
		t.Fatalf("hover tracked below High: %v", m2.hoverDay)
	}
}
