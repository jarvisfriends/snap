// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

//go:build !windows

package winterm

import (
	"errors"
	"testing"
)

// TestDetectAndSetUnsupportedOffWindows pins the non-Windows contract: both
// entry points fail with an error matching errors.ErrUnsupported so callers
// can feature-gate cleanly.
func TestDetectAndSetUnsupportedOffWindows(t *testing.T) {
	t.Parallel()

	d, err := Detect()
	if d != Unknown || !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("Detect() = %v, %v; want Unknown and ErrUnsupported", d, err)
	}
	if err := Set(WindowsTerminal); !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("Set() error = %v; want ErrUnsupported", err)
	}
}
