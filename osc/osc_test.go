package osc

import (
	"bytes"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func emit(t *testing.T, cmd tea.Cmd) string {
	t.Helper()
	if cmd == nil {
		t.Fatal("nil command")
	}
	_ = cmd()
	return ""
}

// TestSequences pins the exact OSC 9;4 wire bytes per state — terminals
// match these literally, so any drift breaks the taskbar indicator silently.
func TestSequences(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(Emitter) tea.Cmd
		want string
	}{
		{"progress", func(e Emitter) tea.Cmd { return e.SetProgress(42) }, "\x1b]9;4;1;42\x07"},
		{"progress clamps high", func(e Emitter) tea.Cmd { return e.SetProgress(150) }, "\x1b]9;4;1;100\x07"},
		{"progress clamps low", func(e Emitter) tea.Cmd { return e.SetProgress(-1) }, "\x1b]9;4;1;0\x07"},
		{"indeterminate", Emitter.Indeterminate, "\x1b]9;4;3;0\x07"},
		{"error", func(e Emitter) tea.Cmd { return e.Error(80) }, "\x1b]9;4;2;80\x07"},
		{"paused", func(e Emitter) tea.Cmd { return e.Paused(50) }, "\x1b]9;4;4;50\x07"},
		{"clear", Emitter.Clear, "\x1b]9;4;0;0\x07"},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		emit(t, c.run(Emitter{W: &buf}))
		if buf.String() != c.want {
			t.Errorf("%s: wrote %q; want %q", c.name, buf.String(), c.want)
		}
	}
}
