package styles

import (
	"testing"

	tint "github.com/lrstanley/bubbletint/v2"
)

// TestBuiltinTintsRegisterAndParse verifies every built-in theme is well-formed
// (all slots parsed by toTint), dark, and registered into the default registry.
func TestBuiltinTintsRegisterAndParse(t *testing.T) {
	tint.NewDefaultRegistry()
	registerBuiltinTints()

	if len(orderedBuiltinTints) != 7 {
		t.Fatalf("expected 7 built-in themes, got %d", len(orderedBuiltinTints))
	}

	for _, bt := range orderedBuiltinTints {
		if bt.ID == "" || bt.DisplayName == "" {
			t.Errorf("built-in tint missing id/display name: %+v", bt)
		}
		if !bt.Dark {
			t.Errorf("built-in tint %q expected to be dark", bt.ID)
		}
		// Every ANSI slot plus fg/bg must be populated for a full 16-color tint.
		for name, c := range map[string]*tint.Color{
			"fg": bt.Fg, "bg": bt.Bg,
			"black": bt.Black, "red": bt.Red, "green": bt.Green, "yellow": bt.Yellow,
			"blue": bt.Blue, "purple": bt.Purple, "cyan": bt.Cyan, "white": bt.White,
			"bright_black": bt.BrightBlack, "bright_red": bt.BrightRed,
			"bright_green": bt.BrightGreen, "bright_yellow": bt.BrightYellow,
			"bright_blue": bt.BrightBlue, "bright_purple": bt.BrightPurple,
			"bright_cyan": bt.BrightCyan, "bright_white": bt.BrightWhite,
		} {
			if c == nil {
				t.Errorf("built-in tint %q: slot %q is nil", bt.ID, name)
			}
		}

		got, ok := tint.GetTint(bt.ID)
		if !ok {
			t.Errorf("built-in tint %q not registered", bt.ID)
			continue
		}
		if got != bt {
			t.Errorf("registered tint %q is not the built-in instance", bt.ID)
		}
	}
}

// TestBuiltinTintOrdering verifies the built-ins lead ThemeTintIDs in the
// documented display order and that BuiltinTintIDs matches.
func TestBuiltinTintOrdering(t *testing.T) {
	want := []string{"deep_space", "starfleet", "lcars", "romulan", "klingon", "darkula", "earthy"}

	ids := BuiltinTintIDs()
	if len(ids) != len(want) {
		t.Fatalf("BuiltinTintIDs len = %d, want %d", len(ids), len(want))
	}
	for i, id := range want {
		if ids[i] != id {
			t.Errorf("BuiltinTintIDs[%d] = %q, want %q", i, ids[i], id)
		}
	}

	all := ThemeTintIDs()
	if len(all) < len(want) {
		t.Fatalf("ThemeTintIDs len = %d, want >= %d", len(all), len(want))
	}
	for i, id := range want {
		if all[i] != id {
			t.Errorf("ThemeTintIDs[%d] = %q, want built-in %q first", i, all[i], id)
		}
	}
}

// TestBuiltinTintActivates confirms a built-in can be selected and produces a
// palette whose accent/border/success come from the mapped ANSI slots.
func TestBuiltinTintActivates(t *testing.T) {
	if err := SetCurrentTint("deep_space"); err != nil {
		t.Fatalf("SetCurrentTint(deep_space): %v", err)
	}
	SetThemePreferences(ThemeModeDark, false, DefaultStylePreset)

	got := Active()
	// Accent maps from the Purple slot (#7d56f4).
	r, g, b, _ := got.Accent.RGBA()
	if r>>8 != 0x7d || g>>8 != 0x56 || b>>8 != 0xf4 {
		t.Errorf("Deep Space accent = #%02x%02x%02x, want #7d56f4", r>>8, g>>8, b>>8)
	}
}
