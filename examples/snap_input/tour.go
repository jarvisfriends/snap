// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package main

import (
	"slices"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jarvisfriends/snap/examples/internal/exui"
	"github.com/jarvisfriends/snap/navigation"
)

// The tour hosts every example in one program: same page models, same shell,
// only the surrounding navigation is new. `snap_input` with no arguments
// tours all of them; `snap_input <name>` runs the identical code path with a
// single page (and, being a one-shot prompt, exits as soon as that page's
// value is confirmed — see exui.Confirm).
//
// Page models are recreated on entry rather than kept alive: components
// snapshot their styles at construction, so recreation is what makes theme
// cycling apply to the page you are looking at. Results are copied out of a
// page before it is dropped, so revisiting never loses an earlier choice.

// Tour chords. Page navigation is on alt+←/→ unconditionally, and on
// tab/shift+tab whenever the page is not capturing keystrokes — the forms
// page needs tab for its own fields, so tab alone could not be the contract.
const (
	keyNextAlt    = "alt+right"
	keyPrevAlt    = "alt+left"
	keyNextTab    = "tab"
	keyPrevTab    = "shift+tab"
	keyCycleTheme = "ctrl+t"
	keyToggleTabs = "ctrl+b"
)

// tour is the multi-page host model.
type tour struct {
	defs  []pageDef
	cur   int
	model exui.Page

	// results holds each page's confirmed fields, captured when the page is
	// left (or at exit) so a recreated model cannot lose them. order is the
	// visit order, which is the order results print in.
	results map[string][]exui.Field
	order   []string

	// cleaners collects every page that allocated teardown-worthy state.
	// Pages are recreated on entry, so one page kind can contribute several.
	cleaners []exui.Cleaner

	tabs     *navigation.Tabs
	showTabs bool
	w, h     int
}

// Cleanup tears down every page the tour built. Called once the program has
// ended, before the result is printed and the process exits.
func (t *tour) Cleanup() {
	for _, c := range t.cleaners {
		c.Cleanup()
	}
	t.cleaners = nil
}

// newTour builds the host over the given pages, starting on the first.
func newTour(defs []pageDef) *tour {
	exui.SetTourMode(len(defs) > 1)
	t := &tour{
		defs:    defs,
		results: make(map[string][]exui.Field, len(defs)),
		tabs:    navigation.NewTabs(),
	}
	pages := make([]navigation.Page, len(defs))
	for i, d := range defs {
		pages[i] = navigation.Page{ID: d.name, Title: d.name}
	}
	t.tabs.SetPages(pages)
	t.tabs.SetActiveIndex(0)
	t.enter(0)
	return t
}

// enter builds page i and records the visit. The previous page's result is
// captured first, since the model is about to be dropped.
func (t *tour) enter(i int) {
	t.capture()
	t.cur = i
	t.model = t.defs[i].new()
	if c, ok := t.model.(exui.Cleaner); ok {
		t.cleaners = append(t.cleaners, c)
	}
	t.tabs.SetActiveIndex(i)
	if name := t.defs[i].name; !slices.Contains(t.order, name) {
		t.order = append(t.order, name)
	}
}

// capture copies the live page's result into results. Called before dropping
// a model and once more at exit.
func (t *tour) capture() {
	if t.model == nil {
		return
	}
	if r := t.model.Result(); len(r) > 0 {
		t.results[t.defs[t.cur].name] = r
	}
}

// Fields returns every visited page's confirmed values in visit order,
// skipping pages the user never confirmed anything on.
//
// Keys are namespaced by page only in a real multi-page tour: a single-page
// run is a scriptable prompt whose output shape is a documented contract, so
// `snap_input datepicker --output values` must stay a bare date.
func (t *tour) Fields() []exui.Field {
	t.capture()
	var out []exui.Field
	for _, name := range t.order {
		r := t.results[name]
		if len(r) == 0 {
			continue
		}
		if len(t.defs) > 1 {
			r = exui.Namespace(name, r)
		}
		out = append(out, r...)
	}
	return out
}

// shell returns the current page's chrome, or nil if it exposes none.
func (t *tour) shell() *exui.Chrome {
	if s, ok := t.model.(exui.Shelled); ok {
		return s.Shell()
	}
	return nil
}

// tabsHeight is the height the tab strip occupies (0 while hidden).
func (t *tour) tabsHeight() int {
	if !t.showTabs || len(t.defs) < 2 {
		return 0
	}
	return lipgloss.Height(t.tabs.View().Content)
}

func (t *tour) Init() tea.Cmd { return t.model.Init() }

// Result satisfies exui.Page so a tour can itself be hosted (single-page mode
// runs one page through this same type).
func (t *tour) Result() []exui.Field { return t.Fields() }

func (t *tour) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.w, t.h = msg.Width, msg.Height
		return t, t.resize()

	case tea.MouseMsg:
		// Pointer input reaches the page through the root view's OnMouse;
		// the runtime delivers it here too, so ignore it (see README's input
		// contract).
		return t, nil

	case tea.KeyPressMsg:
		if cmd, done := t.hotkey(msg); done {
			return t, cmd
		}
	}
	m, cmd := t.model.Update(msg)
	if p, ok := m.(exui.Page); ok {
		t.model = p
	}
	return t, cmd
}

// hotkey handles the tour's own chords. It defers entirely to the page while
// a shell overlay is open, and leaves tab to the page while it is capturing
// keystrokes.
func (t *tour) hotkey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if len(t.defs) < 2 {
		return nil, false
	}
	sh := t.shell()
	if sh.OverlayOpen() {
		return nil, false
	}
	switch s := msg.String(); {
	case s == keyNextAlt, s == keyNextTab && !sh.Capturing():
		return t.goTo(t.cur + 1), true
	case s == keyPrevAlt, s == keyPrevTab && !sh.Capturing():
		return t.goTo(t.cur - 1), true
	case s == keyCycleTheme:
		return t.cycleTheme(), true
	case s == keyToggleTabs:
		t.showTabs = !t.showTabs
		return t.resize(), true
	}
	return nil, false
}

// goTo moves to another page, wrapping at both ends.
func (t *tour) goTo(i int) tea.Cmd {
	n := len(t.defs)
	t.enter(((i % n) + n) % n)
	return tea.Batch(t.model.Init(), t.resize())
}

// cycleTheme advances the shared palette and rebuilds the current page, which
// rebuilds its chrome too — components snapshot styles when constructed, so
// nothing already built would otherwise pick up the new colors.
func (t *tour) cycleTheme() tea.Cmd {
	exui.CycleTheme()
	t.enter(t.cur)
	return tea.Batch(t.model.Init(), t.resize())
}

// resize hands the page the space left under the tab strip. The page keeps
// drawing its own status bar on the last line of that space, which is the
// terminal's last line whenever the strip is hidden.
func (t *tour) resize() tea.Cmd {
	if t.w == 0 && t.h == 0 {
		return nil
	}
	m, cmd := t.model.Update(tea.WindowSizeMsg{
		Width:  t.w,
		Height: max(t.h-t.tabsHeight(), 1),
	})
	if p, ok := m.(exui.Page); ok {
		t.model = p
	}
	return cmd
}

func (t *tour) View() tea.View {
	v := t.model.View()
	th := t.tabsHeight()
	if th == 0 {
		return v
	}
	// Stack the strip above the page and shift pointer coordinates back into
	// page space — Bubble Tea hands the root view absolute coordinates and
	// never translates for children.
	strip := t.tabs.View().Content
	inner := v.OnMouse
	v.SetContent(lipgloss.JoinVertical(lipgloss.Left, strip, v.Content))
	v.OnMouse = func(mm tea.MouseMsg) tea.Cmd {
		if cmd, handled := t.stripMouse(mm, th); handled {
			return cmd
		}
		if inner == nil {
			return nil
		}
		return inner(shiftMouse(mm, th))
	}
	return v
}

// stripMouse selects a page when the tab strip itself is clicked.
func (t *tour) stripMouse(mm tea.MouseMsg, th int) (tea.Cmd, bool) {
	click, ok := mm.(tea.MouseClickMsg)
	if !ok || click.Mouse().Y >= th {
		return nil, false
	}
	m, _ := t.tabs.Update(mm)
	if tabs, ok := m.(*navigation.Tabs); ok {
		t.tabs = tabs
	}
	if i := t.tabs.GetActiveIndex(); i != t.cur && i >= 0 && i < len(t.defs) {
		return t.goTo(i), true
	}
	return nil, true
}

// shiftMouse moves an event up by dy so a child sees page-local coordinates.
func shiftMouse(mm tea.MouseMsg, dy int) tea.MouseMsg {
	switch e := mm.(type) {
	case tea.MouseClickMsg:
		e.Y -= dy
		return e
	case tea.MouseReleaseMsg:
		e.Y -= dy
		return e
	case tea.MouseWheelMsg:
		e.Y -= dy
		return e
	case tea.MouseMotionMsg:
		e.Y -= dy
		return e
	}
	return mm
}
