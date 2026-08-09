// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package status implements the `snap_input status` subcommand, demoing
// snap/status + snap/notifications end to end through the shared example
// shell: the status bar (key help left, live segments right), i/w/e emit
// notifications, p runs a fake download whose progress notification fills,
// and the shell provides ctrl+n history, ctrl+e info, and ctrl+d debug —
// exactly what every other example gets for free. Display-only demo.
package status

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/notifications"
)

// progressTickMsg advances the fake download's progress notification.
type progressTickMsg struct{}

func progressTick() tea.Cmd {
	return tea.Tick(180*time.Millisecond, func(time.Time) tea.Msg { return progressTickMsg{} })
}

type demoApp struct {
	chrome *exui.Chrome
	pct    float64
	dlID   int64
	w, h   int
	start  time.Time
}

func newDemo() *demoApp {
	a := &demoApp{dlID: -1, start: time.Now()}
	a.chrome = exui.NewChrome(
		exui.Bind("i/w/e", "info/warn/error"),
		exui.Bind("p", "progress"),
		exui.Bind("q", "quit"),
	)
	// Right-aligned segments re-evaluate every render — the consumer hook
	// for live widgets like a git branch or connection state.
	a.chrome.SetSegment("branch", func() string { return " master" })
	a.chrome.SetSegment("uptime", func() string {
		return time.Since(a.start).Truncate(time.Second).String()
	})
	return a
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height

	case progressTickMsg:
		if a.dlID < 0 {
			return a, nil
		}
		a.pct += 7
		a.chrome.Manager().SetProgress(a.dlID, a.pct)
		a.chrome.Refresh()
		if a.pct >= 100 {
			a.dlID = -1
			return a, a.chrome.Notify("download complete", notifications.SeverityInfo)
		}
		return a, progressTick()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "i":
			return a, a.chrome.Notify("deploy finished cleanly", notifications.SeverityInfo)
		case "w":
			return a, a.chrome.Notify("disk 82% full on /var", notifications.SeverityWarning)
		case "e":
			return a, a.chrome.Notify("backup job failed (exit 3)", notifications.SeverityError)
		case "p":
			if a.dlID >= 0 {
				return a, nil
			}
			a.pct = 0
			zero := 0.0
			n, cmd := a.chrome.Manager().AddWithOptions("downloading assets", notifications.SeverityInfo, 0,
				notifications.AddOptions{Key: "download", Percent: &zero, RetainInHistory: true})
			a.dlID = n.ID
			a.chrome.Refresh()
			return a, tea.Batch(cmd, progressTick())
		}
	}
	return a, nil
}

func (a *demoApp) View() tea.View {
	dim := exui.Dim()
	line := dim.Render(strings.Repeat("·", max(a.w, 1)))
	paneH := max(a.h-a.chrome.Height(), 1)
	rows := make([]string, 0, paneH)
	rows = append(rows, dim.Render(fmt.Sprintf("notifications in history: %d", a.chrome.Manager().Count())))
	for len(rows) < paneH {
		rows = append(rows, line)
	}
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
	a.chrome.Frame(&v, a.h)
	return v
}

// New builds the status page.
func New() exui.Page { return newDemo() }

// Result is always empty: this page demonstrates notification surfaces
// rather than producing a value.
func (a *demoApp) Result() []exui.Field { return nil }

// Shell exposes this page's chrome to the tour host.
func (a *demoApp) Shell() *exui.Chrome { return a.chrome }
