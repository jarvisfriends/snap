package rendercheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// boxModel renders a bordered box sized to the current window — a
// well-behaved fixture that every conformance check must pass.
type boxModel struct {
	w, h int
}

func (m boxModel) Init() tea.Cmd { return nil }

func (m boxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.w, m.h = ws.Width, ws.Height
	}
	return m, nil
}

func (m boxModel) View() tea.View {
	if m.w < 2 || m.h < 2 {
		return tea.NewView("")
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(m.w - 2).
		Height(m.h - 2).
		Render("ok")
	return tea.NewView(box)
}

// TestWellBehavedModelPassesChecks runs the conformance suite over the
// fixture: a model that sizes itself to the window must sail through every
// checker (the checkers' own loops and size tables get exercised in-repo,
// not only from downstream consumers).
func TestWellBehavedModelPassesChecks(t *testing.T) {
	t.Parallel()

	CheckFitsViewport(t, boxModel{}, tea.KeyPressMsg{Code: tea.KeyRight})
	CheckBorderIntegrity(t, boxModel{}, "│")
	AssertBounds(t, boxModel{}, 80, 24)
	CheckStatusBarVisible(t, boxModel{}, nil) // no StatusProvider: must skip, not fail
}

func TestCheckBorderIntegrityStringCountsGlyphs(t *testing.T) {
	t.Parallel()

	clean := "┌──┐\n│ok│\n└──┘"
	CheckBorderIntegrityString(t, clean, "│")

	// The failure path reports through t.Errorf; drive it with a throwaway
	// recorder T so a wrapped-border frame is detected without failing us.
	rec := &testing.T{}
	CheckBorderIntegrityString(rec, "│bad│wrap│", "│")
	if !rec.Failed() {
		t.Fatal("CheckBorderIntegrityString accepted a line with 3 border glyphs")
	}
}

// TestGoldenRoundTrip covers both Golden paths: UPDATE_GOLDEN writes the
// file, the normal path compares against it.
func TestGoldenRoundTrip(t *testing.T) {
	content := "golden fixture\nline two"

	t.Setenv("UPDATE_GOLDEN", "1")
	Golden(t, "selfcheck_roundtrip", content)

	t.Setenv("UPDATE_GOLDEN", "")
	Golden(t, "selfcheck_roundtrip", content)

	// Mismatch must be flagged (recorder T again, so this test still passes).
	rec := &testing.T{}
	Golden(rec, "selfcheck_roundtrip", content+" drifted")
	if !rec.Failed() {
		t.Fatal("Golden accepted drifted output")
	}

	// The fixture file is a build artifact of this test; drop it so the
	// repo stays clean.
	if err := os.Remove(filepath.Join("testdata", "golden", "selfcheck_roundtrip.golden")); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}

func TestLongestLineAndColorHelpers(t *testing.T) {
	t.Parallel()

	if got := longestLine("ab\nabcd\nc"); got != "abcd" {
		t.Fatalf("longestLine = %q; want %q", got, "abcd")
	}

	red := "\x1b[31mred\x1b[0m plain \x1b[31magain\x1b[0m"
	set := ansiColorSet(red)
	if len(set) == 0 {
		t.Fatal("ansiColorSet found no colors in styled text")
	}
	if !sameSet(set, ansiColorSet(red)) {
		t.Fatal("sameSet(x, x) = false")
	}
	if sameSet(set, ansiColorSet("plain")) {
		t.Fatal("sameSet treated styled and plain frames as equal")
	}
	if strings.Contains(StripANSI(red), "\x1b") {
		t.Fatal("StripANSI left an escape sequence behind")
	}
}
