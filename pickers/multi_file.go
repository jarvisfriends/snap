package pickers

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	huh "charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/keys"
	"github.com/jarvisfriends/snap/uifx"
)

type MultiFileEditor struct {
	paths       []string
	cursor      int
	picking     bool
	pickerForm  *huh.Form
	dirPicker   *DirPicker
	pickerIndex int

	KeyMap *keys.AppKeyMap
	// Styles are the injected style hooks (theme-free; see DefaultStyles).
	Styles Styles
	// HuhTheme styles the embedded huh file-picker form. Nil uses huh's
	// base theme; hosts inject their live theme's huh mapping.
	HuhTheme func() huh.Theme
	// CollapsePath is forwarded to row DirPickers (see DirPicker.CollapsePath).
	CollapsePath func(string) string
	// Effects selects the interaction-feedback tier (see uifx.Level); it is
	// forwarded to row DirPickers.
	Effects uifx.Level

	// Row hit-zone geometry recorded during View: rows start at rowsTopY and
	// row i is paths[i], with the "[ Add Path ]" row at index len(paths).
	rowsTopY  int
	rowsWidth int
	// hoverRow tracks the row under the pointer (-1 none; LevelHigh only).
	hoverRow int
	// DirsOnly makes each row's picker a directory-only DirPicker (with
	// drive navigation) instead of the mixed file/folder browser.
	DirsOnly bool
	Done     bool
	Aborted  bool
	Width    int
	Height   int
}

func NewMultiFileEditor(value string) *MultiFileEditor {
	var paths []string
	if strings.TrimSpace(value) != "" {
		parts := strings.SplitSeq(value, ";")
		for p := range parts {
			paths = append(paths, strings.TrimSpace(p))
		}
	}
	return &MultiFileEditor{
		Styles:   DefaultStyles(),
		hoverRow: -1,
		paths:    paths,
		KeyMap:   keys.DefaultKeyMap(),
	}
}

func (m *MultiFileEditor) Value() string {
	return strings.Join(m.paths, "; ")
}

func (m *MultiFileEditor) Init() tea.Cmd {
	return nil
}

// updatePicker forwards msg to the active picker form (keeping any replacement
// model it returns) and applies the completed/aborted result.
func (m *MultiFileEditor) updatePicker(msg tea.Msg) tea.Cmd {
	model, cmd := m.pickerForm.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		m.pickerForm = f
	}
	switch m.pickerForm.State {
	case huh.StateCompleted:
		m.applyPickedPath(m.pickerForm.GetString("path"))
		m.picking = false
		m.pickerForm = nil
	case huh.StateAborted:
		m.picking = false
		m.pickerForm = nil
	case huh.StateNormal:
		// form still in progress — no action
	}
	return cmd
}

// updateDirPicker forwards msg to the active directory picker and applies
// the completed/aborted result.
func (m *MultiFileEditor) updateDirPicker(msg tea.Msg) tea.Cmd {
	model, cmd := m.dirPicker.Update(msg)
	if dp, ok := model.(*DirPicker); ok {
		m.dirPicker = dp
	}
	m.finishDirPicker()
	return cmd
}

// finishDirPicker applies the browse result once the hosted DirPicker
// reports Done or Aborted — shared by the Update (keys) and onMouse paths.
func (m *MultiFileEditor) finishDirPicker() {
	switch {
	case m.dirPicker == nil:
	case m.dirPicker.Done:
		m.applyPickedPath(m.dirPicker.Value())
		m.picking = false
		m.dirPicker = nil
	case m.dirPicker.Aborted:
		m.picking = false
		m.dirPicker = nil
	}
}

// applyPickedPath stores a non-empty picker result at the pending index,
// appending when the pick targeted the "add new" slot.
func (m *MultiFileEditor) applyPickedPath(val string) {
	if val == "" {
		return
	}
	if m.pickerIndex == len(m.paths) {
		m.paths = append(m.paths, val)
	} else {
		m.paths[m.pickerIndex] = val
	}
}

func (m *MultiFileEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width, m.Height = ws.Width, ws.Height
	}

	if m.picking {
		// Mouse reaches the hosted picker exclusively via onMouse; a host
		// that also feeds Update raw mouse must not double-process it.
		if _, isMouse := msg.(tea.MouseMsg); isMouse {
			return m, nil
		}
		if m.dirPicker != nil {
			return m, m.updateDirPicker(msg)
		}
		if m.pickerForm != nil {
			return m, m.updatePicker(msg)
		}
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Cancel):
			m.Aborted = true
			return m, nil
		case key.Matches(msg, m.KeyMap.Up):
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.paths) // "Add Row" is at the bottom
			}
		case key.Matches(msg, m.KeyMap.Down):
			m.cursor++
			if m.cursor > len(m.paths) {
				m.cursor = 0
			}
		case key.Matches(msg, m.KeyMap.Submit):
			return m, m.startPicking(m.cursor)
		case key.Matches(msg, m.KeyMap.Delete):
			if m.cursor < len(m.paths) {
				m.paths = append(m.paths[:m.cursor], m.paths[m.cursor+1:]...)
				if m.cursor > len(m.paths) {
					m.cursor = len(m.paths)
				}
			}
		case key.Matches(msg, m.KeyMap.Save):
			m.Done = true
			return m, nil
		}
	}

	return m, nil
}

// startPicking opens the per-row picker form and returns its Init command.
// The command must be dispatched to the runtime: it carries the picker's
// readDir, and bubbles' filepicker opens with a collapsed one-row browse
// window until that readDir's message arrives and expands it.
func (m *MultiFileEditor) startPicking(index int) tea.Cmd {
	m.pickerIndex = index
	initialValue := ""
	if index < len(m.paths) {
		initialValue = m.paths[index]
	}

	if m.DirsOnly {
		dp := NewDirPicker(initialValue)
		dp.Styles = m.Styles
		dp.CollapsePath = m.CollapsePath
		dp.Effects = m.Effects
		dp.Width, dp.Height = m.Width, m.Height
		m.dirPicker = dp
		m.picking = true
		return dp.Init()
	}

	fp := huh.NewFilePicker().
		Key("path").
		Title("Select Path").
		DirAllowed(true).
		FileAllowed(true).
		// Open directly in browse mode: the embedded picker otherwise
		// defaults to a one-row list.
		Picking(true).
		Value(&initialValue)
	if m.Height > 0 {
		// Fill most of the available area with the file listing.
		fp.Height(pickerFormHeight(m.Height))
	}

	m.pickerForm = huh.NewForm(huh.NewGroup(fp)).
		WithTheme(m.huhTheme()).
		WithKeyMap(FilePickerKeyMap()).
		WithWidth(overlayContentWidth(m.Width))
	if m.Height > 0 {
		// The form height must be set explicitly: huh freezes each group's
		// height from the fields' first render, which for a picker is the
		// collapsed pre-readDir browse list, and the zoomed picker gets that
		// frozen height re-imposed on every render.
		m.pickerForm = m.pickerForm.WithHeight(pickerFormHeight(m.Height))
	}
	m.picking = true
	initCmd := m.pickerForm.Init()
	if m.Width > 0 && m.Height > 0 {
		// Deliver the current size the way tea.Program would on open:
		// bubbles' filepicker starts with a collapsed one-row browse window
		// that only its WindowSizeMsg handler unconditionally expands —
		// setting heights through huh's builder API alone leaves the window
		// collapsed until the user resizes the terminal or changes
		// directory. Sized here, before initCmd's readDir message arrives,
		// the first directory listing renders at full height.
		model, _ := m.pickerForm.Update(tea.WindowSizeMsg{
			Width:  overlayContentWidth(m.Width),
			Height: pickerFormHeight(m.Height),
		})
		if f, ok := model.(*huh.Form); ok {
			m.pickerForm = f
		}
	}
	return initCmd
}

// rowAt maps content-relative coordinates to a row index (paths rows, then
// the "[ Add Path ]" row at len(paths)); -1 when outside the list.
func (m *MultiFileEditor) rowAt(x, y int) int {
	if x < 0 || x >= m.rowsWidth || y < m.rowsTopY {
		return -1
	}
	i := y - m.rowsTopY
	if i <= len(m.paths) {
		return i
	}
	return -1
}

// onMouse is the View.OnMouse entry point. List-mode events dispatch to the
// row handlers, never through Update. While browsing, our View is the hosted
// DirPicker's view verbatim (same coordinate space), so mouse goes to its
// onMouse; a huh picker form exposes no OnMouse, so its Update is the only
// door for it.
func (m *MultiFileEditor) onMouse(msg tea.MouseMsg) tea.Cmd {
	if m.picking {
		if m.dirPicker != nil {
			cmd := m.dirPicker.onMouse(msg)
			m.finishDirPicker()
			return cmd
		}
		if m.pickerForm != nil {
			return m.updatePicker(msg)
		}
		return nil
	}
	return uifx.MouseHandlers{
		Click:  m.handleClick,
		Wheel:  m.handleWheel,
		Motion: m.handleMotion,
	}.OnMouse(msg)
}

// handleClick moves the highlight to the clicked row; clicking the
// highlighted row activates it (opens its picker / adds a path), matching
// the click-to-highlight, click-again-to-act convention.
func (m *MultiFileEditor) handleClick(me tea.Mouse) tea.Cmd {
	if me.Button != tea.MouseLeft {
		return nil
	}
	i := m.rowAt(me.X, me.Y)
	if i < 0 {
		return nil
	}
	if i == m.cursor {
		return m.startPicking(m.cursor)
	}
	m.cursor = i
	return nil
}

// handleWheel scrolls the highlight through the rows (wrapping like the
// keyboard bindings do).
func (m *MultiFileEditor) handleWheel(me tea.Mouse) tea.Cmd {
	switch me.Button {
	case tea.MouseWheelUp:
		m.cursor--
		if m.cursor < 0 {
			m.cursor = len(m.paths)
		}
	case tea.MouseWheelDown:
		m.cursor++
		if m.cursor > len(m.paths) {
			m.cursor = 0
		}
	}
	return nil
}

// handleMotion tracks drags (highlight follows a held left button,
// LevelMedium+) and hover (LevelHigh).
func (m *MultiFileEditor) handleMotion(me tea.Mouse) tea.Cmd {
	i := m.rowAt(me.X, me.Y)
	if me.Button == tea.MouseLeft {
		if m.Effects.Drag() && i >= 0 {
			m.cursor = i
		}
		return nil
	}
	if m.Effects.Hover() {
		m.hoverRow = i
	}
	return nil
}

func (m *MultiFileEditor) View() tea.View {
	if m.picking && m.dirPicker != nil {
		return m.dirPicker.View()
	}
	if m.picking && m.pickerForm != nil {
		return tea.NewView(m.pickerForm.View())
	}

	maxW := overlayContentWidth(m.Width)
	title := m.Styles.Title.Render("Multi-File Picker")

	var rows []string
	selStyle := m.Styles.Selected
	normStyle := m.Styles.Normal
	delStyle := m.Styles.Dim

	for i, p := range m.paths {
		prefix := "  "
		style := normStyle
		switch {
		case m.cursor == i:
			prefix = "▶ "
			style = selStyle
		case m.Effects.Hover() && m.hoverRow == i:
			style = normStyle.Underline(true)
		}
		rows = append(rows, fitLine(style.Render(prefix+p), maxW))
	}

	addPrefix := "  "
	addStyle := normStyle
	if m.cursor == len(m.paths) {
		addPrefix = "▶ "
		addStyle = selStyle
	}
	rows = append(rows, addStyle.Render(addPrefix+"[ Add Path ]"))

	help := delStyle.MarginTop(1).Render(fitLine(
		"↑/↓: Navigate • Enter: Edit/Add • Del: Remove • Ctrl+S: Save • Esc: Cancel",
		maxW,
	))

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, title, body, help)

	// Record row hit zones: rows start directly under the title.
	m.rowsTopY = lipgloss.Height(title)
	m.rowsWidth = maxW

	v := tea.NewView(content)
	// One mouse path: hosts deliver via OnMouse or Update, never both (the
	// Bubble Tea root gets the raw event on both). While a child picker is
	// open its own View is returned instead, so coordinates always match
	// what is on screen.
	v.OnMouse = m.onMouse
	return v
}

// huhTheme resolves the injected huh theme, defaulting to huh's base theme
// (dark variant — hosts with light palettes inject their own).
func (m *MultiFileEditor) huhTheme() huh.Theme {
	if m.HuhTheme != nil {
		return m.HuhTheme()
	}
	return huh.ThemeFunc(huh.ThemeBase)
}

// pickerFormHeight sizes the embedded file-picker form: most of the host
// area, floored so the picker stays usable in tiny terminals. Mirrors
// the host-overlay sizing convention so the embedded look stays identical.
func pickerFormHeight(termH int) int {
	return max(5, termH-6)
}
