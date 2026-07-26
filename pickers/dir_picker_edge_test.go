// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package pickers

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestDirPickerUnreadableDir: a failed listing keeps the previous location
// and surfaces the error in the view instead of rows.
func TestDirPickerUnreadableDir(t *testing.T) {
	t.Parallel()

	root := makePickerTree(t)
	dp := newTestDirPicker(t, root)

	_, _ = dp.Update(dirEntriesMsg{dir: filepath.Join(root, "gone"), err: errors.New("boom")})
	if dp.dir != root {
		t.Fatalf("failed listing changed dir to %q", dp.dir)
	}
	if !strings.Contains(dp.View().Content, "unreadable") {
		t.Fatal("view must surface the read error")
	}
}

// TestDirPickerBackAtRoot: at a filesystem root the parent equals the dir, so
// Back falls through to the drive list — empty off Windows, leaving the
// picker where it is. From the (hypothetical) drive list Back is a no-op.
func TestDirPickerBackAtRoot(t *testing.T) {
	t.Parallel()

	dp := NewDirPicker("/")
	dp.Width, dp.Height = 80, 24
	if cmd := dp.Init(); cmd != nil {
		_, _ = dp.Update(cmd())
	}
	_, cmd := dp.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if cmd != nil {
		t.Fatal("back at / must not read a parent directory")
	}
	if dp.Aborted || dp.Done {
		t.Fatal("back at / must not complete the picker")
	}
	if backCmd := dp.navigateBack(); backCmd != nil {
		t.Fatal("navigateBack at / must return nil (no drives off Windows)")
	}

	dp.dir = "" // drive list (only reachable on Windows, but Back must be safe)
	_, cmd = dp.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if cmd != nil {
		t.Fatal("back at the drive list must be a no-op")
	}
	if backCmd := dp.navigateBack(); backCmd != nil {
		t.Fatal("navigateBack at the drive list must return nil")
	}
}

// TestDirPickerCollapsePathHook: the injected CollapsePath shortens the
// location line; without it the full path renders.
func TestDirPickerCollapsePathHook(t *testing.T) {
	t.Parallel()

	root := makePickerTree(t)
	dp := newTestDirPicker(t, root)
	dp.CollapsePath = func(string) string { return "~" }
	if !strings.Contains(dp.View().Content, "📁 ~") {
		t.Fatal("view must render the collapsed path")
	}
	dp.CollapsePath = nil
	if got := dp.collapse("/x"); got != "/x" {
		t.Fatalf("collapse without a hook = %q; want the path unchanged", got)
	}
}

// TestDirPickerScrollKeepsCursorVisible: with more subdirectories than list
// rows, moving down scrolls the window and moving back up scrolls it back.
func TestDirPickerScrollKeepsCursorVisible(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, d := range []string{"d0", "d1", "d2", "d3", "d4", "d5"} {
		if err := os.Mkdir(filepath.Join(root, d), 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	dp := NewDirPicker(root)
	dp.Width, dp.Height = 80, 11 // listHeight = 3
	if cmd := dp.Init(); cmd != nil {
		_, _ = dp.Update(cmd())
	}

	for range 5 {
		_, _ = dp.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if dp.cursor != 5 || dp.scrollTop != 3 {
		t.Fatalf("cursor=%d scrollTop=%d; want 5,3 after scrolling down", dp.cursor, dp.scrollTop)
	}
	view := dp.View().Content
	if !strings.Contains(view, "d5") || strings.Contains(view, "d0") {
		t.Fatal("view must show the window around the cursor")
	}

	for range 5 {
		_, _ = dp.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	if dp.cursor != 0 || dp.scrollTop != 0 {
		t.Fatalf("cursor=%d scrollTop=%d; want 0,0 after scrolling back", dp.cursor, dp.scrollTop)
	}

	// rowAt honors the scroll window: the y of the first visible row maps to
	// scrollTop, and points beyond the window miss.
	_, _ = dp.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_ = dp.View()
	if i := dp.rowAt(1, dp.rowsTopY); i != dp.scrollTop {
		t.Fatalf("rowAt first visible row = %d; want %d", i, dp.scrollTop)
	}
	if i := dp.rowAt(1, dp.rowsTopY+dp.listHeight()); i != -1 {
		t.Fatalf("rowAt below the window = %d; want -1", i)
	}
	if i := dp.rowAt(1, dp.rowsTopY-1); i != -1 {
		t.Fatalf("rowAt above the rows = %d; want -1", i)
	}
}

// TestDirPickerHideHelp: HideHelp removes the built-in key-hint line.
func TestDirPickerHideHelp(t *testing.T) {
	t.Parallel()

	root := makePickerTree(t)
	dp := newTestDirPicker(t, root)
	if !strings.Contains(dp.View().Content, "Navigate") {
		t.Fatal("help line must render by default")
	}
	dp.HideHelp = true
	if strings.Contains(dp.View().Content, "Navigate") {
		t.Fatal("HideHelp must suppress the help line")
	}
}
