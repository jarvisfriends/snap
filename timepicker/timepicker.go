package timepicker

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/geom"
	"github.com/jarvisfriends/snap/uifx"
)

type Field int

const (
	FieldHours Field = iota
	FieldMinutes
	FieldSeconds
)

type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Submit key.Binding
	Quit   key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:     key.NewBinding(key.WithKeys("up")),
		Down:   key.NewBinding(key.WithKeys("down")),
		Left:   key.NewBinding(key.WithKeys("left", "shift+tab")),
		Right:  key.NewBinding(key.WithKeys("right", "tab")),
		Submit: key.NewBinding(key.WithKeys("enter")),
		Quit:   key.NewBinding(key.WithKeys("ctrl+c", "esc", "q")),
	}
}

type TimePickerModel struct {
	Duration time.Duration
	KeyMap   KeyMap
	Focused  Field
	Done     bool
	Aborted  bool

	ActiveStyle   lipgloss.Style
	InactiveStyle lipgloss.Style
	HelpStyle     lipgloss.Style

	// Effects selects the interaction-feedback tier (see uifx.Level).
	Effects uifx.Level
	// hoverSeg is the segment under the pointer (-1 none; LevelHigh).
	hoverSeg int

	// segRects are the h/m/s cells' content-relative hit zones, recorded
	// during View for click-to-focus.
	segRects [3]geom.Rect
}

func New(d time.Duration) *TimePickerModel {
	return &TimePickerModel{
		hoverSeg: -1,
		Duration: d,
		KeyMap:   DefaultKeyMap(),
		Focused:  FieldHours,
		ActiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()),
		InactiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()),
		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}

func (m *TimePickerModel) Init() tea.Cmd {
	return nil
}

func (m *TimePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			m.Aborted = true
			return m, nil
		case key.Matches(msg, m.KeyMap.Submit):
			m.Done = true
			return m, nil
		case key.Matches(msg, m.KeyMap.Left):
			m.Focused = (m.Focused - 1)
			if m.Focused < FieldHours {
				m.Focused = FieldSeconds
			}
		case key.Matches(msg, m.KeyMap.Right):
			m.Focused = (m.Focused + 1)
			if m.Focused > FieldSeconds {
				m.Focused = FieldHours
			}
		case key.Matches(msg, m.KeyMap.Up):
			m.increment(1)
		case key.Matches(msg, m.KeyMap.Down):
			m.increment(-1)
		}
	case tea.MouseWheelMsg:
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			m.increment(1)
		case tea.MouseWheelDown:
			m.increment(-1)
		case tea.MouseWheelLeft:
			if m.Focused > FieldHours {
				m.Focused--
			}
		case tea.MouseWheelRight:
			if m.Focused < FieldSeconds {
				m.Focused++
			}
		}

	case tea.MouseClickMsg:
		// Clicking a segment cell focuses it; the wheel then adjusts it.
		me := msg.Mouse()
		if me.Button == tea.MouseLeft {
			for i, r := range m.segRects {
				if r.Contains(me.X, me.Y) {
					m.Focused = Field(i)
					break
				}
			}
		}

	case tea.MouseMotionMsg:
		// Hover (LevelHigh): track the segment under the pointer.
		me := msg.Mouse()
		if me.Button == tea.MouseNone && m.Effects.Hover() {
			m.hoverSeg = -1
			for i, r := range m.segRects {
				if r.Contains(me.X, me.Y) {
					m.hoverSeg = i
					break
				}
			}
		}
	}
	return m, nil
}

func (m *TimePickerModel) increment(dir int) {
	hours := int64(m.Duration.Hours())
	minutes := int64(m.Duration.Minutes()) % 60
	seconds := int64(m.Duration.Seconds()) % 60

	switch m.Focused {
	case FieldHours:
		hours += int64(dir)
		if hours < 0 {
			hours = 0
		}
	case FieldMinutes:
		minutes += int64(dir)
		if minutes > 59 {
			minutes = 0
			hours++
		} else if minutes < 0 {
			minutes = 59
			if hours > 0 {
				hours--
			}
		}
	case FieldSeconds:
		seconds += int64(dir)
		if seconds > 59 {
			seconds = 0
			minutes++
			if minutes > 59 {
				minutes = 0
				hours++
			}
		} else if seconds < 0 {
			seconds = 59
			if minutes > 0 {
				minutes--
			} else if hours > 0 {
				minutes = 59
				hours--
			}
		}
	}
	m.Duration = time.Duration(
		hours,
	)*time.Hour + time.Duration(
		minutes,
	)*time.Minute + time.Duration(
		seconds,
	)*time.Second
}

func (m *TimePickerModel) View() tea.View {
	hours := int64(m.Duration.Hours())
	minutes := int64(m.Duration.Minutes()) % 60
	seconds := int64(m.Duration.Seconds()) % 60

	hStr := fmt.Sprintf("%02dh", hours)
	mStr := fmt.Sprintf("%02dm", minutes)
	sStr := fmt.Sprintf("%02ds", seconds)

	styleFor := func(f Field) lipgloss.Style {
		if m.Focused == f {
			return m.ActiveStyle
		}
		if m.Effects.Hover() && int(f) == m.hoverSeg {
			return m.InactiveStyle.Underline(true)
		}
		return m.InactiveStyle
	}

	title := lipgloss.NewStyle().Bold(true).Padding(0, 1).Render("Duration")

	cells := []string{
		styleFor(FieldHours).Render(hStr),
		styleFor(FieldMinutes).Render(mStr),
		styleFor(FieldSeconds).Render(sStr),
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, cells...)

	help := m.HelpStyle.
		MarginTop(1).
		Render("↑/↓: Adjust • ←/→: Select • Enter: Save")

	content := lipgloss.JoinVertical(lipgloss.Center, title, body, help)

	// Record the segment hit zones: cells sit under the title, offset by the
	// centering indent JoinVertical applies to the (narrower) body row.
	titleH := lipgloss.Height(title)
	offX := max((lipgloss.Width(content)-lipgloss.Width(body))/2, 0)
	x := offX
	for i, c := range cells {
		w := lipgloss.Width(c)
		m.segRects[i] = geom.Rect{X: x, Y: titleH, W: w, H: lipgloss.Height(c)}
		x += w
	}

	v := tea.NewView(content)
	v.OnMouse = uifx.RouteToUpdate(m.Update)
	return v
}
