// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Command snap_input bundles every snap example as a subcommand of one
// small, scriptable input tool: run `snap_input <example>`, the terminal UI
// takes over the screen (on stderr), and the confirmed value is written to
// stdout — so any shell can capture it:
//
//	date=$(snap_input datepicker)   # -> 2026-07-12
//
// Canceling prints nothing and exits 1. See examples/USAGE.md for the full
// calling convention; each subcommand lives in examples/<name>.
package main

import (
	"fmt"
	"os"
	"sort"

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

// commands maps each subcommand to its entry point and the one-line
// description shown by `snap_input help` (kept in sync with examples/USAGE.md).
var commands = map[string]struct {
	run  func()
	desc string
}{
	"cellcanvas":   {cellcanvas.Run, "Whole-cell canvas and gradient helpers for animated truecolor effects."},
	"charts":       {charts.Run, "Sparklines, horizontal bars, pie, and sankey charts, stretch-to-fit sized."},
	"datepicker":   {datepicker.Run, "Calendar date picker: click-to-confirm days, month/year focus, keyboard/wheel paging."},
	"dependencies": {dependencies.Run, "Build info and dependency reader (status info modal)."},
	"forms":        {forms.Run, "Parser-backed form validation: required fields, durations, ISO dates, list splitting."},
	"linechart":    {linechart.Run, "Braille line chart showing rolling streams, compact terminal-cell rendering."},
	"menu":         {menu.Run, "Right-click context menu with keyboard parity and terminal-edge clamping."},
	"navigation":   {navigation.Run, "Tabs, Sidebar, and MinimalTopNav behind one navigator contract."},
	"pickers":      {pickers.Run, "Drive-aware directory picker and path-editing interactions."},
	"pills":        {pills.Run, "Segmented pills, shape variants, and breadcrumb styling helpers."},
	"scrollbar":    {scrollbar.Run, "Scrollbar presets with click/drag mapping through OffsetAt."},
	"status":       {status.Run, "Status bar surfaces plus notification toast/history flows."},
	"table":        {table.Run, "Sortable/filterable table with row activation, keyboard/mouse support."},
	"timepicker":   {timepicker.Run, "HH:MM(:SS) time field: per-column dropdowns, type-ahead, validation."},
}

func usage(w *os.File) {
	_, _ = fmt.Fprintf(w, "Usage: %s <command> [--no-help]\n\nCommands:\n", os.Args[0])
	names := make([]string, 0, len(commands))
	for n := range commands {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		_, _ = fmt.Fprintf(w, "  %-13s %s\n", n, commands[n].desc)
	}
	_, _ = fmt.Fprintf(w, "\nEach command prints the confirmed value to stdout (UI is on stderr);\ncanceling prints nothing and exits 1. See examples/USAGE.md.\n")
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		usage(os.Stderr)
		if len(os.Args) >= 2 {
			os.Exit(0)
		}
		os.Exit(2)
	}
	if os.Args[1] == "--version" {
		fmt.Println(exui.Version)
		return
	}
	cmd, ok := commands[os.Args[1]]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "%s: unknown command %q\n\n", os.Args[0], os.Args[1])
		usage(os.Stderr)
		os.Exit(2)
	}
	// Splice the subcommand out so each example's exui.Init flag parsing
	// (--no-help, --version) sees only its own flags.
	os.Args = append(os.Args[:1], os.Args[2:]...)
	cmd.run()
}
