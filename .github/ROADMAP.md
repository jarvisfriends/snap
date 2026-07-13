
# Tasks to update

## Chart library adoption: ntcharts + asciigraph (2026-07-11)

Evaluated github.com/NimbleMarkets/ntcharts (v2.2.0 — targets charm.land v2;
its go.mod replaces bubbletea with a fork "awaiting upstream merges", which
does NOT propagate to consumers but is worth re-checking at upgrades) and
github.com/guptarohit/asciigraph (v0.10.0, zero-dep ASCII plots with axes).

Duplicates removed:

- **snap `charts/canvas.go` (braille pixel canvas) — deleted.**
  `BrailleLineChart` now plots through ntcharts' `canvas` + `canvas/graph`
  primitives (BrailleGrid per series → dot patterns merged onto one canvas;
  interpolated segments replace the old vertical-run fill). Same signature,
  window, NaN gaps, and scale reporting; per-cell multi-series color
  *blending* was the one lost feature (overlap cells now take the later
  series' color — dot patterns still merge).
- **dash `creator/{charts,canvas,linechart}.go` — deleted** (the original
  copies snap was extracted from). creator now re-exports snap/charts via
  aliases; the game's pixel canvas became `examples/gamecanvas.go` on
  ntcharts primitives (SetPixel/Line/ThickLine/Text preserved).

Kept, deliberately (not duplicates):

- **Sparkline** — ntcharts sparklines have no directional coloring; snap's
  braille-up/down styles color glyphs green/red by value direction and dash
  exposes the style names in settings. Feature superset stays ours.
- **Pie, Sankey, HBar, CellCanvas+Gradient** — no ntcharts/asciigraph
  equivalent (sankey keeps per-cell color blending via its own hexRGB).

Available to dash now (direct deps in dash go.mod): the full ntcharts option
surface — barchart (grouped/horizontal), heatmap, linechart with axes/tick
labels/mouse zones, OHLC candlesticks, streamline/timeseries/waveline
variants, canvas/graph primitives (wired: examples/gamecanvas.go) — and all
asciigraph options (height/width/bounds/precision/series colors/legends/
captions; wired: the gallery's asciigraph section).

## Tape coverage scan (2026-07-11)

Swept every package for demoable capabilities without a tape. Four new
examples (each with a tape and a render_check test that drives the same flow
headlessly): `examples/dependencies` (the build-info reader rendered by
status.InfoModal — wheel scroll + click-outside via its HandleMouse),
`examples/linechart` (the one chart model the charts demo didn't show:
braille 2x4 dots, per-cell blend, rolling stream), `examples/cellcanvas`
(CellCanvas + Gradient truecolor plasma, two pixels per cell via '▀'), and
`examples/forms` (live parse-and-validate with field-naming errors).
README gallery updated; gifs pending a Docker-equipped machine
(`go -C tools/rendertapes run .`).

Deliberately not taped: navigation/status-bar/notification toasts (demoed
host-shaped in tui-base, decided 2026-07-10); gate (no visual surface —
worth revisiting if a settings page ever renders it); osc (writes to the
terminal's taskbar/tab, which a gif can't capture); winterm (registry
repair); geom/keys/layout/page/uifx/rendercheck (infrastructure);
logging (placeholder).

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

## Keyboard/mouse parity pass (2026-07-10)

Audited every interactive component for the "keyboard and mouse" design
rule. Already at parity: datepicker, timepicker, table, dir/multi-file
pickers, tabs, topnav (each handles keys + click + wheel). Gaps found and
fixed:

- **menu**: mouse was one `HandleMouse` call but keyboard forced every host
  to hand-wire up/down/enter/esc from the movement primitives. Added
  `menu.KeyMap` (rebindable, `DefaultKeyMap`: arrows/jk, Enter, Esc) and
  `HandleKey` — the keyboard twin of `HandleMouse`, modal while open.
  examples/menu now uses it.
- **scrollbar**: was render-only — a scrollbar you couldn't click. Added
  `OffsetAt` (pure, like the rest of the package): maps a click/drag row on
  the bar to the offset that centers the thumb there; the inverse of
  `Vertical`'s placement (round-trip pinned by test). examples/scrollbar
  wires click + drag-to-scrub on all three bars.
- **status.InfoModal**: keyboard-complete, but pointer support required
  each host to hand-roll Bounds hit-testing, wheel routing, and
  outside-click detection. Added `HandleMouse`: wheel scrolls, click
  outside closes (returning the same CloseInfoModalMsg cmd as Dismiss),
  everything else consumed while open. tui-base's router can drop its
  hand-rolled version next tag flip (additive; nothing breaks meanwhile).
- **navigation.Sidebar**: tabs/topnav wheel-cycle pages but the sidebar
  ignored the wheel. Added `verticalWheelDelta` (twin of
  `horizontalWheelDelta`) and wheel-steps-pages in the sidebar's
  handleMouse, wrapping like the others.

Tape/gallery pass in the same sitting: the datepicker tape now shows off
month/year paging (`]`/`[`, `}`/`{`) — its headline feature was missing
from the gif — and the table tape now demos the 3-state `s` sort cycle and
live `/` filtering before opening a row. README gallery text updated to
match; gifs pending a Docker-equipped machine (`go -C tools/rendertapes
run .`). VHS has no mouse-click/drag commands, so scrollbar click/drag and
InfoModal outside-click stay keyboard-demoed in tapes.

## Byte-vs-cell string hygiene sweep (2026-07-11)

Removed every byte-based string operation from render paths, repo-wide
(started from examples/pills' `%-9s` + `name[2:]`):

- **No `fmt.Sprintf` outside rendercheck's failure messages.** Byte-padded
  verbs became lipgloss cell padding (`PlaceHorizontal`/`Style.Width`);
  value formatting became `strconv` + concatenation; `%02d` became local
  `pad2` helpers (datepicker, timepicker); the triplicated `#%02x%02x%02x`
  hex builders became fmt-free helpers (`charts.hexRGB`, `styles.ColorHex`).
- **`strings.Join(rows, "\n")` → `lipgloss.JoinVertical`** everywhere
  (menu.Render's builder loop, scrollbar ×2, status bar rows, notification
  history, menu/scrollbar examples).
- **Space-run gaps → lipgloss.** status-bar and table footers now
  right-align via `PlaceHorizontal` (styled whitespace, so backgrounds run
  unbroken); the overlay slide-in indent is `PaddingLeft`; example gaps are
  `Width(n)` blank blocks.
- Kept, deliberately: `strings.Split(x, "\n")` for *parsing* rendered
  output (table border scan, rendercheck helpers — no lipgloss equivalent);
  non-newline `strings.Join` on data (key lists, "; " paths, " • "
  segments); `strings.Repeat` of non-space glyphs ("─" rules, "·" fill) and
  standalone blank fills (sparkline) — all cell-safe.

Enforcement (rendercheck, UI packages = imports lipgloss/bubbletea):
`checkSprintfBytePadding` (width-flagged %s/%q/%v), `checkJoinNewline`,
and `checkRepeatSpaceConcat` (space runs concatenated as gaps) joined the
existing len()-as-width and strings.Count checks — and snap now runs
`CheckCodeStandards` on its own whole module
(`rendercheck.TestSnapMeetsOwnCodeStandards`), which immediately caught the
menu KeyMap's vim fallbacks (removed; hosts rebind). tui-base picks the new
checks up automatically at the next tag flip; its tree may need the same
sweep before it goes green.

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
- [x] Decided 2026-07-10: **Generic list-picker overlay** — not ported.
  tribble `ui/overlay_picker.go`'s three ingredients are all covered now:
  overlay positioning + opaque UserData ≈ `menu` (Tag, and since the input
  parity pass keyboard is one `HandleKey` call, same as mouse); form-style
  selects with descriptions ≈ huh Select with `styles.Huh` theming.
  Revisit only if an app actually needs numbered quick-jump items, as a
  small `menu` extension rather than a new package.
- [x] Done 2026-07-10: **Notification progress** — `Notification.Percent
  *float64` (0–100, the charts.HBar scale; nil = not a progress
  notification), carried by `AddMsg`/`AddOptions`, updated in place via
  `ProgressMsg` (by ID, or Key when ID is zero) / `SetProgress` /
  `SetProgressKey` (clamped; re-shows a toast-hidden notification so live
  progress stays visible; stored value is copied so callers can't mutate
  through the pointer). The history panel renders an inline `charts.HBar`
  + percent after the row content. ~~Remaining for the next tag flip:
  tui-base's toast overlay should draw the bar too, and its router must
  route `notifications.ProgressMsg` alongside the other notification
  messages (router.go's Handle forwarding list).~~ Both landed in
  tui-base 2026-07-10 against v0.1.6 (severity-tinted HBar under the
  toast message + ProgressMsg in the forwarding list).
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
- [x] Done 2026-07-10: **Box layout helpers** — checked overlap first: geom
  is pure cell rectangles (no lipgloss) and page is colors/size only, so no
  overlap. Ported as `layout/` (ContentOrigin, InnerSize, RenderInBox) —
  the style-dependent half of hit-testing/sizing that components were
  hand-summing (GetBorderLeftSize+GetPaddingLeft etc.).
- [x] Done 2026-07-10: **Input parse helpers** — ported as `forms/`
  (ParseRequired, ParseDuration, ParseISODate, SplitAndClean), trim-and-
  validate parsers whose errors name the field. w keeps its copy until it
  can depend on snap (added to the not-yet-removable table).
  HUMAN: Yes! please bring this over
- [x] Done 2026-07-10: **Cell canvas + gradients** — `charts.CellCanvas`
  (whole-cell rune+fg+bg surface: Set/SetFG/Clear/Rune, `String()` with
  batched truecolor escapes — colors re-emitted only on change, per-line
  resets) + `charts.Gradient` (HSV blend between two `color.Color`s,
  nil-safe). Ported from brick-breaker's `gameRenderer`; the game keeps
  its drawing primitives and can flip to `charts.CellCanvas` when it
  adopts snap (added to the "copied but not yet removable" table).
  Re-backing the braille `Canvas` with CellCanvas was considered and
  skipped: braille composes dot bitmasks per cell with transparent
  background — a different cell model, and the shared surface would
  complicate both.

### Copied to snap but not yet removable from the source

| Source | Snap home | Why it can't be removed yet |
| --- | --- | --- |
| ~~dash `creator/charts.go` (+ canvas/linechart)~~ | `charts/` | **Flipped 2026-07-11**: dash's creator is now type/func aliases over snap/charts; its braille game canvas re-ported onto ntcharts canvas/graph primitives (examples/gamecanvas.go). |
| weaver_base `gate.go`/`registry.go`/`feature.go` | `gate/` | weaver_base lives on the work Bitbucket network; it can't pull public GitHub modules until its module proxy allows it (or snap is mirrored internally). |
| tui-base `theme/` (alias shim over `snap/styles`) | `styles/` | Kept intentionally as compat aliases until downstream apps migrate their imports. |
| tui-base settings picker tests (makePickerTree/assertFrameFits copies) | `pickers/` tests | Intentional duplication: snap owns component-level tests, tui-base keeps integration coverage of the same flows. |
| tribble `ui/panel_zone.go` | superseded by `uifx.Zones` | tribble is on the work Bitbucket network (same constraint as weaver_base); swap in uifx.Zones when it can depend on snap. |
| aSettings `pages/ui/table_mouse.go` | superseded by `table/` (HandleClick) | aSettings hasn't adopted snap/table; remove when it does. |
| w `ui/shared/notification.go` | `notifications/` (Percent ported 2026-07-10) | w isn't on snap; flip its notification model when it can depend on snap. |
| w `ui/shared/input_validation.go` | `forms/` | w isn't on snap; flip when it can depend on snap. |
| w `ui/shared/layout.go` | `layout/` | w isn't on snap; flip when it can depend on snap. |
| brick-breaker `brick/render.go` gameRenderer | `charts/cellcanvas.go` | brick-breaker isn't on snap; flip its renderer to `charts.CellCanvas` + `charts.Gradient` when it adopts the dependency. |

### Not worth moving (checked, domain-specific)

w tabs (timecard/jira/finance...), w task_suggestions (jira-typed),
aSettings cui (alias scanning), anvil bios/software features (WMI +
installers), verify_setup pages/overview, tribble dashboards/telemetry,
brick-breaker game logic, weaver_base settings renderer (superseded by
tui-base settings).
