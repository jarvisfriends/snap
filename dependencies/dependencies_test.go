package dependencies

import "testing"

func TestDependenciesAndBuildInfo(t *testing.T) {
	t.Parallel()

	deps := Dependencies()
	t.Logf("Found %d dependencies", len(deps))

	info := ExpandedBuildInfo()
	if info != nil {
		t.Logf("Go Version: %s", info.GoVersion)
		t.Logf("OS: %s, Arch: %s", info.Runtime.GOOS, info.Runtime.GOARCH)
	} else {
		t.Log("ExpandedBuildInfo returned nil (not in module context)")
	}
}

func TestNormalizeVersion(t *testing.T) {
	t.Parallel()

	if normalizeVersion("(devel)") != "development" {
		t.Fatal("expected normalizeVersion to normalize devel builds")
	}
}
