// Command scrollbar demos snap/scrollbar's three presets side by side over
// the same scrolling text: Smooth (sub-cell glide), Line (thin default), and
// Classic (retro blocks). Wheel or arrows scroll; q quits.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/scrollbar"
)

const totalLines = 120

type demoApp struct {
	offset int
	w, h   int
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) visible() int { return max(a.h-3, 4) }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up":
			a.offset--
		case "down":
			a.offset++
		case "pgup":
			a.offset -= a.visible()
		case "pgdown":
			a.offset += a.visible()
		default:
			return a, tea.Quit
		}
		a.offset = scrollbar.ClampOffset(a.offset, totalLines, a.visible())
	}
	return a, nil
}

func (a *demoApp) onMouse(mm tea.MouseMsg) tea.Cmd {
	if wheel, ok := mm.(tea.MouseWheelMsg); ok {
		switch wheel.Button {
		case tea.MouseWheelUp:
			a.offset--
		case tea.MouseWheelDown:
			a.offset++
		}
		a.offset = scrollbar.ClampOffset(a.offset, totalLines, a.visible())
	}
	return nil
}

func (a *demoApp) View() tea.View {
	visible := a.visible()
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	lines := make([]string, 0, visible)
	for i := a.offset; i < min(a.offset+visible, totalLines); i++ {
		marker := "  "
		if i%10 == 0 {
			marker = "──"
		}
		lines = append(lines, fmt.Sprintf(" %3d %s scrolling content", i+1, dim.Render(marker)))
	}
	content := strings.Join(lines, "\n")

	bar := func(p scrollbar.Preset) string {
		st := scrollbar.DefaultStyles()
		st.Preset = p
		return scrollbar.Vertical(totalLines, visible, a.offset, visible, st)
	}
	gap := strings.Repeat(" ", 2)
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		content, gap,
		bar(scrollbar.PresetSmooth), gap,
		bar(scrollbar.PresetLine), gap,
		bar(scrollbar.PresetClassic),
	)
	header := dim.Render("wheel/↑↓/PgUp/PgDn scroll — bars: smooth · line · classic — q quits")
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, header, body))
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	v.OnMouse = a.onMouse
	return v
}

func main() {
	if _, err := tea.NewProgram(&demoApp{}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
