package menu

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func keyPress(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: string(code)}
}

func specialKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func TestHandleKeyClosedMenuIgnores(t *testing.T) {
	var m Menu
	chosen, handled := m.HandleKey(specialKey(tea.KeyDown))
	if chosen != nil || handled {
		t.Fatalf("closed menu HandleKey = (%v, %v), want (nil, false)", chosen, handled)
	}
}

func TestHandleKeyNavigateAndChoose(t *testing.T) {
	var m Menu
	m.Open(0, 0, []Item{
		{ID: "a", Label: "A"},
		{ID: "b", Label: "B", Disabled: true},
		{ID: "c", Label: "C"},
	}, nil)

	// Down skips the disabled item and lands on "c".
	if _, handled := m.HandleKey(specialKey(tea.KeyDown)); !handled {
		t.Fatal("down not consumed while open")
	}
	chosen, handled := m.HandleKey(specialKey(tea.KeyEnter))
	if !handled || chosen == nil || chosen.ID != "c" {
		t.Fatalf("enter = (%v, %v), want item c consumed", chosen, handled)
	}
	if m.IsOpen() {
		t.Fatal("menu still open after choosing")
	}
}

func TestHandleKeyVimBindingsAndDismiss(t *testing.T) {
	var m Menu
	m.Open(0, 0, []Item{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}, nil)

	m.HandleKey(keyPress('j')) // vim down
	if sel := m.Selected(); sel == nil || sel.ID != "b" {
		t.Fatalf("after j Selected = %v, want b", sel)
	}
	m.HandleKey(keyPress('k')) // vim up
	if sel := m.Selected(); sel == nil || sel.ID != "a" {
		t.Fatalf("after k Selected = %v, want a", sel)
	}

	chosen, handled := m.HandleKey(specialKey(tea.KeyEscape))
	if chosen != nil || !handled || m.IsOpen() {
		t.Fatalf("esc = (%v, %v, open=%v), want (nil, true, closed)", chosen, handled, m.IsOpen())
	}
}

func TestHandleKeyModalSwallowsUnknownKeys(t *testing.T) {
	var m Menu
	m.Open(0, 0, []Item{{ID: "a", Label: "A"}}, nil)
	if _, handled := m.HandleKey(keyPress('x')); !handled {
		t.Fatal("open menu should consume unbound keys (modal)")
	}
	if !m.IsOpen() {
		t.Fatal("unbound key must not close the menu")
	}
}
