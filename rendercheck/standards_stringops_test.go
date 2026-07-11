package rendercheck

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestFormatHasBytePadding pins which printf verbs count as byte padding on
// display text (width-flagged string verbs) versus content-independent
// numeric formatting, which stays allowed.
func TestFormatHasBytePadding(t *testing.T) {
	t.Parallel()

	flagged := []string{
		`"  %-9s"`,     // the pills-example bug this check was written for
		`"%10s"`,       // right-pad string
		`"%-3q"`,       // quoted string with width
		`"%8v"`,        // widthed generic verb
		`"%08.4v"`,     // flags + precision
		`"x %s y %5s"`, // one safe verb, one padded
	}
	for _, f := range flagged {
		if !formatHasBytePadding(f) {
			t.Errorf("formatHasBytePadding(%s) = false, want true", f)
		}
	}

	allowed := []string{
		`"%s %s\n"`,       // plain string verbs, no width
		`"%02d"`,          // zero-padded number: digits are single-cell
		`"%3.0f%%"`,       // widthed float
		`"#%02x%02x%02x"`, // hex color
		`"%d:%s"`,
		`"state[%d] %T"`,
	}
	for _, f := range allowed {
		if formatHasBytePadding(f) {
			t.Errorf("formatHasBytePadding(%s) = true, want false", f)
		}
	}
}

// concatAncestorsFor parses src as an expression and returns the ancestor
// chain (outermost first) for the strings.Repeat call inside it, mirroring
// what CheckCodeStandards' Inspect walk hands to checkRepeatSpaceConcat.
func concatAncestorsFor(t *testing.T, src string) []ast.Node {
	t.Helper()
	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	var (
		ancestors []ast.Node
		found     []ast.Node
		matched   bool
	)
	ast.Inspect(expr, func(n ast.Node) bool {
		if n == nil {
			ancestors = ancestors[:len(ancestors)-1]
			return true
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Repeat" {
				found = append([]ast.Node(nil), ancestors...)
				matched = true
			}
		}
		ancestors = append(ancestors, n)
		return true
	})
	if !matched {
		t.Fatalf("no strings.Repeat call in %q", src)
	}
	return found
}

// TestConcatenatedWithContent pins the alignment-gap classifier: a space run
// concatenated between content is flagged, a standalone fill is not.
func TestConcatenatedWithContent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		src  string
		want bool
	}{
		{`left + strings.Repeat(" ", gap) + right`, true},
		{`prefix + (strings.Repeat(" ", n))`, true},
		{`strings.Repeat(" ", width)`, false},         // standalone blank fill
		{`render(strings.Repeat(" ", width))`, false}, // argument, not concat
	}
	for _, c := range cases {
		if got := concatenatedWithContent(concatAncestorsFor(t, c.src)); got != c.want {
			t.Errorf("concatenatedWithContent(%q) = %v, want %v", c.src, got, c.want)
		}
	}
}

// TestJoinNewlineLiteralDetection pins the separator matching used by
// checkJoinNewline: only newline-bearing separators are flagged.
func TestJoinNewlineLiteralDetection(t *testing.T) {
	t.Parallel()

	// The check reads the literal's source text, so exercise it the same way.
	newlineSeps := []string{`"\n"`, `"\n\n"`}
	safeSeps := []string{`", "`, `" • "`, `";"`}

	lit := func(v string) *ast.BasicLit { return &ast.BasicLit{Kind: token.STRING, Value: v} }
	contains := func(l *ast.BasicLit) bool {
		// mirror of the check's condition
		return containsEscapedNewline(l.Value)
	}
	for _, s := range newlineSeps {
		if !contains(lit(s)) {
			t.Errorf("separator %s not detected as newline join", s)
		}
	}
	for _, s := range safeSeps {
		if contains(lit(s)) {
			t.Errorf("separator %s wrongly detected as newline join", s)
		}
	}
}
