// Package datepicker provides a Bubble Tea v2 component for viewing and selecting
// a date from a monthly view.
package datepicker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/uifx"
)

// Focus is a value passed to `model.SetFocus` to indicate what component
// controls should be available.
type Focus int

const (
	// FocusNone is a value passed to `model.SetFocus` to ignore all date altering key msgs
	FocusNone Focus = iota
	// FocusHeaderMonth is a value passed to `model.SetFocus` to accept key msgs that change the month
	FocusHeaderMonth
	// FocusHeaderYear is a value passed to `model.SetFocus` to accept key msgs that change the year
	FocusHeaderYear
	// FocusCalendar is a value passed to `model.SetFocus` to accept key msgs that change the week or date
	FocusCalendar
)

//go:generate stringer -type=Focus

// KeyMap is the key bindings for different actions within the datepicker.
type KeyMap struct {
	Up        key.Binding
	Right     key.Binding
	Down      key.Binding
	Left      key.Binding
	FocusPrev key.Binding
	FocusNext key.Binding
	Select    key.Binding
	Quit      key.Binding
}

// DefaultKeyMap returns a KeyMap struct with default values
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:        key.NewBinding(key.WithKeys("up")),
		Right:     key.NewBinding(key.WithKeys("right")),
		Down:      key.NewBinding(key.WithKeys("down")),
		Left:      key.NewBinding(key.WithKeys("left")),
		FocusPrev: key.NewBinding(key.WithKeys("shift+tab")),
		FocusNext: key.NewBinding(key.WithKeys("tab")),
		Select:    key.NewBinding(key.WithKeys("enter")),
		Quit:      key.NewBinding(key.WithKeys("ctrl+c", "q")),
	}
}

// Styles is a struct of lipgloss styles to apply to various elements of the datepicker
type Styles struct {
	Header lipgloss.Style
	Date   lipgloss.Style

	HeaderText   lipgloss.Style
	Text         lipgloss.Style
	SelectedText lipgloss.Style
	FocusedText  lipgloss.Style
}

// DefaultStyles returns a default `Styles` struct
func DefaultStyles() Styles {
	// TODO: refactor for adaptive colors
	r := lipgloss.NewStyle()
	return Styles{
		Header:       r.Padding(1, 0, 0),
		Date:         r.Padding(0, 1, 1),
		HeaderText:   r.Bold(true),
		Text:         r.Foreground(lipgloss.Color("247")),
		SelectedText: r.Reverse(true).Bold(true),
		FocusedText:  r.Bold(true).Foreground(lipgloss.Color("212")),
	}
}

// Model is a struct that contains the state of the datepicker component and satisfies
// the `tea.Model` interface
type DatePickerModel struct {
	// Time is the `time.Time` struct that represents the selected date month and year
	Time time.Time

	// KeyMap encodes the keybindings recognized by the model
	KeyMap KeyMap

	// Styles represent the Styles struct used to render the datepicker
	Styles Styles

	// Focused indicates the component which the end user is focused on
	Focused Focus

	// Selected indicates whether a date is Selected in the datepicker
	Selected bool

	// Effects selects the interaction-feedback tier (see uifx.Level).
	Effects uifx.Level

	// hoverDay is the date under the pointer (zero when none; LevelHigh).
	hoverDay time.Time

	// Mouse hit-zone geometry recorded during View (content-relative cells).
	// dayGrid[row][col] is the date shown in that cell (zero for blanks).
	dayGrid  [][]time.Time
	titleH   int
	gridTopY int
	gridOffX int
	cellW    int
	cellH    int
	totalW   int
}

// New returns the Model of the datepicker
func New(initial time.Time) *DatePickerModel {
	return &DatePickerModel{
		Time:   initial,
		KeyMap: DefaultKeyMap(),
		Styles: DefaultStyles(),

		Focused:  FocusCalendar,
		Selected: false,
	}
}

// Init satisfies the `tea.Model` interface. This sends a nil cmd
func (m *DatePickerModel) Init() tea.Cmd {
	return nil
}

// Update changes the state of the datepicker. Update satisfies the `tea.Model` interface
func (m *DatePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.KeyMap.Select):
			// Confirm the highlighted date. The embedding app decides what
			// happens next (the example prints it and quits).
			m.Selected = true

		case key.Matches(msg, m.KeyMap.Up):
			m.updateUp()

		case key.Matches(msg, m.KeyMap.Right):
			m.updateRight()

		case key.Matches(msg, m.KeyMap.Down):
			m.updateDown()

		case key.Matches(msg, m.KeyMap.Left):
			m.updateLeft()

		case key.Matches(msg, m.KeyMap.FocusPrev):
			switch m.Focused {
			case FocusHeaderYear:
				m.SetFocus(FocusHeaderMonth)
			case FocusCalendar:
				m.SetFocus(FocusHeaderYear)
			case FocusNone, FocusHeaderMonth:
				// no previous field to move to
			}

		case key.Matches(msg, m.KeyMap.FocusNext):
			switch m.Focused {
			case FocusHeaderMonth:
				m.SetFocus(FocusHeaderYear)
			case FocusHeaderYear:
				m.SetFocus(FocusCalendar)
			case FocusNone, FocusCalendar:
				// no next field to move to
			}
		}
	case tea.MouseClickMsg:
		m.handleClick(msg.Mouse())

	case tea.MouseWheelMsg:
		// Vertical wheel pages weeks (mirroring up/down); horizontal wheel
		// pages whole months, so the wheel alone can reach any date.
		switch msg.Mouse().Button {
		case tea.MouseWheelUp:
			m.Time = m.Time.AddDate(0, 0, -7)
		case tea.MouseWheelDown:
			m.Time = m.Time.AddDate(0, 0, 7)
		case tea.MouseWheelLeft:
			m.Time = m.Time.AddDate(0, -1, 0)
		case tea.MouseWheelRight:
			m.Time = m.Time.AddDate(0, 1, 0)
		}

	case tea.MouseMotionMsg:
		m.handleMotion(msg.Mouse())
	}
	return m, nil
}

// handleMotion tracks drags (the highlight follows a held left button
// across day cells, LevelMedium+) and hover (LevelHigh: the day under the
// pointer renders underlined so the click target reads before committing).
func (m *DatePickerModel) handleMotion(me tea.Mouse) {
	day := m.dayAt(me.X, me.Y)
	if me.Button == tea.MouseLeft {
		if m.Effects.Drag() && !day.IsZero() {
			m.Time = day
			m.SetFocus(FocusCalendar)
		}
		return
	}
	if m.Effects.Hover() {
		m.hoverDay = day
	}
}

// dayAt maps content-relative coordinates to the date in that grid cell
// (zero when outside the grid or on a blank cell).
func (m *DatePickerModel) dayAt(x, y int) time.Time {
	if y < m.gridTopY || m.cellW == 0 || m.cellH == 0 {
		return time.Time{}
	}
	col := (x - m.gridOffX) / m.cellW
	row := (y - m.gridTopY) / m.cellH
	if col < 0 || col > 6 || row < 0 || row >= len(m.dayGrid) || col >= len(m.dayGrid[row]) {
		return time.Time{}
	}
	return m.dayGrid[row][col]
}

// handleClick routes a content-relative left click: a day cell moves the
// highlight there (clicking the highlighted day again confirms it, setting
// Selected); the title line focuses the month (left half) or year (right).
func (m *DatePickerModel) handleClick(me tea.Mouse) {
	if me.Button != tea.MouseLeft {
		return
	}
	if me.Y < m.titleH {
		if me.X < m.totalW/2 {
			m.SetFocus(FocusHeaderMonth)
		} else {
			m.SetFocus(FocusHeaderYear)
		}
		return
	}
	day := m.dayAt(me.X, me.Y)
	if day.IsZero() {
		return
	}
	sameDay := day.Day() == m.Time.Day() && day.Month() == m.Time.Month() && day.Year() == m.Time.Year()
	if sameDay && m.Focused == FocusCalendar {
		m.Selected = true
		return
	}
	m.Time = day
	m.SetFocus(FocusCalendar)
}

func (m *DatePickerModel) updateUp() {
	switch m.Focused {
	case FocusHeaderYear:
		m.LastYear()
	case FocusHeaderMonth:
		m.LastMonth()
	case FocusCalendar:
		m.LastWeek()
	case FocusNone:
		// do nothing
	}
}

func (m *DatePickerModel) updateRight() {
	switch m.Focused {
	case FocusHeaderYear:
		// do nothing
	case FocusHeaderMonth:
		m.SetFocus(FocusHeaderYear)
	case FocusCalendar:
		m.Tomorrow()
	case FocusNone:
		// do nothing
	}
}

func (m *DatePickerModel) updateDown() {
	switch m.Focused {
	case FocusHeaderYear:
		m.NextYear()
	case FocusHeaderMonth:
		m.NextMonth()
	case FocusCalendar:
		m.NextWeek()
	case FocusNone:
		// do nothing
	}
}

func (m *DatePickerModel) updateLeft() {
	switch m.Focused {
	case FocusHeaderYear:
		m.SetFocus(FocusHeaderMonth)
	case FocusHeaderMonth:
		// do nothing
	case FocusCalendar:
		m.Yesterday()
	case FocusNone:
		// do nothing
	}
}

// View renders a month view as a multiline string in the bubbletea application.
// View satisfies the `tea.Model` interface.
func (m *DatePickerModel) View() tea.View {
	b := strings.Builder{}
	month := m.Time.Month()
	year := m.Time.Year()

	tMonth, tYear := month.String(), strconv.Itoa(year)

	if m.Focused == FocusHeaderMonth {
		tMonth = m.Styles.FocusedText.Render(tMonth)
	} else {
		tMonth = m.Styles.HeaderText.Render(tMonth)
	}

	if m.Focused == FocusHeaderYear {
		tYear = m.Styles.FocusedText.Render(tYear)
	} else {
		tYear = m.Styles.HeaderText.Render(tYear)
	}

	title := m.Styles.Header.Render(fmt.Sprintf("%s %s\n", tMonth, tYear))

	// get all the dates of the current month
	firstDayOfTheMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	lastSundayOfLastMonth := firstDayOfTheMonth.AddDate(0, 0, -1)
	for lastSundayOfLastMonth.Weekday() != time.Sunday {
		lastSundayOfLastMonth = lastSundayOfLastMonth.AddDate(0, 0, -1)
	}

	lastDayOfTheMonth := firstDayOfTheMonth.AddDate(0, 1, -1)

	firstSundayOfNextMonth := lastDayOfTheMonth.AddDate(0, 0, 1)
	for firstSundayOfNextMonth.Weekday() != time.Sunday {
		firstSundayOfNextMonth = firstSundayOfNextMonth.AddDate(0, 0, 1)
	}

	day := lastSundayOfLastMonth
	if firstDayOfTheMonth.Weekday() == time.Sunday {
		day = firstDayOfTheMonth
	}

	weekHeaders := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for i, h := range weekHeaders {
		weekHeaders[i] = m.Styles.Date.Inherit(m.Styles.HeaderText).Render(h)
	}

	cal := [][]string{weekHeaders}
	dayGrid := [][]time.Time{}
	j := 1

	for day.Before(firstSundayOfNextMonth) {
		if j >= len(cal) {
			cal = append(cal, []string{})
			dayGrid = append(dayGrid, make([]time.Time, 0, 7))
		}
		if len(dayGrid) < j {
			dayGrid = append(dayGrid, make([]time.Time, 0, 7))
		}
		out := "  "
		cellDate := time.Time{}
		if day.Month() == month {
			out = fmt.Sprintf("%02d", day.Day())
			cellDate = day
		}

		style := m.Styles.Date
		textStyle := m.Styles.Text
		switch {
		case day.Day() == m.Time.Day() && day.Month() == m.Time.Month() && m.Focused == FocusCalendar:
			// The cursor day is always visibly highlighted (inverted colors by
			// default) — regression: it used to render like every other day
			// until Selected was set, which no binding ever did.
			textStyle = m.Styles.FocusedText
		case day.Day() == m.Time.Day() && day.Month() == m.Time.Month():
			textStyle = m.Styles.SelectedText
		case m.Effects.Hover() && !m.hoverDay.IsZero() &&
			day.Day() == m.hoverDay.Day() && day.Month() == m.hoverDay.Month():
			textStyle = m.Styles.Text.Underline(true)
		}

		out = style.Inherit(textStyle).Render(out)
		cal[j] = append(cal[j], out)
		dayGrid[j-1] = append(dayGrid[j-1], cellDate)
		if m.cellW == 0 {
			m.cellW = lipgloss.Width(out)
			m.cellH = lipgloss.Height(out)
		}

		if day.Weekday() == time.Saturday {
			j++
		}
		day = day.AddDate(0, 0, 1)
	}

	rows := make([]string, 0, 1+len(cal))
	rows = append(rows, title)
	for _, row := range cal {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Center, row...))
	}
	content := lipgloss.JoinVertical(lipgloss.Center, rows...)
	b.WriteString(content)

	// Record hit-zone geometry for handleClick: the title block, then the
	// weekday-header row, then the day grid; JoinVertical(Center) indents
	// narrower rows, so the grid's x offset is derived from the final width.
	m.dayGrid = dayGrid
	m.titleH = lipgloss.Height(title)
	m.totalW = lipgloss.Width(content)
	headerH := lipgloss.Height(rows[1])
	m.gridTopY = m.titleH + headerH
	m.gridOffX = max((m.totalW-7*m.cellW)/2, 0)

	v := tea.NewView(b.String())
	v.OnMouse = uifx.RouteToUpdate(m.Update)
	return v
}

// SetsFocus focuses one of the datepicker components. This can also be used to blur
// the datepicker by passing the Focus `FocusNone`.
func (m *DatePickerModel) SetFocus(f Focus) {
	m.Focused = f
}

// Blur sets the datepicker focus to `FocusNone`
func (m *DatePickerModel) Blur() {
	m.Focused = FocusNone
}

// SetTime sets the model's `Time` struct and is used as reference to the selected date
func (m *DatePickerModel) SetTime(t time.Time) {
	m.Time = t
}

// LastWeek sets the model's `Time` struct back 7 days
func (m *DatePickerModel) LastWeek() {
	m.Time = m.Time.AddDate(0, 0, -7)
}

// NextWeek sets the model's `Time` struct forward 7 days
func (m *DatePickerModel) NextWeek() {
	m.Time = m.Time.AddDate(0, 0, 7)
}

// Yesterday sets the model's `Time` struct back 1 day
func (m *DatePickerModel) Yesterday() {
	m.Time = m.Time.AddDate(0, 0, -1)
}

// Tomorrow sets the model's `Time` struct forward 1 day
func (m *DatePickerModel) Tomorrow() {
	m.Time = m.Time.AddDate(0, 0, 1)
}

// LastMonth sets the model's `Time` struct back 1 month
func (m *DatePickerModel) LastMonth() {
	m.Time = m.Time.AddDate(0, -1, 0)
}

// NextMonth sets the model's `Time` struct forward 1 month
func (m *DatePickerModel) NextMonth() {
	m.Time = m.Time.AddDate(0, 1, 0)
}

// LastYear sets the model's `Time` struct back 1 year
func (m *DatePickerModel) LastYear() {
	m.Time = m.Time.AddDate(-1, 0, 0)
}

// NextYear sets the model's `Time` struct forward 1 year
func (m *DatePickerModel) NextYear() {
	m.Time = m.Time.AddDate(1, 0, 0)
}

// SelectDate changes the model's Selected to true
func (m *DatePickerModel) SelectDate() {
	m.Selected = true
}

// UnselectDate changes the model's Selected to false
func (m *DatePickerModel) UnselectDate() {
	m.Selected = false
}
