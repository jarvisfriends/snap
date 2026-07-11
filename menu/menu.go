// Package menu provides a right-click context menu: a small pop-up action
// list opened at the pointer position, clamped to the terminal, and driven
// by mouse or keyboard. Ported from the tribble console and decoupled from
// its model: styles are injectable, widths are unicode-safe, and input
// arrives through the host's OnMouse (per the snap input contract — a menu
// is always hosted, so the parent translates nothing: it hands the menu
// screen coordinates and composites the rendered box over its own frame).
package menu

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/geom"
)

// Item is one action in the menu.
type Item struct {
	ID       string // action identifier the host dispatches on (e.g. "delete")
	Label    string // display text
	Disabled bool   // greyed out and not selectable
}

// Styles selects the menu's appearance.
type Styles struct {
	Box         lipgloss.Style // border + colors around the item list
	Text        lipgloss.Style // enabled item labels
	Dim         lipgloss.Style // disabled item labels
	Cursor      lipgloss.Style // the selection glyph
	CursorGlyph string         // rendered before the selected item ("▸")
	MinWidth    int            // minimum inner width so short menus don't shrink to a sliver
}

// DefaultStyles returns a rounded-border menu on the terminal's colors.
func DefaultStyles() Styles {
	return Styles{
		Box:         lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
		Text:        lipgloss.NewStyle(),
		Dim:         lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		CursorGlyph: "▸",
		MinWidth:    14,
	}
}

// KeyMap is the menu's key bindings while it is open. Use DefaultKeyMap for
// the conventional set; hosts rebind per field.
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Choose  key.Binding // choose the item under the cursor (menu closes)
	Dismiss key.Binding // close without choosing
}

// DefaultKeyMap returns the standard bindings: arrows/jk move, Enter
// chooses, Esc dismisses.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Choose:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "choose")),
		Dismiss: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "dismiss")),
	}
}

// Menu is the pop-up state, designed to be embedded in a host model. The
// zero value is a closed menu with default styles and keys applied on first
// use.
type Menu struct {
	Styles Styles
	Keys   KeyMap

	visible bool
	items   []Item
	cursor  int
	x, y    int // screen position where the menu was opened
	tag     any // opaque host context: what was right-clicked
}

// Open shows the menu at screen position (x, y) with the given items. tag is
// an opaque value handed back with the chosen item so the host knows what
// the click targeted (a row index, an entity ID, …). The cursor starts on
// the first enabled item.
func (m *Menu) Open(x, y int, items []Item, tag any) {
	m.visible = true
	m.items = items
	m.x, m.y = x, y
	m.tag = tag
	m.cursor = 0
	for i, it := range items {
		if !it.Disabled {
			m.cursor = i
			break
		}
	}
}

// Close hides the menu.
func (m *Menu) Close() { m.visible = false }

// IsOpen reports whether the menu is showing.
func (m *Menu) IsOpen() bool { return m.visible }

// Tag returns the opaque context passed to Open.
func (m *Menu) Tag() any { return m.tag }

// Selected returns the item under the cursor, or nil when the cursor rests
// on nothing selectable.
func (m *Menu) Selected() *Item {
	if m.cursor >= 0 && m.cursor < len(m.items) && !m.items[m.cursor].Disabled {
		return &m.items[m.cursor]
	}
	return nil
}

// MoveUp moves the cursor to the previous enabled item, if any.
func (m *Menu) MoveUp() {
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.items[i].Disabled {
			m.cursor = i
			return
		}
	}
}

// MoveDown moves the cursor to the next enabled item, if any.
func (m *Menu) MoveDown() {
	for i := m.cursor + 1; i < len(m.items); i++ {
		if !m.items[i].Disabled {
			m.cursor = i
			return
		}
	}
}

// keys returns the effective key bindings (defaults for the zero value).
func (m *Menu) keys() KeyMap {
	if len(m.Keys.Up.Keys()) == 0 && len(m.Keys.Down.Keys()) == 0 {
		return DefaultKeyMap()
	}
	return m.Keys
}

// HandleKey processes a key press while the menu is open — the keyboard twin
// of HandleMouse: Up/Down move the cursor, Choose picks the item under the
// cursor (menu closes), Dismiss closes without choosing. It returns the
// chosen item (nil when nothing was chosen) and whether the event was
// consumed; while open the menu is modal, so every key is consumed and an
// unconsumed key means the menu was already closed.
func (m *Menu) HandleKey(msg tea.KeyPressMsg) (chosen *Item, handled bool) {
	if !m.visible {
		return nil, false
	}
	km := m.keys()
	switch {
	case key.Matches(msg, km.Up):
		m.MoveUp()
	case key.Matches(msg, km.Down):
		m.MoveDown()
	case key.Matches(msg, km.Choose):
		if sel := m.Selected(); sel != nil {
			it := *sel
			chosen = &it
		}
		m.Close()
	case key.Matches(msg, km.Dismiss):
		m.Close()
	}
	return chosen, true
}

// styles returns the effective styles (defaults for the zero value).
func (m *Menu) styles() Styles {
	if m.Styles.CursorGlyph == "" && m.Styles.MinWidth == 0 {
		return DefaultStyles()
	}
	return m.Styles
}

// innerWidth is the content width inside the border: the widest label plus
// the cursor gutter, at least MinWidth.
func (m *Menu) innerWidth() int {
	st := m.styles()
	w := 0
	for _, it := range m.items {
		if lw := lipgloss.Width(it.Label); lw > w {
			w = lw
		}
	}
	return max(w+4, st.MinWidth)
}

// Rect returns the screen rectangle the rendered menu occupies, clamped so
// it never runs off the right or bottom edge of a termW x termH terminal.
func (m *Menu) Rect(termW, termH int) geom.Rect {
	w := m.innerWidth() + m.styles().Box.GetHorizontalFrameSize()
	h := len(m.items) + m.styles().Box.GetVerticalFrameSize()
	x := geom.Clamp(m.x, 0, max(termW-w, 0))
	y := geom.Clamp(m.y, 0, max(termH-h, 0))
	return geom.Rect{X: x, Y: y, W: w, H: h}
}

// HandleMouse processes a screen-coordinate mouse event while the menu is
// open, per the context-menu conventions: hover and drag move the cursor,
// wheel scrolls it, a left click on an enabled item chooses it (menu
// closes), and any click outside dismisses the menu. It returns the chosen
// item (nil when nothing was chosen) and whether the event was consumed —
// an unconsumed event means the menu is closed and the host should process
// the event itself.
func (m *Menu) HandleMouse(msg tea.MouseMsg, termW, termH int) (chosen *Item, handled bool) {
	if !m.visible {
		return nil, false
	}
	me := msg.Mouse()
	switch msg.(type) {
	case tea.MouseClickMsg:
		idx := m.itemAt(me.X, me.Y, termW, termH)
		if idx < 0 {
			m.Close()
			// A click outside both dismisses and stays with the host (it
			// may open a different menu or act on what was clicked).
			return nil, false
		}
		if me.Button != tea.MouseLeft || m.items[idx].Disabled {
			return nil, true
		}
		m.cursor = idx
		it := m.items[idx]
		m.Close()
		return &it, true
	case tea.MouseWheelMsg:
		switch me.Button {
		case tea.MouseWheelUp:
			m.MoveUp()
		case tea.MouseWheelDown:
			m.MoveDown()
		}
		return nil, true
	case tea.MouseMotionMsg:
		if idx := m.itemAt(me.X, me.Y, termW, termH); idx >= 0 && !m.items[idx].Disabled {
			m.cursor = idx
		}
		return nil, true
	}
	return nil, true
}

// itemAt maps a screen coordinate to an item index, or -1 outside the menu.
func (m *Menu) itemAt(x, y, termW, termH int) int {
	r := m.Rect(termW, termH)
	if !r.Contains(x, y) {
		return -1
	}
	idx := y - r.Y - m.styles().Box.GetBorderTopSize()
	if idx < 0 || idx >= len(m.items) {
		return -1
	}
	return idx
}

// Render produces the styled menu box.
func (m *Menu) Render() string {
	st := m.styles()
	var b strings.Builder
	for i, it := range m.items {
		label := st.Text.Render(it.Label)
		if it.Disabled {
			label = st.Dim.Render(it.Label)
		}
		if i > 0 {
			b.WriteString("\n")
		}
		if i == m.cursor && !it.Disabled {
			b.WriteString(st.Cursor.Render(st.CursorGlyph))
			b.WriteString(" ")
			b.WriteString(label)
		} else {
			b.WriteString("  ")
			b.WriteString(label)
		}
	}
	return st.Box.Width(m.innerWidth()).Render(b.String())
}

// Composite draws the open menu over base (a full termW x termH frame) at
// its clamped position; with the menu closed it returns base unchanged.
func (m *Menu) Composite(base string, termW, termH int) string {
	if !m.visible {
		return base
	}
	r := m.Rect(termW, termH)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(base),
		lipgloss.NewLayer(m.Render()).X(r.X).Y(r.Y).Z(20),
	).Render()
}
