// Command scrollbar demos snap/scrollbar's three presets side by side over
// the same scrolling text: Smooth (sub-cell glide), Line (thin default), and
// Classic (retro blocks). Wheel or arrows scroll; clicking or dragging on
// any bar jumps the view there (scrollbar.OffsetAt); q quits.
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
	// barCols are the screen columns of the three rendered bars, recorded by
	// View so onMouse can hit-test clicks and drags against them.
	barCols [3]int
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
	switch ev := mm.(type) {
	case tea.MouseWheelMsg:
		switch ev.Button {
		case tea.MouseWheelUp:
			a.offset--
		case tea.MouseWheelDown:
			a.offset++
		}
		a.offset = scrollbar.ClampOffset(a.offset, totalLines, a.visible())
	case tea.MouseClickMsg, tea.MouseMotionMsg:
		// Click a bar to jump; keep dragging (motion with the button held)
		// to scrub. OffsetAt maps the bar row back to a scroll offset.
		me := mm.Mouse()
		if me.Button != tea.MouseLeft {
			return nil
		}
		for _, col := range a.barCols {
			if me.X == col {
				a.offset = scrollbar.OffsetAt(me.Y-1, a.visible(), totalLines, a.visible())
				break
			}
		}
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
	// Bars sit after the content block and a 2-col gap, 3 columns apart
	// (1-col bar + 2-col gap); onMouse hit-tests against these columns.
	contentW := lipgloss.Width(content)
	a.barCols = [3]int{contentW + 2, contentW + 5, contentW + 8}
	header := dim.Render("wheel/↑↓/PgUp/PgDn scroll · click/drag a bar — smooth · line · classic — q quits")
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
