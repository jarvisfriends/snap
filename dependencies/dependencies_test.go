// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package dependencies

import (
	"runtime/debug"
	"testing"
	"time"
)

// syntheticInfo builds a BuildInfo exercising the branches the test binary's
// own build info never hits: a replaced module, every recognized VCS
// setting, an unparsable vcs.time, and a passthrough setting.
func syntheticInfo() *debug.BuildInfo {
	return &debug.BuildInfo{
		GoVersion: "go1.27.0",
		Main:      debug.Module{Path: "example.com/app", Version: "(devel)"},
		Deps: []*debug.Module{
			{Path: "example.com/zeta", Version: "v1.0.0", Sum: "h1:zzz"},
			{
				Path: "example.com/alpha", Version: "v2.3.4", Sum: "h1:aaa",
				Replace: &debug.Module{Path: "example.com/fork", Version: "v0.0.1"},
			},
		},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abc123"},
			{Key: "vcs.time", Value: "2026-08-22T12:00:00Z"},
			{Key: "vcs.modified", Value: "true"},
			{Key: "CGO_ENABLED", Value: "0"},
		},
	}
}

func TestDependenciesFromSortsAndRecordsReplacements(t *testing.T) {
	deps := dependenciesFrom(syntheticInfo())
	if len(deps) != 2 {
		t.Fatalf("got %d deps, want 2", len(deps))
	}
	// Sorted by path: alpha before zeta regardless of input order.
	if deps[0].Path != "example.com/alpha" || deps[1].Path != "example.com/zeta" {
		t.Errorf("deps not sorted by path: %+v", deps)
	}
	if deps[0].Replace != "example.com/fork@v0.0.1" {
		t.Errorf("replacement not recorded: %q", deps[0].Replace)
	}
	if deps[1].Replace != "" {
		t.Errorf("unreplaced module carries a replacement: %q", deps[1].Replace)
	}
}

func TestExpandedFromDecodesVCSSettings(t *testing.T) {
	info := expandedFrom(syntheticInfo())
	if info.App.Path != "example.com/app" {
		t.Errorf("app path = %q", info.App.Path)
	}
	if info.App.Version != "development" {
		t.Errorf("(devel) should normalize to development, got %q", info.App.Version)
	}
	if info.GoVersion != "go1.27.0" {
		t.Errorf("go version = %q", info.GoVersion)
	}
	if info.VCS.Revision != "abc123" {
		t.Errorf("revision = %q", info.VCS.Revision)
	}
	want := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	if info.VCS.Time == nil || !info.VCS.Time.Equal(want) {
		t.Errorf("vcs.time = %v, want %v", info.VCS.Time, want)
	}
	if info.VCS.Modified == nil || !*info.VCS.Modified {
		t.Errorf("vcs.modified = %v, want true", info.VCS.Modified)
	}
	if info.Settings["CGO_ENABLED"] != "0" {
		t.Errorf("passthrough setting missing: %v", info.Settings)
	}
	if len(info.Dependencies) != 2 {
		t.Errorf("dependencies not filled: %+v", info.Dependencies)
	}
	if info.Runtime.CPUs < 1 || info.Runtime.GOOS == "" || info.Runtime.GOARCH == "" {
		t.Errorf("runtime facts missing: %+v", info.Runtime)
	}
}

func TestExpandedFromToleratesBadTime(t *testing.T) {
	info := syntheticInfo()
	info.Settings[1].Value = "not-a-timestamp"
	got := expandedFrom(info)
	if got.VCS.Time != nil {
		t.Errorf("unparsable vcs.time should stay nil, got %v", got.VCS.Time)
	}
}

func TestNormalizeVersion(t *testing.T) {
	cases := map[string]string{
		"":         "development",
		"(devel)":  "development",
		" (devel)": "development",
		"v1.2.3":   "v1.2.3",
	}
	for in, want := range cases {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

// The public wrappers read the test binary's real build info; what must hold
// everywhere is that they return without error and sort stably.
func TestPublicWrappersOnRealBuildInfo(t *testing.T) {
	deps := Dependencies()
	for i := 1; i < len(deps); i++ {
		if deps[i-1].Path > deps[i].Path {
			t.Fatalf("Dependencies() not sorted at %d: %q > %q", i, deps[i-1].Path, deps[i].Path)
		}
	}
	if info := ExpandedBuildInfo(); info == nil {
		t.Fatal("ExpandedBuildInfo() = nil for a test binary with build info")
	}
}
