// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package navigation

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/jarvisfriends/snap/styles"
)

// TestDigitIndex pins the direct-select key mapping: single digits 1-9 map
// to zero-based indices, everything else misses.
func TestDigitIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		want   int
		wantOK bool
	}{
		{"1", 0, true},
		{"9", 8, true},
		{"0", 0, false},
		{"a", 0, false},
		{"12", 0, false},
		{"", 0, false},
	}
	for _, test := range tests {
		got, ok := digitIndex(test.input)
		if got != test.want || ok != test.wantOK {
			t.Errorf("digitIndex(%q) = %d, %v; want %d, %v", test.input, got, ok, test.want, test.wantOK)
		}
	}
}

// TestMinimalTopNav_NonDigitKeyIgnored: unbound keys fall through the digit
// path without selecting anything.
func TestMinimalTopNav_NonDigitKeyIgnored(t *testing.T) {
	t.Parallel()

	m := NewMinimalTopNav()
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if cmd != nil || m.GetActiveIndex() != 0 {
		t.Fatal("a non-digit key must be ignored")
	}

	m.SetPages(nil)
	if _, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyRight}); cmd != nil {
		t.Fatal("keys with no pages must be ignored")
	}
}

// TestMinimalTopNav_PillShapeOverride: a configured PillShape reaches the
// pill styles; unset falls back to the diagonal default.
func TestMinimalTopNav_PillShapeOverride(t *testing.T) {
	t.Parallel()

	m := NewMinimalTopNav()
	c := styles.Active()
	if got := m.pillStyles(c).Shape; got != styles.PillDiagonal {
		t.Fatalf("default pill shape = %q; want %q", got, styles.PillDiagonal)
	}
	m.PillShape = styles.PillRound
	if got := m.pillStyles(c).Shape; got != styles.PillRound {
		t.Fatalf("pill shape = %q; want the configured %q", got, styles.PillRound)
	}
	if m.View().Content == "" {
		t.Fatal("top nav must render with a configured pill shape")
	}
}
