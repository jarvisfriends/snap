package pickers

import (
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/x/ansi"
	"github.com/jarvisfriends/snap/keys"
	"github.com/jarvisfriends/snap/uifx"
)

// overlayContentWidth returns how many columns a model hosted in a
// ModelOverlayHost may use: the page width minus the overlay box chrome
// (2 border + 4 padding) and a 2-column margin. Content wider than this wraps
// inside the box and corrupts the layout. Zero/unset widths fall back to a
// conservative 74 (an 80-column terminal minus the chrome).
func overlayContentWidth(pageWidth int) int {
	if pageWidth <= 0 {
		return 74
	}
	return max(20, pageWidth-8)
}

// fitLine truncates a styled line to w display cells with an ellipsis; lines
// already within w are returned unchanged.
func fitLine(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	return ansi.Truncate(s, w, "…")
}

// dirEntriesMsg carries the subdirectory listing produced by readDirCmd.
type dirEntriesMsg struct {
	dir     string
	entries []string
	err     error
}

// DirPicker is a directory-only browser overlay: unlike the huh/bubbles file
// picker it lists no files at all, which is the right presentation when the
// user can only choose a directory. Selection follows the file-picker split
// (Enter/→ browses, Space selects the highlighted folder) plus Ctrl+S to
// select the directory currently being browsed.
type DirPicker struct {
	dir       string   // directory currently being browsed (absolute)
	entries   []string // subdirectory names of dir
	cursor    int
	scrollTop int
	err       error
	selected  string

	KeyMap  *keys.AppKeyMap
	Done    bool
	Aborted bool
	Width   int
	Height  int

	// Styles are the injected style hooks (theme-free; hosts map their
	// palette on from their live theme).
	Styles Styles
	// CollapsePath, when set, shortens the displayed directory (e.g.
	// substituting %USERPROFILE% or ~). Defaults to showing the full path.
	CollapsePath func(string) string

	// Effects selects the interaction-feedback tier (see uifx.Level):
	// hover highlighting renders only at LevelHigh, drag tracking at
	// LevelMedium and above.
	Effects uifx.Level

	// HideHelp suppresses the picker's built-in key-hint line. Hosts that
	// surface the picker's KeyMap in their own help/status bar set this so
	// the hints aren't shown twice.
	HideHelp bool

	// Mouse hit-zone geometry recorded during View: the y of the first
	// visible row and the list width. Row i on screen is entries[scrollTop+i].
	rowsTopY  int
	rowsWidth int
	// hoverRow is the entries index under the pointer (-1 none; LevelHigh).
	hoverRow int
}

// NewDirPicker returns a picker browsing from initial: the value itself when
// it is an existing directory, its parent when it is a file path, and the
// working directory when it is empty or invalid.
func NewDirPicker(initial string) *DirPicker {
	dir := initial
	if dir != "" {
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			dir = filepath.Dir(dir)
		}
	}
	if st, err := os.Stat(dir); dir == "" || err != nil || !st.IsDir() {
		if wd, wdErr := os.Getwd(); wdErr == nil {
			dir = wd
		} else {
			dir = "."
		}
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return &DirPicker{
		Styles:   DefaultStyles(),
		hoverRow: -1,
		dir:      dir,
		KeyMap:   keys.DefaultKeyMap(),
	}
}

// Value returns the selected directory path (empty until Done).
// collapse applies the CollapsePath hook, defaulting to the full path.
func (m *DirPicker) collapse(dir string) string {
	if m.CollapsePath != nil {
		return m.CollapsePath(dir)
	}
	return dir
}

func (m *DirPicker) Value() string {
	return m.selected
}

func (m *DirPicker) Init() tea.Cmd {
	return readDirCmd(m.dir)
}

// listDrives returns the mounted drive roots (e.g. ["C:\", "E:\"]) by
// probing every letter; on non-Windows systems no letter resolves so the
// list is empty and the drive view is never entered ("/" is its own parent,
// but Back at "/" simply finds nothing to list).
func listDrives() []string {
	var drives []string
	for l := 'A'; l <= 'Z'; l++ {
		root := string(l) + `:\`
		if st, err := os.Stat(root); err == nil && st.IsDir() {
			drives = append(drives, root)
		}
	}
	return drives
}

// readDirCmd lists the subdirectories of dir, hiding all files.
func readDirCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		ents, err := os.ReadDir(dir)
		if err != nil {
			return dirEntriesMsg{dir: dir, err: err}
		}
		var names []string
		for _, e := range ents {
			if e.IsDir() {
				names = append(names, e.Name())
			}
		}
		return dirEntriesMsg{dir: dir, entries: names}
	}
}

func (m *DirPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width, m.Height = msg.Width, msg.Height
		m.ensureCursorVisible()

	case dirEntriesMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.dir = msg.dir
		m.entries = msg.entries
		m.cursor = 0
		m.scrollTop = 0
		m.err = nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Cancel):
			m.Aborted = true
		case key.Matches(msg, m.KeyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			m.ensureCursorVisible()
		case key.Matches(msg, m.KeyMap.Down):
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
			m.ensureCursorVisible()
		case key.Matches(msg, m.KeyMap.Open, m.KeyMap.Right):
			if m.cursor < len(m.entries) {
				return m, readDirCmd(filepath.Join(m.dir, m.entries[m.cursor]))
			}
		case key.Matches(msg, m.KeyMap.Left, m.KeyMap.Delete):
			if m.dir == "" {
				break // already at the drive list
			}
			if parent := filepath.Dir(m.dir); parent != m.dir {
				return m, readDirCmd(parent)
			}
			// At a filesystem root: offer the available drives so the user
			// can navigate anywhere, not just below the starting directory.
			if drives := listDrives(); len(drives) > 0 {
				m.dir = ""
				m.entries = drives
				m.cursor, m.scrollTop = 0, 0
				m.err = nil
			}
		case key.Matches(msg, m.KeyMap.Select):
			if m.cursor < len(m.entries) {
				m.selected = filepath.Join(m.dir, m.entries[m.cursor])
				m.Done = true
			}
		case key.Matches(msg, m.KeyMap.Save):
			if m.dir != "" {
				m.selected = m.dir
				m.Done = true
			}
		}
	}
	return m, nil
}

// rowAt maps a content-relative y to an entries index, or -1.
func (m *DirPicker) rowAt(x, y int) int {
	if x < 0 || x >= m.rowsWidth || y < m.rowsTopY {
		return -1
	}
	i := m.scrollTop + (y - m.rowsTopY)
	if i < m.scrollTop+m.listHeight() && i < len(m.entries) {
		return i
	}
	return -1
}

// handleClick: clicking a row moves the highlight there; clicking the
// already-highlighted row opens it (mirroring the datepicker's
// click-to-highlight, click-again-to-act convention).
func (m *DirPicker) handleClick(me tea.Mouse) tea.Cmd {
	if me.Button != tea.MouseLeft {
		return nil
	}
	i := m.rowAt(me.X, me.Y)
	if i < 0 {
		return nil
	}
	if i == m.cursor {
		return readDirCmd(filepath.Join(m.dir, m.entries[i]))
	}
	m.cursor = i
	m.ensureCursorVisible()
	return nil
}

// handleWheel: up/down move the highlight; left goes to the parent
// directory, right opens the highlighted one — the wheel alone can walk the
// whole tree.
func (m *DirPicker) handleWheel(me tea.Mouse) tea.Cmd {
	switch me.Button {
	case tea.MouseWheelUp:
		if m.cursor > 0 {
			m.cursor--
		}
		m.ensureCursorVisible()
	case tea.MouseWheelDown:
		if m.cursor < len(m.entries)-1 {
			m.cursor++
		}
		m.ensureCursorVisible()
	case tea.MouseWheelLeft:
		return m.navigateBack()
	case tea.MouseWheelRight:
		if m.cursor < len(m.entries) {
			return readDirCmd(filepath.Join(m.dir, m.entries[m.cursor]))
		}
	}
	return nil
}

// handleMotion: while the left button is held the highlight follows the
// pointer (LevelMedium+); with no button held the hovered row is tracked for
// the LevelHigh hover highlight.
func (m *DirPicker) handleMotion(me tea.Mouse) tea.Cmd {
	i := m.rowAt(me.X, me.Y)
	if me.Button == tea.MouseLeft {
		if m.Effects.Drag() && i >= 0 {
			m.cursor = i
			m.ensureCursorVisible()
		}
		return nil
	}
	if m.Effects.Hover() {
		m.hoverRow = i
	}
	return nil
}

// navigateBack mirrors the Back key: parent directory, or the drive list at
// a filesystem root.
func (m *DirPicker) navigateBack() tea.Cmd {
	if m.dir == "" {
		return nil
	}
	if parent := filepath.Dir(m.dir); parent != m.dir {
		return readDirCmd(parent)
	}
	if drives := listDrives(); len(drives) > 0 {
		m.dir = ""
		m.entries = drives
		m.cursor, m.scrollTop = 0, 0
		m.err = nil
	}
	return nil
}

// listHeight returns how many directory rows fit between the header and help
// chrome for the current overlay height.
func (m *DirPicker) listHeight() int {
	return max(3, m.Height-8)
}

func (m *DirPicker) ensureCursorVisible() {
	h := m.listHeight()
	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	}
	if m.cursor >= m.scrollTop+h {
		m.scrollTop = m.cursor - h + 1
	}
	if m.scrollTop < 0 {
		m.scrollTop = 0
	}
}

// onMouse is the View.OnMouse entry point: mouse events dispatch straight to
// the handler methods, never through Update, so hosts (and the Bubble Tea
// runtime) deliver pointer input through exactly one door. Parents hosting
// this component should call onMouse with translated coordinates.
func (m *DirPicker) onMouse(msg tea.MouseMsg) tea.Cmd {
	return uifx.MouseHandlers{
		Click:  m.handleClick,
		Wheel:  m.handleWheel,
		Motion: m.handleMotion,
	}.OnMouse(msg)
}

func (m *DirPicker) View() tea.View {
	maxW := overlayContentWidth(m.Width)
	title := m.Styles.Title.Render("Select Directory")
	pathStyle := m.Styles.Path
	selStyle := m.Styles.Selected
	normStyle := m.Styles.Normal
	dimStyle := m.Styles.Dim

	location := "📁 " + m.collapse(m.dir)
	if m.dir == "" {
		location = "💾 Drives"
	}
	current := fitLine(pathStyle.Render(location), maxW)

	var rows []string
	switch {
	case m.err != nil:
		rows = append(rows, fitLine(dimStyle.Render("(unreadable: "+m.err.Error()+")"), maxW))
	case len(m.entries) == 0:
		rows = append(rows, dimStyle.Render("(no subdirectories)"))
	default:
		h := m.listHeight()
		end := min(m.scrollTop+h, len(m.entries))
		for i := m.scrollTop; i < end; i++ {
			prefix, style := "  ", normStyle
			switch {
			case m.cursor == i:
				prefix, style = "▶ ", selStyle
			case m.Effects.Hover() && m.hoverRow == i:
				// LevelHigh: the row under the pointer renders underlined so
				// the click target reads before committing.
				style = normStyle.Underline(true)
			}
			rows = append(rows, fitLine(style.Render(prefix+m.entries[i]), maxW))
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	content := lipgloss.JoinVertical(lipgloss.Left, title, current, body)
	if !m.HideHelp {
		help := dimStyle.MarginTop(1).Render(fitLine(
			"↑/↓: Navigate • Enter/→: Open • ←: Up • Space: Select • Ctrl+S: Select This Folder • Esc: Cancel",
			maxW,
		))
		content = lipgloss.JoinVertical(lipgloss.Left, content, help)
	}

	// Record the row hit zones for the mouse handlers: rows start under the
	// title and location lines.
	m.rowsTopY = lipgloss.Height(title) + lipgloss.Height(current)
	m.rowsWidth = maxW

	v := tea.NewView(content)
	// Route mouse through Update so hosts honoring View.OnMouse get clicks,
	// wheel navigation, drag, and hover with no extra wiring. Hosts must
	// deliver mouse via exactly one path (OnMouse or Update), never both —
	// Bubble Tea itself sends the raw event to both at the root.
	v.OnMouse = m.onMouse
	return v
}
