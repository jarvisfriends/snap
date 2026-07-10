package pickers

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/uifx"
)

// pumpDir builds a DirPicker over the fixture tree with its listing loaded
// and geometry recorded, ready for coordinate-based input.
func pumpDir(t *testing.T, effects uifx.Level) *DirPicker {
	t.Helper()
	dp := NewDirPicker(makePickerTree(t))
	dp.Effects = effects
	_, _ = dp.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	cmd := dp.Init()
	if cmd != nil {
		if msg := cmd(); msg != nil {
			_, _ = dp.Update(msg)
		}
	}
	_ = dp.View()
	if len(dp.entries) != 2 {
		t.Fatalf("fixture listing = %v; want the two subdirectories", dp.entries)
	}
	return dp
}

// rowPoint returns coordinates inside visible row i.
func (m *DirPicker) rowPoint(i int) (x, y int) { return 1, m.rowsTopY + (i - m.scrollTop) }

func TestDirPickerViewSetsOnMouse(t *testing.T) {
	t.Parallel()

	if NewDirPicker("").View().OnMouse == nil {
		t.Fatal("DirPicker View must set OnMouse")
	}
}

// TestDirPickerClickHighlightsThenOpens pins the click convention: first
// click moves the highlight, clicking the highlighted row opens it.
func TestDirPickerClickHighlightsThenOpens(t *testing.T) {
	t.Parallel()

	dp := pumpDir(t, uifx.LevelMedium)
	x, y := dp.rowPoint(1)
	cmd := dp.View().OnMouse(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	if dp.cursor != 1 {
		t.Fatalf("click on row 1 moved cursor to %d", dp.cursor)
	}
	if cmd != nil {
		t.Fatal("first click must only highlight, not open")
	}
	cmd = dp.View().OnMouse(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("clicking the highlighted row must open it (readDir command)")
	}
}

// TestDirPickerWheelWalksTheTree: vertical wheel moves the highlight,
// horizontal wheel navigates up (left) and into (right) directories.
func TestDirPickerWheelWalksTheTree(t *testing.T) {
	t.Parallel()

	dp := pumpDir(t, uifx.LevelMedium)
	_ = dp.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if dp.cursor != 1 {
		t.Fatalf("wheel down cursor = %d; want 1", dp.cursor)
	}
	_ = dp.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if dp.cursor != 0 {
		t.Fatalf("wheel up cursor = %d; want 0", dp.cursor)
	}
	if cmd := dp.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelRight}); cmd == nil {
		t.Fatal("wheel right must open the highlighted directory")
	}
	if cmd := dp.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelLeft}); cmd == nil {
		t.Fatal("wheel left must navigate to the parent directory")
	}
}

// TestDirPickerEffectTiers pins the uifx contract: drag follows the pointer
// at Medium but not Minimal; hover tracks only at High.
func TestDirPickerEffectTiers(t *testing.T) {
	t.Parallel()

	// Medium: drag moves the cursor, hover does not track.
	dp := pumpDir(t, uifx.LevelMedium)
	x, y := dp.rowPoint(1)
	_ = dp.View().OnMouse(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseLeft})
	if dp.cursor != 1 {
		t.Fatalf("drag at Medium did not move cursor (got %d)", dp.cursor)
	}
	_ = dp.View().OnMouse(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseNone})
	if dp.hoverRow != -1 {
		t.Fatalf("hover tracked at Medium (hoverRow=%d)", dp.hoverRow)
	}

	// Minimal: drag is cosmetic-only feedback — suppressed.
	dp = pumpDir(t, uifx.LevelMinimal)
	x, y = dp.rowPoint(1)
	_ = dp.View().OnMouse(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseLeft})
	if dp.cursor != 0 {
		t.Fatalf("drag at Minimal moved cursor (got %d)", dp.cursor)
	}

	// High: hover tracks the row under the pointer.
	dp = pumpDir(t, uifx.LevelHigh)
	x, y = dp.rowPoint(1)
	_ = dp.View().OnMouse(tea.MouseMotionMsg{X: x, Y: y, Button: tea.MouseNone})
	if dp.hoverRow != 1 {
		t.Fatalf("hover at High tracked row %d; want 1", dp.hoverRow)
	}
}

// TestMultiFileEditorMouseList pins list-mode mouse behavior: click
// highlights, click-again activates ([ Add Path ] opens a picker), the wheel
// wraps through rows, hover tracks only at High.
func TestMultiFileEditorMouseList(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor("a;b")
	e.Effects = uifx.LevelHigh
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_ = e.View()

	// Click row 1 highlights it.
	cmd := e.View().OnMouse(tea.MouseClickMsg{X: 1, Y: e.rowsTopY + 1, Button: tea.MouseLeft})
	if e.cursor != 1 || cmd != nil {
		t.Fatalf("click row 1: cursor=%d cmd=%v; want 1,nil", e.cursor, cmd)
	}
	// Click it again: activates (opens its picker).
	cmd = e.View().OnMouse(tea.MouseClickMsg{X: 1, Y: e.rowsTopY + 1, Button: tea.MouseLeft})
	if cmd == nil {
		t.Fatal("clicking the highlighted row must activate it")
	}

	// Fresh editor: wheel wraps and hover tracks.
	e = NewMultiFileEditor("a;b")
	e.Effects = uifx.LevelHigh
	_, _ = e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_ = e.View()
	_ = e.View().OnMouse(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if e.cursor != len(e.paths) {
		t.Fatalf("wheel up from row 0 should wrap to the Add row (got %d)", e.cursor)
	}
	_ = e.View().OnMouse(tea.MouseMotionMsg{X: 1, Y: e.rowsTopY, Button: tea.MouseNone})
	if e.hoverRow != 0 {
		t.Fatalf("hover row = %d; want 0", e.hoverRow)
	}
	if e.View().OnMouse == nil {
		t.Fatal("MultiFileEditor View must set OnMouse")
	}
}
