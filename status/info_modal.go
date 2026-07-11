package status

import (
	"fmt"
	"strings"

	"github.com/jarvisfriends/snap/dependencies"
	"github.com/jarvisfriends/snap/geom"
	"github.com/jarvisfriends/snap/keys"

	"github.com/jarvisfriends/snap/styles"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// CloseInfoModalMsg is sent (via tea.Cmd) when the user clicks outside the
// info modal. The router handles it in Update and calls m.infoModal.Close().
type CloseInfoModalMsg struct{}

// InfoModalScrollMsg is sent when the user scrolls the mouse wheel inside the
// info modal bounds. The router forwards it to the modal's scroll methods.
type InfoModalScrollMsg struct{ Up bool }

// infoModalMaxW is the maximum terminal columns the modal box will occupy.
const infoModalMaxW = 90

const (
	// Keep a minimum breathing room around the modal when the terminal allows it.
	infoModalMinMarginX = 2
	infoModalMinMarginY = 2
)

// InfoModal is a full-screen centered modal that shows the app version, build
// metadata, and a scrollable dependency list. It is owned by the router, which
// composites it on the canvas and routes key/mouse events into it.
type InfoModal struct {
	visible    bool
	vp         viewport.Model
	availableW int
	availableH int
	keys       *keys.AppKeyMap
	appName    string
	appVersion string
}

func NewInfoModal() *InfoModal {
	return &InfoModal{
		keys: keys.DefaultKeyMap(),
		vp:   viewport.New(),
	}
}

// IsVisible reports whether the modal is currently open.
func (m *InfoModal) IsVisible() bool { return m.visible }

func (m *InfoModal) Init() tea.Cmd {
	return nil
}

func (m *InfoModal) Name() string               { return "InfoModal" }
func (m *InfoModal) SetKeys(km *keys.AppKeyMap) { m.keys = km }
func (m *InfoModal) SetAppName(name string)     { m.appName = name }
func (m *InfoModal) SetVersion(v string)        { m.appVersion = v }

var _ tea.Model = (*InfoModal)(nil)

// Open opens the modal. It rebuilds viewport content to match the current
// screen dimensions and styles.
func (m *InfoModal) Open(screenW, screenH int) {
	m.availableW = screenW
	m.availableH = screenH
	m.visible = true
	m.rebuildContent()
}

// Close hides the modal.
func (m *InfoModal) Close() { m.visible = false }

// Toggle opens the modal when closed and closes it when open.
func (m *InfoModal) Toggle(screenW, screenH int) {
	if m.visible {
		m.Close()
	} else {
		m.Open(screenW, screenH)
	}
}

// Resize updates the stored screen dimensions and rebuilds content if the
// modal is currently visible (so the viewport fills the new terminal size).
func (m *InfoModal) Resize(screenW, screenH int) {
	m.availableW = screenW
	m.availableH = screenH
	if m.visible {
		m.rebuildContent()
	}
}

// Bounds returns the screen-space bounding box (x, y, width, height) of the
// rendered modal. The router uses this for outside-click detection.
func (m *InfoModal) Bounds() (x, y, w, h int) {
	bw, bh, bx, by := m.boxDims()
	return bx, by, bw, bh
}

// ScrollUp scrolls the viewport up by 3 lines.
func (m *InfoModal) ScrollUp() { m.vp.ScrollUp(3) }

// ScrollDown scrolls the viewport down by 3 lines.
func (m *InfoModal) ScrollDown() { m.vp.ScrollDown(3) }

// PageUp scrolls the viewport up by one page.
func (m *InfoModal) PageUp() { m.vp.PageUp() }

// PageDown scrolls the viewport down by one page.
func (m *InfoModal) PageDown() { m.vp.PageDown() }

// GotoTop scrolls to the top of the content.
func (m *InfoModal) GotoTop() { m.vp.GotoTop() }

// GotoBottom scrolls to the bottom of the content.
func (m *InfoModal) GotoBottom() { m.vp.GotoBottom() }

// HandleMouse processes a screen-coordinate mouse event while the modal is
// open — the pointer twin of the key handling in Update, so hosts forward
// events here instead of hand-rolling Bounds hit-tests, wheel routing, and
// outside-click detection: the wheel scrolls the content (anywhere — the
// modal owns the screen), a click outside the box closes the modal and
// returns the same CloseInfoModalMsg cmd Dismiss produces, and every other
// event is consumed. handled=false means the modal is closed and the host
// should process the event itself.
func (m *InfoModal) HandleMouse(msg tea.MouseMsg) (cmd tea.Cmd, handled bool) {
	if !m.visible {
		return nil, false
	}
	switch ev := msg.(type) {
	case tea.MouseWheelMsg:
		switch ev.Mouse().Button {
		case tea.MouseWheelUp:
			m.ScrollUp()
		case tea.MouseWheelDown:
			m.ScrollDown()
		}
		return nil, true
	case tea.MouseClickMsg:
		me := ev.Mouse()
		bx, by, bw, bh := m.Bounds()
		if !(geom.Rect{X: bx, Y: by, W: bw, H: bh}).Contains(me.X, me.Y) {
			m.Close()
			return func() tea.Msg { return CloseInfoModalMsg{} }, true
		}
		return nil, true
	}
	return nil, true
}

// boxDims computes the box total width, total height, and its top-left corner
// position on the terminal screen.
//
//	Total layout:
//	  ╭──────────────────────────────╮  ← row `by`
//	  │ title                        │
//	  │ ─────────────────────────── │
//	  │ <viewport lines>             │
//	  │ ─────────────────────────── │
//	  │ footer                       │
//	  ╰──────────────────────────────╯  ← row `by + boxH - 1`
func (m *InfoModal) boxDims() (boxW, boxH, boxX, boxY int) {
	if m.availableW <= 0 || m.availableH <= 0 {
		return 0, 0, 0, 0
	}

	// Width is clamped by max modal width while preserving side margins.
	boxW = min(infoModalMaxW, m.availableW-infoModalMinMarginX*2)
	if boxW < 20 {
		boxW = m.availableW
	}
	if boxW > m.availableW {
		boxW = m.availableW
	}

	// Height fills the available terminal area minus top/bottom margins,
	// then remains centered. This keeps the box usable across small and large
	// terminals and keeps hit-tests aligned with what is rendered.
	boxH = m.availableH - infoModalMinMarginY*2
	if boxH < 10 {
		boxH = m.availableH
	}
	if boxH > m.availableH {
		boxH = m.availableH
	}

	boxX = max((m.availableW-boxW)/2, 0)
	boxY = max((m.availableH-boxH)/2, 0)
	return boxW, boxH, boxX, boxY
}

// modalChromeRows is the number of rendered lines inside the box border that
// are not part of the scrollable viewport: the title line, two separator
// rules, and the footer line (see the boxDims diagram above).
const modalChromeRows = 4

// modalFrameStyle returns the border+padding configuration shared by every
// caller that needs to measure or draw the modal's outer frame. Keeping the
// geometry in one place (rather than duplicating the border type and padding
// amounts as separate literals in vpDims and View) means GetHorizontalFrameSize
// / GetVerticalFrameSize always reflect what actually gets rendered. It is a
// var (not a func literal) so tests can swap in other border/padding
// combinations and confirm vpDims and View still agree.
var modalFrameStyle = func() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
}

// vpDims returns the width and height of the inner viewport (content area
// minus the outer box's border/padding frame and the fixed chrome rows).
func (m *InfoModal) vpDims() (vpW, vpH int) {
	boxW, boxH, _, _ := m.boxDims()
	frame := modalFrameStyle()
	vpW = boxW - frame.GetHorizontalFrameSize()
	vpH = boxH - frame.GetVerticalFrameSize() - modalChromeRows
	if vpW < 10 {
		vpW = 10
	}
	if vpH < 1 {
		vpH = 1
	}
	return vpW, vpH
}

// rebuildContent recreates the viewport with content matching the current
// dimensions and active styles. Called on Open() and Resize().
func (m *InfoModal) rebuildContent() {
	vpW, vpH := m.vpDims()
	c := styles.Active()

	mutedStyle := c.Styles.Subtitle
	accentStyle := c.Styles.Title
	dimStyle := c.Styles.FilterDim

	lines := m.buildInfoLines(vpW, accentStyle, mutedStyle, dimStyle)

	m.vp = viewport.New(viewport.WithWidth(vpW), viewport.WithHeight(vpH))
	m.vp.SetContentLines(lines)
}

func (m *InfoModal) buildInfoLines(
	vpW int,
	accentStyle, mutedStyle, dimStyle lipgloss.Style,
) []string {
	var lines []string
	info := dependencies.ExpandedBuildInfo()
	if info == nil {
		return append(lines, mutedStyle.Render("  (build info unavailable)"))
	}
	rev := info.VCS.Revision
	if lipgloss.Width(rev) > 8 {
		rev = string([]rune(rev)[:8])
	}
	builtAt := ""
	if info.VCS.Time != nil {
		builtAt = " built " + info.VCS.Time.Format("2006-01-02")
	}
	modified := ""
	if info.VCS.Modified != nil && *info.VCS.Modified {
		modified = " (modified)"
	}
	lines = append(lines, dimStyle.Render(fmt.Sprintf(
		"  Go: %-10s  OS: %s/%s  CPUs: %d",
		info.GoVersion, info.Runtime.GOOS, info.Runtime.GOARCH, info.Runtime.CPUs,
	)))
	// "Executable", st.Launch.Executable,
	// "Args", fmt.Sprintf("%v", st.Launch.Args),
	// "Work Dir", st.Launch.WorkDir,
	// "User@Host", fmt.Sprintf("%s@%s", st.Launch.Username, st.Launch.Hostname),
	if rev != "" {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  Rev: %s%s%s", rev, builtAt, modified)))
	}
	// Package column padded by display cells (A-5): fmt's %-50s pads by
	// bytes, which misaligns the Version column for any non-ASCII path.
	const pkgColW = 50
	pkgCell := lipgloss.NewStyle().Width(pkgColW)
	lines = append(
		lines,
		"",
		accentStyle.Render(
			"  Dependencies",
		)+dimStyle.Render(
			fmt.Sprintf("  (total: %d)", len(info.Dependencies)),
		),
		mutedStyle.Render("  "+pkgCell.Render("Package")+"  Version"),
	)
	sepLen := min(vpW-2, 72)
	lines = append(lines, mutedStyle.Render("  "+strings.Repeat("─", sepLen)))
	for _, dep := range info.Dependencies {
		path := dep.Path
		if w := ansi.StringWidth(path); w > pkgColW {
			// Keep the tail — the module name matters more than the host.
			path = ansi.TruncateLeft(path, w-pkgColW+1, "…")
		}
		line := "  " + pkgCell.Render(path) + "  " + dep.Version
		if dep.Replace != "" {
			line += "  ⇒ " + dep.Replace
		}
		lines = append(lines, mutedStyle.Render(line))
	}
	return lines
}

func (m *InfoModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyPressMsg:
		keyMsg := msg
		switch {
		case key.Matches(keyMsg, m.keys.Dismiss):
			m.Close()
			return m, func() tea.Msg { return CloseInfoModalMsg{} }
		case key.Matches(keyMsg, m.keys.Up):
			m.ScrollUp()
			return m, nil
		case key.Matches(keyMsg, m.keys.Down):
			m.ScrollDown()
			return m, nil
		case key.Matches(keyMsg, m.keys.PageUp):
			m.PageUp()
			return m, nil
		case key.Matches(keyMsg, m.keys.PageDown):
			m.PageDown()
			return m, nil
		case key.Matches(keyMsg, m.keys.Top):
			m.GotoTop()
			return m, nil
		case key.Matches(keyMsg, m.keys.Bottom):
			m.GotoBottom()
			return m, nil
		}
	}

	return m, nil
}

// Render returns the fully rendered modal string and its top-left position
// on the terminal screen. Returns ("") when the modal is not visible.
func (m *InfoModal) View() (content tea.View) {
	if !m.visible || m.availableW == 0 || m.availableH == 0 {
		return tea.NewView("")
	}
	c := styles.Active()
	boxW, _, _, _ := m.boxDims()
	vpW, _ := m.vpDims()

	headerStyle := c.Styles.Title.Bold(true)
	mutedStyle := c.Styles.Subtitle
	sepStyle := c.Styles.Title

	// Title centered in the content area.
	name := m.appName
	if name == "" {
		name = "TUI Base"
	}
	version := m.appVersion
	if version == "" {
		// Hosts inject their build version (tui-base does via SetAppVersion);
		// standalone consumers see the development placeholder.
		version = "(dev)"
	}
	titleText := "ℹ  " + name + "  " + version
	titleLine := lipgloss.PlaceHorizontal(vpW, lipgloss.Center, headerStyle.Render(titleText))

	sep := sepStyle.Render(strings.Repeat("─", vpW))

	// Scroll percentage badge shown when the content overflows.
	scrollBadge := ""
	if m.vp.TotalLineCount() > m.vp.VisibleLineCount() {
		scrollBadge = fmt.Sprintf("  %d%%", int(m.vp.ScrollPercent()*100))
	}
	footerText := "↑/↓ • PgUp/PgDn • Esc or click outside to close" + scrollBadge
	footerLine := lipgloss.PlaceHorizontal(vpW, lipgloss.Center, mutedStyle.Render(footerText))

	inner := lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine,
		sep,
		m.vp.View(),
		sep,
		footerLine,
	)

	// Width(boxW) = total rendered width including border+padding, so the
	// content area lipgloss produces here always matches vpDims (same
	// modalFrameStyle backs both).
	borderStyle := modalFrameStyle().
		BorderForeground(c.Accent).
		Background(c.Styles.TextOnBg.GetBackground()).
		Foreground(c.Styles.TextOnBg.GetForeground())

	rendered := borderStyle.Width(boxW).Render(inner)
	return tea.NewView(rendered)
}
