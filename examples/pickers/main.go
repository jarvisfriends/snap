// Command pickers is a script-usable directory prompt built on snap/pickers'
// DirPicker: walk the tree, Space selects, Ctrl+S picks the browsed folder,
// and the chosen path (relative to the demo tree) is written to stdout (the
// TUI itself renders on stderr):
//
//	dir=$(go run ./examples/pickers)
//
// --no-help hides the status bar. Esc aborts: nothing printed, exit 1.
package main

import (
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/pickers"
	"github.com/jarvisfriends/snap/uifx"
)

type demoApp struct {
	dp     *pickers.DirPicker
	chrome *exui.Chrome
	height int
}

func newDemo(root string) demoApp {
	return demoApp{
		dp: pickers.NewDirPicker(root),
		chrome: exui.NewChrome(
			exui.Bind("↑/↓", "move"),
			exui.Bind("←/→", "close/open"),
			exui.Bind("space", "select"),
			exui.Bind("ctrl+s", "pick browsed"),
			exui.Bind("esc", "cancel"),
		),
	}
}

func (a demoApp) Init() tea.Cmd { return a.dp.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.height = msg.Height
		a.chrome.SetWidth(msg.Width)
	case tea.MouseMsg:
		// Mouse arrives via the root view's OnMouse (the picker's); the
		// runtime also delivers it here — ignore to avoid double handling.
		return a, nil
	}
	model, cmd := a.dp.Update(msg)
	if dp, ok := model.(*pickers.DirPicker); ok {
		a.dp = dp
	}
	if a.dp.Done || a.dp.Aborted {
		return a, tea.Quit
	}
	return a, cmd
}

func (a demoApp) View() tea.View {
	v := a.dp.View()
	v.SetContent(a.chrome.Attach(v.Content, a.height))
	v.MouseMode = uifx.LevelMedium.MouseMode()
	v.AltScreen = true
	return v
}

// makeTree builds a deterministic little directory tree so the demo (and
// its VHS tape) always shows the same content.
func makeTree() string {
	root, err := os.MkdirTemp("", "snap-pickers-demo")
	if err != nil {
		panic(err)
	}
	for _, d := range []string{
		"projects/alpha", "projects/beta", "projects/gamma",
		"documents/reports", "music",
	} {
		_ = os.MkdirAll(filepath.Join(root, d), 0o750)
	}
	return root
}

func main() {
	exui.Init()
	picked, ok, err := run()
	if err != nil {
		exui.Fatal(err)
	}
	exui.Finish(ok, picked)
}

// run holds the body so the temp-tree cleanup runs on every path (os.Exit
// in exui.Finish would skip a defer in main).
func run() (picked string, ok bool, err error) {
	root := makeTree()
	defer os.RemoveAll(root) //nolint:errcheck // temp demo tree

	final, err := exui.Program(newDemo(root)).Run()
	if err != nil {
		return "", false, err
	}
	if a, ok := final.(demoApp); ok && a.dp.Done {
		rel, relErr := filepath.Rel(root, a.dp.Value())
		if relErr != nil {
			rel = a.dp.Value()
		}
		return rel, true, nil
	}
	return "", false, nil
}
