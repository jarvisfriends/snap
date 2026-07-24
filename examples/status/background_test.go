package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestViewSetsThemeCanvasColors(t *testing.T) {
	t.Parallel()

	a := newDemo()
	_, _ = a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	v := a.View()
	if v.BackgroundColor == nil {
		t.Fatal("status demo view missing BackgroundColor")
	}
	if v.ForegroundColor == nil {
		t.Fatal("status demo view missing ForegroundColor")
	}
}
