package menu

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func testItems() []Item {
	return []Item{
		{ID: "edit", Label: "Edit"},
		{ID: "sep", Label: "Rename", Disabled: true},
		{ID: "delete", Label: "Delete"},
	}
}

func TestOpenSkipsDisabledAndNavWraps(t *testing.T) {
	t.Parallel()

	var m Menu
	m.Open(5, 5, []Item{{ID: "a", Label: "A", Disabled: true}, {ID: "b", Label: "B"}}, 7)
	if got := m.Selected(); got == nil || got.ID != "b" {
		t.Fatalf("Open cursor = %+v; want first enabled item b", got)
	}
	if m.Tag() != 7 {
		t.Fatalf("Tag() = %v; want 7", m.Tag())
	}

	m.Open(0, 0, testItems(), nil)
	m.MoveDown() // skips the disabled middle item
	if got := m.Selected(); got == nil || got.ID != "delete" {
		t.Fatalf("MoveDown landed on %+v; want delete", got)
	}
	m.MoveUp()
	if got := m.Selected(); got == nil || got.ID != "edit" {
		t.Fatalf("MoveUp landed on %+v; want edit", got)
	}
}

func TestRectClampsToTerminal(t *testing.T) {
	t.Parallel()

	var m Menu
	m.Open(78, 22, testItems(), nil)
	r := m.Rect(80, 24)
	if r.X+r.W > 80 || r.Y+r.H > 24 {
		t.Fatalf("menu rect %+v overflows an 80x24 terminal", r)
	}
	if r.X < 0 || r.Y < 0 {
		t.Fatalf("menu rect %+v clamped past the origin", r)
	}
}

func TestMouseChoosesAndDismisses(t *testing.T) {
	t.Parallel()

	var m Menu
	m.Open(10, 5, testItems(), "row-3")
	r := m.Rect(80, 24)

	// Click the last item ("delete"): row = border + index 2.
	clickY := r.Y + 1 + 2
	chosen, handled := m.HandleMouse(tea.MouseClickMsg{X: r.X + 2, Y: clickY, Button: tea.MouseLeft}, 80, 24)
	if !handled || chosen == nil || chosen.ID != "delete" {
		t.Fatalf("click chose %+v (handled=%v); want delete", chosen, handled)
	}
	if m.IsOpen() {
		t.Fatal("choosing an item must close the menu")
	}

	// Click outside: dismiss, unhandled (host may act on the click).
	m.Open(10, 5, testItems(), nil)
	chosen, handled = m.HandleMouse(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft}, 80, 24)
	if chosen != nil || handled || m.IsOpen() {
		t.Fatalf("outside click: chosen=%v handled=%v open=%v; want nil,false,false", chosen, handled, m.IsOpen())
	}

	// Clicking a disabled item neither chooses nor closes.
	m.Open(10, 5, testItems(), nil)
	r = m.Rect(80, 24)
	chosen, handled = m.HandleMouse(tea.MouseClickMsg{X: r.X + 2, Y: r.Y + 1 + 1, Button: tea.MouseLeft}, 80, 24)
	if chosen != nil || !handled || !m.IsOpen() {
		t.Fatalf("disabled click: chosen=%v handled=%v open=%v; want nil,true,true", chosen, handled, m.IsOpen())
	}

	// Hover moves the cursor; wheel navigates.
	_, _ = m.HandleMouse(tea.MouseMotionMsg{X: r.X + 2, Y: r.Y + 1 + 2}, 80, 24)
	if got := m.Selected(); got == nil || got.ID != "delete" {
		t.Fatalf("hover moved cursor to %+v; want delete", got)
	}
	_, _ = m.HandleMouse(tea.MouseWheelMsg{Button: tea.MouseWheelUp}, 80, 24)
	if got := m.Selected(); got == nil || got.ID != "edit" {
		t.Fatalf("wheel up moved cursor to %+v; want edit", got)
	}

	// A closed menu consumes nothing.
	m.Close()
	if _, handled := m.HandleMouse(tea.MouseClickMsg{X: 1, Y: 1, Button: tea.MouseLeft}, 80, 24); handled {
		t.Fatal("closed menu consumed a mouse event")
	}
}

func TestRenderAndComposite(t *testing.T) {
	t.Parallel()

	var m Menu
	m.Open(4, 2, testItems(), nil)
	frame := ansi.Strip(m.Render())
	for _, want := range []string{"Edit", "Rename", "Delete", "▸"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("rendered menu missing %q:\n%s", want, frame)
		}
	}

	base := strings.TrimRight(strings.Repeat(strings.Repeat("x", 40)+"\n", 12), "\n")
	over := ansi.Strip(m.Composite(base, 40, 12))
	if !strings.Contains(over, "Edit") {
		t.Fatalf("composite lost the menu:\n%s", over)
	}
	if lines := strings.Split(over, "\n"); !strings.HasPrefix(lines[0], "xxxx") {
		t.Fatalf("composite lost the base frame:\n%s", over)
	}
	m.Close()
	if got := m.Composite(base, 40, 12); got != base {
		t.Fatal("closed menu must return the base unchanged")
	}
}
