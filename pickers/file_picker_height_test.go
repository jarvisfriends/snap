package pickers

import (
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// TestFilePickerKeyMapSplitsOpenAndSelect guards the enter/space split: with
// huh's defaults both actions were enter, and the picker resolved the overlap
// by submitting — folders could never be browsed into.
func TestFilePickerKeyMapSplitsOpenAndSelect(t *testing.T) {
	t.Parallel()

	km := filePickerKeyMap()

	if slices.Contains(km.FilePicker.Select.Keys(), "enter") {
		t.Fatal(
			"Select must not include enter: enter on a folder would submit it instead of opening it",
		)
	}
	if !slices.Contains(km.FilePicker.Open.Keys(), "enter") {
		t.Fatal("Open must include enter so enter descends into folders")
	}
	if !slices.Contains(km.FilePicker.Select.Keys(), "space") {
		t.Fatal("Select must include space so entries can be chosen")
	}
	// bubbles' filepicker only honors Select for keys that also match Open.
	for _, k := range km.FilePicker.Select.Keys() {
		if !slices.Contains(km.FilePicker.Open.Keys(), k) {
			t.Fatalf("Select key %q must also be in Open, or it can never fire", k)
		}
	}
}

// TestMultiFilePickerEnterBrowsesSpaceSelects drives the real picker: enter
// must keep the user browsing (descend or no-op), and space must select the
// highlighted entry and close the picker.
func TestMultiFilePickerEnterBrowsesSpaceSelects(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor("")

	// press sends the key and then executes returned commands, feeding their
	// messages back into Update — the job tea.Program does in the real app
	// (huh advances its form state via command-produced messages).
	var pump func(cmd tea.Cmd, depth int)
	pump = func(cmd tea.Cmd, depth int) {
		if cmd == nil || depth > 16 {
			return
		}
		msg := cmd()
		if msg == nil {
			return
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				pump(c, depth+1)
			}
			return
		}
		_, next := e.Update(msg)
		pump(next, depth+1)
	}
	press := func(k tea.KeyPressMsg) {
		_, cmd := e.Update(k)
		pump(cmd, 0)
	}

	_, _ = e.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	press(tea.KeyPressMsg{Code: tea.KeyEnter}) // open the picker
	if !e.picking {
		t.Fatal("expected Enter on [ Add Path ] to start picking")
	}

	// Enter on the highlighted entry: descends into a folder or is a no-op on
	// a file — it must never submit the picker.
	press(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !e.picking {
		t.Fatal("enter while browsing submitted the picker; it must open folders instead")
	}

	// Space selects the highlighted entry (this package dir is never empty).
	press(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
	if e.picking {
		t.Fatal("space while browsing did not select the highlighted entry")
	}
	if len(e.paths) == 0 || e.paths[0] == "" {
		t.Fatalf("expected a selected path after space; got %q", e.Value())
	}
}

// TestMultiFilePickerFillsAvailableHeight verifies the multi-file editor's
// per-row picker opens in browse mode sized to the editor's known height.
func TestMultiFilePickerFillsAvailableHeight(t *testing.T) {
	t.Parallel()

	e := NewMultiFileEditor("")
	_, _ = e.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	// Cursor starts on the "[ Add Path ]" row; Enter opens the picker.
	_, _ = e.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if !e.picking || e.pickerForm == nil {
		t.Fatal("expected Enter to start picking")
	}
	got := lipgloss.Height(e.View().Content)
	want := pickerFormHeight(e.Height)
	if got < want {
		t.Fatalf(
			"multi-file picker view height = %d; want at least %d (editor height %d)",
			got, want, e.Height,
		)
	}
}
