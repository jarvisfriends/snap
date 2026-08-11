# Roadmap

Completed work is pruned from this file (see git history and
`docs/examples-architecture.md` for the record). Items below are open.

## NEXT: finish the tour's expanded help

The tour itself shipped 2026-08-01 (design and deviations recorded in
`docs/examples-architecture.md`, "One page, one host"). One piece of that
design is still open:

- [ ] Expanded help (ctrl+h) does not yet grow rows for the tour chords:
      page nav (tab/shift+tab, alt+←/→), the ctrl+b page strip, and a
      ctrl+t theme row naming the CURRENT tint. `exui.TintID()` already
      exposes the name; the bar's full-help rows are the missing part.

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

Gifs are no longer committed anywhere: `tools/rendertapes` renders every
tape into `dist/`, and the README gallery points at release assets
(`releases/latest/download/<name>.gif`). Gif paths come from each tape's
own `Output` directive. Because the asset URL is fixed per demo, nothing
rewrites markdown and no sidecar records a link — the earlier Charm
`vhs publish` flow, its `<name>.gif.url` records, and the rate-limit
pacing it needed are all gone.

`.github/workflows/demos.yml` owns the pipeline: PRs touching tapes or
example sources render every tape in the vhs container (build artifact
only); a published release gets its own gifs attached, which is what gives
each version a visual history; merges to the default branch refresh the
gifs on the latest release. Open items:

- [ ] Land the demos workflow on the default branch, then run it via
      `workflow_dispatch` so the current release carries gifs — until then
      the gallery URLs 404 and `.lycheeignore` skips them.
- [ ] Same release-asset flow still needed in inspector and tui-base.

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
