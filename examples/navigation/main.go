// Command navigation is a script-usable page picker that shows all three
// snap/navigation styles behind the one Navigator contract: n swaps between
// Sidebar, Tabs, and MinimalTopNav at runtime (the swap the contract exists
// for), arrows/clicks/wheel move between pages, and Enter writes the active
// page's ID to stdout (the TUI itself renders on stderr):
//
//	page=$(go run ./examples/navigation)
//
// --no-help hides the status bar. Quitting (q/esc) prints nothing, exit 1.
package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/navigation"
)

// demoPages is the page set every navigator style shows.
func demoPages() []navigation.Page {
	return []navigation.Page{
		{ID: "home", Title: "Home"},
		{ID: "metrics", Title: "Metrics"},
		{ID: "logs", Title: "Logs"},
		{ID: "settings", Title: "Settings"},
		{ID: "about", Title: "About"},
	}
}

// navStyles builds the three navigators sharing one page set; the active
// index carries over when the user swaps styles.
func navStyles() []navigation.Navigator {
	styles := []navigation.Navigator{
		navigation.New(), // sidebar (docks left)
		navigation.NewTabs(),
		navigation.NewMinimalTopNav(),
	}
	for _, n := range styles {
		n.SetPages(demoPages())
	}
	return styles
}

var styleNames = []string{"Sidebar", "Tabs", "MinimalTopNav"}

type demoApp struct {
	navs   []navigation.Navigator
	style  int
	picked string
	chrome *exui.Chrome
	w, h   int
}

func newDemo() *demoApp {
	return &demoApp{
		navs: navStyles(),
		chrome: exui.NewChrome(
			exui.Bind("↑/↓/←/→", "page"),
			exui.Bind("click/wheel", "page"),
			exui.Bind("n", "nav style"),
			exui.Bind("enter", "pick"),
			exui.Bind("q", "quit"),
		),
	}
}

func (a *demoApp) nav() navigation.Navigator { return a.navs[a.style] }

func (a *demoApp) Init() tea.Cmd { return a.nav().Init() }

func (a *demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.w, a.h = msg.Width, msg.Height
		a.chrome.SetWidth(msg.Width)
		return a, a.forwardSize()
	case tea.MouseMsg:
		// Pointer input arrives via the root view's OnMouse below.
		return a, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return a, tea.Quit
		case "enter":
			pages := a.nav().GetPages()
			if i := a.nav().GetActiveIndex(); i >= 0 && i < len(pages) {
				a.picked = pages[i].ID
			}
			return a, tea.Quit
		case "n":
			active := a.nav().GetActiveIndex()
			a.style = (a.style + 1) % len(a.navs)
			a.nav().SetActiveIndex(active)
			return a, tea.Batch(a.nav().Init(), a.forwardSize())
		}
	}
	m, cmd := a.nav().Update(msg)
	if n, ok := m.(navigation.Navigator); ok {
		a.navs[a.style] = n
	}
	return a, cmd
}

// forwardSize hands the navigator the space it may occupy (the window minus
// the help bar).
func (a *demoApp) forwardSize() tea.Cmd {
	m, cmd := a.nav().Update(tea.WindowSizeMsg{Width: a.w, Height: max(a.h-a.chrome.Height(), 1)})
	if n, ok := m.(navigation.Navigator); ok {
		a.navs[a.style] = n
	}
	return cmd
}

// body renders the fake page pane for the active page.
func (a *demoApp) body(w, h int) string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	pages := a.nav().GetPages()
	title := ""
	if i := a.nav().GetActiveIndex(); i >= 0 && i < len(pages) {
		title = pages[i].Title
	}
	pane := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(title),
		dim.Render("navigator: "+styleNames[a.style]+" (swappable at runtime — same Navigator contract)"),
	)
	return lipgloss.NewStyle().Width(max(w, 1)).Height(max(h, 1)).Padding(1, 2).Render(pane)
}

func (a *demoApp) View() tea.View {
	nav := a.nav()
	nv := nav.View()

	availH := max(a.h-a.chrome.Height(), 1)
	var content string
	if nav.Dock() == navigation.DockLeft {
		content = lipgloss.JoinHorizontal(lipgloss.Top,
			nv.Content, a.body(a.w-nav.Width(), availH))
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left,
			nv.Content, a.body(a.w, availH-nav.Height()))
	}

	v := tea.NewView(a.chrome.Attach(content, a.h))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	// The navigator hit-tests its own rendered region; it sits at the frame
	// origin for every dock, so coordinates pass through untranslated.
	v.OnMouse = nv.OnMouse
	return v
}

func main() {
	exui.Init()
	final, err := exui.Program(newDemo()).Run()
	if err != nil {
		exui.Fatal(err)
	}
	if a, ok := final.(*demoApp); ok && a.picked != "" {
		exui.Finish(true, a.picked)
	}
	exui.Finish(false)
}
