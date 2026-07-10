
# Tasks to update

## Overall
- [x] ~~Height issue~~ Fixed 2026-07-10: demo roots render inline by default, pinning content to the prompt line — they now set AltScreen. (The empty gifs had a second cause: the vhs container has no Go toolchain, so in-tape `go build` failed; rendertapes now cross-compiles demo binaries on the host.)
- [x] Done 2026-07-10: `tools/rendertapes` (own module) renders all tapes through the official vhs container via the Docker/Podman Go client, worker pool = NumCPU.
- [x] Done 2026-07-10: both tapes showcase ScrollUp/ScrollDown alongside keyboard input.

## Date Picker


## Time Picker

- [ ] Add in an optional seconds value, make private or remove hour and minute, use time.Time to store and retrieve Hour Minute and Second
- [ ]

## Maintenance pass (2026-07-10)

Done: `.golangci.yml` ported from tui-base (same linters, snap module path);
strict CI in `.github/workflows/ci.yml` (build/race/vet + lint both GOOS +
tidy/gofmt/go-fix/generate drift + docs/shell/workflow lint — all hard
failures); `tools/local_verify.sh` rewritten continue-through (drift =
WARN locally, everything runs, failures summarized at the end).

Dedupe (tests pinned behavior before each removal): `clamp` →
`geom.Clamp`; timepicker `cellRect` → `geom.Rect`; the five identical
`View.OnMouse` closures → `uifx.RouteToUpdate`; the three 16-color palette
blocks in `styles/accessibility.go` → one `tintPairs` helper (fallback
palette label "Magenta" is now "Purple", matching themed palettes).

Unit-test coverage snapshot (after this pass):

| Package | Coverage | Notes |
| --- | --- | --- |
| geom, keys, page, uifx | 100% | |
| gate | 96.5% | was 63.5% |
| charts | 92.7% | was 46.8% — pie/sankey/smoothstep now covered |
| styles | 91.1% | |
| timepicker | 90.6% | |
| pickers | 81.5% | |
| notifications | 80.4% | |
| navigation | 78.4% | |
| datepicker | 75.9% | |
| status, table | ~75% | |
| dependencies | 63.9% | build-info paths need a real binary |
| examples/datepicker | 60.9% | demo `main` glue |
| rendercheck | 56.2% | was 38.7%; AST standards checkers exercised by tui-base's suite |
| winterm | 38.1% | registry I/O untestable without mutating HKCU; pure GUID helpers covered |
