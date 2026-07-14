package styles

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestBorderBoxStylesDrawABorder locks in that BoarderActive / BoarderInactive
// are complete bordered-box styles — they carry a real border SHAPE, not just a
// border color. A style with only BorderForeground set draws nothing, which
// silently breaks any consumer that sizes content or maps mouse clicks against
// an assumed border (the frame getters report 0). This test guards that
// invariant so the mouse-hit-testing bug class cannot return.
func TestBorderBoxStylesDrawABorder(t *testing.T) {
	app := Active()
	cases := []struct {
		name  string
		style lipgloss.Style
	}{
		{"BoarderActive", app.Styles.BoarderActive},
		{"BoarderInactive", app.Styles.BoarderInactive},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The frame getters must report a real border (content sizing and
			// hit-test offsets are derived from these).
			if got := tc.style.GetHorizontalFrameSize(); got < 2 {
				t.Errorf("%s.GetHorizontalFrameSize() = %d, want >= 2 (a border adds 1 cell per side)", tc.name, got)
			}
			if got := tc.style.GetVerticalFrameSize(); got < 2 {
				t.Errorf("%s.GetVerticalFrameSize() = %d, want >= 2", tc.name, got)
			}
			if got := tc.style.GetBorderTopSize(); got != 1 {
				t.Errorf("%s.GetBorderTopSize() = %d, want 1", tc.name, got)
			}

			// And the rendered box must actually contain border glyphs.
			out := tc.style.Width(10).Height(3).Render("x")
			if !strings.ContainsAny(out, "╭╮╰╯│─┌┐└┘") {
				t.Errorf("%s rendered no border glyphs:\n%q", tc.name, out)
			}
		})
	}
}
