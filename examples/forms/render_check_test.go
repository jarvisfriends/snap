package main

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
)

// drive feeds msg to the model and then runs the commands it queues the way
// the runtime would (unpacking batches), so huh's field-focus transitions
// land. Cursor-blink messages are dropped instead of fed back — following
// them re-queues a fresh blink tick forever — and the total command budget
// is capped as a backstop.
func drive(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()
	m, cmd := m.Update(msg)
	queue := []tea.Cmd{cmd}
	for i := 0; i < len(queue) && i < 32; i++ {
		if queue[i] == nil {
			continue
		}
		switch out := queue[i]().(type) {
		case nil, tea.QuitMsg:
		case tea.BatchMsg:
			queue = append(queue, out...)
		default:
			if strings.Contains(fmt.Sprintf("%T", out), "Blink") {
				continue
			}
			var next tea.Cmd
			m, next = m.Update(out)
			queue = append(queue, next)
		}
	}
	return m
}

// typeText feeds plain rune presses without chasing commands: typing only
// queues cursor-blink ticks, and executing each one costs real tick time.
func typeText(t *testing.T, m tea.Model, s string) tea.Model {
	t.Helper()
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

// TestFormsDemoValidatesWithSnapParsers drives the real huh form: a bad
// duration surfaces the snap/forms field-naming error inline, fixing it
// clears the error, and completing every field finishes the form.
func TestFormsDemoValidatesWithSnapParsers(t *testing.T) {
	a := newDemo()
	var m tea.Model = a
	if cmd := a.Init(); cmd != nil {
		if out := cmd(); out != nil {
			m = drive(t, m, out)
		}
	}
	m = drive(t, m, tea.WindowSizeMsg{Width: 80, Height: 30})

	m = typeText(t, m, "ship it")
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) // -> duration

	// An unparsable duration must block Enter and show the parser's error.
	m = typeText(t, m, "soon")
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	frame := ansi.Strip(m.View().Content)
	if !strings.Contains(frame, "duration") {
		t.Fatalf("frame missing the duration field error:\n%s", frame)
	}
	app, ok := m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	if app.form.State == huh.StateCompleted {
		t.Fatal("form completed despite invalid duration")
	}

	// Fix it and finish the remaining fields.
	for range 4 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	m = typeText(t, m, "7h30m")
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) // -> due
	m = typeText(t, m, "2026-07-14")
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) // -> tags
	m = typeText(t, m, "go,  tui ,, release")
	m = drive(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}) // submit

	app, ok = m.(*demoApp)
	if !ok {
		t.Fatalf("Update returned %T; want *demoApp", m)
	}
	if app.form.State != huh.StateCompleted {
		t.Fatalf("form did not complete; state=%v errors=%v", app.form.State, app.form.Errors())
	}
	if got := app.form.GetString("task"); got != "ship it" {
		t.Errorf("task = %q", got)
	}
	if got := app.form.GetString("duration"); got != "7h30m" {
		t.Errorf("duration = %q", got)
	}
}
