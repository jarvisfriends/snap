// Command status demos snap/status + snap/notifications end to end: the
// status bar (key help on the left, live segments and a summary on the
// right) with notification toasts, a progress notification that fills as a
// fake download runs, and the ctrl+n history panel with severity badges.
// It is a display-only demo (no value is written to stdout); --no-help is
// accepted for consistency but this demo IS the status bar, so it only
// hides nothing extra.
package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/notifications"
	"github.com/jarvisfriends/snap/status"
)

// progressTickMsg advances the fake download's progress notification.
type progressTickMsg struct{}

func progressTick() tea.Cmd {
	return tea.Tick(180*time.Millisecond, func(time.Time) tea.Msg { return progressTickMsg{} })
}

type demoApp struct {
	bar   *status.BarModel
	mgr   *notifications.Manager
	pct   float64
	dlID  int64
	w, h  int
	start time.Time
}

func newDemo() *demoApp {
	mgr := notifications.NewManager()
	bar := status.New()
	bar.SetNotifManager(mgr)
	bar.SetPageBindings(exui.Bindings{
		exui.Bind("i/w/e", "info/warn/error"),
		exui.Bind("p", "progress"),
		exui.Bind("ctrl+n", "history"),
		exui.Bind("q", "quit"),
	})
	a := &demoApp{bar: bar, mgr: mgr, dlID: -1, start: time.Now()}
	// Right-aligned segments re-evaluate every render — the consumer hook
	// for live widgets like a git branch or connection state.
	bar.SetSegment("branch", func() string { return " master" })
	bar.SetSegment("uptime", func() string {
		return time.Since(a.start).Truncate(time.Second).String()
	})
	return a
}

func (a *demoApp) Init() tea.Cmd { return a.bar.Init() }

func (a *demoApp) notify(content string, sev notifications.Severity) tea.Cmd {
	_, cmd := a.mgr.Add(content, sev, sev.DefaultTTL())
	a.refresh()
	return cmd
}

// refresh re-renders the bar (toasts live inside it) after manager changes.
func (a *demoApp) refresh() { a.bar.SetWidth(a.w) }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.bar.SetWidth(msg.Width)
		return a, nil

	case progressTickMsg:
		if a.dlID < 0 {
			return a, nil
		}
		a.pct += 7
		a.mgr.SetProgress(a.dlID, a.pct)
		if a.pct >= 100 {
			a.dlID = -1
			cmd := a.notify("download complete", notifications.SeverityInfo)
			return a, cmd
		}
		a.refresh()
		return a, progressTick()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return a, tea.Quit
		case "i":
			return a, a.notify("deploy finished cleanly", notifications.SeverityInfo)
		case "w":
			return a, a.notify("disk 82% full on /var", notifications.SeverityWarning)
		case "e":
			return a, a.notify("backup job failed (exit 3)", notifications.SeverityError)
		case "p":
			if a.dlID >= 0 {
				return a, nil
			}
			a.pct = 0
			zero := 0.0
			n, cmd := a.mgr.AddWithOptions("downloading assets", notifications.SeverityInfo, 0,
				notifications.AddOptions{Key: "download", Percent: &zero, RetainInHistory: true})
			a.dlID = n.ID
			a.refresh()
			return a, tea.Batch(cmd, progressTick())
		case "ctrl+n":
			cmd := a.bar.ToggleNotifications()
			a.refresh()
			return a, cmd
		case "up":
			a.bar.NotifHistoryCursorUp()
			return a, nil
		case "down":
			a.bar.NotifHistoryCursorDown(1)
			return a, nil
		}
	}
	// Everything else (toast TTL expiries, animation ticks) belongs to the
	// bar's own machinery.
	m, cmd := a.bar.Update(msg)
	if b, ok := m.(*status.BarModel); ok {
		a.bar = b
	}
	a.refresh()
	return a, cmd
}

func (a *demoApp) View() tea.View {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	line := dim.Render(strings.Repeat("·", max(a.w, 1)))
	paneH := max(a.h-a.bar.Height(), 1)
	rows := make([]string, 0, paneH+1)
	rows = append(rows, dim.Render(fmt.Sprintf("notifications in history: %d", a.mgr.Count())))
	for len(rows) < paneH {
		rows = append(rows, line)
	}
	base := lipgloss.JoinVertical(lipgloss.Left, append(rows, a.bar.View().Content)...)

	// The history panel composites above the page, anchored by the router in
	// real apps; here the demo is the router.
	if overlay := a.bar.RenderHistoryOverlay(a.w, a.h); overlay != "" {
		base = lipgloss.NewCompositor(
			lipgloss.NewLayer(base),
			lipgloss.NewLayer(overlay).X(max(a.w-lipgloss.Width(overlay)-1, 0)).
				Y(max(a.h-lipgloss.Height(overlay)-1, 0)).Z(5),
		).Render()
	}

	v := tea.NewView(base)
	v.AltScreen = true
	return v
}

func main() {
	exui.Init()
	if _, err := exui.Program(newDemo()).Run(); err != nil {
		exui.Fatal(err)
	}
}
