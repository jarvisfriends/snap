// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package pickers

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	huh "charm.land/huh/v2"

	"github.com/jarvisfriends/snap/uifx"
)

func TestMultiFileEditorValueJoinsPaths(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor(" a ; b ")
	if got := e.Value(); got != "a; b" {
		t.Fatalf("Value() = %q; want %q", got, "a; b")
	}
	if cmd := e.Init(); cmd != nil {
		t.Fatal("Init must return a nil command")
	}
	if NewMultiFileEditor("   ").Value() != "" {
		t.Fatal("blank input must yield no paths")
	}
}

// TestMultiFileEditorKeys drives the list-mode bindings: up/down wrap through
// the rows (including the Add row), Del removes the highlighted path, Ctrl+S
// finishes, and Esc aborts.
func TestMultiFileEditorKeys(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor("a;b")
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Up from row 0 wraps to the Add row; down from there wraps back to 0.
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if e.cursor != 2 {
		t.Fatalf("up from row 0: cursor = %d; want the Add row (2)", e.cursor)
	}
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if e.cursor != 0 {
		t.Fatalf("down from the Add row: cursor = %d; want 0", e.cursor)
	}
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if e.cursor != 1 {
		t.Fatalf("down: cursor = %d; want 1", e.cursor)
	}

	// Del removes the highlighted path.
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	if got := e.Value(); got != "a" {
		t.Fatalf("after delete Value() = %q; want %q", got, "a")
	}

	// Ctrl+S saves, Esc aborts.
	_, _ = e.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if !e.Done {
		t.Fatal("ctrl+s did not finish the editor")
	}
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if !e.Aborted {
		t.Fatal("esc did not abort the editor")
	}
}

// TestMultiFileEditorDirPickerFlow walks the DirsOnly browse cycle end to
// end: Enter opens a DirPicker for the row, Ctrl+S selects the browsed
// directory and replaces the row's path, and a second browse on the Add row
// aborted with Esc leaves the list unchanged.
func TestMultiFileEditorDirPickerFlow(t *testing.T) {
	t.Parallel()

	root := makePickerTree(t)
	alpha := filepath.Join(root, "alpha")
	e := NewMultiFileEditor(alpha)
	e.DirsOnly = true
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Enter on row 0 opens the row's directory picker.
	_, cmd := e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || e.dirPicker == nil {
		t.Fatal("enter must open a DirPicker and return its readDir command")
	}
	// Deliver the listing (as tea.Program would) and confirm the browsed dir.
	_, _ = e.Update(cmd())
	_, _ = e.Update(tea.KeyPressMsg{Text: "ctrl+s"})
	if e.picking || e.dirPicker != nil {
		t.Fatal("ctrl+s must close the picker")
	}
	if got := e.Value(); got != alpha {
		t.Fatalf("Value() = %q; want the picked dir %q", got, alpha)
	}

	// Browse from the Add row, then abort: no path is added.
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyUp}) // wraps to the Add row
	_, cmd = e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || e.dirPicker == nil {
		t.Fatal("enter on the Add row must open a DirPicker")
	}
	_, _ = e.Update(cmd())
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if e.picking || e.dirPicker != nil {
		t.Fatal("esc must close the picker")
	}
	if got := e.Value(); got != alpha {
		t.Fatalf("aborted browse changed Value() to %q", got)
	}
}

// TestMultiFileEditorAddAppends: a pick confirmed from the Add row appends
// instead of replacing, and an empty pick result is ignored.
func TestMultiFileEditorAddAppends(t *testing.T) {
	t.Parallel()

	root := makePickerTree(t)
	e := NewMultiFileEditor(root)
	e.DirsOnly = true
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyUp}) // Add row
	_, cmd := e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on the Add row must open a picker")
	}
	_, _ = e.Update(cmd())
	_, _ = e.Update(tea.KeyPressMsg{Text: "ctrl+s"})
	if len(e.paths) != 2 {
		t.Fatalf("paths = %v; want the original plus the appended pick", e.paths)
	}

	// An empty result (defensive: pickers report "" when nothing was chosen)
	// must not append or replace.
	e.pickerIndex = len(e.paths)
	e.applyPickedPath("")
	if len(e.paths) != 2 {
		t.Fatalf("empty pick changed paths: %v", e.paths)
	}
}

// TestMultiFileEditorHuhFormAbort: without DirsOnly the row picker is a huh
// file-picker form; while it is open messages route to it, and its quit key
// closes it without touching the paths.
func TestMultiFileEditorHuhFormAbort(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor("a")
	e.HuhTheme = func() huh.Theme { return huh.ThemeFunc(huh.ThemeBase) }
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, cmd := e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil || e.pickerForm == nil {
		t.Fatal("enter must open the huh picker form")
	}

	// Raw mouse must not reach the form through Update (onMouse is the only
	// mouse door while picking).
	_, mouseCmd := e.Update(tea.MouseClickMsg{X: 1, Y: 1, Button: tea.MouseLeft})
	if mouseCmd != nil {
		t.Fatal("raw mouse while picking must be dropped by Update")
	}

	// The form's view is what the editor shows while picking.
	if e.View().Content == "" {
		t.Fatal("picking view must render the hosted form")
	}

	// The form's quit key aborts it and returns to the list unchanged.
	_, _ = e.Update(tea.KeyPressMsg{Text: "ctrl+c"})
	if e.picking || e.pickerForm != nil {
		t.Fatal("ctrl+c must abort the hosted form")
	}
	if got := e.Value(); got != "a" {
		t.Fatalf("aborted form changed Value() to %q", got)
	}
}

// TestMultiFileEditorMouseEdges covers the pointer paths the happy-path
// mouse test misses: non-left clicks and out-of-list points are ignored,
// wheel-down wraps from the Add row, drag follows the tier contract, and
// while a DirPicker is open onMouse routes to it.
func TestMultiFileEditorMouseEdges(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor("a;b")
	e.Effects = uifx.LevelMedium
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_ = e.View()

	// Non-left click and clicks outside the rows do nothing.
	_ = e.View().OnMouse(tea.MouseClickMsg{X: 1, Y: e.rowsTopY + 1, Button: tea.MouseRight})
	if e.cursor != 0 {
		t.Fatalf("right click moved the cursor to %d", e.cursor)
	}
	_ = e.View().OnMouse(tea.MouseClickMsg{X: 1, Y: e.rowsTopY + 10, Button: tea.MouseLeft})
	if e.cursor != 0 {
		t.Fatalf("click below the list moved the cursor to %d", e.cursor)
	}
	if i := e.rowAt(e.rowsWidth+1, e.rowsTopY); i != -1 {
		t.Fatalf("rowAt right of the list = %d; want -1", i)
	}

	// Wheel down from the Add row wraps to the top.
	e.cursor = len(e.paths)
	_ = e.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if e.cursor != 0 {
		t.Fatalf("wheel down from the Add row: cursor = %d; want 0", e.cursor)
	}

	// Drag moves the highlight at Medium, hover does not track below High.
	_ = e.View().OnMouse(tea.MouseMotionMsg{X: 1, Y: e.rowsTopY + 1, Button: tea.MouseLeft})
	if e.cursor != 1 {
		t.Fatalf("drag at Medium: cursor = %d; want 1", e.cursor)
	}
	_ = e.View().OnMouse(tea.MouseMotionMsg{X: 1, Y: e.rowsTopY, Button: tea.MouseNone})
	if e.hoverRow != -1 {
		t.Fatalf("hover tracked at Medium (hoverRow=%d)", e.hoverRow)
	}

	// Drag is suppressed at Minimal.
	e.Effects = uifx.LevelMinimal
	e.cursor = 0
	_ = e.View().OnMouse(tea.MouseMotionMsg{X: 1, Y: e.rowsTopY + 1, Button: tea.MouseLeft})
	if e.cursor != 0 {
		t.Fatalf("drag at Minimal moved the cursor to %d", e.cursor)
	}

	// While a DirPicker is open, onMouse routes to it (wheel moves its
	// cursor), and a defensive picking state with no child yields nil.
	root := makePickerTree(t)
	e = NewMultiFileEditor(root)
	e.DirsOnly = true
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, cmd := e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter must open the DirPicker")
	}
	_, _ = e.Update(cmd())
	_ = e.View() // record the child picker's geometry
	_ = e.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if e.dirPicker == nil || e.dirPicker.cursor != 1 {
		t.Fatal("wheel while browsing must reach the hosted DirPicker")
	}

	e.dirPicker = nil // picking with no child: nothing to route to
	if got := e.onMouse(tea.MouseWheelMsg{Button: tea.MouseWheelDown}); got != nil {
		t.Fatal("onMouse with no hosted picker must return nil")
	}
}

// TestMultiFileEditorHuhThemeDefault: huhTheme uses the injected hook when
// set and falls back to huh's base theme otherwise.
func TestMultiFileEditorHuhThemeDefault(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor("")
	if e.huhTheme() == nil {
		t.Fatal("default huh theme must not be nil")
	}
	called := false
	e.HuhTheme = func() huh.Theme {
		called = true
		return huh.ThemeFunc(huh.ThemeBase)
	}
	if e.huhTheme() == nil || !called {
		t.Fatal("huhTheme must use the injected hook")
	}
}
