package uifx

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestLevelContract pins the tier semantics components rely on: the zero
// value is Medium (drag yes, hover no), High adds hover and needs AllMotion,
// Minimal suppresses both cosmetic tiers but keeps CellMotion so clicks and
// wheel still arrive.
func TestLevelContract(t *testing.T) {
	t.Parallel()

	var def Level
	if def != LevelMedium {
		t.Fatal("zero value must be LevelMedium")
	}
	cases := []struct {
		l     Level
		hover bool
		drag  bool
		mode  tea.MouseMode
	}{
		{LevelMinimal, false, false, tea.MouseModeCellMotion},
		{LevelMedium, false, true, tea.MouseModeCellMotion},
		{LevelHigh, true, true, tea.MouseModeAllMotion},
	}
	for _, c := range cases {
		if c.l.Hover() != c.hover || c.l.Drag() != c.drag || c.l.MouseMode() != c.mode {
			t.Errorf("%v: hover=%v drag=%v mode=%v; want %v/%v/%v",
				c.l, c.l.Hover(), c.l.Drag(), c.l.MouseMode(), c.hover, c.drag, c.mode)
		}
	}
}

// TestRouteToUpdateForwardsMouseAndCmd pins the OnMouse adapter contract:
// the mouse message reaches Update unchanged and Update's command comes back.
func TestRouteToUpdateForwardsMouseAndCmd(t *testing.T) {
	t.Parallel()

	var got tea.Msg
	want := func() tea.Msg { return nil }
	h := RouteToUpdate(func(msg tea.Msg) (tea.Model, tea.Cmd) {
		got = msg
		return nil, want
	})
	click := tea.MouseClickMsg{X: 3, Y: 4, Button: tea.MouseLeft}
	if cmd := h(click); cmd == nil {
		t.Fatal("RouteToUpdate dropped Update's command")
	}
	if got != tea.MouseMsg(click) {
		t.Fatalf("Update received %#v; want the click", got)
	}
}

func TestLevelString(t *testing.T) {
	t.Parallel()

	for l, want := range map[Level]string{
		LevelMinimal: "minimal",
		LevelMedium:  "medium",
		LevelHigh:    "high",
		Level(99):    "medium", // unknown values read as the default tier
	} {
		if got := l.String(); got != want {
			t.Errorf("Level(%d).String() = %q; want %q", l, got, want)
		}
	}
}
