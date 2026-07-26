// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package styles

import (
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	tint "github.com/lrstanley/bubbletint/v2"
)

// TestScrollbarStylesFromTheme: the themed scrollbar takes the frame border
// color for the thumb and the muted slot for the track.
func TestScrollbarStylesFromTheme(t *testing.T) {
	c := Active()
	st := ScrollbarStyles(c)
	if got := st.Thumb.GetForeground(); got != c.Border {
		t.Errorf("thumb fg = %v; want Border %v", got, c.Border)
	}
	if got := st.Track.GetForeground(); got != c.Muted {
		t.Errorf("track fg = %v; want Muted %v", got, c.Muted)
	}
}

// TestVerifyRegistryIdempotent: repeated calls keep the registry live and
// never discard tints already registered (other tests reset the registry, so
// the built-ins are re-asserted first, as ThemeTints does).
func TestVerifyRegistryIdempotent(t *testing.T) {
	tint.NewDefaultRegistry()
	registerBuiltinTints()
	VerifyRegistry()
	VerifyRegistry()
	if len(tint.Tints()) == 0 {
		t.Fatal("registry must hold tints after VerifyRegistry")
	}
	if _, ok := tint.GetTint("deep_space"); !ok {
		t.Fatal("built-in deep_space must be registered")
	}
}

// TestRegisterYAMLTints: themes from a directory land in the global registry;
// a missing directory registers nothing and reports no errors.
func TestRegisterYAMLTints(t *testing.T) {
	dir := t.TempDir()
	writeTheme(t, dir, "ocean.yaml", strings.Replace(
		validTheme, "id: test-ocean", "id: test-ocean-registered", 1,
	))

	n, errs := RegisterYAMLTints(dir)
	if n != 1 || len(errs) != 0 {
		t.Fatalf("RegisterYAMLTints = %d, %v; want 1 tint, no errors", n, errs)
	}
	if _, ok := tint.GetTint("test-ocean-registered"); !ok {
		t.Fatal("registered YAML tint not found in the registry")
	}

	n, errs = RegisterYAMLTints(filepath.Join(dir, "missing"))
	if n != 0 || errs != nil {
		t.Fatalf("missing dir = %d, %v; want 0 tints and no errors", n, errs)
	}
}

// TestBuiltinTintPanicsOnBadSchema: builtinTint guards this file's constants
// — malformed hex is a programming error and must panic, not limp along.
func TestBuiltinTintPanicsOnBadSchema(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("builtinTint must panic on a malformed built-in")
		}
	}()
	_ = builtinTint(yamlTint{ID: "bad", Fg: "not-a-color", Bg: "#000000"})
}

// TestTableHeaderColorsCollisionFallback: when the selection colors collide
// (bg == fg) the header falls back to Bg, and past that to the body text
// color, so the header never renders invisibly.
func TestTableHeaderColorsCollisionFallback(t *testing.T) {
	same := lipgloss.Color("#123456")
	c := &AppStyle{
		Bg: same,
		Styles: &Styles{
			SelectedItem: lipgloss.NewStyle().Background(same).Foreground(same),
			TextOnBg:     lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")),
		},
	}
	bg, fg := tableHeaderColors(c)
	if bg != "#123456" || fg != "#ffffff" {
		t.Fatalf("tableHeaderColors = %q, %q; want #123456 and the TextOnBg fallback #ffffff", bg, fg)
	}
}

// TestResolveTintIDForModeFallbacks: a requested ID that matches the mode is
// kept; an ID of the wrong mode resolves to some tint of the requested mode.
func TestResolveTintIDForModeFallbacks(t *testing.T) {
	tint.NewDefaultRegistry()
	registerBuiltinTints()

	if got := ResolveTintIDForMode("deep_space", ThemeModeDark); got != "deep_space" {
		t.Fatalf("dark deep_space resolved to %q; want it kept", got)
	}
	got := ResolveTintIDForMode("deep_space", ThemeModeLight)
	if got == "deep_space" {
		t.Fatal("a dark tint must not satisfy a light-mode request")
	}
	if resolved, ok := tint.GetTint(got); !ok || resolved.Dark {
		t.Fatalf("light-mode request resolved to %q (dark=%v); want a light tint", got, ok && resolved.Dark)
	}
	if got := ResolveTintIDForMode("", ThemeModeDark); got == "" {
		t.Fatal("empty request must resolve to some dark tint")
	}
}
