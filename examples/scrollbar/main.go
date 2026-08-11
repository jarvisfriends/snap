// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package scrollbar implements the `snap_input scrollbar` subcommand, demoing snap/scrollbar's three presets side by side over
// the same scrolling text: Smooth (sub-cell glide), Line (thin default), and
// Classic (retro blocks). Wheel or arrows scroll; clicking or dragging on
// any bar jumps the view there (scrollbar.OffsetAt); q quits. It is a
// display-only demo (no value is written to stdout); --no-help hides the
// status bar.
package scrollbar

import (
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/scrollbar"
)

const totalLines = 120

type demoApp struct {
	offset int
	w, h   int
	chrome *exui.Chrome
	// barCols are the screen columns of the three rendered bars, recorded by
	// View so onMouse can hit-test clicks and drags against them.
	barCols [3]int
}

func (a *demoApp) Init() tea.Cmd { return nil }

func (a *demoApp) visible() int { return max(a.h-1-a.chrome.Height(), 4) }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
	case tea.KeyPressMsg:
		// Unhandled keys are ignored: quit is the shell's (q/ctrl+c), and in
		// the tour any other key would otherwise end the whole run.
		switch msg.String() {
		case "up":
			a.offset--
		case "down":
			a.offset++
		case "pgup":
			a.offset -= a.visible()
		case "pgdown":
			a.offset += a.visible()
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
	dim := exui.Dim()

	// Line numbers right-align in a fixed cell-width column via lipgloss —
	// no printf byte padding.
	numStyle := lipgloss.NewStyle().Width(4).Align(lipgloss.Right)
	lines := make([]string, 0, visible)
	for i := a.offset; i < min(a.offset+visible, totalLines); i++ {
		marker := "  "
		if i%10 == 0 {
			marker = "──"
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top,
			numStyle.Render(strconv.Itoa(i+1)), " ", dim.Render(marker), " scrolling content"))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)

	bar := func(p scrollbar.Preset) string {
		st := scrollbar.DefaultStyles()
		st.Preset = p
		return scrollbar.Vertical(totalLines, visible, a.offset, visible, st)
	}
	// A two-cell blank block between columns — a lipgloss gap, not spaces.
	gap := lipgloss.NewStyle().Width(2).Render("")
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		content, gap,
		bar(scrollbar.PresetSmooth), gap,
		bar(scrollbar.PresetLine), gap,
		bar(scrollbar.PresetClassic),
	)
	// Bars sit after the content block and a 2-col gap, 3 columns apart
	// (1-col bar + 2-col gap); onMouse hit-tests against these columns.
	contentW := lipgloss.Width(content)
	a.barCols = [3]int{contentW + 2, contentW + 5, contentW + 8}
	labels := dim.Render("bars: smooth · line · classic")
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, body, labels))
	v.OnMouse = a.onMouse
	a.chrome.Frame(&v, a.h)
	return v
}

// New builds the scrollbar page.
func New() exui.Page {
	return &demoApp{chrome: exui.NewChrome(
		exui.Bind("wheel ↑↓", "move"),
		exui.Bind("pgup pgdn", "page"),
		exui.Bind("click drag bar", "jump"),
		exui.Bind("q", "quit"),
	)}
}

// Result is always empty: this page demonstrates scrolling rather than
// producing a value.
func (a *demoApp) Result() []exui.Field { return nil }

// Shell exposes this page's chrome to the tour host.
func (a *demoApp) Shell() *exui.Chrome { return a.chrome }
