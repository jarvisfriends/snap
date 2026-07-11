package navigation

import (
	"image/color"
	"strconv"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"github.com/jarvisfriends/snap/page"
	"github.com/jarvisfriends/snap/styles"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// MinimalTopNav is a compact top-docked navigator rendered as one segmented
// pill: each tab is a color segment (theme palette colors, cycled when there
// are more tabs than colors) separated by the pill shape's slanted divider,
// with the active tab marked. A leading per-tab number ("1:", "2:", …) is
// optional and defaults to hidden. The number keys 1–9 select a tab directly
// only when the user has enabled "Number Key Select" in Settings (off by
// default); the router maps the digits for top-docked navs independently of
// whether the prefix is shown.
type MinimalTopNav struct {
	Pages       []Page
	ActiveIndex int
	ShowNumbers bool
	KeyMap      NavKeyMap

	// PillShape selects the segment geometry. The default, PillDiagonal, is
	// the pure-Unicode slant that renders in any font; hosts with a Nerd
	// Font can set styles.PillSlant for the seamless Powerline version.
	PillShape styles.PillShape

	width  int
	height int
	starts []int // per-tab click ranges, rebuilt in View()
	ends   []int
	page.Base
}

// NewMinimalTopNav returns a minimal top nav with the default page set and the
// number prefixes hidden.
func NewMinimalTopNav() *MinimalTopNav {
	return &MinimalTopNav{
		Pages: []Page{
			{ID: pageIDHome, Title: pageHome},
			{ID: pageIDInspector, Title: pageInspector},
			{ID: pageIDSettings, Title: pageSettings},
		},
		ActiveIndex: 0,
		ShowNumbers: false,
		KeyMap:      DefaultNavKeyMap(),
		PillShape:   styles.PillDiagonal,
	}
}

func (m *MinimalTopNav) Init() tea.Cmd { return nil }

func (m *MinimalTopNav) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		if len(m.Pages) == 0 {
			return m, nil
		}
		switch {
		case key.Matches(msg, m.KeyMap.PreviousPage):
			m.ActiveIndex = (m.ActiveIndex - 1 + len(m.Pages)) % len(m.Pages)
			return m, m.selectCmd()
		case key.Matches(msg, m.KeyMap.NextPage):
			m.ActiveIndex = (m.ActiveIndex + 1) % len(m.Pages)
			return m, m.selectCmd()
		case key.Matches(msg, m.KeyMap.Select):
			return m, m.selectCmd()
		default:
			// Number keys select directly (always active, even when prefixes hidden).
			if i, ok := digitIndex(msg.Text); ok && i < len(m.Pages) {
				m.ActiveIndex = i
				return m, m.selectCmd()
			}
		}
	}
	return m, nil
}

func (m *MinimalTopNav) selectCmd() tea.Cmd {
	idx := m.ActiveIndex
	return func() tea.Msg { return SelectedMsg{PageIndex: idx} }
}

func (m *MinimalTopNav) label(i int, title string) string {
	if m.ShowNumbers {
		return strconv.Itoa(i+1) + ":" + title
	}
	return title
}

// tabPalette returns the theme colors tabs cycle through (index mod length),
// so any number of tabs stays colored without configuration.
func (m *MinimalTopNav) tabPalette(c *styles.AppStyle) []color.Color {
	return []color.Color{c.Accent, c.Success, c.Warning, c.Error, c.SelectionBg, c.Muted}
}

// pillStyles is the segmented-pill configuration for the current theme: the
// host page's background behind the caps so the slants sit on the page color.
func (m *MinimalTopNav) pillStyles(c *styles.AppStyle) styles.PillStyles {
	shape := m.PillShape
	if shape == "" {
		shape = styles.PillDiagonal
	}
	return styles.PillStyles{Shape: shape, Base: c.Bg}
}

func (m *MinimalTopNav) View() tea.View {
	c := m.Colors()
	st := m.pillStyles(c)
	palette := m.tabPalette(c)

	segs := make([]styles.PillSegment, 0, len(m.Pages))
	texts := make([]string, 0, len(m.Pages))
	for i, p := range m.Pages {
		text := " " + m.label(i, p.Title) + " "
		if i == m.ActiveIndex {
			text = "▶" + text
		}
		texts = append(texts, text)
		segs = append(segs, styles.PillSegment{Text: text, Bg: palette[i%len(palette)]})
	}
	row := styles.SegmentedPill(segs, st)

	// Click ranges: measure the shape's cap and divider footprints once by
	// rendering trivial pills, then walk the segment texts — cell math that
	// holds for every shape (Fade's caps are two cells, the rest one). The
	// caps and dividers are folded into the adjacent tab's range so every
	// cell of the row is clickable: the left cap belongs to the first tab,
	// each divider to the tab before it, the right cap to the last.
	capsW := lipgloss.Width(styles.SegmentedPill(
		[]styles.PillSegment{{Bg: palette[0]}}, st))
	divW := lipgloss.Width(styles.SegmentedPill(
		[]styles.PillSegment{{Bg: palette[0]}, {Bg: palette[1%len(palette)]}}, st)) - capsW
	capL := capsW / 2
	m.starts = make([]int, len(m.Pages))
	m.ends = make([]int, len(m.Pages))
	x := 0
	for i, text := range texts {
		w := lipgloss.Width(text)
		if i == 0 {
			w += capL
		}
		if i == len(texts)-1 {
			w += capsW - capL // right cap
		} else {
			w += divW
		}
		m.starts[i] = x
		if w > 0 {
			m.ends[i] = x + w - 1
		} else {
			m.ends[i] = x
		}
		x += w
	}

	v := tea.NewView(row)
	v.BackgroundColor = c.Styles.TextOnBg.GetBackground()
	v.ForegroundColor = c.Styles.TextOnBg.GetForeground()
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.OnMouse = func(mm tea.MouseMsg) tea.Cmd {
		// Horizontal wheel scrolls through the pages, matching the tab bar.
		if d := horizontalWheelDelta(mm); d != 0 && len(m.Pages) > 0 {
			me := mm.Mouse()
			if me.Y < 0 || me.Y >= lipgloss.Height(v.Content) {
				return nil
			}
			m.ActiveIndex = (m.ActiveIndex + d + len(m.Pages)) % len(m.Pages)
			return m.selectCmd()
		}
		switch ev := mm.(type) {
		case tea.MouseClickMsg, tea.MouseReleaseMsg:
			me := ev.Mouse()
			if me.Button != tea.MouseLeft {
				return nil
			}
			if me.Y < 0 || me.Y >= lipgloss.Height(v.Content) {
				return nil
			}
			for i := range m.Pages {
				if me.X >= m.starts[i] && me.X <= m.ends[i] {
					m.ActiveIndex = i
					return m.selectCmd()
				}
			}
		}
		return nil
	}
	return v
}

// SetShowNumbers toggles the leading per-tab number prefix.
func (m *MinimalTopNav) SetShowNumbers(show bool) { m.ShowNumbers = show }

// Width reports the horizontal layout space consumed; a top nav stacks above
// content and reserves no side width.
func (m *MinimalTopNav) Width() int  { return 0 }
func (m *MinimalTopNav) Height() int { return lipgloss.Height(m.View().Content) }

func (m *MinimalTopNav) Dock() Side           { return DockTop }
func (m *MinimalTopNav) GetPages() []Page     { return m.Pages }
func (m *MinimalTopNav) SetPages(p []Page)    { m.Pages = p }
func (m *MinimalTopNav) SetActiveIndex(i int) { m.ActiveIndex = i }
func (m *MinimalTopNav) GetActiveIndex() int  { return m.ActiveIndex }

// digitIndex maps a single key string "1".."9" to a zero-based index.
func digitIndex(s string) (int, bool) {
	if utf8.RuneCountInString(s) == 1 && s[0] >= '1' && s[0] <= '9' {
		return int(s[0] - '1'), true
	}
	return 0, false
}

var (
	_ Navigator     = (*MinimalTopNav)(nil)
	_ NumberLabeled = (*MinimalTopNav)(nil)
)
