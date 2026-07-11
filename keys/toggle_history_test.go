package keys

import "testing"

// TestToggleHistoryBinding pins the I-12 history-panel binding: ctrl+n by
// default, customizable through ApplyCustomizations, listed in BindingDefs
// so the settings keybindings section shows it.
func TestToggleHistoryBinding(t *testing.T) {
	km := DefaultKeyMap()
	if got := km.ToggleHistory.Keys(); len(got) != 1 || got[0] != "ctrl+n" {
		t.Fatalf("default ToggleHistory keys: %v", got)
	}

	km.ApplyCustomizations(map[string]string{"ToggleHistory": "ctrl+y"})
	if got := km.ToggleHistory.Keys(); len(got) != 1 || got[0] != "ctrl+y" {
		t.Fatalf("customized ToggleHistory keys: %v", got)
	}

	found := false
	for _, def := range km.BindingDefs() {
		if def.ID == "ToggleHistory" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("ToggleHistory missing from BindingDefs")
	}
}
