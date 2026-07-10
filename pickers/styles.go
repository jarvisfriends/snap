// Package pickers holds directory/file selection components extracted from
// tui-base (ROADMAP SP-7): a drive-aware directory browser (DirPicker) and a
// multi-path editor (MultiFileEditor) whose rows open per-row pickers. Both
// are theme-free with injected style hooks, per snap's design rules.
package pickers

import "charm.land/lipgloss/v2"

// Styles are the style hooks shared by the pickers. Hosts map their palette
// onto these; tui-base does so from its live theme.
type Styles struct {
	// Title styles the picker heading line.
	Title lipgloss.Style
	// Path styles the current-location line (DirPicker).
	Path lipgloss.Style
	// Selected styles the cursor row.
	Selected lipgloss.Style
	// Normal styles non-cursor rows.
	Normal lipgloss.Style
	// Dim styles help text and empty/error notices.
	Dim lipgloss.Style
}

// DefaultStyles returns neutral, terminal-palette styles.
func DefaultStyles() Styles {
	r := lipgloss.NewStyle()
	return Styles{
		Title:    r.Bold(true).Padding(0, 1),
		Path:     r.Foreground(lipgloss.Color("245")),
		Selected: r.Foreground(lipgloss.Color("212")).Bold(true),
		Normal:   r.Foreground(lipgloss.Color("252")),
		Dim:      r.Foreground(lipgloss.Color("240")),
	}
}
