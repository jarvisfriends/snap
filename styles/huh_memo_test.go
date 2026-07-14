package styles

import (
	"testing"

	tint "github.com/lrstanley/bubbletint/v2"
)

// TestHuhThemeFuncMemoized verifies HuhThemeFunc returns a stable, shared
// *huh.Styles across calls for the same theme (so huh's per-option render loop
// does not pay a full struct copy each call), and a fresh one after the theme
// changes.
func TestHuhThemeFuncMemoized(t *testing.T) {
	tint.NewDefaultRegistry()
	fn := HuhThemeFunc()

	if err := SetCurrentTint("dracula"); err != nil {
		t.Fatalf("SetCurrentTint(dracula): %v", err)
	}
	SetThemePreferences(ThemeModeDark, false, DefaultStylePreset)

	a1 := fn(true)
	a2 := fn(true)
	if a1 != a2 {
		t.Errorf("expected same styles pointer for unchanged theme; got %p and %p", a1, a2)
	}

	// Changing the tint must invalidate the cache and yield a different pointer.
	if err := SetCurrentTint("nord"); err != nil {
		t.Fatalf("SetCurrentTint(nord): %v", err)
	}
	b1 := fn(true)
	if b1 == a1 {
		t.Error("expected a new styles pointer after theme change")
	}
	if got := fn(true); got != b1 {
		t.Errorf("expected memoized pointer after theme settled; got %p want %p", got, b1)
	}
}
