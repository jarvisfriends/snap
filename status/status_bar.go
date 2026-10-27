// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package status

import (
	"strings"

	"github.com/jarvisfriends/snap/keys"
	"github.com/jarvisfriends/snap/notifications"
	"github.com/jarvisfriends/snap/page"
	"github.com/jarvisfriends/snap/styles"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// BarModel is the bar at the bottom of our App that displays helpful info like key combos
// available for the current view. This implementation wraps the more featureful
// statusbar model (animation, overlay, click regions) and exposes the same
// tea.Model surface used elsewhere in the app.
type BarModel struct {
	page.Base
	help     *help.Model
	helpView tea.View
	keys     *keys.AppKeyMap

	// pageBindings, when non-nil, overrides the global keys display so each
	// page can show its own relevant shortcuts in the status bar.
	pageBindings help.KeyMap
	isVisible    bool

	sb          *UserNotificationOverlay
	lastRegions []ClickRegion // regions from the last SetWidth render

	// summaryProvider, when set, supplies the right-aligned status bar text
	// (e.g. compact runtime summary shown when active). It is evaluated on
	// every render so the text stays current without manual refresh calls.
	summaryProvider func() string

	// segmentOrder/segments hold named consumer-provided right-aligned
	// segments (git branch, connection state, …), rendered before the summary
	// in registration order and re-evaluated every render (E-1).
	segmentOrder []string
	segments     map[string]func() string
}

// SetSummaryProvider sets a callback that supplies the right-aligned status bar
// text, evaluated on every render. Pass nil to clear it.
func (b *BarModel) SetSummaryProvider(fn func() string) {
	b.summaryProvider = fn
	b.SetWidth(b.help.Width())
}

// SetSegment registers (or replaces) a named right-aligned status bar segment
// whose text is re-evaluated on every render — the consumer hook for live
// widgets like a git branch or connection state. Segments render in first-
// registration order, separated by " • ", before the summary text. A segment
// returning "" is skipped for that frame; passing a nil fn removes the
// segment entirely.
func (b *BarModel) SetSegment(name string, fn func() string) {
	if fn == nil {
		if _, ok := b.segments[name]; ok {
			delete(b.segments, name)
			for i, n := range b.segmentOrder {
				if n == name {
					b.segmentOrder = append(b.segmentOrder[:i], b.segmentOrder[i+1:]...)
					break
				}
			}
		}
		b.SetWidth(b.help.Width())
		return
	}
	if b.segments == nil {
		b.segments = make(map[string]func() string)
	}
	if _, exists := b.segments[name]; !exists {
		b.segmentOrder = append(b.segmentOrder, name)
	}
	b.segments[name] = fn
	b.SetWidth(b.help.Width())
}

// summary returns the full right-aligned text: registered segments in order,
// then the summary provider's text, joined with " • ".
func (b *BarModel) summary() string {
	parts := make([]string, 0, len(b.segmentOrder)+1)
	for _, name := range b.segmentOrder {
		if s := b.segments[name](); s != "" {
			parts = append(parts, s)
		}
	}
	if b.summaryProvider != nil {
		if s := b.summaryProvider(); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " • ")
}

// Regions returns the interactive click regions computed during the last
// SetWidth call. The router uses these for hit-testing without re-parsing
// ANSI-encoded output.
func (b *BarModel) Regions() []ClickRegion { return b.lastRegions }

// applyHelpStyles updates the help widget's internal styles to match the
// current theme palette. SetColors/Colors are inherited from page.Base.
func (b *BarModel) applyHelpStyles() {
	b.help.Styles = b.Colors().Styles.Help
}

func New() *BarModel {
	h := help.New()
	h.Styles = styles.Active().Styles.Help

	return &BarModel{
		help:      &h,
		helpView:  tea.NewView(""),
		sb:        NewUserNotificationOverlay(),
		isVisible: true,
	}
}

// ClickRegionMsg is emitted when the user clicks a named interactive region
// on the status bar (for example, the settings icon). The router will handle
// this message and perform navigation.
type ClickRegionMsg struct{ Name string }

// KeyActivatedMsg is emitted when the user clicks a status-bar help hint. The
// router converts this back into the equivalent key press so a mouse click and a
// physical key press reach the same handlers.
type KeyActivatedMsg struct {
	Binding key.Binding
	// Keys is retained for compatibility with downstream repos still
	// constructing KeyActivatedMsg directly.
	Keys []string
	Help    key.Help
	// Press is the precomputed key press corresponding to the activated hint.
	// When set, consumers should prefer it over reparsing Keys.
	Press *tea.KeyPressMsg
}

var keySpecToCode = map[string]rune{
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"tab":       tea.KeyTab,
	"enter":     tea.KeyEnter,
	"esc":       tea.KeyEsc,
	"escape":    tea.KeyEsc,
	"space":     tea.KeySpace,
	"pgup":      tea.KeyPgUp,
	"pageup":    tea.KeyPgUp,
	"pgdown":    tea.KeyPgDown,
	"pagedown":  tea.KeyPgDown,
	"home":      tea.KeyHome,
	"end":       tea.KeyEnd,
	"delete":    tea.KeyDelete,
	"del":       tea.KeyDelete,
	"backspace": tea.KeyBackspace,
	"backtab":   tea.KeyTab,
}

var keySpecModMask = map[string]tea.KeyMod{
	"ctrl":    tea.ModCtrl,
	"control": tea.ModCtrl,
	"alt":     tea.ModAlt,
	"shift":   tea.ModShift,
	"meta":    tea.ModMeta,
	"cmd":     tea.ModMeta,
	"super":   tea.ModMeta,
}

// KeyMsgFor converts key binding spec strings into a Bubble Tea key message.
// Deprecated: prefer passing key.Binding in KeyActivatedMsg and using
// key.Matches where possible.
func KeyMsgFor(keysList []string) (tea.KeyMsg, bool) {
	for _, spec := range keysList {
		if press, ok := keyMsgForSpec(spec); ok {
			return press, true
		}
	}
	return nil, false
}

func keyMsgForSpec(spec string) (tea.KeyPressMsg, bool) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return tea.KeyPressMsg{}, false
	}

	if rs := []rune(spec); len(rs) == 1 {
		return tea.KeyPressMsg{Code: rs[0], Text: spec}, true
	}

	base, mod, ok := parseKeySpec(spec)
	if !ok {
		return tea.KeyPressMsg{}, false
	}
	return tea.KeyPressMsg{Code: base, Mod: mod, Text: textForKey(base, mod)}, true
}

func parseKeySpec(spec string) (code rune, mod tea.KeyMod, ok bool) {
	parts := strings.Split(strings.ToLower(spec), "+")
	if len(parts) == 0 {
		return 0, 0, false
	}

	mod = 0
	for _, p := range parts[:len(parts)-1] {
		if m, found := keySpecModMask[strings.TrimSpace(p)]; found {
			mod |= m
		} else {
			return 0, 0, false
		}
	}

	base := strings.TrimSpace(parts[len(parts)-1])
	if named, found := keySpecToCode[base]; found {
		if base == "backtab" {
			return named, mod | tea.ModShift, true
		}
		return named, mod, true
	}

	if rs := []rune(base); len(rs) == 1 {
		return rs[0], mod, true
	}
	return 0, 0, false
}

func textForKey(code rune, mod tea.KeyMod) string {
	if code == tea.KeySpace {
		return " "
	}
	if mod == 0 && code >= 32 && code != 127 {
		return string(code)
	}
	return ""
}

func (b *BarModel) activeKeyRegionAtX(x int) (key.Binding, bool) {
	bindings := b.pageBindings
	if bindings == nil {
		if b.keys == nil {
			return key.Binding{}, false
		}
		bindings = b.keys
	}
	if bindings == nil {
		return key.Binding{}, false
	}

	pos := 0
	for _, binding := range bindings.ShortHelp() {
		if !binding.Enabled() {
			continue
		}
		helpText := binding.Help().Key
		if binding.Help().Desc != "" {
			helpText += " " + binding.Help().Desc
		}
		start := pos
		end := start + lipgloss.Width(helpText) - 1
		if x >= start && x <= end {
			return binding, true
		}
		pos = end + 2
	}
	return key.Binding{}, false
}

func (b *BarModel) SetKeys(km *keys.AppKeyMap) {
	b.keys = km
	b.SetWidth(b.help.Width())
}

// SetPageBindings supplies page-specific key bindings that are shown in the
// status bar instead of the global router shortcuts. Pass nil to revert to
// the global key map.
func (b *BarModel) SetPageBindings(bindings help.KeyMap) {
	b.pageBindings = bindings
	b.SetWidth(b.help.Width())
}

// Init implements [tea.Model].
func (b *BarModel) Init() tea.Cmd { return nil }

// Update implements [tea.Model].
func (b *BarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Forward control messages to the internal statusbar model to drive
	// animations and TTLs. The statusbar.Handle method returns a tea.Cmd
	// when it wants the runtime to schedule ticks or timeouts.
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		b.SetWidth(m.Width)
		return b, nil
	}
	// Allow the internal statusbar to handle other control messages like
	// TickMsg, ToggleVisibilityMsg, AddMessageMsg, etc.
	return b, b.sb.Update(msg)
}

// View implements [tea.Model]. It composes overlay (if any) above the status
// line and exposes an OnMouse handler that maps clicks to ClickRegionMsg.
func (b *BarModel) View() tea.View {
	if !b.IsVisible() {
		return tea.NewView("")
	}
	return b.helpView
}

func (b *BarModel) SetWidth(width int) {
	// Apply theme colors to the help widget before rendering.
	b.applyHelpStyles()
	// verify help width is up to date for the left content
	b.help.SetWidth(width)
	c := b.Colors()

	// When the active page provides its own key bindings show those; otherwise
	// fall back to the global router shortcuts.
	var left string
	if b.pageBindings != nil {
		left = b.help.View(b.pageBindings)
	} else {
		left = b.help.View(b.keys)
	}
	// The bubbles help widget emits \x1b[m (reset) before each inter-element
	// space, stripping the background. Re-apply StatusBg after every reset so
	// those spaces stay on the status bar background, not the terminal default.
	left = styles.ReapplyBg(left, c.Styles.StatusBase.GetBackground())

	// Render the status line using the internal statusbar renderer. The right
	// segment carries the optional summary (e.g. runtime stats).
	statusLine, regions := b.sb.Render(width, left, b.summary())

	// Store regions so the router can access them without re-parsing ANSI output.
	b.lastRegions = regions

	// Icons live on the last row (row 0 in short-help, last row in full-help).
	statusLineRow := max(lipgloss.Height(statusLine)-1, 0)

	// Update the tea.View content and attach an OnMouse handler to map
	// clicks on interactive regions back into ClickRegionMsg messages.
	b.helpView.SetContent(c.Styles.StatusBase.Width(width).Render(statusLine))
	b.helpView.BackgroundColor = c.Styles.StatusBase.GetBackground()
	b.helpView.OnMouse = func(mm tea.MouseMsg) tea.Cmd {
		return func() tea.Msg {
			switch ev := mm.(type) {
			case tea.MouseReleaseMsg:
				me := ev.Mouse()
				// Only respond to clicks on the status bar row itself.
				// lipgloss.Height("") == 1, so never derive the row from an
				// empty overlay string — use the pre-computed row index instead.
				if me.Y != statusLineRow {
					return nil
				}
				// Use lipgloss.Width for ANSI-safe region hit-testing.
				x := me.X
				for _, r := range regions {
					if x >= r.Start && x <= r.End {
						return ClickRegionMsg{Name: r.Name}
					}
				}
				if binding, ok := b.activeKeyRegionAtX(x); ok {
					msg := KeyActivatedMsg{Binding: binding, Keys: binding.Keys(), Help: binding.Help()}
					if press, ok := KeyMsgFor(binding.Keys()); ok {
						if keyPress, ok := press.(tea.KeyPressMsg); ok {
							msg.Press = &keyPress
						}
					}
					return msg
				}
			}
			return nil
		}
	}
}

// Height returns the height of the status bar view (always 1 when visible).
func (b *BarModel) Height() int {
	if b.sb.Visible() {
		return lipgloss.Height(b.helpView.Content)
	}
	return 0
}

func (b *BarModel) ToggleFullHelpVisible() {
	b.help.ShowAll = !b.help.ShowAll
	b.SetWidth(b.help.Width())
}

func (b *BarModel) IsVisible() bool { return b.isVisible }

func (b *BarModel) ToggleVisible() tea.Cmd {
	b.isVisible = !b.isVisible
	// if cmd := b.sb.ToggleVisibility(); cmd != nil {
	// 	return cmd
	// }
	// animation was in-flight; force immediate toggle for deterministic tests
	b.sb.ForceToggleVisibility()
	return nil
}

// ToggleNotifications toggles the notification history panel.
func (b *BarModel) ToggleNotifications() tea.Cmd { return b.sb.ToggleHistory() }

// IsHistoryVisible reports whether the history panel is currently open.
func (b *BarModel) IsHistoryVisible() bool { return b.sb.ShowHistory() }

// NotifHistoryCursorUp moves the history cursor up.
func (b *BarModel) NotifHistoryCursorUp() { b.sb.HistoryCursorUp() }

// NotifHistoryCursorDown moves the history cursor down.
func (b *BarModel) NotifHistoryCursorDown(n int) { b.sb.HistoryCursorDown(n) }

// HistoryCursor returns the current notification history cursor index.
func (b *BarModel) HistoryCursor() int { return b.sb.HistoryCursor() }

// CloseHistory closes the notification history panel.
func (b *BarModel) CloseHistory() { b.sb.CloseHistory() }

// SetNotifManager wires the shared notification manager to the status bar.
func (b *BarModel) SetNotifManager(nm *notifications.Manager) { b.sb.SetNotifManager(nm) }

// RenderHistoryOverlay returns the rendered notification history panel string
// (or "") so the router can composite it as a canvas layer. screenW and screenH
// are the full terminal dimensions used to cap the panel size.
func (b *BarModel) RenderHistoryOverlay(screenW, screenH int) string {
	return b.sb.RenderHistoryOverlay(screenW, screenH)
}

var _ tea.Model = (*BarModel)(nil)
