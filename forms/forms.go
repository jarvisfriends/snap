// Package forms provides small input-parsing helpers for text fields:
// trim-and-validate parsers that return user-facing error messages naming
// the field, so every form page doesn't re-invent "required", duration, and
// date validation with slightly different wording. Ported from w's
// ui/shared/input_validation.go as the seed of snap's form tooling.
package forms

import (
	"fmt"
	"strings"
	"time"
)

// ParseRequired trims raw and rejects empty input. fieldName appears in the
// error message shown to the user (e.g. "project is required").
func ParseRequired(raw, fieldName string) (string, error) {
	val := strings.TrimSpace(raw)
	if val == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}
	return val, nil
}

// ParseDuration parses a Go duration string ("5m", "1h", "7h30m") with a
// friendly, example-carrying error naming the field.
func ParseDuration(raw, fieldName string) (time.Duration, error) {
	val := strings.TrimSpace(raw)
	dur, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q (examples: 5m, 1h, 7h30m)", fieldName, raw)
	}
	return dur, nil
}

// ParseISODate parses a YYYY-MM-DD date with a friendly error naming the
// field. The result is midnight UTC of that date (time.Parse semantics).
func ParseISODate(raw, fieldName string) (time.Time, error) {
	val := strings.TrimSpace(raw)
	date, err := time.Parse("2006-01-02", val)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s %q (expected YYYY-MM-DD)", fieldName, raw)
	}
	return date, nil
}

// SplitAndClean splits raw on sep, trims each piece, and drops empties —
// the usual treatment for comma-separated tag/list fields.
func SplitAndClean(raw, sep string) []string {
	parts := strings.Split(raw, sep)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
