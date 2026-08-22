// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package charts

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestHBarGradient_ZeroWidth(t *testing.T) {
	if got := HBarGradient(50, 0, lipgloss.Color("1"), lipgloss.Color("2")); got != "" {
		t.Errorf("HBarGradient(_, 0, …) = %q, want empty", got)
	}
}

func TestHBarGradient_NilColorsFallBackToPlainBar(t *testing.T) {
	want := HBar(50, 10)
	if got := HBarGradient(50, 10, nil, nil); got != want {
		t.Errorf("nil colors: got %q, want the plain HBar %q", got, want)
	}
	if got := HBarGradient(50, 10, lipgloss.Color("1"), nil); got != want {
		t.Errorf("one nil color: got %q, want the plain HBar %q", got, want)
	}
}

func TestHBarGradient_HalfFull(t *testing.T) {
	got := ansi.Strip(HBarGradient(50, 10, lipgloss.Color("1"), lipgloss.Color("2")))
	runes := []rune(got)
	if len(runes) != 10 {
		t.Fatalf("stripped width = %d, want 10 (%q)", len(runes), got)
	}
	filled := strings.Count(got, "█")
	if filled < 4 || filled > 6 {
		t.Errorf("filled = %d, want ~5 (%q)", filled, got)
	}
	if !strings.Contains(got, "░") {
		t.Errorf("half bar has no empty cells: %q", got)
	}
}

func TestHBarGradient_ClampsOutOfRange(t *testing.T) {
	low := ansi.Strip(HBarGradient(-25, 8, lipgloss.Color("1"), lipgloss.Color("2")))
	if strings.Contains(low, "█") {
		t.Errorf("negative pct should render empty, got %q", low)
	}
	high := ansi.Strip(HBarGradient(250, 8, lipgloss.Color("1"), lipgloss.Color("2")))
	if strings.Contains(high, "░") {
		t.Errorf("pct above 100 should render full, got %q", high)
	}
}
