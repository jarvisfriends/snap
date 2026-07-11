package forms

import (
	"strings"
	"testing"
	"time"
)

func TestParseRequired(t *testing.T) {
	t.Parallel()

	got, err := ParseRequired("  hello  ", "name")
	if err != nil || got != "hello" {
		t.Fatalf("ParseRequired trimmed = (%q, %v), want (\"hello\", nil)", got, err)
	}
	for _, raw := range []string{"", "   ", "\t\n"} {
		if _, err := ParseRequired(raw, "name"); err == nil {
			t.Errorf("ParseRequired(%q) = nil error, want required error", raw)
		} else if !strings.Contains(err.Error(), "name is required") {
			t.Errorf("ParseRequired(%q) error %q does not name the field", raw, err)
		}
	}
}

func TestParseDuration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want time.Duration
	}{
		{"5m", 5 * time.Minute},
		{" 1h ", time.Hour}, // surrounding whitespace is trimmed
		{"7h30m", 7*time.Hour + 30*time.Minute},
	}
	for _, c := range cases {
		got, err := ParseDuration(c.raw, "interval")
		if err != nil || got != c.want {
			t.Errorf("ParseDuration(%q) = (%v, %v), want (%v, nil)", c.raw, got, err, c.want)
		}
	}
	for _, raw := range []string{"", "5 minutes", "abc"} {
		if _, err := ParseDuration(raw, "interval"); err == nil {
			t.Errorf("ParseDuration(%q) = nil error, want error", raw)
		} else if !strings.Contains(err.Error(), "interval") {
			t.Errorf("ParseDuration(%q) error %q does not name the field", raw, err)
		}
	}
}

func TestParseISODate(t *testing.T) {
	t.Parallel()

	got, err := ParseISODate(" 2026-07-10 ", "start date")
	if err != nil {
		t.Fatalf("ParseISODate valid = error %v", err)
	}
	if got.Year() != 2026 || got.Month() != time.July || got.Day() != 10 {
		t.Fatalf("ParseISODate = %v, want 2026-07-10", got)
	}
	for _, raw := range []string{"", "07/10/2026", "2026-13-40", "yesterday"} {
		if _, err := ParseISODate(raw, "start date"); err == nil {
			t.Errorf("ParseISODate(%q) = nil error, want error", raw)
		} else if !strings.Contains(err.Error(), "start date") {
			t.Errorf("ParseISODate(%q) error %q does not name the field", raw, err)
		}
	}
}

func TestSplitAndClean(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw  string
		want []string
	}{
		{"a, b ,c", []string{"a", "b", "c"}},
		{" a ,, ,b", []string{"a", "b"}},
		{"", nil},
		{" , , ", nil},
		{"solo", []string{"solo"}},
	}
	for _, c := range cases {
		got := SplitAndClean(c.raw, ",")
		if len(got) != len(c.want) {
			t.Errorf("SplitAndClean(%q) = %v, want %v", c.raw, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("SplitAndClean(%q)[%d] = %q, want %q", c.raw, i, got[i], c.want[i])
			}
		}
	}
}
