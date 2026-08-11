// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Command snap_input bundles every snap example as a subcommand of one
// small, scriptable input tool: run `snap_input <example>`, the terminal UI
// takes over the screen (on stderr), and the confirmed value is written to
// stdout — so any shell can capture it:
//
//	date=$(snap_input datepicker)   # -> 2026-07-12
//
// Canceling prints nothing and exits 1.
//
// With no arguments it instead tours every example in one program: tab and
// shift+tab (or alt+←/→) move between pages, ctrl+b shows the page strip,
// ctrl+t cycles the theme, and q ends the tour, printing each visited page's
// confirmed value namespaced by page. See examples/USAGE.md for the full
// calling convention; each subcommand lives in examples/<name>.
package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/cellcanvas"
	"github.com/jarvisfriends/snap/examples/charts"
	"github.com/jarvisfriends/snap/examples/datepicker"
	"github.com/jarvisfriends/snap/examples/dependencies"
	"github.com/jarvisfriends/snap/examples/forms"
	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/examples/linechart"
	"github.com/jarvisfriends/snap/examples/menu"
	"github.com/jarvisfriends/snap/examples/navigation"
	"github.com/jarvisfriends/snap/examples/pickers"
	"github.com/jarvisfriends/snap/examples/pills"
	"github.com/jarvisfriends/snap/examples/scrollbar"
	"github.com/jarvisfriends/snap/examples/status"
	"github.com/jarvisfriends/snap/examples/table"
	"github.com/jarvisfriends/snap/examples/timepicker"
)

// pageDef is one subcommand: its name, the one-line description shown by
// `snap_input help` (kept in sync with examples/USAGE.md), and the
// constructor that builds a fresh model for it.
type pageDef struct {
	name string
	desc string
	new  func() exui.Page
}

// pages is the tour order, which is also the order `help` lists commands in.
// It leads with the value-producing pickers, then the read-only visual demos.
var pages = []pageDef{
	{"datepicker", "Calendar date picker: click-to-confirm days, month/year focus, keyboard/wheel paging.", datepicker.New},
	{"timepicker", "HH:MM(:SS) time field: per-column dropdowns, type-ahead, validation.", timepicker.New},
	{"table", "Sortable/filterable table with row activation, keyboard/mouse support.", table.New},
	{"pickers", "Drive-aware directory picker and path-editing interactions.", pickers.New},
	{"forms", "Parser-backed form validation: required fields, durations, ISO dates, list splitting.", forms.New},
	{"menu", "Right-click context menu with keyboard parity and terminal-edge clamping.", menu.New},
	{"navigation", "Tabs, Sidebar, and MinimalTopNav behind one navigator contract.", navigation.New},
	{"pills", "Segmented pills, shape variants, and breadcrumb styling helpers.", pills.New},
	{"scrollbar", "Scrollbar presets with click/drag mapping through OffsetAt.", scrollbar.New},
	{"charts", "Sparklines, horizontal bars, pie, and sankey charts, stretch-to-fit sized.", charts.New},
	{"linechart", "Braille line chart showing rolling streams, compact terminal-cell rendering.", linechart.New},
	{"cellcanvas", "Whole-cell canvas and gradient helpers for animated truecolor effects.", cellcanvas.New},
	{"status", "Status bar surfaces plus notification toast/history flows.", status.New},
	{"dependencies", "Build info and dependency reader (status info modal).", dependencies.New},
}

// find returns the page with the given name.
func find(name string) (pageDef, bool) {
	for _, p := range pages {
		if p.name == name {
			return p, true
		}
	}
	return pageDef{}, false
}

// nameCol is the width of the command column in `help`. Padding goes through
// lipgloss so it is measured in terminal cells, like every other aligned
// column in this module.
var nameCol = lipgloss.NewStyle().Width(14)

func usage(w *os.File) {
	_, _ = fmt.Fprintf(w, "Usage: %s [command] [--no-help] [--output FORMAT]\n\nCommands:\n", os.Args[0])
	for _, p := range pages {
		_, _ = fmt.Fprintf(w, "  %s%s\n", nameCol.Render(p.name), p.desc)
	}
	_, _ = fmt.Fprintf(w, "\nEach command prints the confirmed value to stdout (UI is on stderr);\ncanceling prints nothing and exits 1.\n")
	_, _ = fmt.Fprintf(w, "With no command, every example runs as one tour: tab/shift+tab (or\nalt+←/→) change page, ctrl+b shows the page strip, ctrl+t cycles the\ntheme, q ends it. See examples/USAGE.md.\n")
}

func main() {
	args := os.Args[1:]
	switch {
	case len(args) > 0 && (args[0] == "help" || args[0] == "-h" || args[0] == "--help"):
		usage(os.Stderr)
		os.Exit(0)
	case len(args) > 0 && args[0] == "--version":
		fmt.Println(exui.Version)
		return
	}

	// No command tours everything; a command runs that one page. Either way
	// the same host model drives it (see tour.go).
	defs := pages
	if len(args) > 0 && args[0][0] != '-' {
		p, ok := find(args[0])
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "%s: unknown command %q\n\n", os.Args[0], args[0])
			usage(os.Stderr)
			os.Exit(2)
		}
		defs = []pageDef{p}
		// Splice the subcommand out so exui's flag parsing (--no-help,
		// --output, --version) sees only its own flags.
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}
	run(defs)
}

// run drives the host to completion and writes whatever was confirmed.
// Results are read and pages torn down before any exui.Finish call, since
// those exit the process and would skip a defer here.
func run(defs []pageDef) {
	exui.Init()
	t := newTour(defs)
	final, err := exui.Program(t).Run()
	if ft, ok := final.(*tour); ok {
		t = ft
	}
	fields := t.Fields()
	t.Cleanup()
	if err != nil {
		exui.Fatal(err)
	}
	if len(fields) == 0 {
		exui.Finish(false) // nothing confirmed: print nothing, exit 1
	}
	exui.FinishFields(true, fields...)
}
