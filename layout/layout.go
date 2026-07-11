// Package layout provides lipgloss-frame arithmetic helpers: where content
// starts inside a bordered/padded style, how much room it has, and rendering
// content into a fixed outer box. They complement geom (pure cell rectangles,
// no lipgloss) by answering the style-dependent half of hit-testing and
// sizing, so components stop hand-summing GetBorderLeftSize+GetPaddingLeft.
// Ported from w's ui/shared/layout.go.
package layout

import "charm.land/lipgloss/v2"

// ContentOrigin returns the (x, y) offset of the content area inside a
// styled box: the left border plus left padding, and the top border plus top
// padding. Add it to a box's screen position to translate outer coordinates
// into content coordinates for mouse hit-testing.
func ContentOrigin(style lipgloss.Style) (x, y int) {
	return style.GetBorderLeftSize() + style.GetPaddingLeft(),
		style.GetBorderTopSize() + style.GetPaddingTop()
}

// InnerSize returns the content area of a box rendered at outerWidth x
// outerHeight with the style's border and padding, floored at 1x1 so
// downstream layout math never divides by or renders into zero.
func InnerSize(style lipgloss.Style, outerWidth, outerHeight int) (width, height int) {
	frameWidth, frameHeight := style.GetFrameSize()
	return max(outerWidth-frameWidth, 1), max(outerHeight-frameHeight, 1)
}

// RenderInBox renders content into a box of exactly outerWidth x outerHeight
// (frame included). A non-positive dimension is derived from the content size
// plus the style's frame, so callers can fix one axis and let the other fit.
func RenderInBox(style lipgloss.Style, outerWidth, outerHeight int, content string) string {
	if outerWidth <= 0 || outerHeight <= 0 {
		contentWidth, contentHeight := lipgloss.Size(content)
		if outerWidth <= 0 {
			outerWidth = contentWidth + style.GetHorizontalFrameSize()
		}
		if outerHeight <= 0 {
			outerHeight = contentHeight + style.GetVerticalFrameSize()
		}
	}
	return style.Width(outerWidth).Height(outerHeight).Render(content)
}
