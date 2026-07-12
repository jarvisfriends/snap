package rendercheck

import "testing"

// TestSnapMeetsOwnCodeStandards runs the AST code-standard checks over every
// package in this module (examples included) — the same gate tui-base runs
// on its own tree. It exists so a printf byte-pad, a hand-joined newline
// stack, a space-run alignment gap, or a len()-as-width can't slip back into
// snap after the 2026-07-11 sweep that removed them all.
func TestSnapMeetsOwnCodeStandards(t *testing.T) {
	t.Parallel()
	CheckCodeStandards(t, "github.com/jarvisfriends/snap/...")
}
