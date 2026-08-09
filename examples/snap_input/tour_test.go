// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
)

// stubPage is a minimal exui.Page: it records what it was told to confirm and
// how many times it was built, so tests can watch recreate-on-entry.
type stubPage struct {
	key, value string
	confirmed  bool
	cleanups   *int
}

func (p *stubPage) Init() tea.Cmd { return nil }

func (p *stubPage) Update(tea.Msg) (tea.Model, tea.Cmd) { return p, nil }

func (p *stubPage) View() tea.View { return tea.NewView("stub") }

func (p *stubPage) Result() []exui.Field {
	if !p.confirmed {
		return nil
	}
	return []exui.Field{exui.F(p.key, p.value)}
}

func (p *stubPage) Cleanup() { *p.cleanups++ }

// stubDef builds a pageDef whose model confirms immediately when confirm is
// true, counting constructions and cleanups through the returned pointers.
func stubDef(name, key, value string, confirm bool) (def pageDef, builds, cleanups *int) {
	builds, cleanups = new(int), new(int)
	return pageDef{
		name: name,
		desc: name,
		new: func() exui.Page {
			*builds++
			return &stubPage{key: key, value: value, confirmed: confirm, cleanups: cleanups}
		},
	}, builds, cleanups
}

// restoreTourMode keeps the process-global confirm semantics from leaking
// between tests (newTour sets it from the page count).
func restoreTourMode(t *testing.T) {
	t.Helper()
	prev := exui.InTour()
	t.Cleanup(func() { exui.SetTourMode(prev) })
}

// TestSinglePageResultIsNotNamespaced pins the scripting contract: one page
// prints a bare key, so `date=$(snap_input datepicker --output values)` keeps
// working exactly as documented.
func TestSinglePageResultIsNotNamespaced(t *testing.T) {
	restoreTourMode(t)

	def, _, _ := stubDef("datepicker", "date", "2026-07-12", true)
	got := newTour([]pageDef{def}).Fields()

	want := []exui.Field{{Key: "date", Value: "2026-07-12"}}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("single-page fields = %v; want %v", got, want)
	}
	if exui.InTour() {
		t.Fatal("a single page must not enable tour confirm semantics")
	}
}

// TestTourNamespacesInVisitOrder covers the multi-page contract: keys carry
// their page name and appear in the order the pages were visited, not the
// order they were declared.
func TestTourNamespacesInVisitOrder(t *testing.T) {
	restoreTourMode(t)

	a, _, _ := stubDef("datepicker", "date", "2026-07-12", true)
	b, _, _ := stubDef("table", "service", "api", true)
	c, _, _ := stubDef("charts", "unused", "", false) // never confirms

	tr := newTour([]pageDef{a, b, c})
	if !exui.InTour() {
		t.Fatal("multiple pages must enable tour confirm semantics")
	}
	tr.goTo(2) // charts, confirms nothing
	tr.goTo(1) // table

	got := tr.Fields()
	want := []exui.Field{
		{Key: "datepicker.date", Value: "2026-07-12"},
		{Key: "table.service", Value: "api"},
	}
	if len(got) != len(want) {
		t.Fatalf("fields = %v; want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("fields[%d] = %v; want %v", i, got[i], want[i])
		}
	}
}

// TestResultSurvivesPageRecreation is the regression test for losing a choice
// by walking away from a page: models are rebuilt on entry, so the result has
// to be copied out before the old model is dropped.
func TestResultSurvivesPageRecreation(t *testing.T) {
	restoreTourMode(t)

	a, aBuilds, _ := stubDef("datepicker", "date", "2026-07-12", true)
	b, _, _ := stubDef("charts", "unused", "", false)

	tr := newTour([]pageDef{a, b})
	tr.goTo(1) // leave the confirmed page
	tr.goTo(0) // and come back: a fresh, unconfirmed model

	if *aBuilds != 2 {
		t.Fatalf("page rebuilt %d times; want 2 (once per entry)", *aBuilds)
	}
	got := tr.Fields()
	if len(got) != 1 || got[0].Value != "2026-07-12" {
		t.Fatalf("fields = %v; want the earlier confirmation to survive", got)
	}
}

// TestGoToWraps: page navigation is a ring in both directions.
func TestGoToWraps(t *testing.T) {
	restoreTourMode(t)

	a, _, _ := stubDef("a", "k", "v", false)
	b, _, _ := stubDef("b", "k", "v", false)
	tr := newTour([]pageDef{a, b})

	tr.goTo(tr.cur - 1)
	if tr.cur != 1 {
		t.Fatalf("stepping back from page 0 landed on %d; want 1", tr.cur)
	}
	tr.goTo(tr.cur + 1)
	if tr.cur != 0 {
		t.Fatalf("stepping forward from the last page landed on %d; want 0", tr.cur)
	}
}

// TestCleanupRunsForEveryBuiltPage: a page kind that allocates (the pickers
// temp tree) is rebuilt on each entry, so every instance must be torn down.
func TestCleanupRunsForEveryBuiltPage(t *testing.T) {
	restoreTourMode(t)

	a, _, aCleanups := stubDef("a", "k", "v", false)
	b, _, _ := stubDef("b", "k", "v", false)

	tr := newTour([]pageDef{a, b})
	tr.goTo(1)
	tr.goTo(0) // second instance of page a
	tr.Cleanup()

	if *aCleanups != 2 {
		t.Fatalf("cleaned up %d instances of page a; want 2", *aCleanups)
	}
}

// TestSinglePageIgnoresTourChords: with one page there is nowhere to go, so
// tab must reach the page instead of being swallowed by the host.
func TestSinglePageIgnoresTourChords(t *testing.T) {
	restoreTourMode(t)

	def, _, _ := stubDef("only", "k", "v", false)
	tr := newTour([]pageDef{def})

	if _, done := tr.hotkey(tea.KeyPressMsg{Code: tea.KeyTab}); done {
		t.Fatal("tab was consumed by the host in single-page mode")
	}
}

// TestEveryRealPageSizesAndRenders walks the actual tour over every declared
// page, with the tab strip shown, and renders each one. It is the cheap
// stand-in for launching the TUI: a page that panics on size or paints
// nothing would otherwise only show up by hand.
func TestEveryRealPageSizesAndRenders(t *testing.T) {
	restoreTourMode(t)

	tr := newTour(pages)
	t.Cleanup(tr.Cleanup)
	tr.showTabs = true
	tr.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	for i, p := range pages {
		tr.goTo(i)
		v := tr.View()
		if v.Content == "" {
			t.Fatalf("page %q rendered nothing", p.name)
		}
		// The strip is stacked above the page, so the frame must be taller
		// than the page alone and still fit the window.
		if h := lipgloss.Height(v.Content); h > 30 {
			t.Fatalf("page %q rendered %d lines into a 30-line window", p.name, h)
		}
		if v.OnMouse == nil {
			t.Fatalf("page %q lost its mouse handler through the tour frame", p.name)
		}
	}
}

// TestTabStripShiftsPointerIntoPageSpace: Bubble Tea hands the root view
// absolute coordinates, so the tour must translate them past its own strip
// before the page sees them, or every click lands rows too low.
func TestTabStripShiftsPointerIntoPageSpace(t *testing.T) {
	t.Parallel()

	got := shiftMouse(tea.MouseClickMsg{X: 4, Y: 9, Button: tea.MouseLeft}, 3)
	click, ok := got.(tea.MouseClickMsg)
	if !ok {
		t.Fatalf("shiftMouse changed the event type to %T", got)
	}
	if click.Y != 6 || click.X != 4 {
		t.Fatalf("shifted click to (%d,%d); want (4,6)", click.X, click.Y)
	}
}

// TestPagesAreUniqueAndComplete guards the command table itself: duplicate or
// empty names would make `snap_input <name>` ambiguous or unreachable.
func TestPagesAreUniqueAndComplete(t *testing.T) {
	t.Parallel()

	seen := map[string]bool{}
	for _, p := range pages {
		switch {
		case p.name == "":
			t.Fatal("a page has an empty name")
		case p.desc == "":
			t.Fatalf("page %q has no description for `help`", p.name)
		case p.new == nil:
			t.Fatalf("page %q has no constructor", p.name)
		case seen[p.name]:
			t.Fatalf("page %q is declared twice", p.name)
		}
		seen[p.name] = true
		if _, ok := find(p.name); !ok {
			t.Fatalf("find(%q) did not resolve a declared page", p.name)
		}
	}
	if _, ok := find("nope"); ok {
		t.Fatal("find resolved a name that is not a page")
	}
}
