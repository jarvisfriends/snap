
# Tasks to update

## Overall
- [x] ~~Height issue~~ Fixed 2026-07-10: demo roots render inline by default, pinning content to the prompt line — they now set AltScreen. (The empty gifs had a second cause: the vhs container has no Go toolchain, so in-tape `go build` failed; rendertapes now cross-compiles demo binaries on the host.)
- [x] Done 2026-07-10: `tools/rendertapes` (own module) renders all tapes through the official vhs container via the Docker/Podman Go client, worker pool = NumCPU.
- [x] Done 2026-07-10: both tapes showcase ScrollUp/ScrollDown alongside keyboard input.
- [ ] Create high level README.md file that includes the gif created for each snap with a brief description of what it is
- [ ] Add all remaining demo.tape files for those components that don't have it
- [ ] Search for tapes by *.tape instead of just demo.tape, useful for things like the charts folder
- [ ] Whats the difference between the osc showing progress and the tea.View.ProgressBar showing it? does the osc offer any that we can't get with the tea.ProgressBar? 
- [ ] My desire with the running of the lint twice was to catch issues that are present on one OS but not another... What if we always run the linux one, then do a grep for the build flags and only run the golangci linters against those files that call out a different OS than linux? There is probably a nicer way to do that... Can you research for a bit to figure out what people usually do and then go for that route? Just make sure we are not running things multiple times that we don't need to... 

## Date Picker
- [ ] Let the user change the year and Month as well instead of having to go through each week if they wanted to for instance set their birthday


## Time Picker

- [ ] Add in an optional seconds value, make private or remove hour and minute, use time.Time to store and retrieve Hour Minute and Second


## Charts

- [ ] Convert each chart to a compatible model (so init, update, view)
- [ ] ID or some other way to determine which of multiple of the same chart types incoming data is for, even if we require the parent to do this, make sure we have an example thats the perfect way to do it in case they just c/p it each time.
- [ ] set height and width maximums that the charts are allowed to use and report the actual height and width being used
  - [ ] Providing used space allows our UI to be a bit more flexible on where things are placed especially when we are adjusting the terminal sizes (we can roll charts down that don't fit on lines this way)
- [ ] Expand out/stretch any of our single column or single row values to fill the area provided (where possible, favor staying within bounds over having to scroll outside the bounds)
  - [ ] This includes the Pie, Sankey, Sparkline, and hbar snaps
- [ ] Handle resize messages now that we are a true model
- [ ] Allow dynamic number of values to be provided when it makes since, for example, Pie chart might be only 2 values or it might be provided 8, do the adjustments internal. If the slices wouldn't be visible, then combine them together with others until they can be visible and somehow supply a key/legend back to the user so they can handle that case if they need to (can be just a function call or something similar)

## Scrollbar
This still looks old school, and not in a good retro way... Can we figure out some better characters to use, or show it on the screen differently... I am not sure what would make it look better, you will have to be the expert on this one after you research a bit on some cool looking scroll bars

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
- [ ] **Badge/pill styles** — aSettings `pages/ui/badges.go` BadgeStyle +
  categorical palette. Generic pill helper belongs in `styles`; the
  Catppuccin categorical palette stays app-side.
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
