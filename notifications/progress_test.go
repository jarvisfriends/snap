package notifications_test

import (
	"testing"

	"github.com/jarvisfriends/snap/notifications"
)

func pctPtr(v float64) *float64 { return &v }

// TestProgressAddAndUpdate: a notification created with Percent carries it,
// SetProgress updates it in place (clamped to 0–100), and the stored value is
// a copy the caller's pointer can't reach.
func TestProgressAddAndUpdate(t *testing.T) {
	m := notifications.NewManager()
	src := 40.0
	n, _ := m.AddWithOptions("copying", notifications.SeverityInfo, 0,
		notifications.AddOptions{Key: "copy", Percent: &src})
	if n.Percent == nil || *n.Percent != 40 {
		t.Fatalf("expected 40%% at creation, got %+v", n.Percent)
	}
	src = 99 // mutating the caller's value must not touch the stored copy
	if got := m.Active()[0].Percent; got == nil || *got != 40 {
		t.Fatalf("stored percent aliased the caller's pointer: %+v", got)
	}

	m.SetProgress(n.ID, 75)
	if got := m.Active()[0].Percent; got == nil || *got != 75 {
		t.Fatalf("SetProgress by ID: got %+v want 75", got)
	}
	m.SetProgressKey("copy", 150)
	if got := m.Active()[0].Percent; got == nil || *got != 100 {
		t.Fatalf("SetProgressKey should clamp to 100, got %+v", got)
	}
	m.SetProgress(n.ID, -5)
	if got := m.Active()[0].Percent; got == nil || *got != 0 {
		t.Fatalf("SetProgress should clamp to 0, got %+v", got)
	}
}

// TestProgressReshowsHiddenToast: updating progress on a toast-hidden
// notification re-shows the toast so ongoing progress stays visible.
func TestProgressReshowsHiddenToast(t *testing.T) {
	m := notifications.NewManager()
	n, _ := m.AddWithOptions("job", notifications.SeverityInfo, 0,
		notifications.AddOptions{Percent: pctPtr(10), RetainInHistory: true})
	m.Handle(notifications.ExpireMsg{ID: n.ID})
	if len(m.Visible()) != 0 {
		t.Fatal("expected toast hidden after expiry")
	}
	m.SetProgress(n.ID, 50)
	if len(m.Visible()) != 1 {
		t.Fatal("progress update should re-show the toast")
	}
}

// TestProgressMsgRouting: Handle routes ProgressMsg by ID, or by Key when the
// ID is zero, and AddMsg carries Percent through.
func TestProgressMsgRouting(t *testing.T) {
	m := notifications.NewManager()
	m.Handle(notifications.AddMsg{
		Key:      "dl",
		Content:  "downloading",
		Severity: notifications.SeverityInfo,
		Percent:  pctPtr(5),
	})
	got := m.Active()
	if len(got) != 1 || got[0].Percent == nil || *got[0].Percent != 5 {
		t.Fatalf("AddMsg should carry Percent, got %+v", got)
	}

	m.Handle(notifications.ProgressMsg{Key: "dl", Percent: 42})
	if got := m.Active()[0].Percent; got == nil || *got != 42 {
		t.Fatalf("ProgressMsg by key: got %+v want 42", got)
	}
	m.Handle(notifications.ProgressMsg{ID: m.Active()[0].ID, Percent: 88})
	if got := m.Active()[0].Percent; got == nil || *got != 88 {
		t.Fatalf("ProgressMsg by ID: got %+v want 88", got)
	}
}

// TestProgressPersistsAcrossSaveLoad: Percent round-trips through the JSON
// persistence file.
func TestProgressPersistsAcrossSaveLoad(t *testing.T) {
	dir := t.TempDir()
	m := notifications.NewManager()
	m.AddWithOptions("syncing", notifications.SeverityInfo, 0,
		notifications.AddOptions{Key: "sync", Percent: pctPtr(66), RetainInHistory: true})
	if err := m.Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}

	m2 := notifications.NewManager()
	if err := m2.Load(dir); err != nil {
		t.Fatalf("load: %v", err)
	}
	active := m2.Active()
	if len(active) != 1 || active[0].Percent == nil || *active[0].Percent != 66 {
		t.Fatalf("percent lost in persistence round-trip: %+v", active)
	}
}
