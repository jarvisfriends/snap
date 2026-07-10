// Command pickers demos snap/pickers' DirPicker: keyboard and wheel walk a
// small generated directory tree; Space selects, Ctrl+S picks the browsed
// folder, Esc aborts.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/pickers"
	"github.com/jarvisfriends/snap/uifx"
)

type demoApp struct {
	dp *pickers.DirPicker
}

func (a demoApp) Init() tea.Cmd { return a.dp.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.MouseMsg); ok {
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
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run holds the body so the temp-tree cleanup runs on every path (os.Exit
// in main would skip the defer).
func run() error {
	root := makeTree()
	defer os.RemoveAll(root) //nolint:errcheck // temp demo tree

	app := demoApp{dp: pickers.NewDirPicker(root)}
	final, err := tea.NewProgram(app).Run()
	if err != nil {
		return err
	}
	if a, ok := final.(demoApp); ok && a.dp.Done {
		rel, _ := filepath.Rel(root, a.dp.Value())
		fmt.Printf("Selected: %s\n", rel)
	}
	return nil
}
