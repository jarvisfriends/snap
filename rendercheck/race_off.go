// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

//go:build !race

package rendercheck

// raceEnabled reports whether this binary was built with the race detector.
// See loadConformancePackages for why the package-loading checks care.
const raceEnabled = false
