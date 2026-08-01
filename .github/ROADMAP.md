# Roadmap

Completed work is pruned from this file (see git history and
`docs/examples-architecture.md` for the record). Items below are open.

## NEXT: snap_input tour mode (pages + theme cycling) — designed 2026-08-01

One program, one code path — kills the current split brain where each
example owns its own Run/quit/output logic.

Contract changes (a major rev; breaking the examples' API is fine):

- Each example package exports `New() tea.Model` (its current `newDemo`
  with the fixed demo args) and a `Result() []exui.Field` method returning
  the value(s) the user has confirmed so far (empty until confirmed).
  Per-package `Run()` is deleted.
- `snap_input` with no args runs the tour over ALL examples;
  `snap_input <name>` runs the same tour with a single page. Same logic,
  page count 1..N.
- Quit semantics invert: ONLY `q`/`ctrl+c` (shell-owned, in
  `exui.Chrome.Update`) quits the whole app. Selection actions (enter,
  click-confirm, second-click on a date) record the page's result and stay
  running — no component or click ever quits.
- Tab / Shift+Tab move between pages. Page models are recreated on entry
  (recreation is the sanctioned pattern here; tui-base instead re-themes
  live via messages — that stays out of scope for the examples).
- A `navigation.Tabs` strip above the content shows the pages; HIDDEN by
  default, toggled with `keys.ToggleNav` (ctrl+b).
- Theme cycling on a key: needs `styles.TintIDs()` (new accessor over the
  tint registry) and `exui` theme cache made mutable (`sync.Once` →
  guarded rebuild). Chrome + current page rebuild on cycle since
  components snapshot styles at construction.
- Expanded help (ctrl+h) grows rows for the less-common actions: toggle
  tabs, toggle status bar (`keys.ToggleStatus`), theme cycle showing the
  CURRENT TINT NAME, notifications (ctrl+n), info (ctrl+e), debug (ctrl+d).
- On exit, emit `Result()` from every VISITED page (in visit order,
  skipping unvisited and empty pages) through the existing `--output`
  formatter (pretty/values/json/yaml/xml), namespaced per page — e.g.
  `datepicker.date` in pretty, nested objects in json/yaml/xml.

Estimated shape: 14 small per-package edits (New/Result, drop Run), one
new `examples/snap_input/tour.go`, ~40 lines in `exui` chrome, the styles
accessor, and USAGE.md/README updates.

## Cross-cutting conflicts queued for the same major rev

Tracked in detail in `docs/examples-architecture.md` ("Cross-cutting
conflicts"): InfoModal ignoring injected colors; history-panel actions
(Enter/dismiss-all) living in the exui shell instead of the status
package; `NotifHistoryCursorDown(n)` requiring callers to pass counts;
toasts documented but unrendered; legacy `styles/status.go` renderers;
menu's private KeyMap/Styles vs keys/styles adapters (datepicker,
timepicker done via `ApplyTimeFieldTheme` — menu, charts/pie hardcodes
remain); deriving tool version pins from go.mod in CI.

## Demo pipeline (state as of 2026-08-01)

Gifs are no longer committed anywhere: `tools/rendertapes -publish`
renders every tape, uploads via `vhs publish`, writes `<name>.gif.url`
beside each tape, and rewrites markdown image links to the hosted URLs.
Open items:

- [ ] Run `-publish` once on a Docker-equipped machine so README image
      links point at live hosted URLs (they reference removed files until
      then). Same one-time publish needed in inspector and tui-base.
- [ ] Port the `-publish` flow into tui-base's rendertapes copy.

## Ports & adoption still open

- [ ] **Elevation / privilege** — merge verify_setup `internal/privilege`
      (no-admin-by-default, relaunch-elevated) with anvil's
      ShellExecuteExW runas + wait-for-exit-code into one `snap/elevate`.
      Blocked on elevated-Windows manual testing.
- [ ] Flip the "copied but not yet removable" sources when they can depend
      on snap: weaver_base gate (work Bitbucket proxy), tribble panel_zone
      (→ uifx.Zones), aSettings table_mouse (→ table.HandleClick), w
      notification/input_validation/layout (→ notifications/forms/layout),
      brick-breaker gameRenderer (→ charts.CellCanvas), tui-base `theme/`
      alias shim (until downstream imports migrate).
- [ ] ntcharts fork-replace check: NimbleMarkets/ntcharts go.mod replaces
      bubbletea with a fork "awaiting upstream merges" (doesn't propagate
      to consumers) — re-verify at each ntcharts upgrade.
- [ ] rendercheck's new byte-vs-cell checks reach tui-base at the next tag
      flip; its tree may need the same sweep before it goes green.
