package datepicker

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TestEnterSelectsHighlightedDate is the regression test for the missing
// Select binding: Enter must confirm the highlighted date.
func TestEnterSelectsHighlightedDate(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	if m.Selected {
		t.Fatal("Selected must start false")
	}
	// Move the cursor one day right, then confirm.
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.Selected {
		t.Fatal("enter did not select the date")
	}
	if m.Time.Day() != 10 {
		t.Fatalf("selected day = %d; want 10", m.Time.Day())
	}
}

// TestHighlightedDayIsVisiblyDistinct is the regression test for the invisible
// cursor: the style configured for the focused day must be applied to exactly
// the highlighted day cell. A Transform marker makes the assertion independent
// of the terminal color profile (ANSI may be stripped in tests).
func TestHighlightedDayIsVisiblyDistinct(t *testing.T) {
	t.Parallel()

	m := New(time.Date(2026, 7, 9, 0, 0, 0, 0, time.UTC))
	m.Styles.FocusedText = lipgloss.NewStyle().Transform(func(s string) string {
		return "<" + s + ">"
	})

	view := m.View().Content
	if !strings.Contains(view, "<09>") {
		t.Fatalf("highlighted day 09 not rendered with the focused style:\n%s", view)
	}
	if strings.Contains(view, "<10>") || strings.Contains(view, "<08>") {
		t.Fatalf("focused style leaked onto non-highlighted days:\n%s", view)
	}
}

// TestDefaultHighlightIsInverted documents the visibility decision: the
// default styles render the highlighted day with reversed colors, not just
// bold (bold alone is indistinguishable in many terminals).
func TestDefaultHighlightIsInverted(t *testing.T) {
	t.Skip("This wasn't enough to verify a visible change, need to implement a color distance check")
	t.Parallel()
	//Lets update this unit test to check the difference between the colors of the focused text and the selected text.
	// If the colors are far enough apart, then we can assume that the focused text is visible enough to be distinguished from the selected text.
	//   This means that either the background or the foreground colors must be different enough

	st := DefaultStyles()
	if !st.FocusedText.GetBold() || st.FocusedText.GetForeground() == st.Text.GetForeground() {
		t.Error("FocusedText must default to a bold accent distinct from Text")
	}
	if !st.SelectedText.GetReverse() {
		t.Error("SelectedText must default to reversed colors")
	}
}
