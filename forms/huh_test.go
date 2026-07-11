package forms

import (
	"testing"

	"charm.land/huh/v2"
)

// TestHuhValidateAdaptsParsers proves the parsers plug into huh unchanged:
// the adapter satisfies huh's Validate signature (compile-level guarantee via
// the NewInput call) and surfaces the parser's field-naming errors.
func TestHuhValidateAdaptsParsers(t *testing.T) {
	t.Parallel()

	// Compile-level: a snap/forms parser is a valid huh field validator.
	_ = huh.NewInput().Key("due").Validate(HuhValidate(ParseISODate, "due date"))

	cases := []struct {
		name     string
		validate func(string) error
		good     string
		bad      string
	}{
		{"required", HuhValidate(ParseRequired, "task"), "ship it", "   "},
		{"duration", HuhValidate(ParseDuration, "duration"), "7h30m", "soon"},
		{"date", HuhValidate(ParseISODate, "due date"), "2026-07-14", "next week"},
	}
	for _, tc := range cases {
		if err := tc.validate(tc.good); err != nil {
			t.Errorf("%s: valid input %q rejected: %v", tc.name, tc.good, err)
		}
		if err := tc.validate(tc.bad); err == nil {
			t.Errorf("%s: invalid input %q accepted", tc.name, tc.bad)
		}
	}
}
