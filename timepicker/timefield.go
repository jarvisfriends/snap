package timepicker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	SideSeconds
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
	// ShowSeconds adds a third seconds column. Set it before the first View.
	ShowSeconds bool

	// base carries the date (and location) the clock values are applied to,
	// so Time() round-trips the full timestamp handed to NewTimeField or
	// SetTime with only the clock edited.
	base                 time.Time
	hour, minute, second int

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

	// zones holds the named hit zones recorded during View: the two column
	// cells and the visible dropdown rows, built from the same blocks the
	// View renders (uifx.Zones).
	zones *uifx.Zones
}

// NewTimeField returns a time field editing t's clock. The date part of t
// is preserved: Time() returns it with the edited hour/minute/second, so the
// field pairs naturally with the datepicker when editing a full timestamp.
// Seconds are hidden until ShowSeconds is set.
func NewTimeField(t time.Time) *TimeFieldModel {
	m := &TimeFieldModel{
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
	m.SetTime(t)
	return m
}

// SetTime replaces the edited timestamp: the clock loads into the columns
// and the date part is kept for Time().
func (m *TimeFieldModel) SetTime(t time.Time) {
	m.base = t
	m.hour, m.minute, m.second = t.Clock()
}

// Time returns the edited timestamp: the date (and location) given to
// NewTimeField/SetTime with the current hour, minute, and second.
func (m *TimeFieldModel) Time() time.Time {
	y, mo, d := m.base.Date()
	return time.Date(y, mo, d, m.hour, m.minute, m.second, 0, m.base.Location())
}

// sides lists the visible columns in order.
func (m *TimeFieldModel) sides() []Side {
	if m.ShowSeconds {
		return []Side{SideHours, SideMinutes, SideSeconds}
	}
	return []Side{SideHours, SideMinutes}
}

// lastSide is the right-most visible column.
func (m *TimeFieldModel) lastSide() Side {
	if m.ShowSeconds {
		return SideSeconds
	}
	return SideMinutes
}

// Zone IDs for the field's interactive regions (see uifx.Zones).
const (
	zoneHours   = "hours"
	zoneMinutes = "minutes"
	zoneSeconds = "seconds"
	zoneRow     = "row-" // + visible dropdown row index
)

// zoneFor is the hit-zone ID of a column.
func zoneFor(s Side) string {
	switch s {
	case SideHours:
		return zoneHours
	case SideMinutes:
		return zoneMinutes
	case SideSeconds:
		return zoneSeconds
	}
	return ""
}

// sideForZone is the inverse of zoneFor ( -1 when id is not a column).
func sideForZone(id string) Side {
	switch id {
	case zoneHours:
		return SideHours
	case zoneMinutes:
		return SideMinutes
	case zoneSeconds:
		return SideSeconds
	}
	return -1
}

// dropRow parses a dropdown-row zone ID ("row-3") into its visible index.
func dropRow(id string) (int, bool) {
	num, ok := strings.CutPrefix(id, zoneRow)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(num)
	return n, err == nil
}

func sideMax(s Side) int {
	if s == SideHours {
		return 23
	}
	return 59
}

func (m *TimeFieldModel) value(s Side) int {
	switch s {
	case SideMinutes:
		return m.minute
	case SideSeconds:
		return m.second
	case SideHours:
		return m.hour
	default:
		return m.hour
	}
}

func (m *TimeFieldModel) setValue(s Side, v int) {
	v = geom.Clamp(v, 0, sideMax(s))
	switch s {
	case SideMinutes:
		m.minute = v
	case SideSeconds:
		m.second = v
	case SideHours:
		m.hour = v
	default:
		m.hour = v
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
	}
	return m, nil
}

// handleMotion tracks drags (dropdown highlight follows a held left button,
// LevelMedium+) and hover (LevelHigh: hovered column or dropdown row).
func (m *TimeFieldModel) handleMotion(me tea.Mouse) tea.Cmd {
	id := m.zones.Hit(me.X, me.Y)
	if me.Button == tea.MouseLeft {
		if m.Effects.Drag() && m.open >= 0 {
			if i, ok := dropRow(id); ok {
				m.cursor = m.top + i
			}
		}
		return nil
	}
	if !m.Effects.Hover() {
		return nil
	}
	m.hoverSide, m.hoverRow = -1, -1
	switch {
	case m.open >= 0:
		if i, ok := dropRow(id); ok {
			m.hoverRow = m.top + i
		}
	default:
		m.hoverSide = sideForZone(id)
	}
	return nil
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
		m.focusSide(min(m.Focused+1, m.lastSide()))
	case key.Matches(msg, m.KeyMap.PrevField):
		m.closeDropdown()
		m.focusSide(max(m.Focused-1, SideHours))
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
func (m *TimeFieldModel) handleClick(me tea.Mouse) tea.Cmd {
	if me.Button != tea.MouseLeft {
		return nil
	}
	id := m.zones.Hit(me.X, me.Y)
	if i, ok := dropRow(id); ok && m.open >= 0 {
		m.setValue(m.open, m.top+i)
		m.closeDropdown()
		return nil
	}
	if side := sideForZone(id); side >= 0 {
		m.openDropdown(side)
		return nil
	}
	if m.open >= 0 {
		// Click elsewhere closes the dropdown without committing.
		m.closeDropdown()
	}
	return nil
}

// handleWheel: vertical scroll moves the open dropdown window (or spins the
// focused column when closed); horizontal wheel hops between the hour and
// minute columns.
func (m *TimeFieldModel) handleWheel(me tea.Mouse) tea.Cmd {
	switch me.Button {
	case tea.MouseWheelLeft:
		m.closeDropdown()
		m.focusSide(max(m.Focused-1, SideHours))
		return nil
	case tea.MouseWheelRight:
		m.closeDropdown()
		m.focusSide(min(m.Focused+1, m.lastSide()))
		return nil
	}
	delta := 1
	if me.Button == tea.MouseWheelUp {
		delta = -1
	}
	if m.open >= 0 {
		m.top = geom.Clamp(m.top+delta, 0, sideMax(m.open)+1-dropdownVisibleRows)
		return nil
	}
	m.setValue(m.Focused, m.value(m.Focused)-delta)
	return nil
}

// displayValue is the column's text, showing the in-progress digit buffer on
// the focused column.
func (m *TimeFieldModel) displayValue(s Side) string {
	if s == m.Focused && m.typed != "" {
		return fmt.Sprintf("%2s", m.typed)
	}
	return fmt.Sprintf("%02d", m.value(s))
}

// onMouse is the View.OnMouse entry point: mouse events dispatch straight to
// the handler methods, never through Update, so hosts (and the Bubble Tea
// runtime) deliver pointer input through exactly one door. Parents hosting
// this component should call onMouse with translated coordinates.
func (m *TimeFieldModel) onMouse(msg tea.MouseMsg) tea.Cmd {
	return uifx.MouseHandlers{
		Click:  m.handleClick,
		Wheel:  m.handleWheel,
		Motion: m.handleMotion,
	}.OnMouse(msg)
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

	// Build the colon-separated column cells; hit zones come from the same
	// blocks this frame renders (dropdown rows below via renderDropdown).
	colon := m.ColonStyle.Render(":")
	cw := lipgloss.Width(colon)
	var (
		blocks   []string
		layers   []*lipgloss.Layer
		x, cellH int
		openX    int // x offset of the open side's cell, for the dropdown
	)
	for i, side := range m.sides() {
		if i > 0 {
			blocks = append(blocks, colon)
			x += cw
		}
		cell := styleFor(side).Render(m.displayValue(side))
		blocks = append(blocks, cell)
		layers = append(layers, lipgloss.NewLayer(cell).ID(zoneFor(side)).X(x))
		if side == m.open {
			openX = x
		}
		x += lipgloss.Width(cell)
		cellH = lipgloss.Height(cell)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, blocks...)

	parts := []string{row}
	if m.open >= 0 {
		drop, rowLayers := m.renderDropdown(cellH, openX)
		parts = append(parts, drop)
		layers = append(layers, rowLayers...)
	} else {
		parts = append(parts, m.HelpStyle.Render("type/↑↓ set • space/click list • enter save"))
	}
	m.zones = uifx.NewZones(layers...)
	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left, parts...))
	v.OnMouse = m.onMouse
	return v
}

// renderDropdown draws the open side's scrollable value list under its
// column (indent = that cell's x offset) and returns the visible rows'
// hit-zone layers alongside it.
func (m *TimeFieldModel) renderDropdown(cellH, indent int) (string, []*lipgloss.Layer) {
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

	// Row hit zones: the rendered row blocks, placed inside the list border
	// (+1,+1) exactly where the composed frame shows them.
	rowLayers := make([]*lipgloss.Layer, 0, len(rows))
	for i, r := range rows {
		rowLayers = append(rowLayers,
			lipgloss.NewLayer(r).ID(fmt.Sprintf("%s%d", zoneRow, i)).
				X(indent+1).Y(cellH+1+i))
	}
	if indent > 0 {
		return lipgloss.NewStyle().MarginLeft(indent).Render(list), rowLayers
	}
	return list, rowLayers
}
