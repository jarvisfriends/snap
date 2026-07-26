// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package osc

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestPackageShorthands: the package-level helpers must hand back a command
// on the default Emitter. The commands are not executed here — they would
// write to os.Stderr — the wire bytes are pinned by TestSequences.
func TestPackageShorthands(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cmd  tea.Cmd
	}{
		{"SetProgress", SetProgress(42)},
		{"Indeterminate", Indeterminate()},
		{"Error", Error(80)},
		{"Paused", Paused(50)},
		{"Clear", Clear()},
	}
	for _, c := range cases {
		if c.cmd == nil {
			t.Errorf("%s returned a nil command", c.name)
		}
	}
}
