
# Tasks to update

## Overall
- [x] ~~Height issue~~ Fixed 2026-07-10: demo roots render inline by default, pinning content to the prompt line — they now set AltScreen. (The empty gifs had a second cause: the vhs container has no Go toolchain, so in-tape `go build` failed; rendertapes now cross-compiles demo binaries on the host.)
- [x] Done 2026-07-10: `tools/rendertapes` (own module) renders all tapes through the official vhs container via the Docker/Podman Go client, worker pool = NumCPU.
- [x] Done 2026-07-10: both tapes showcase ScrollUp/ScrollDown alongside keyboard input.
- [x] Done 2026-07-10: README gained a Gallery section — every demo gif (datepicker, timepicker, charts, pickers, menu, scrollbar, table) with a brief description; gifs regenerate via `go -C tools/rendertapes run .`.
- [x] Mostly done 2026-07-10: new demos + tapes for pickers, menu, scrollbar,
  table, and the charts model example (7 tapes render clean end to end).
  ~~Remaining: navigation, status, and notifications need a host-shaped app
  (router + pages) to demo meaningfully — their tape should live in the
  tui-base reference app instead of a synthetic snap example.~~ Done
  2026-07-10 in tui-base: `examples/multipage/demo.tape` (navigation) and
  `cmd/tui-base/notifications.tape` (toasts + status bar), rendered by
  tui-base's own `tools/rendertapes` port; gifs pending a Docker-equipped
  machine.
- [x] Done 2026-07-10: rendertapes walks the repo for every `*.tape` at any depth (skipping .git/tools), so `charts/sparkline.tape` etc. render too.
- [x] Answered 2026-07-10: they are the **same wire protocol** — Bubble Tea
  v2's renderer emits OSC 9;4 for `tea.View.ProgressBar` (states None/
  Default/Error/Indeterminate/Warning map to protocol states 0/1/2/3/4,
  see bubbletea `cursed_renderer.go` `setProgressBar`). Inside a running
  program, prefer `View.ProgressBar`: the renderer diffs it (re-emits only
  on change), serializes it with frame output, and resets it on exit. The
  `osc` package offers exactly one thing `View.ProgressBar` can't: progress
  **outside** a running Bubble Tea program — cobra commands, pre-`tea.Run`
  setup phases, post-exit work, plain scripts. Its doc comment now says so.
- [x] Done 2026-07-10: cross-GOOS lint is now targeted. Research notes:
  most Go projects either lint once per CI-runner OS (paying full double
  runs) or ignore the problem; golangci-lint has no per-file GOOS mode
  (typechecking needs whole packages), so the sweet spot is per-package.
  We now run one **full** lint on the native GOOS, then detect packages
  whose sources actually diverge by OS (`*_windows.go`-style suffixes or
  `//go:build` lines naming an OS — currently just `winterm/`) and lint
  **only those** under the other GOOS. Implemented identically in
  `tools/local_verify.sh` and `.github/workflows/ci.yml` (detection logic
  kept in sync, commented on both sides).

## Date Picker
- [x] Done 2026-07-10: month/year jump without week-walking, three ways —
  PgUp/PgDn (or `[` / `]`) page months and Shift+PgUp/PgDn (`{` / `}`) page
  years from any focus; the mouse wheel over the title line pages months on
  its left half and years on its right; and the existing Tab-to-header +
  arrows still works. All in DefaultKeyMap so apps can rebind.


## Time Picker

- [x] Done 2026-07-10: `NewTimeField(t time.Time)` + `SetTime`/`Time()` —
  the date part (and location) of the timestamp round-trips untouched with
  only the clock edited, so the field pairs with the datepicker for full
  timestamps. `Hour`/`Minute` are private; `ShowSeconds` adds an optional
  third column (dropdown, type-ahead, wheel, hover — all side-generic now).
  Breaking change, absorbed by the example + tests (no other consumers).


## Charts

All done 2026-07-10 — the pure render functions stay (and stay tested); the
new `*Model` types wrap them per chart file:

- [x] Models: `SparklineModel`, `HBarModel`, `PieModel`, `SankeyModel`,
  `LineChartModel` (Init/Update/View), each in its chart's `_model.go`.
- [x] ID routing: every data message carries the target chart's ID
  (`SparklineDataMsg`/`SparklinePointMsg`, `HBarDataMsg`, `PieDataMsg`,
  `SankeyDataMsg`, `LineDataMsg`); charts ignore other IDs, so hosts just
  forward everything. **examples/charts is the canonical copy-paste
  wiring** — two sparklines + pie + sankey + hbar, one producer stream.
- [x] Sizing: the shared `Frame` gives every model
  `MaxWidth`/`MaxHeight`/`SetSize` caps and `Used()` reporting, so layouts
  can pack charts and roll the ones that don't fit.
- [x] Stretch-to-fill: sparkline/hbar fill their width, sankey/linechart
  fill both axes, pie fits its radius to the box (height/2 vs width/4).
- [x] Resize: every model consumes `tea.WindowSizeMsg`; multi-chart hosts
  re-split via `SetSize` (shown in the example).
- [x] Dynamic pie values: slices under `MinSliceFrac` (default 2%) fold
  into a dim "Other" slice; `Combined()` returns what was folded so the
  host renders a legend (the example does).

## Scrollbar

- [x] Restyled 2026-07-10. Research: modern TUIs converge on two looks —
  fzf/yazi use a thin line track with a heavier thumb (or a floating thumb
  and no track), and btop gets its polish from **sub-cell precision**:
  partial cells drawn with eighth-block glyphs so the thumb glides instead
  of jumping a whole cell at a time. Both are in as presets:
  - `PresetLine` (new default): dim `│` track, bright heavy `┃` thumb.
  - `PresetSmooth`: floating block thumb with 1/8-cell resolution — the
    thumb's bottom edge renders as a lower-eighth block and its top edge as
    the inverted complement (there are no upper-eighth glyphs, so the style
    reverses to paint the remainder), giving 8x the positional steps.
  - `PresetClassic`: the old `░` / `█` for the actually-retro moods.
  Glyphs and colors stay overridable per Styles.

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

## Repo sweep for forgotten snaps (2026-07-10)

Swept: `w`, `anvil`, `verify_setup`, `weaver_base`, `brick-breaker`,
`aSettings`, `tribble/tui` (plus dash, already mined for charts).

### Ported this pass

- `scrollbar/` — tribble's dashboard scrollbar, decoupled (pure geometry +
  injectable styles, `ClampOffset` helper).
- `menu/` — tribble's right-click context menu: disabled items, cursor skip,
  terminal-edge clamping, hover/wheel/click via the host's OnMouse,
  compositor-based overlay. Widths are now unicode-safe (`lipgloss.Width`,
  the original used `len`).
- `osc/` — aSettings' OSC 9;4 taskbar progress, extended from
  indeterminate/clear to the full protocol (determinate, error, paused) with
  an injectable writer for tests.

### Candidates found, not yet ported

- [ ] **Elevation / privilege** — TWO re-implementations: verify_setup
  `internal/privilege` (IsElevated + relaunch-elevated with env marker,
  cross-platform) and anvil `tui/features/software/elevate_windows.go`
  (ShellExecuteExW runas + wait + exit code). Merge into one `snap/elevate`
  taking the best of both: verify_setup's no-admin-by-default design +
  anvil's wait-for-exit-code. Blocked on: needs elevated-Windows manual
  testing before it ships.
- [ ] **Generic list-picker overlay** — tribble `ui/overlay_picker.go`
  (numbered items, descriptions, opaque UserData). Overlaps huh selects and
  snap pickers; decide whether it becomes `snap/listpicker` or a huh recipe.
- [ ] **Notification progress** — w `ui/shared/notification.go` carries a
  `Percent` field for progress-bar notifications; snap/notifications has no
  progress concept. Add Percent + a bar renderer (charts.HBar) to the
  notification model and history panel.
- [x] Done 2026-07-10: **Badge/pill styles** — `styles/pill.go`. Six
  user-selectable `PillShape`s (string-preset pattern like `StylePreset`):
  Round half-circles (default), Arrow, Slant, Flame — Powerline-extras
  glyphs, `NeedsNerdFont()` — plus pure-Unicode Block and padded Plain.
  `Pill` renders one badge, `SegmentedPill` divides one pill by color
  (solid divider = prev bg over next bg, thin variant when bgs match),
  `Breadcrumbs` joins items on the thin glyph; nil Fg auto-picks
  black/white by bg luminance. `examples/pills` demos every shape as
  badges, segmented status runs, nav items, and breadcrumbs. The
  Catppuccin categorical palette stays app-side (demo hardcodes its own).
  NOTE: examples/pills/demo.gif is not rendered yet — this machine has no
  Docker/Podman; run `go -C tools/rendertapes run .` where one exists.
- [ ] **Box layout helpers** — w `ui/shared/layout.go` (ContentOrigin,
  InnerSize, RenderInBox). Check overlap with page/geom before porting.
- [ ] **Input parse helpers** — w `ui/shared/input_validation.go`
  (required/duration/… parsers with friendly errors) — possible
  `snap/forms` seed.
- [ ] **Cell canvas + gradients** — brick-breaker's `gameRenderer` is a
  colored cell canvas (set/setFG per cell, color gradients); charts.Canvas
  is braille-only. A shared cell canvas could back both, and `gradient()`
  is useful for charts.

### Copied to snap but not yet removable from the source

| Source | Snap home | Why it can't be removed yet |
| --- | --- | --- |
| dash `creator/charts.go` (+ canvas/linechart) | `charts/` | dash isn't flipped to depend on snap yet — same tag-then-flip flow as tui-base (snap PR + v0.1.x tag first). |
| weaver_base `gate.go`/`registry.go`/`feature.go` | `gate/` | weaver_base lives on the work Bitbucket network; it can't pull public GitHub modules until its module proxy allows it (or snap is mirrored internally). |
| tui-base `theme/` (alias shim over `snap/styles`) | `styles/` | Kept intentionally as compat aliases until downstream apps migrate their imports. |
| tui-base settings picker tests (makePickerTree/assertFrameFits copies) | `pickers/` tests | Intentional duplication: snap owns component-level tests, tui-base keeps integration coverage of the same flows. |
| tribble `ui/panel_zone.go` | superseded by `uifx.Zones` | tribble is on the work Bitbucket network (same constraint as weaver_base); swap in uifx.Zones when it can depend on snap. |
| aSettings `pages/ui/table_mouse.go` | superseded by `table/` (HandleClick) | aSettings hasn't adopted snap/table; remove when it does. |
| w `ui/shared/notification.go` | `notifications/` (partial) | w isn't on snap; also carries the Percent feature snap doesn't have yet (see candidate above). |

### Not worth moving (checked, domain-specific)

w tabs (timecard/jira/finance...), w task_suggestions (jira-typed),
aSettings cui (alias scanning), anvil bios/software features (WMI +
installers), verify_setup pages/overview, tribble dashboards/telemetry,
brick-breaker game logic, weaver_base settings renderer (superseded by
tui-base settings).
