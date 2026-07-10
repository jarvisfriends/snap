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

// TestMouseHandlersDispatch pins the OnMouse dispatch contract: each event
// kind reaches only its handler with the unwrapped tea.Mouse, the handler's
// command comes back, and nil handlers (and releases here) fall through to a
// nil command instead of panicking.
func TestMouseHandlersDispatch(t *testing.T) {
	t.Parallel()

	var got []string
	rec := func(kind string) func(tea.Mouse) tea.Cmd {
		return func(m tea.Mouse) tea.Cmd {
			got = append(got, kind)
			return func() tea.Msg { return nil }
		}
	}
	h := MouseHandlers{Click: rec("click"), Wheel: rec("wheel"), Motion: rec("motion")}

	if cmd := h.OnMouse(tea.MouseClickMsg{X: 1, Y: 2, Button: tea.MouseLeft}); cmd == nil {
		t.Fatal("click handler's command was dropped")
	}
	_ = h.OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	_ = h.OnMouse(tea.MouseMotionMsg{X: 3, Y: 4})
	// Release has no handler: must be a silent nil, not a panic.
	if cmd := h.OnMouse(tea.MouseReleaseMsg{Button: tea.MouseLeft}); cmd != nil {
		t.Fatal("nil Release handler must yield a nil command")
	}

	want := []string{"click", "wheel", "motion"}
	if len(got) != len(want) {
		t.Fatalf("handlers called: %v; want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("handlers called: %v; want %v", got, want)
		}
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
