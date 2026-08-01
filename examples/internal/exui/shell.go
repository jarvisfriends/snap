// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package exui

import (
	"fmt"
	"runtime"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/dependencies"
	"github.com/jarvisfriends/snap/geom"
	"github.com/jarvisfriends/snap/keys"
	"github.com/jarvisfriends/snap/notifications"
	"github.com/jarvisfriends/snap/status"
)

// This file is the shared "shell" behind every example: the status bar's
// three surfaces (notifications history, info modal, debug overlay) plus the
// global chords and click regions that drive them. Examples only supply
// their component and its data; the shell owns the rest, mirroring how a
// real app (tui-base) would mount these same snap components.
//
// Everything here is lazy: overlays render only while visible, the bar
// re-renders only when refresh() marks it dirty, and no per-frame
// allocations happen while all surfaces are closed.

// infoKey opens the info modal; keys.AppKeyMap has no binding for it (the
// bar's ⓘ region is the primary affordance), so the shell adds one chord
// that can never collide with text input.
const infoKey = "ctrl+e"

// Manager exposes the shell's notification manager so examples can emit
// toasts/history entries (data in, chrome handled here).
func (c *Chrome) Manager() *notifications.Manager {
	if c == nil {
		return nil
	}
	return c.mgr
}

// Notify adds a notification and refreshes the bar's bell badge.
func (c *Chrome) Notify(content string, sev notifications.Severity) tea.Cmd {
	if c == nil {
		return nil
	}
	_, cmd := c.mgr.Add(content, sev, sev.DefaultTTL())
	c.refresh()
	return cmd
}

// SetSegment registers a live right-aligned bar segment (see status.BarModel).
func (c *Chrome) SetSegment(name string, fn func() string) {
	if c == nil {
		return
	}
	c.bar.SetSegment(name, fn)
}

// Bar exposes the underlying status bar (tests and advanced examples).
func (c *Chrome) Bar() *status.BarModel {
	if c == nil {
		return nil
	}
	return c.bar
}

// Info exposes the shared info modal (tests and advanced examples).
func (c *Chrome) Info() *status.InfoModal {
	if c == nil {
		return nil
	}
	return c.modal
}

// OpenInfo opens the info modal (version + full dependency list).
func (c *Chrome) OpenInfo() {
	if c == nil {
		return
	}
	c.modal.Open(c.width, c.height)
}

// SetSize informs the shell of the terminal size. Examples call it from
// their WindowSizeMsg case; SetWidth remains for bar-only layouts.
func (c *Chrome) SetSize(w, h int) {
	if c == nil {
		return
	}
	c.height = h
	c.modal.Resize(w, h)
	c.SetWidth(w)
}

// Refresh re-renders the bar; call after mutating state through Manager()
// directly (Notify and the shell's own paths refresh automatically).
func (c *Chrome) Refresh() {
	if c == nil {
		return
	}
	c.refresh()
}

// refresh re-renders the bar on the next SetWidth; cheap enough to call
// after every state change (SetWidth is the bar's single render entry).
func (c *Chrome) refresh() { c.bar.SetWidth(c.width) }

// Update gives the shell first look at every message. done=true means the
// shell consumed it (an overlay owned the key, a bar surface toggled, a
// notification message was routed) and the example should return cmd as-is.
// done=false (always with a nil cmd) means the example proceeds normally.
func (c *Chrome) Update(msg tea.Msg) (cmd tea.Cmd, done bool) {
	if c == nil {
		return nil, false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.SetSize(msg.Width, msg.Height)
		return nil, false // the example sizes its component too

	case status.ClickRegionMsg:
		switch msg.Name {
		case status.NotificationsRegionName:
			cmd := c.bar.ToggleNotifications()
			c.refresh()
			return cmd, true
		case status.InfoRegionName:
			c.modal.Toggle(c.width, c.height)
			return nil, true
		case status.SettingsRegionName:
			return c.Notify("settings are wired by the host app", notifications.SeverityInfo), true
		}
		return nil, true

	case status.CloseInfoModalMsg:
		return nil, true

	// Notification lifecycle messages route to the manager (TTL expiry,
	// programmatic add/dismiss); the bell badge refreshes after.
	case notifications.ExpireMsg, notifications.AddMsg, notifications.ProgressMsg,
		notifications.DismissMsg, notifications.DismissKeyMsg, notifications.DismissAllMsg:
		cmd := c.mgr.Handle(msg)
		c.refresh()
		return cmd, true

	case tea.KeyPressMsg:
		// Modal-first: an open overlay owns the keyboard.
		if c.modal.IsVisible() {
			m, cmd := c.modal.Update(msg)
			if im, ok := m.(*status.InfoModal); ok {
				c.modal = im
			}
			return cmd, true
		}
		if c.debug {
			if key.Matches(msg, c.km.Debug, c.km.Dismiss) {
				c.debug = false
			}
			return nil, true
		}
		if c.bar.IsHistoryVisible() {
			c.historyKey(msg)
			return nil, true
		}
		// Global chords, chosen to never collide with typing components.
		switch {
		case key.Matches(msg, c.km.ToggleHistory):
			cmd := c.bar.ToggleNotifications()
			c.refresh()
			return cmd, true
		case key.Matches(msg, c.km.Debug):
			c.debug = !c.debug
			return nil, true
		case key.Matches(msg, c.km.ToggleFullHelp):
			c.bar.ToggleFullHelpVisible()
			c.refresh()
			return nil, true
		case msg.String() == infoKey:
			c.modal.Toggle(c.width, c.height)
			return nil, true
		}
	}
	return nil, false
}

// historyKey drives the open notifications panel: move, dismiss one,
// dismiss all, close — the actions the panel's own footer advertises.
func (c *Chrome) historyKey(msg tea.KeyPressMsg) {
	switch {
	case key.Matches(msg, c.km.Up):
		c.bar.NotifHistoryCursorUp()
	case key.Matches(msg, c.km.Down):
		c.bar.NotifHistoryCursorDown(len(c.mgr.Active()))
	case key.Matches(msg, c.km.Select): // enter: dismiss the highlighted entry
		if act := c.mgr.Active(); len(act) > 0 {
			i := geom.Clamp(c.bar.HistoryCursor(), 0, len(act)-1)
			c.mgr.Dismiss(act[i].ID)
		}
	case key.Matches(msg, c.km.DismissAll):
		c.mgr.DismissAll(nil)
	case key.Matches(msg, c.km.Dismiss), key.Matches(msg, c.km.ToggleHistory):
		c.bar.CloseHistory()
	}
	c.refresh()
}

// wrapMouse chains the shell's pointer handling in front of the example's
// own handler: an open info modal takes the event first (wheel scroll,
// outside-click close), then bar icon clicks, then the example.
func (c *Chrome) wrapMouse(next func(tea.MouseMsg) tea.Cmd) func(tea.MouseMsg) tea.Cmd {
	return func(mm tea.MouseMsg) tea.Cmd {
		if c == nil {
			if next != nil {
				return next(mm)
			}
			return nil
		}
		if c.modal.IsVisible() {
			if cmd, handled := c.modal.HandleMouse(mm); handled {
				return cmd
			}
		}
		if c.historyMouse(mm) {
			return nil
		}
		if cmd := c.barRegionCmd(mm); cmd != nil {
			return cmd
		}
		if next != nil {
			return next(mm)
		}
		return nil
	}
}

// historyMouse closes the open notifications panel on an outside click —
// the same affordance the info modal has. Reports whether it consumed the
// event.
func (c *Chrome) historyMouse(mm tea.MouseMsg) bool {
	if !c.bar.IsHistoryVisible() {
		return false
	}
	click, ok := mm.(tea.MouseClickMsg)
	if !ok {
		return false
	}
	overlay := c.bar.RenderHistoryOverlay(c.width, c.height)
	if overlay == "" {
		return false
	}
	ow, oh := lipgloss.Width(overlay), lipgloss.Height(overlay)
	r := geom.Rect{X: max(c.width-ow-1, 0), Y: max(c.height-oh-1, 0), W: ow, H: oh}
	m := click.Mouse()
	if r.Contains(m.X, m.Y) {
		return true // panel owns clicks inside it
	}
	c.bar.CloseHistory()
	c.refresh()
	return true
}

// barRegionCmd maps a release on the bar's bottom row to the matching
// ClickRegionMsg (settings / notifications / info icon), or nil.
func (c *Chrome) barRegionCmd(mm tea.MouseMsg) tea.Cmd {
	rel, ok := mm.(tea.MouseReleaseMsg)
	if !ok || c.hidden || c.height <= 0 {
		return nil
	}
	m := rel.Mouse()
	if m.Y != c.height-1 {
		return nil
	}
	for _, r := range c.bar.Regions() {
		if m.X >= r.Start && m.X < r.End {
			name := r.Name
			return func() tea.Msg { return status.ClickRegionMsg{Name: name} }
		}
	}
	return nil
}

// overlays composites any open shell surface over the finished frame.
// Called by Frame after Apply; no-ops (and allocates nothing) when closed.
func (c *Chrome) overlays(content string) string {
	if c == nil {
		return content
	}
	var layers []*lipgloss.Layer
	if hist := c.bar.RenderHistoryOverlay(c.width, c.height); hist != "" {
		layers = append(layers, lipgloss.NewLayer(hist).
			X(max(c.width-lipgloss.Width(hist)-1, 0)).
			Y(max(c.height-lipgloss.Height(hist)-1, 0)).Z(5))
	}
	if c.modal.IsVisible() {
		bx, by, _, _ := c.modal.Bounds()
		layers = append(layers, lipgloss.NewLayer(c.modal.View().Content).X(bx).Y(by).Z(10))
	}
	if c.debug {
		box := c.debugView()
		r := geom.Rect{W: lipgloss.Width(box), H: lipgloss.Height(box)}.
			CenteredIn(c.width, c.height)
		layers = append(layers, lipgloss.NewLayer(box).X(r.X).Y(r.Y).Z(20))
	}
	if layers == nil {
		return content
	}
	return lipgloss.NewCompositor(
		append([]*lipgloss.Layer{lipgloss.NewLayer(content)}, layers...)...,
	).Render()
}

// debugView renders the ctrl+d quick-debug overlay: build, runtime, and
// shell state at a glance. Built only while the overlay is open.
func (c *Chrome) debugView() string {
	t := Theme()
	ver := Version
	if info := dependencies.ExpandedBuildInfo(); info != nil && info.App.Version != "" {
		ver = info.App.Version
	}
	body := fmt.Sprintf(
		"version   %s\ngo        %s %s/%s\nterm      %d×%d\ntint      %s\nnotifs    %d active / %d pending\nuptime    %s\n\nctrl+d/esc close",
		ver, runtime.Version(), runtime.GOOS, runtime.GOARCH,
		c.width, c.height, themeTint,
		len(c.mgr.Active()), c.mgr.PendingCount(),
		time.Since(c.start).Truncate(time.Second),
	)
	return t.Styles.OverlayBorder.Render(" debug \n\n" + body)
}

// newShellParts builds the notification manager, keymap, and info modal a
// Chrome carries; split out so NewChrome (exui.go) stays declarative.
func newShellParts(bar *status.BarModel) (*notifications.Manager, *keys.AppKeyMap, *status.InfoModal) {
	mgr := notifications.NewManager()
	km := keys.DefaultKeyMap()
	bar.SetNotifManager(mgr)
	bar.SetKeys(km)
	modal := status.NewInfoModal()
	modal.SetKeys(km)
	modal.SetAppName("snap_input")
	if info := dependencies.ExpandedBuildInfo(); info != nil && info.App.Version != "" {
		modal.SetVersion(info.App.Version)
	} else {
		modal.SetVersion(Version)
	}
	return mgr, km, modal
}
