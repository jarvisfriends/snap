package styles

import (
	"sync"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// BuildHuhStyles produces huh form styles by combining three orthogonal axes:
//   - structure (borders, prefixes, indicators, padding, glyphs) from the chosen
//     StylePreset's huh built-in theme,
//   - colors from the application palette c (derived from the active tint),
//   - the light/dark variant via isDark.
//
// The preset theme is used as the structural starting point; overlayPaletteColors
// then re-tints every foreground/background/border so any of the 342 tints works
// with any preset. Because lipgloss styles are immutable values, re-applying a
// color preserves the preset's BorderStyle, padding, and SetString glyphs.
func BuildHuhStyles(c *AppStyle, preset StylePreset, isDark bool) *huh.Styles {
	fn := presetThemes[NormalizePreset(string(preset))]
	t := fn(isDark)
	overlayPaletteColors(t, c)
	return t
}

// overlayPaletteColors re-applies the application palette's colors onto a huh
// theme in place, mapping each semantic color to the matching form role while
// leaving the theme's structural attributes untouched.
func overlayPaletteColors(t *huh.Styles, c *AppStyle) {
	t.Focused.Base = t.Focused.Base.BorderForeground(c.Accent)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(c.Accent).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(c.Accent).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(c.Muted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(c.Error)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(c.Error)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(c.Warning)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(c.Warning)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(c.Warning)
	t.Focused.Option = t.Focused.Option.Foreground(c.Fg)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(c.SelectionFg).
		Background(c.SelectionBg)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(c.Success)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(c.Fg)
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(c.Muted)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(c.Bg).
		Background(c.Accent).
		Bold(true)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(c.Fg).Background(c.Border)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(c.Warning)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(c.Muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(c.Warning)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = c.Styles.TextOnBg
	t.Blurred.PrevIndicator = c.Styles.TextOnBg
	t.Blurred.Title = t.Focused.Title.Foreground(c.Muted).Bold(false)
	t.Blurred.Description = c.Styles.Dim
	t.Blurred.SelectedOption = t.Focused.SelectedOption.Foreground(c.Muted)
	t.Blurred.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(c.Muted)

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
}

// huhStylesCache memoizes the *huh.Styles returned by HuhThemeFunc. huh calls
// the ThemeFunc once per option per render (see Select.activeStyles), and
// huh.Styles is a large struct, so copying it every call dominated theme-picker
// scrolling (a 300-option list re-rendered options up to the cursor on every
// keypress → ~240ms/key, most of it this copy). huh only ever reads the returned
// styles — no field type mutates them — so we can hand out a shared pointer and
// rebuild it only when the active palette's precomputed styles change (detected
// by pointer identity, since fromTint caches one *huh.Styles per theme).
var (
	huhStylesMu     sync.Mutex
	huhStylesSrc    *huh.Styles // active.HuhStyles pointer the cache was built from
	huhStylesShared *huh.Styles // shared, read-only styles handed to huh
)

// HuhThemeFunc returns a ThemeFunc backed by the styles precomputed in the
// active palette (see fromTint). The result is memoized and shared: huh treats
// it as read-only, so a per-call copy is unnecessary and was prohibitively
// expensive on hot render paths.
func HuhThemeFunc() huh.ThemeFunc {
	return func(_ bool) *huh.Styles {
		active := Active()
		if active.HuhStyles == nil {
			return BuildHuhStyles(active, DefaultStylePreset, true)
		}
		huhStylesMu.Lock()
		defer huhStylesMu.Unlock()
		if huhStylesSrc != active.HuhStyles {
			cp := *active.HuhStyles
			huhStylesShared = &cp
			huhStylesSrc = active.HuhStyles
		}
		return huhStylesShared
	}
}
