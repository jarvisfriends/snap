// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Package pickers implements the `snap_input pickers` subcommand: a script-usable directory prompt built on snap/pickers'
// DirPicker: walk the tree, Space selects, Ctrl+S picks the browsed folder,
// and the chosen path (relative to the demo tree) is written to stdout (the
// TUI itself renders on stderr):
//
//	dir=$(go run ./examples/snap_input pickers)
//
// --no-help hides the status bar. Esc aborts: nothing printed, exit 1.
package pickers

import (
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/pickers"
	"github.com/jarvisfriends/snap/styles"
)

type demoApp struct {
	dp     *pickers.DirPicker
	chrome *exui.Chrome
	height int
}

func newDemo(root string) demoApp {
	dp := pickers.NewDirPicker(root)
	// Theme the picker from the shared palette instead of its fixed ANSI
	// defaults, so it follows tint changes like every other component.
	dp.Styles = styles.PickerStyles(exui.Theme())
	// The status bar below carries the key hints; the picker's own help
	// line would show the same thing twice.
	dp.HideHelp = true
	return demoApp{
		dp: dp,
		chrome: exui.NewChrome(
			exui.Bind("↑↓", "move"),
			exui.Bind("←→", "open back"),
			exui.Bind("space", "select"),
			exui.Bind("ctrl+s", "pick browsed"),
			exui.Bind("esc", "cancel"),
		),
	}
}

func (a demoApp) Init() tea.Cmd { return a.dp.Init() }

func (a demoApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd, done := a.chrome.Update(msg); done {
		return a, cmd
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.height = msg.Height
		a.chrome.SetSize(msg.Width, msg.Height)
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
	a.chrome.Frame(&v, a.height)
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
		_ = os.MkdirAll(filepath.Join(root, d), 0o700)
	}
	return root
}

// Run is the snap_input subcommand entry point.
func Run() {
	exui.Init()
	picked, ok, err := run()
	if err != nil {
		exui.Fatal(err)
	}
	exui.FinishFields(ok, exui.F("dir", picked))
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
