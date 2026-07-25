package notifications

import (
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func TestSeverity_KnownAndUnknownFallbacks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sev   Severity
		str   string
		badge string
		color string
		ttl   time.Duration
	}{
		{SeverityInfo, "Info", "INFO", "#4FC3F7", 5 * time.Second},
		{SeverityWarning, "Warning", "WARN", "#F9C513", 10 * time.Second},
		{SeverityError, "Error", "ERR ", "#FF5757", 15 * time.Second},
		// An unknown severity falls back to the Info values.
		{Severity(99), "Info", "INFO", "#4FC3F7", 5 * time.Second},
	}
	for _, tc := range cases {
		if got := tc.sev.String(); got != tc.str {
			t.Errorf("Severity(%d).String() = %q; want %q", tc.sev, got, tc.str)
		}
		if got := tc.sev.Badge(); got != tc.badge {
			t.Errorf("Severity(%d).Badge() = %q; want %q", tc.sev, got, tc.badge)
		}
		if got := ColorForSeverity(tc.sev); got != tc.color {
			t.Errorf("ColorForSeverity(%d) = %q; want %q", tc.sev, got, tc.color)
		}
		if got := tc.sev.DefaultTTL(); got != tc.ttl {
			t.Errorf("Severity(%d).DefaultTTL() = %v; want %v", tc.sev, got, tc.ttl)
		}
	}
}

func TestManager_ActionRegistration(t *testing.T) {
	t.Parallel()

	m := NewManager()
	if _, ok := m.Action("o"); ok {
		t.Error("no handler should be registered before OnAction")
	}

	called := false
	m.OnAction("o", func(Notification) tea.Cmd {
		called = true
		return nil
	})
	fn, ok := m.Action("o")
	if !ok || fn == nil {
		t.Fatal(`Action("o") should return the registered handler`)
	}
	fn(Notification{}) // invoking the stored handler exercises it
	if !called {
		t.Error("registered handler was not invoked")
	}

	// Passing nil removes the handler.
	m.OnAction("o", nil)
	if _, ok := m.Action("o"); ok {
		t.Error("handler should be removed after OnAction(key, nil)")
	}
}

func TestManager_Pending(t *testing.T) {
	t.Parallel()

	m := NewManager()
	m.Add("normal", SeverityInfo, 5*time.Second) // not pending
	m.AddWithOptions("waiting", SeverityWarning, 0, AddOptions{Pending: true, Key: "job"})

	if got := m.PendingCount(); got != 1 {
		t.Errorf("PendingCount() = %d; want 1", got)
	}
	p := m.Pending()
	if len(p) != 1 {
		t.Fatalf("Pending() returned %d items; want 1", len(p))
	}
	if p[0].Content != "waiting" || !p[0].Pending {
		t.Errorf("Pending()[0] = %+v; want the pending \"waiting\" item", p[0])
	}
}

func TestManager_SaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	// A nested path confirms Save creates the directory tree.
	dir := filepath.Join(t.TempDir(), "nested", "state")

	src := NewManager()
	src.Add("first", SeverityInfo, 5*time.Second)
	src.Add("second", SeverityError, 15*time.Second)
	if err := src.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// A fresh manager restores the persisted notifications.
	dst := NewManager()
	if err := dst.Load(dir); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := len(dst.All()); got != 2 {
		t.Fatalf("loaded %d notifications; want 2", got)
	}

	// Loading from a directory with no state file is a silent no-op.
	empty := NewManager()
	if err := empty.Load(t.TempDir()); err != nil {
		t.Errorf("Load(missing file) = %v; want nil", err)
	}
	if got := len(empty.All()); got != 0 {
		t.Errorf("Load(missing file) populated %d items; want 0", got)
	}
}
