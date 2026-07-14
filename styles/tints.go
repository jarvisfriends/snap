package styles

import (
	tint "github.com/lrstanley/bubbletint/v2"
)

// This file defines snap's own built-in color themes: the seven schemes ported
// from the tribble TUI (its eight minus the plain "Light" theme). They register
// into the bubbletint default registry alongside the library's tints (see
// verifyRegistryUnsafe) and, via BuiltinTints/BuiltinTintIDs, provide a stable
// display order that puts them first in any theme picker — mirroring the
// orderedPresets pattern in presets.go.
//
// The tribble themes are semantic palettes (title, border, success, …). Each
// role is mapped onto the exact bubbletint slot that [fromTint] reads for the
// same role, so the ported look survives the round-trip:
//
//	Accent  -> Purple        Muted    -> BrightBlack
//	Border  -> BrightPurple  StatusBg -> Black
//	Success -> Green         Error    -> Red        Warning -> Yellow
//	Selection background -> SelectionBg
//
// The remaining ANSI slots (cyan, white, and the bright variants) are filled
// with hues drawn from each theme's own accent palette so a full 16-color tint
// is available for any consumer that wants it.

// builtinTint constructs a tint from the compact hex schema used by yamlTint,
// panicking on malformed input. Every value here is a compile-time constant
// authored in this file, so a panic means a typo in this file — never a runtime
// condition a caller could hit.
func builtinTint(y yamlTint) *tint.Tint {
	t, err := y.toTint()
	if err != nil {
		panic("styles: invalid built-in tint " + y.ID + ": " + err.Error())
	}
	return t
}

// The seven built-in themes, in display order. Dark is derived from the
// background luminance by toTint (all seven are dark), so it is left unset.
var (
	// TintDeepSpace is the default violet-on-charcoal theme.
	TintDeepSpace = builtinTint(yamlTint{
		ID: "deep_space", DisplayName: "Deep Space",
		Fg: "#fafafa", Bg: "#1d2021",
		SelectionBg: "#7d56f4", Cursor: "#d75fd7",
		Black: "#282828", Red: "#d70000", Green: "#00d787", Yellow: "#ffff00",
		Blue: "#00afff", Purple: "#7d56f4", Cyan: "#56b6c2", White: "#abb2bf",
		BrightBlack: "#585858", BrightRed: "#e06c75", BrightGreen: "#98c379", BrightYellow: "#e5c07b",
		BrightBlue: "#61afef", BrightPurple: "#5f5fff", BrightCyan: "#9fb8e0", BrightWhite: "#ffffff",
	})

	// TintStarfleet is a gold-on-deep-blue LCARS-adjacent bridge theme.
	TintStarfleet = builtinTint(yamlTint{
		ID: "starfleet", DisplayName: "Starfleet",
		Fg: "#c0c0e0", Bg: "#12122a",
		SelectionBg: "#003366", Cursor: "#ffd700",
		Black: "#0d0d1a", Red: "#ff0000", Green: "#00d787", Yellow: "#ffff00",
		Blue: "#00afff", Purple: "#ffd700", Cyan: "#5fafff", White: "#c0c0e0",
		BrightBlack: "#6c6c6c", BrightRed: "#e06c75", BrightGreen: "#98c379", BrightYellow: "#ffd700",
		BrightBlue: "#87d7ff", BrightPurple: "#0087ff", BrightCyan: "#9fb8e0", BrightWhite: "#ffffff",
	})

	// TintLCARS is the amber/lavender Okudagram palette on pure black.
	TintLCARS = builtinTint(yamlTint{
		ID: "lcars", DisplayName: "LCARS",
		Fg: "#ff9900", Bg: "#000000",
		SelectionBg: "#cc99cc", Cursor: "#cc99cc",
		Black: "#000000", Red: "#cc6666", Green: "#99cc66", Yellow: "#ffc66d",
		Blue: "#9999ff", Purple: "#ff9900", Cyan: "#cc6699", White: "#ff9900",
		BrightBlack: "#996633", BrightRed: "#ff6666", BrightGreen: "#99cc66", BrightYellow: "#ffc66d",
		BrightBlue: "#9999ff", BrightPurple: "#cc99cc", BrightCyan: "#cc99cc", BrightWhite: "#ffffff",
	})

	// TintRomulan is a green-on-near-black theme.
	TintRomulan = builtinTint(yamlTint{
		ID: "romulan", DisplayName: "Romulan",
		Fg: "#8fbc8f", Bg: "#0a150a",
		SelectionBg: "#1a3a1a", Cursor: "#00ff00",
		Black: "#0a0a0a", Red: "#d70000", Green: "#00ff00", Yellow: "#d7d700",
		Blue: "#00afaf", Purple: "#00ff00", Cyan: "#5faf87", White: "#8fbc8f",
		BrightBlack: "#4e4e4e", BrightRed: "#e06c75", BrightGreen: "#98c379", BrightYellow: "#d19a66",
		BrightBlue: "#87d787", BrightPurple: "#008700", BrightCyan: "#56b6c2", BrightWhite: "#ffffff",
	})

	// TintKlingon is a red/rust theme on near-black.
	TintKlingon = builtinTint(yamlTint{
		ID: "klingon", DisplayName: "Klingon",
		Fg: "#cd853f", Bg: "#150a0a",
		SelectionBg: "#3a1a1a", Cursor: "#ff0000",
		Black: "#0a0a0a", Red: "#ff0000", Green: "#d7d700", Yellow: "#ff8700",
		Blue: "#9fb8e0", Purple: "#ff4444", Cyan: "#56b6c2", White: "#cd853f",
		BrightBlack: "#4e4e4e", BrightRed: "#e06c75", BrightGreen: "#98c379", BrightYellow: "#d19a66",
		BrightBlue: "#61afef", BrightPurple: "#af0000", BrightCyan: "#9fb8e0", BrightWhite: "#ffffff",
	})

	// TintDarkula is the JetBrains Darkula-inspired IDE theme.
	TintDarkula = builtinTint(yamlTint{
		ID: "darkula", DisplayName: "Darkula",
		Fg: "#a9b7c6", Bg: "#2b2b2b",
		SelectionBg: "#214283", Cursor: "#ffc66d",
		Black: "#3c3f41", Red: "#ff6b68", Green: "#6a8759", Yellow: "#ffc66d",
		Blue: "#6897bb", Purple: "#ffc66d", Cyan: "#56b6c2", White: "#a9b7c6",
		BrightBlack: "#606060", BrightRed: "#ff6b68", BrightGreen: "#6a8759", BrightYellow: "#ffc66d",
		BrightBlue: "#6897bb", BrightPurple: "#6897bb", BrightCyan: "#56b6c2", BrightWhite: "#ffffff",
	})

	// TintEarthy is a warm tan/brown theme.
	TintEarthy = builtinTint(yamlTint{
		ID: "earthy", DisplayName: "Earthy",
		Fg: "#c4a882", Bg: "#1e1209",
		SelectionBg: "#3e2723", Cursor: "#d4a574",
		Black: "#1a0f08", Red: "#cd5c5c", Green: "#6b8e23", Yellow: "#b8860b",
		Blue: "#9fb8e0", Purple: "#d4a574", Cyan: "#c2c6a8", White: "#c4a882",
		BrightBlack: "#5c4033", BrightRed: "#cd5c5c", BrightGreen: "#6b8e23", BrightYellow: "#d4a574",
		BrightBlue: "#61afef", BrightPurple: "#8b6914", BrightCyan: "#56b6c2", BrightWhite: "#ffffff",
	})
)

// orderedBuiltinTints is the stable display order for the built-in themes,
// matching the source TUI's ordering.
var orderedBuiltinTints = []*tint.Tint{
	TintDeepSpace,
	TintStarfleet,
	TintLCARS,
	TintRomulan,
	TintKlingon,
	TintDarkula,
	TintEarthy,
}

// BuiltinTints returns snap's built-in themes in display order. These are the
// first entries returned by ThemeTints; use this to build a theme picker that
// leads with the built-ins.
func BuiltinTints() []*tint.Tint {
	out := make([]*tint.Tint, len(orderedBuiltinTints))
	copy(out, orderedBuiltinTints)
	return out
}

// BuiltinTintIDs returns the built-in theme IDs in display order.
func BuiltinTintIDs() []string {
	out := make([]string, len(orderedBuiltinTints))
	for i, t := range orderedBuiltinTints {
		out[i] = t.ID
	}
	return out
}

// registerBuiltinTints adds the built-in themes to the bubbletint default
// registry. Callers must hold tintMu and must call it after the default
// registry has been created (see verifyRegistryUnsafe).
func registerBuiltinTints() {
	tint.Register(orderedBuiltinTints...)
}

// ThemeTints returns every registered tint with snap's built-ins first (in
// display order) followed by the remaining tints sorted alphabetically by ID —
// the ordering a theme picker should present. It initializes the registry if
// needed, so it is safe to call before any explicit theme selection.
func ThemeTints() []*tint.Tint {
	tintMu.Lock()
	verifyRegistryUnsafe()
	// Re-assert the built-ins every call (idempotent): verifyRegistryUnsafe
	// short-circuits once initialized, so if the registry was reset externally
	// after that point the built-ins could be missing — and a listed theme that
	// isn't registered can't be selected (SetCurrentTint would fail).
	registerBuiltinTints()
	tintMu.Unlock()

	builtinSet := make(map[string]bool, len(orderedBuiltinTints))
	for _, t := range orderedBuiltinTints {
		builtinSet[t.ID] = true
	}

	out := BuiltinTints()
	for _, t := range tint.Tints() { // already sorted alphabetically by ID
		if !builtinSet[t.ID] {
			out = append(out, t)
		}
	}
	return out
}

// ThemeTintIDs returns the IDs from ThemeTints in the same order.
func ThemeTintIDs() []string {
	tints := ThemeTints()
	out := make([]string, len(tints))
	for i, t := range tints {
		out[i] = t.ID
	}
	return out
}
