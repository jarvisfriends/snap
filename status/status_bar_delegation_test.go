// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package status

import (
	"testing"
	"time"

	"github.com/jarvisfriends/snap/keys"
	"github.com/jarvisfriends/snap/notifications"

	tea "charm.land/bubbletea/v2"
)

// newBarWithNotifs returns a keyed, sized status bar wired to a manager holding
// the given notification contents.
func newBarWithNotifs(t *testing.T, contents ...string) (*BarModel, *notifications.Manager) {
	t.Helper()
	b := New()
	b.SetKeys(keys.DefaultKeyMap())
	nm := notifications.NewManager()
	b.SetNotifManager(nm)
	for _, c := range contents {
		nm.Add(c, notifications.SeverityInfo, 5*time.Second)
	}
	b.SetWidth(80)
	return b, nm
}

func TestBarModel_InitAndVisibility(t *testing.T) {
	t.Parallel()

	b := New()
	b.SetKeys(keys.DefaultKeyMap())
	if cmd := b.Init(); cmd != nil {
		t.Errorf("Init() = %v; want nil", cmd)
	}
	if !b.IsVisible() {
		t.Fatal("New() bar should start visible")
	}
	b.SetWidth(80)
	if b.View().Content == "" {
		t.Error("visible bar View() content should be non-empty")
	}

	// Hidden bar renders empty content.
	b.ToggleVisible()
	if b.IsVisible() {
		t.Fatal("ToggleVisible should hide the bar")
	}
	if got := b.View().Content; got != "" {
		t.Errorf("hidden bar View() content = %q; want empty", got)
	}
}

func TestBarModel_UpdateWindowSizePropagatesWidth(t *testing.T) {
	t.Parallel()

	b := New()
	b.SetKeys(keys.DefaultKeyMap())

	model, cmd := b.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if cmd != nil {
		t.Errorf("WindowSizeMsg Update cmd = %v; want nil", cmd)
	}
	if model != b {
		t.Error("Update should return the same model instance")
	}
	// SetWidth ran as a side effect, so the three click regions are computed.
	if len(b.Regions()) != 3 {
		t.Errorf("Regions() = %d entries; want 3 (settings/notifications/info)", len(b.Regions()))
	}

	// Non-size messages are delegated to the notification overlay without panic.
	if _, cmd := b.Update(TickMsg{}); cmd != nil {
		t.Errorf("idle TickMsg delegated cmd = %v; want nil", cmd)
	}
}

func TestBarModel_SetPageBindingsRenders(t *testing.T) {
	t.Parallel()

	b := New()
	b.SetKeys(keys.DefaultKeyMap())
	b.SetWidth(80)

	// Page-specific bindings override the global key display.
	b.SetPageBindings(keys.DefaultKeyMap())
	if b.View().Content == "" {
		t.Error("bar with page bindings should render non-empty content")
	}
	// Passing nil reverts to the global key map.
	b.SetPageBindings(nil)
	if b.View().Content == "" {
		t.Error("bar reverted to global keys should render non-empty content")
	}
}

func TestBarModel_NotificationHistoryDelegation(t *testing.T) {
	t.Parallel()

	b, _ := newBarWithNotifs(t, "one", "two", "three")

	if b.IsHistoryVisible() {
		t.Fatal("history should start closed")
	}
	b.ToggleNotifications()
	if !b.IsHistoryVisible() {
		t.Fatal("ToggleNotifications should open the history panel")
	}
	if b.HistoryCursor() != 0 {
		t.Errorf("cursor after open = %d; want 0", b.HistoryCursor())
	}

	// Move down within the 3-item bound, then confirm it clamps at max-1.
	b.NotifHistoryCursorDown(3)
	b.NotifHistoryCursorDown(3)
	b.NotifHistoryCursorDown(3)
	if b.HistoryCursor() != 2 {
		t.Errorf("cursor clamped = %d; want 2", b.HistoryCursor())
	}

	// Move up, then confirm it floors at 0.
	b.NotifHistoryCursorUp()
	if b.HistoryCursor() != 1 {
		t.Errorf("cursor after up = %d; want 1", b.HistoryCursor())
	}
	b.NotifHistoryCursorUp()
	b.NotifHistoryCursorUp()
	if b.HistoryCursor() != 0 {
		t.Errorf("cursor floored = %d; want 0", b.HistoryCursor())
	}

	// The overlay renders while open and is empty once closed.
	if b.RenderHistoryOverlay(100, 24) == "" {
		t.Error("RenderHistoryOverlay should be non-empty while open with notifications")
	}
	b.CloseHistory()
	if b.IsHistoryVisible() {
		t.Fatal("CloseHistory should close the panel")
	}
	if got := b.RenderHistoryOverlay(100, 24); got != "" {
		t.Errorf("RenderHistoryOverlay after close = %q; want empty", got)
	}
}

func TestUserNotificationOverlay_VisibilityAndAnimation(t *testing.T) {
	t.Parallel()

	o := NewUserNotificationOverlay()
	o.SetWidth(60)
	if !o.Visible() || !o.ShouldShow() {
		t.Fatal("overlay should start visible")
	}

	// An animated toggle schedules a tick, hides immediately, but keeps
	// ShouldShow true for the duration of the slide-out.
	cmd := o.ToggleVisibility()
	if cmd == nil {
		t.Fatal("ToggleVisibility should schedule a tick cmd")
	}
	if o.Visible() {
		t.Error("overlay should be hidden after toggling")
	}
	if !o.ShouldShow() {
		t.Error("ShouldShow should remain true while animating out")
	}
	// A second toggle mid-animation is a no-op.
	if o.ToggleVisibility() != nil {
		t.Error("ToggleVisibility while animating should return nil")
	}

	// Drive the animation to completion.
	for range o.animFrames + 1 {
		o.Update(TickMsg{})
	}
	if o.animating {
		t.Error("animation should have finished after animFrames ticks")
	}

	// ForceToggleVisibility flips instantly with no animation.
	o.ForceToggleVisibility()
	if !o.Visible() || o.animating {
		t.Error("ForceToggleVisibility should flip visible immediately, no animation")
	}
}

func TestUserNotificationOverlay_UpdateMessages(t *testing.T) {
	t.Parallel()

	o := NewUserNotificationOverlay()
	// ToggleVisibilityMsg drives an animated toggle (non-nil tick).
	if cmd := o.Update(ToggleVisibilityMsg{}); cmd == nil {
		t.Error("ToggleVisibilityMsg should return a tick cmd")
	}

	// WindowSizeMsg updates the stored width.
	o.Update(tea.WindowSizeMsg{Width: 133, Height: 10})
	if o.width != 133 {
		t.Errorf("width after WindowSizeMsg = %d; want 133", o.width)
	}

	// TickMsg while idle is a no-op.
	idle := NewUserNotificationOverlay()
	if cmd := idle.Update(TickMsg{}); cmd != nil {
		t.Errorf("idle TickMsg cmd = %v; want nil", cmd)
	}
}

func TestUserNotificationOverlay_HistoryCursor(t *testing.T) {
	t.Parallel()

	o := NewUserNotificationOverlay()
	if o.ShowHistory() {
		t.Fatal("history should start closed")
	}

	// Opening history resets the cursor to the top.
	o.historyCursor = 5
	if cmd := o.ToggleHistory(); cmd != nil {
		t.Errorf("ToggleHistory cmd = %v; want nil", cmd)
	}
	if !o.ShowHistory() {
		t.Fatal("ToggleHistory should open the panel")
	}
	if o.HistoryCursor() != 0 {
		t.Errorf("opening history should reset cursor, got %d", o.HistoryCursor())
	}

	// Bounds: down clamps at maxItems-1, up floors at 0.
	o.HistoryCursorDown(2)
	if o.HistoryCursor() != 1 {
		t.Errorf("cursor = %d; want 1", o.HistoryCursor())
	}
	o.HistoryCursorDown(2)
	if o.HistoryCursor() != 1 {
		t.Errorf("clamped cursor = %d; want 1", o.HistoryCursor())
	}
	o.HistoryCursorUp()
	o.HistoryCursorUp()
	if o.HistoryCursor() != 0 {
		t.Errorf("floored cursor = %d; want 0", o.HistoryCursor())
	}

	o.CloseHistory()
	if o.ShowHistory() {
		t.Fatal("CloseHistory should close the panel")
	}
}

func TestColorForSeverity(t *testing.T) {
	t.Parallel()

	o := NewUserNotificationOverlay()
	cases := map[notifications.Severity]string{
		notifications.SeverityWarning: "#F9C513",
		notifications.SeverityError:   "#FF5757",
		notifications.SeverityInfo:    "#4FC3F7",
	}
	for sev, want := range cases {
		if got := o.colorForSeverity(sev); got != want {
			t.Errorf("colorForSeverity(%v) = %s; want %s", sev, got, want)
		}
	}
	// Unknown severities fall back to the info blue.
	if got := o.colorForSeverity(notifications.Severity(99)); got != "#4FC3F7" {
		t.Errorf("colorForSeverity(unknown) = %s; want #4FC3F7 fallback", got)
	}
}

func TestFormatAge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"seconds", 30 * time.Second, "30s ago"},
		{"minutes", 5 * time.Minute, "5m ago"},
		{"hours", 3 * time.Hour, "3h ago"},
	}
	for _, tc := range cases {
		if got := formatAge(tc.d); got != tc.want {
			t.Errorf("formatAge(%s) = %q; want %q", tc.name, got, tc.want)
		}
	}
}

func TestUserNotificationOverlay_Render(t *testing.T) {
	t.Parallel()

	o := NewUserNotificationOverlay()

	// Static (not animating): full-width bar with the three click regions.
	line, regions := o.Render(80, "help text", "summary")
	if line == "" {
		t.Error("static Render should produce a non-empty status line")
	}
	if len(regions) != 3 {
		t.Errorf("static Render regions = %d; want 3", len(regions))
	}

	// Mid slide-in animation exercises the indent branch.
	o.animating = true
	o.animFrames = 8
	o.animFrame = 2
	o.animDirection = 1
	if line, _ := o.Render(80, "help", "sum"); line == "" {
		t.Error("animating Render should still produce a non-empty line")
	}

	// With a manager wired, Render reads Enabled()/PendingCount(); a disabled
	// manager drives the muted-icon branch.
	nm := notifications.NewManager()
	nm.SetEnabled(false)
	o.animating = false
	o.SetNotifManager(nm)
	if line, _ := o.Render(80, "help", "sum"); line == "" {
		t.Error("Render with a manager should produce a non-empty line")
	}
}

func TestInfoModal_Geometry(t *testing.T) {
	t.Parallel()

	m := NewInfoModal()
	if m.IsVisible() {
		t.Fatal("info modal should start closed")
	}

	// Zero dimensions: nothing renders and the bounds collapse to the origin.
	m.Open(0, 0)
	if got := m.View().Content; got != "" {
		t.Errorf("0x0 modal View content = %q; want empty", got)
	}
	if bx, by, bw, bh := m.Bounds(); bw != 0 || bh != 0 || bx != 0 || by != 0 {
		t.Errorf("0x0 modal Bounds = (%d,%d,%d,%d); want all zero", bx, by, bw, bh)
	}

	// A very small terminal forces the minimum-size clamps in boxDims/vpDims.
	m.Open(12, 6)
	vpW, vpH := m.vpDims()
	if vpW < 10 || vpH < 1 {
		t.Errorf("vpDims clamps not applied: got %dx%d", vpW, vpH)
	}

	// A normal terminal builds content and renders a bordered box.
	m.Resize(100, 30)
	if m.View().Content == "" {
		t.Error("normal-size modal should render non-empty content")
	}

	// Toggle closes an open modal and reopens a closed one.
	m.Toggle(100, 30)
	if m.IsVisible() {
		t.Error("Toggle should close the open modal")
	}
	m.Toggle(100, 30)
	if !m.IsVisible() {
		t.Error("Toggle should reopen the closed modal")
	}
}
