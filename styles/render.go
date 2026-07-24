package styles

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// ColorHex formats a color.Color as a "#rrggbb" string. A nil color yields
// "#000000". Shared by the router for OSC/diagnostic output.
func ColorHex(c color.Color) string {
	if c == nil {
		return "#000000"
	}
	const digits = "0123456789abcdef"
	r, g, b, _ := c.RGBA()
	buf := [7]byte{'#'}
	for i, v := range [3]uint32{r >> 8, g >> 8, b >> 8} {
		buf[1+i*2] = digits[v>>4]
		buf[2+i*2] = digits[v&0xf]
	}
	return string(buf[:])
}

// ReapplyBg re-applies the background color after every ANSI reset in s. Child
// components emit `\x1b[m`/`\x1b[0m` resets mid-line; over SSH (where the OSC
// background is stripped) those resets expose the terminal default in unstyled
// gaps. Inserting the background SGR after each reset keeps the fill consistent.
// Shared by the router's main layout and the status bar.
func ReapplyBg(s string, bg color.Color) string {
	bgCode := firstEscapeFromStyle(lipgloss.NewStyle().Background(bg).Render("X"))
	if bgCode == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\x1b[0m", "\x1b[0m"+bgCode)
	s = strings.ReplaceAll(s, "\x1b[m", "\x1b[m"+bgCode)
	return s
}

// firstEscapeFromStyle extracts the first ANSI escape sequence from a lipgloss
// render result (used to recover the background SGR code).
func firstEscapeFromStyle(s string) string {
	i := strings.Index(s, "\x1b[")
	if i < 0 {
		return ""
	}
	j := strings.Index(s[i:], "m")
	if j < 0 {
		return ""
	}
	return s[i : i+j+1]
}
