package status

import (
	"strings"
	"testing"

	"github.com/jarvisfriends/snap/notifications"
)

// TestHistoryOverlayRendersProgressBar: a progress notification's row carries
// an inline HBar and the percent label; plain notifications don't.
func TestHistoryOverlayRendersProgressBar(t *testing.T) {
	t.Parallel()

	overlay := NewUserNotificationOverlay()
	nm := notifications.NewManager()
	overlay.SetNotifManager(nm)
	pct := 50.0
	nm.AddWithOptions("copying files", notifications.SeverityInfo, 0,
		notifications.AddOptions{Key: "copy", Percent: &pct})
	nm.Add("plain note", notifications.SeverityInfo, 0)
	overlay.showHistory = true

	stripped := stripANSI(overlay.RenderHistoryOverlay(100, 20))

	if !strings.Contains(stripped, "█████░░░░░  50%") {
		t.Errorf("progress row missing half-filled bar; got:\n%s", stripped)
	}
	if strings.Count(stripped, "░") != 5 {
		t.Errorf("plain row should carry no bar; got:\n%s", stripped)
	}

	nm.SetProgressKey("copy", 100)
	stripped = stripANSI(overlay.RenderHistoryOverlay(100, 20))
	if !strings.Contains(stripped, "██████████ 100%") {
		t.Errorf("full bar missing after update; got:\n%s", stripped)
	}
}
