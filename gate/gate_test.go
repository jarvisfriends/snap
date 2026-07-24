package gate

import (
	"sync"
	"testing"
)

func TestGateRegistry_RegisterAndValue(t *testing.T) {
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "debug", Default: false, Description: "debug mode"})

	if g.Value("debug") {
		t.Error("expected default value of 'debug' to be false")
	}
	if !g.Value("unknown") {
		t.Error("expected unknown gate to default to true")
	}
}

func TestGateRegistry_RegisterDefault_True(t *testing.T) {
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "feature-a", Default: true})
	if !g.Value("feature-a") {
		t.Error("expected feature-a default value to be true")
	}
}

func TestGateRegistry_Set(t *testing.T) {
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "aaa", Default: false})

	g.Set("aaa", true)
	if !g.Value("aaa") {
		t.Error("expected Value('aaa') to be true after Set(true)")
	}

	g.Set("aaa", false)
	if g.Value("aaa") {
		t.Error("expected Value('aaa') to be false after Set(false)")
	}
}

func TestGateRegistry_Toggle(t *testing.T) {
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "x", Default: false})

	g.Toggle("x")
	if !g.Value("x") {
		t.Error("expected Value('x') to be true after first Toggle")
	}
	g.Toggle("x")
	if g.Value("x") {
		t.Error("expected Value('x') to be false after second Toggle")
	}
}

func TestGateRegistry_DuplicatePanics(t *testing.T) {
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "dup"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected duplicate registration to panic")
		}
	}()
	g.Register(FeatureGate{Name: "dup"})
}

func TestGateRegistry_RegisterAll(t *testing.T) {
	g := NewGateRegistry()
	g.RegisterAll([]FeatureGate{
		{Name: "a", Default: true},
		{Name: "b", Default: false},
	})

	if !g.Value("a") {
		t.Error("expected Value('a') to be true")
	}
	if g.Value("b") {
		t.Error("expected Value('b') to be false")
	}
}

func TestGateRegistry_LoadFromEnv_Standard(t *testing.T) {
	t.Setenv("TESTAPP_GATE_DEBUG_FLAG", "true")
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "debug-flag", Default: false})
	g.LoadFromEnv("testapp")
	if !g.Value("debug-flag") {
		t.Error("expected env var to override default false to true")
	}
}

func TestGateRegistry_LoadFromEnv_DevOverride(t *testing.T) {
	t.Setenv("TESTAPP_DEV_FEATURE_X", "true")
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "feature-x", Default: false})
	g.LoadFromEnv("testapp")
	if !g.Value("feature-x") {
		t.Error("expected DEV override env var to override default false to true")
	}
}

func TestGateRegistry_LoadFromEnv_DevOverridePrecedence(t *testing.T) {
	// standard is false, dev override is true
	t.Setenv("TESTAPP_GATE_FEATURE_Y", "false")
	t.Setenv("TESTAPP_DEV_FEATURE_Y", "true")
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "feature-y", Default: false})
	g.LoadFromEnv("testapp")
	if !g.Value("feature-y") {
		t.Error("expected DEV override to take precedence over standard env var")
	}

	// standard is true, dev override is false
	t.Setenv("TESTAPP_GATE_FEATURE_Z", "true")
	t.Setenv("TESTAPP_DEV_FEATURE_Z", "false")
	g.Register(FeatureGate{Name: "feature-z", Default: true})
	g.LoadFromEnv("testapp")
	if g.Value("feature-z") {
		t.Error("expected DEV override false to take precedence over standard true")
	}
}

func TestGateRegistry_ConcurrentAccess(t *testing.T) {
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "shared", Default: false})

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = g.Value("shared")
		}()
		go func() {
			defer wg.Done()
			g.Toggle("shared")
		}()
	}
	wg.Wait()
}

// TestGateRegistryIntrospection covers the registry's read-side API the
// settings Feature Flags UI is built on: Defs/Order preserve registration
// order, Snapshot copies values, ApplyMap only touches registered gates.
func TestGateRegistryIntrospection(t *testing.T) {
	g := NewGateRegistry()
	g.RegisterAll([]FeatureGate{
		{Name: "zeta", Default: true},
		{Name: "alpha", Default: false},
	})

	if got := g.Order(); len(got) != 2 || got[0] != "zeta" || got[1] != "alpha" {
		t.Fatalf("Order() = %v; want registration order [zeta alpha]", got)
	}
	defs := g.Defs()
	if len(defs) != 2 || defs[0].Name != "zeta" || defs[1].Name != "alpha" {
		t.Fatalf("Defs() = %v; want registration order", defs)
	}

	snap := g.Snapshot()
	if !snap["zeta"] || snap["alpha"] {
		t.Fatalf("Snapshot() = %v; want defaults zeta=true alpha=false", snap)
	}
	// The snapshot is a copy: mutating it must not touch the registry.
	snap["zeta"] = false
	if !g.Value("zeta") {
		t.Fatal("mutating a Snapshot changed the registry")
	}

	g.ApplyMap(map[string]bool{"alpha": true, "ghost": true})
	if !g.Value("alpha") {
		t.Fatal("ApplyMap did not apply a registered gate")
	}
	if _, ok := g.Snapshot()["ghost"]; ok {
		t.Fatal("ApplyMap invented an unregistered gate")
	}
}

// TestLoadFromEnvPrefix pins the prefix form used by branded apps: gate names
// are upcased with non-alphanumerics mapped to underscores.
func TestLoadFromEnvPrefix(t *testing.T) {
	g := NewGateRegistry()
	g.Register(FeatureGate{Name: "router.accessibility-tab", Default: false})
	t.Setenv("MYAPP_GATE_ROUTER_ACCESSIBILITY_TAB", "1")
	g.LoadFromEnvPrefix("MYAPP_GATE_")
	if !g.Value("router.accessibility-tab") {
		t.Fatal("LoadFromEnvPrefix did not apply the env override")
	}

	// Unset / junk values leave the default alone.
	g2 := NewGateRegistry()
	g2.Register(FeatureGate{Name: "other", Default: true})
	t.Setenv("MYAPP_GATE_OTHER", "banana")
	g2.LoadFromEnvPrefix("MYAPP_GATE_")
	if !g2.Value("other") {
		t.Fatal("junk env value changed a gate")
	}
}
