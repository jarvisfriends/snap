package status

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func wheelMsg(up bool) tea.MouseWheelMsg {
	b := tea.MouseWheelDown
	if up {
		b = tea.MouseWheelUp
	}
	return tea.MouseWheelMsg{Button: b}
}

func clickAt(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

func TestHandleMouseClosedModalIgnores(t *testing.T) {
	m := NewInfoModal()
	cmd, handled := m.HandleMouse(clickAt(0, 0))
	if cmd != nil || handled {
		t.Fatalf("closed modal HandleMouse = (%v, %v), want (nil, false)", cmd, handled)
	}
}

func TestHandleMouseWheelScrolls(t *testing.T) {
	m := NewInfoModal()
	m.Open(100, 30)
	// Force scrollable content: many lines beyond the viewport height.
	if m.vp.TotalLineCount() <= m.vp.VisibleLineCount() {
		lines := make([]string, 200)
		for i := range lines {
			lines[i] = "line"
		}
		m.vp.SetContentLines(lines)
	}

	before := m.vp.YOffset()
	if cmd, handled := m.HandleMouse(wheelMsg(false)); !handled || cmd != nil {
		t.Fatal("wheel not consumed by open modal")
	}
	if m.vp.YOffset() <= before {
		t.Fatalf("wheel down did not scroll: offset %d -> %d", before, m.vp.YOffset())
	}
	m.HandleMouse(wheelMsg(true))
	if m.vp.YOffset() != before {
		t.Fatalf("wheel up did not scroll back: %d, want %d", m.vp.YOffset(), before)
	}
}

func TestHandleMouseClickOutsideCloses(t *testing.T) {
	m := NewInfoModal()
	m.Open(100, 30)
	bx, by, _, _ := m.Bounds()

	// A click inside is consumed and keeps the modal open.
	cmd, handled := m.HandleMouse(clickAt(bx+1, by+1))
	if !handled || cmd != nil || !m.IsVisible() {
		t.Fatalf("inside click = (%v, %v, visible=%v), want consumed + open", cmd, handled, m.IsVisible())
	}

	// A click outside closes and emits CloseInfoModalMsg (as Dismiss does).
	cmd, handled = m.HandleMouse(clickAt(0, 0))
	if !handled || cmd == nil || m.IsVisible() {
		t.Fatalf("outside click = (%v, %v, visible=%v), want cmd + closed", cmd, handled, m.IsVisible())
	}
	if _, ok := cmd().(CloseInfoModalMsg); !ok {
		t.Fatalf("outside click cmd produced %T, want CloseInfoModalMsg", cmd())
	}
}
