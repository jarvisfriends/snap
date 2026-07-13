// Command dependencies demos snap/dependencies rendered by snap/status's
// InfoModal: the running binary's build info (Go version, OS, VCS revision)
// above a scrollable dependency list read via dependencies.ExpandedBuildInfo.
// The wheel or ↑/↓/PgUp/PgDn scroll the list (mouse arrives through the
// modal's HandleMouse — one call, no host hit-testing), Esc or a click
// outside closes the modal, i reopens it, q quits.
package main

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/dependencies"
	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/status"
)

type demoApp struct {
	modal  *status.InfoModal
	chrome *exui.Chrome
	w, h   int
}

func newDemo() *demoApp {
	m := status.NewInfoModal()
	m.SetAppName("snap dependencies demo")
	if info := dependencies.ExpandedBuildInfo(); info != nil && info.App.Version != "" {
		m.SetVersion(info.App.Version)
	}
	return &demoApp{modal: m, chrome: exui.NewChrome(
		exui.Bind("i", "info modal"),
		exui.Bind("wheel/↑/↓", "scroll"),
		exui.Bind("esc/outside click", "close"),
		exui.Bind("q", "quit"),
	)}
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetWidth(msg.Width)
		if !a.modal.IsVisible() {
			a.modal.Open(a.w, a.h)
		}
	case status.CloseInfoModalMsg:
		return a, nil
	case tea.KeyPressMsg:
		if a.modal.IsVisible() {
			// The open modal owns the keyboard (↑/↓, PgUp/PgDn, Home/End,
			// Esc); everything else falls through to it harmlessly.
			m, cmd := a.modal.Update(msg)
			if im, ok := m.(*status.InfoModal); ok {
				a.modal = im
			}
			return a, cmd
		}
		switch msg.String() {
		case "i":
			a.modal.Open(a.w, a.h)
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

// onMouse forwards pointer input to the modal's HandleMouse: the wheel
// scrolls the dependency list, a click outside closes.
func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	cmd, _ := a.modal.HandleMouse(mm)
	return cmd
}

func (a *demoApp) View() tea.View {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	line := dim.Render(strings.Repeat("·", max(a.w, 1)))
	paneH := max(a.h-a.chrome.Height(), 1)
	rows := make([]string, 0, paneH)
	for range paneH {
		rows = append(rows, line)
	}
	base := lipgloss.JoinVertical(lipgloss.Left, rows...)

	content := base
	if a.modal.IsVisible() {
		bx, by, _, _ := a.modal.Bounds()
		content = lipgloss.NewCompositor(
			lipgloss.NewLayer(base),
			lipgloss.NewLayer(a.modal.View().Content).X(bx).Y(by).Z(10),
		).Render()
	}

	v := tea.NewView(content)
	a.chrome.Apply(&v, a.h)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.OnMouse = a.onMouse
	return v
}

func main() {
	exui.Init()
	if _, err := exui.Program(newDemo()).Run(); err != nil {
		exui.Fatal(err)
	}
}
