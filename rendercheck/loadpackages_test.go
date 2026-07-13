package rendercheck

import (
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func loadTestConfig() *packages.Config {
	return &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles,
		Tests: false,
	}
}

// TestBogusPatternIsDetectedAsVacuous pins the exact failure mode that let a
// consumer's conformance test pass while checking nothing: a pattern naming a
// module that doesn't exist loads zero clean packages, and the failure
// message tells the author how to fix the pattern (go.mod module path or
// "./..."). Before loadConformancePackages, the load "errors" were logged
// and skipped, so the check walked zero packages and passed green.
func TestBogusPatternIsDetectedAsVacuous(t *testing.T) {
	pkgs, err := packages.Load(loadTestConfig(), "github.com/jarvisfriends/does-not-exist/...")
	if err != nil {
		t.Fatalf("Load itself should not error on an unknown pattern: %v", err)
	}
	clean, loadErrors := partitionLoadedPackages(pkgs)
	if len(clean) != 0 {
		t.Fatalf("a nonexistent module pattern should load no clean packages, got %d", len(clean))
	}

	msg := noPackagesMessage([]string{"github.com/jarvisfriends/does-not-exist/..."}, loadErrors)
	for _, want := range []string{
		"NOTHING was checked",
		"go.mod",
		`"./..."`,
		"github.com/jarvisfriends/does-not-exist/...",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("vacuous-pattern message missing %q:\n%s", want, msg)
		}
	}
}

// TestValidPatternLoadsCleanPackages: the happy path partitions everything
// into clean with no load errors, so hardened checks keep working unchanged.
func TestValidPatternLoadsCleanPackages(t *testing.T) {
	pkgs, err := packages.Load(loadTestConfig(), "github.com/jarvisfriends/snap/rendercheck")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	clean, loadErrors := partitionLoadedPackages(pkgs)
	if len(clean) == 0 {
		t.Fatal("expected the rendercheck package itself to load cleanly")
	}
	if len(loadErrors) != 0 {
		t.Fatalf("unexpected load errors: %v", loadErrors)
	}
}
