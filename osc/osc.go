// Package osc emits terminal OSC escape sequences for taskbar / tab progress
// (the ConEmu OSC 9;4 protocol, honored by Windows Terminal and others).
//
// Inside a running Bubble Tea program, set tea.View.ProgressBar instead —
// the v2 renderer speaks this exact protocol, diffs the value so sequences
// are only re-emitted on change, and resets the indicator on exit. This
// package covers the phases a program's View cannot: cobra/CLI commands,
// setup work before tea.Run, and cleanup after it returns. Ported from
// aSettings' page-level helpers and extended with the full protocol states.
package osc

import (
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
)

// Progress states defined by the ConEmu OSC 9;4 protocol.
const (
	stateClear         = 0 // remove the indicator
	stateNormal        = 1 // determinate progress (green)
	stateError         = 2 // error highlight (red), keeps the last percentage
	stateIndeterminate = 3 // marquee / unknown duration
	statePaused        = 4 // paused highlight (yellow), keeps the percentage
)

// Emitter writes progress sequences to a terminal. The zero value writes to
// os.Stderr (the conventional escape-sequence channel while Bubble Tea owns
// stdout); tests point W at a buffer.
type Emitter struct {
	W io.Writer
}

func (e Emitter) write(state, pct int) tea.Cmd {
	w := e.W
	if w == nil {
		w = os.Stderr
	}
	return func() tea.Msg {
		_, _ = fmt.Fprintf(w, "\x1b]9;4;%d;%d\x07", state, pct)
		return nil
	}
}

// SetProgress shows determinate progress (pct clamped to 0-100).
func (e Emitter) SetProgress(pct int) tea.Cmd {
	return e.write(stateNormal, min(max(pct, 0), 100))
}

// Indeterminate shows a marquee for work of unknown duration.
func (e Emitter) Indeterminate() tea.Cmd { return e.write(stateIndeterminate, 0) }

// Error switches the indicator to the error state at pct.
func (e Emitter) Error(pct int) tea.Cmd { return e.write(stateError, min(max(pct, 0), 100)) }

// Paused switches the indicator to the paused state at pct.
func (e Emitter) Paused(pct int) tea.Cmd { return e.write(statePaused, min(max(pct, 0), 100)) }

// Clear removes the indicator. Always send this when the work finishes —
// the terminal keeps showing the last state otherwise.
func (e Emitter) Clear() tea.Cmd { return e.write(stateClear, 0) }

// Package-level shorthands using the default Emitter (os.Stderr).

// SetProgress shows determinate progress on the terminal's taskbar/tab icon.
func SetProgress(pct int) tea.Cmd { return Emitter{}.SetProgress(pct) }

// Indeterminate shows a marquee progress indicator.
func Indeterminate() tea.Cmd { return Emitter{}.Indeterminate() }

// Error switches the indicator to the error state.
func Error(pct int) tea.Cmd { return Emitter{}.Error(pct) }

// Paused switches the indicator to the paused state.
func Paused(pct int) tea.Cmd { return Emitter{}.Paused(pct) }

// Clear removes the indicator.
func Clear() tea.Cmd { return Emitter{}.Clear() }
