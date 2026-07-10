// Package rendercheck holds test helpers that catch rendered-string building
// mistakes in Bubble Tea v2 apps: layout math that guesses frame sizes,
// borders that lose their edges, lines that overflow the viewport, display-
// width errors around emoji/CJK, golden-file comparisons, and key-binding
// hygiene (CheckCodeStandards). Extracted from tui-base's testutil per its
// ROADMAP SP-14 so every snap consumer can gate on the same checks.
package rendercheck
