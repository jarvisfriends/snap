package timepicker

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/geom"
	"github.com/jarvisfriends/snap/uifx"
)

// Side identifies one of the two time columns.
type Side int

const (
	SideHours Side = iota
	SideMinutes
)

// dropdownVisibleRows is how many values the open dropdown shows at once.
const dropdownVisibleRows = 7

// TimeFieldKeyMap holds the key bindings for the two-column time field.
type TimeFieldKeyMap struct {
	NextField key.Binding
	PrevField key.Binding
	Up        key.Binding
	Down      key.Binding
	Open      key.Binding
	Submit    key.Binding
	Quit      key.Binding
}

// DefaultTimeFieldKeyMap returns the standard bindings.
func DefaultTimeFieldKeyMap() TimeFieldKeyMap {
	return TimeFieldKeyMap{
		NextField: key.NewBinding(key.WithKeys("tab", "right")),
		PrevField: key.NewBinding(key.WithKeys("shift+tab", "left")),
		Up:        key.NewBinding(key.WithKeys("up")),
		Down:      key.NewBinding(key.WithKeys("down")),
		Open:      key.NewBinding(key.WithKeys("space")),
		Submit:    key.NewBinding(key.WithKeys("enter")),
		Quit:      key.NewBinding(key.WithKeys("esc", "ctrl+c")),
	}
}

// TimeFieldModel is the redesigned time picker (tui-base ROADMAP SP-8/Q-24):
// two columns — hours and minutes — separated by a highlighted colon. Clicking
// a column (or pressing space) opens a scrollable dropdown of its valid
// values; a click or Enter commits the highlighted value. Digits typed into
// the focused column edit it directly, and the value is validated (clamped
// into range) whenever the column loses focus.
//
// Mouse events must arrive with coordinates relative to the component's
// top-left cell (the standard tui-base overlay convention); the hit zones are
// recorded during View.
type TimeFieldModel struct {
	Hour   int // 0–23
	Minute int // 0–59

	KeyMap  TimeFieldKeyMap
	Focused Side
	Done    bool
	Aborted bool

	// Style hooks (theme-free; consumers map their palette on).
	ActiveStyle   lipgloss.Style // focused column cell
	InactiveStyle lipgloss.Style // unfocused column cell
	ColonStyle    lipgloss.Style // the ":" separator — the highlight color
	ListStyle     lipgloss.Style // dropdown frame
	SelectedStyle lipgloss.Style // dropdown highlighted row
	RowStyle      lipgloss.Style // dropdown regular row
	HelpStyle     lipgloss.Style

	// open is the side whose dropdown is showing, or -1 for none.
	open Side
	// cursor is the highlighted dropdown value.
	cursor int
	// top is the first visible dropdown row (scroll window).
	top int
	// typed buffers digits entered into the focused column; committed (with
	// validation) when the column loses focus or the buffer fills.
	typed string

	// Effects selects the interaction-feedback tier (see uifx.Level).
	Effects uifx.Level
	// hoverSide is the column under the pointer (-1 none; LevelHigh).
	hoverSide Side
	// hoverRow is the dropdown value under the pointer (-1 none; LevelHigh).
	hoverRow int

	// Hit zones recorded during View (component-relative).
	hourRect, minuteRect geom.Rect
	rowRects             []geom.Rect // visible dropdown rows, top first
}

// NewTimeField returns a two-column time field initialized to hour:minute
// (values are clamped into range).
func NewTimeField(hour, minute int) *TimeFieldModel {
	m := &TimeFieldModel{
		Hour:      geom.Clamp(hour, 0, 23),
		Minute:    geom.Clamp(minute, 0, 59),
		KeyMap:    DefaultTimeFieldKeyMap(),
		Focused:   SideHours,
		open:      -1,
		hoverSide: -1,
		hoverRow:  -1,
		ActiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()),
		InactiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()),
		ColonStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true),
		ListStyle:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
		SelectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Padding(0, 1),
		RowStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Padding(0, 1),
		HelpStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
	return m
}

func sideMax(s Side) int {
	if s == SideHours {
		return 23
	}
	return 59
}

func (m *TimeFieldModel) value(s Side) int {
	if s == SideHours {
		return m.Hour
	}
	return m.Minute
}

func (m *TimeFieldModel) setValue(s Side, v int) {
	v = geom.Clamp(v, 0, sideMax(s))
	if s == SideHours {
		m.Hour = v
	} else {
		m.Minute = v
	}
}

// DropdownOpen reports whether a dropdown is showing and for which side.
func (m *TimeFieldModel) DropdownOpen() (Side, bool) { return m.open, m.open >= 0 }

// openDropdown shows side's value list with the current value highlighted.
func (m *TimeFieldModel) openDropdown(s Side) {
	m.commitTyped()
	m.Focused = s
	m.open = s
	m.cursor = m.value(s)
	m.top = geom.Clamp(m.cursor-dropdownVisibleRows/2, 0, sideMax(s)+1-dropdownVisibleRows)
}

func (m *TimeFieldModel) closeDropdown() { m.open = -1 }

// commitTyped validates and applies the digit buffer to the focused column —
// the "verify when we leave the field" rule. Values outside the range clamp
// into it (e.g. minutes "75" become 59).
func (m *TimeFieldModel) commitTyped() {
	if m.typed == "" {
		return
	}
	if n, err := strconv.Atoi(m.typed); err == nil {
		m.setValue(m.Focused, n)
	}
	m.typed = ""
}

// focusSide moves focus, validating the column being left.
func (m *TimeFieldModel) focusSide(s Side) {
	if s != m.Focused {
		m.commitTyped()
		m.Focused = s
	}
}

func (m *TimeFieldModel) Init() tea.Cmd { return nil }

func (m *TimeFieldModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m.handleKey(msg)
	case tea.MouseClickMsg:
		m.handleClick(msg.Mouse())
	case tea.MouseWheelMsg:
		m.handleWheel(msg.Mouse())
	case tea.MouseMotionMsg:
		m.handleMotion(msg.Mouse())
	}
	return m, nil
}

// handleMotion tracks drags (dropdown highlight follows a held left button,
// LevelMedium+) and hover (LevelHigh: hovered column or dropdown row).
func (m *TimeFieldModel) handleMotion(me tea.Mouse) {
	if me.Button == tea.MouseLeft {
		if !m.Effects.Drag() || m.open < 0 {
			return
		}
		for i, r := range m.rowRects {
			if r.Contains(me.X, me.Y) {
				m.cursor = m.top + i
			}
		}
		return
	}
	if !m.Effects.Hover() {
		return
	}
	m.hoverSide, m.hoverRow = -1, -1
	switch {
	case m.open >= 0:
		for i, r := range m.rowRects {
			if r.Contains(me.X, me.Y) {
				m.hoverRow = m.top + i
			}
		}
	case m.hourRect.Contains(me.X, me.Y):
		m.hoverSide = SideHours
	case m.minuteRect.Contains(me.X, me.Y):
		m.hoverSide = SideMinutes
	}
}

func (m *TimeFieldModel) handleKey(msg tea.KeyPressMsg) {
	// Digit type-ahead into the focused column (dropdown closed): two digits
	// fill the column; validation runs when the buffer fills or focus leaves.
	if m.open < 0 && len(msg.Text) == 1 && msg.Text >= "0" && msg.Text <= "9" {
		m.typed += msg.Text
		if len(m.typed) >= 2 {
			m.commitTyped()
		}
		return
	}

	switch {
	case key.Matches(msg, m.KeyMap.Quit):
		if m.open >= 0 {
			m.closeDropdown()
			return
		}
		m.commitTyped()
		m.Aborted = true
	case key.Matches(msg, m.KeyMap.Submit):
		if m.open >= 0 {
			m.setValue(m.open, m.cursor)
			m.closeDropdown()
			return
		}
		m.commitTyped()
		m.Done = true
	case key.Matches(msg, m.KeyMap.Open):
		if m.open >= 0 {
			m.closeDropdown()
		} else {
			m.openDropdown(m.Focused)
		}
	case key.Matches(msg, m.KeyMap.NextField):
		m.closeDropdown()
		m.focusSide(SideMinutes)
	case key.Matches(msg, m.KeyMap.PrevField):
		m.closeDropdown()
		m.focusSide(SideHours)
	case key.Matches(msg, m.KeyMap.Up):
		if m.open >= 0 {
			m.moveCursor(-1)
		} else {
			m.setValue(m.Focused, m.value(m.Focused)+1) // spinner behavior stays
		}
	case key.Matches(msg, m.KeyMap.Down):
		if m.open >= 0 {
			m.moveCursor(1)
		} else {
			m.setValue(m.Focused, m.value(m.Focused)-1)
		}
	}
}

func (m *TimeFieldModel) moveCursor(delta int) {
	m.cursor = geom.Clamp(m.cursor+delta, 0, sideMax(m.open))
	// Keep the cursor inside the scroll window.
	if m.cursor < m.top {
		m.top = m.cursor
	}
	if m.cursor >= m.top+dropdownVisibleRows {
		m.top = m.cursor - dropdownVisibleRows + 1
	}
}

// handleClick routes component-relative clicks: a column cell opens (or
// focuses) its dropdown; a dropdown row commits its value.
func (m *TimeFieldModel) handleClick(me tea.Mouse) {
	if me.Button != tea.MouseLeft {
		return
	}
	if m.open >= 0 {
		for i, r := range m.rowRects {
			if r.Contains(me.X, me.Y) {
				m.setValue(m.open, m.top+i)
				m.closeDropdown()
				return
			}
		}
	}
	switch {
	case m.hourRect.Contains(me.X, me.Y):
		m.openDropdown(SideHours)
	case m.minuteRect.Contains(me.X, me.Y):
		m.openDropdown(SideMinutes)
	default:
		if m.open >= 0 {
			// Click elsewhere closes the dropdown without committing.
			m.closeDropdown()
		}
	}
}

// handleWheel: vertical scroll moves the open dropdown window (or spins the
// focused column when closed); horizontal wheel hops between the hour and
// minute columns.
func (m *TimeFieldModel) handleWheel(me tea.Mouse) {
	switch me.Button {
	case tea.MouseWheelLeft:
		m.closeDropdown()
		m.focusSide(SideHours)
		return
	case tea.MouseWheelRight:
		m.closeDropdown()
		m.focusSide(SideMinutes)
		return
	}
	delta := 1
	if me.Button == tea.MouseWheelUp {
		delta = -1
	}
	if m.open >= 0 {
		m.top = geom.Clamp(m.top+delta, 0, sideMax(m.open)+1-dropdownVisibleRows)
		return
	}
	m.setValue(m.Focused, m.value(m.Focused)-delta)
}

// displayValue is the column's text, showing the in-progress digit buffer on
// the focused column.
func (m *TimeFieldModel) displayValue(s Side) string {
	if s == m.Focused && m.typed != "" {
		return fmt.Sprintf("%2s", m.typed)
	}
	return fmt.Sprintf("%02d", m.value(s))
}

func (m *TimeFieldModel) View() tea.View {
	styleFor := func(s Side) lipgloss.Style {
		if m.Focused == s {
			return m.ActiveStyle
		}
		if m.Effects.Hover() && m.hoverSide == s {
			return m.InactiveStyle.Underline(true)
		}
		return m.InactiveStyle
	}

	hourCell := styleFor(SideHours).Render(m.displayValue(SideHours))
	colon := m.ColonStyle.Render(":")
	minuteCell := styleFor(SideMinutes).Render(m.displayValue(SideMinutes))

	// Record component-relative hit zones for the two cells.
	hw, hh := lipgloss.Width(hourCell), lipgloss.Height(hourCell)
	cw := lipgloss.Width(colon)
	mw := lipgloss.Width(minuteCell)
	m.hourRect = geom.Rect{X: 0, Y: 0, W: hw, H: hh}
	m.minuteRect = geom.Rect{X: hw + cw, Y: 0, W: mw, H: hh}

	row := lipgloss.JoinHorizontal(lipgloss.Center, hourCell, colon, minuteCell)

	parts := []string{row}
	m.rowRects = nil
	if m.open >= 0 {
		parts = append(parts, m.renderDropdown(hh, hw, cw))
	} else {
		parts = append(parts, m.HelpStyle.Render("type/↑↓ set • space/click list • enter save"))
	}
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, parts...))
	v.OnMouse = uifx.RouteToUpdate(m.Update)
	return v
}

// renderDropdown draws the open side's scrollable value list under its column
// and records the visible rows' hit zones.
func (m *TimeFieldModel) renderDropdown(cellH, hourW, colonW int) string {
	lastTop := sideMax(m.open) + 1 - dropdownVisibleRows
	m.top = geom.Clamp(m.top, 0, lastTop)

	rows := make([]string, 0, dropdownVisibleRows)
	for i := range dropdownVisibleRows {
		v := m.top + i
		if v > sideMax(m.open) {
			break
		}
		st := m.RowStyle
		switch {
		case v == m.cursor:
			st = m.SelectedStyle
		case m.Effects.Hover() && v == m.hoverRow:
			st = m.RowStyle.Underline(true)
		}
		rows = append(rows, st.Render(fmt.Sprintf("%02d", v)))
	}
	list := m.ListStyle.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))

	// Align the list under its column: hours at x=0, minutes after hour+colon.
	indent := 0
	if m.open == SideMinutes {
		indent = hourW + colonW
	}

	// Row hit zones: inside the list border (+1,+1), one per visible row.
	listW := lipgloss.Width(list)
	for i := range rows {
		m.rowRects = append(m.rowRects, geom.Rect{
			X: indent + 1,
			Y: cellH + 1 + i,
			W: listW - 2,
			H: 1,
		})
	}
	if indent > 0 {
		return lipgloss.NewStyle().MarginLeft(indent).Render(list)
	}
	return list
}
