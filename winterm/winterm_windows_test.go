// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package winterm

import "testing"

// Detect only reads the per-user registry key, so it is safe to run on any
// Windows machine; whatever the machine's delegation is, the answer must be
// a known enum value with no error.
func TestDetectReturnsKnownDelegation(t *testing.T) {
	d, err := Detect()
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	switch d {
	case LetWindowsDecide, WindowsTerminal, LegacyConsole, Unknown:
	default:
		t.Errorf("Detect() = %v, not a known delegation", d)
	}
}

// Set(Unknown) must refuse before touching the registry — the full write
// path is deliberately untested here, because a unit test has no business
// changing the developer's real default-terminal delegation.
func TestSetRejectsUnknown(t *testing.T) {
	if err := Set(Unknown); err == nil {
		t.Fatal("Set(Unknown) succeeded, want an error")
	}
}
