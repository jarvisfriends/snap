// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package exui

import (
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
)

// This file defines the contract between an example package and its two
// hosts: `snap_input <name>`, which runs one page as a scriptable prompt, and
// `snap_input` with no arguments, which tours every page in one program.
//
// Each example exports New() returning a Page. The only difference between
// the two hosts is what confirming means — see Confirm.

// Page is one example: a tea.Model that can report what the user confirmed.
// Result is read when the page is left (the tour) or when the program ends
// (single-page mode), and is empty until the user actually confirms.
type Page interface {
	tea.Model

	// Result returns the confirmed value(s) for this page, or nil if the
	// user has not confirmed anything.
	Result() []Field
}

// Shelled is implemented by every example page: it hands the host the page's
// Chrome. The tour uses it to defer to open shell overlays and to the page's
// capture predicate before claiming its own chords.
type Shelled interface {
	Shell() *Chrome
}

// TextCapturer is implemented by pages that can be capturing keystrokes — a
// table filter box, a path editor, a huh form. While Capturing reports true
// the shell leaves plain-letter chords alone, so typing "q" into a filter
// cannot quit the program, and the tour does not steal tab from the page.
//
// A page that never captures text simply omits the method.
type TextCapturer interface {
	Capturing() bool
}

// Cleaner is implemented by pages that allocate something needing teardown —
// the pickers page builds a temp directory tree to browse. The host calls
// Cleanup once the program has ended, because Finish's os.Exit would skip
// any defer the page set up itself.
type Cleaner interface {
	Cleanup()
}

// tourMode reports whether pages are hosted by the multi-page tour. It is
// process-global (one program, one host) and atomic only because Confirm can
// be reached from a View's mouse callback as well as Update.
var tourMode atomic.Bool

// SetTourMode switches the confirm semantics for every page in this process.
// The tour sets it before constructing any page.
func SetTourMode(on bool) { tourMode.Store(on) }

// InTour reports whether the multi-page tour is hosting.
func InTour() bool { return tourMode.Load() }

// Confirm is what a page returns as its tea.Cmd when the user confirms a
// value (enter, a click-confirm, a second click on a date).
//
// Single-page mode quits, so `date=$(snap_input datepicker)` stays a one-shot
// prompt that prints and exits. In the tour, confirming records the page's
// result and leaves it running — only q/ctrl+c ends the program there, and
// every visited page's result is printed at the end.
func Confirm() tea.Cmd {
	if tourMode.Load() {
		return nil
	}
	return tea.Quit
}

// Namespace prefixes a page's fields with its command name, so the tour can
// merge results from several pages without key collisions:
//
//	datepicker.date   2026-07-12
//	table.service     api
func Namespace(name string, fields []Field) []Field {
	out := make([]Field, len(fields))
	for i, f := range fields {
		out[i] = Field{Key: name + "." + f.Key, Value: f.Value}
	}
	return out
}
