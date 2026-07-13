package forms

// This package extends charm.land/huh/v2 — it does not replace it. The
// parsers in forms.go are plain (raw, fieldName) → (value, error) functions,
// and HuhValidate adapts any of them to the func(string) error signature huh
// fields take, so they drop straight into a form next to every other huh
// component:
//
//	huh.NewInput().
//	    Key("due").
//	    Title("Due date").
//	    Validate(forms.HuhValidate(forms.ParseISODate, "due date"))
//
// The same parser then turns the submitted string into its typed value
// (time.Time, time.Duration, …), so validation and parsing can never drift
// apart.

// HuhValidate adapts a snap/forms parser into a huh field validator: the
// parser's field-naming error is what huh shows inline under the field.
func HuhValidate[T any](parse func(raw, fieldName string) (T, error), fieldName string) func(string) error {
	return func(raw string) error {
		_, err := parse(raw, fieldName)
		return err
	}
}
