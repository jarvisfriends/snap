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

Gifs are no longer committed anywhere: `tools/rendertapes -publish`
renders every tape, uploads via `vhs publish`, writes `<name>.gif.url`
beside each tape, and rewrites markdown image links to the hosted URLs.
Gif paths come from each tape's own `Output` directive, and a re-publish
repoints links from the previously published URL, so the rewrite is
repeatable rather than one-shot. `-relink` replays the recorded URLs into
markdown with no Docker and no network.

`.github/workflows/demos.yml` owns the pipeline: PRs touching tapes or
example sources render every tape in the vhs container (artifact only, no
upload); merges to the default branch render, publish, and open a PR with
the refreshed `.gif.url` files and rewritten links. Open items:

- [ ] Land the demos workflow on the default branch, then run it via
      `workflow_dispatch` against any branch whose README still names local
      `dist/` gif paths. Those paths are placeholders the publish step
      rewrites, so `.lycheeignore` skips them until the real URLs land.
- [ ] Same publish flow still needed in inspector and tui-base.
- [ ] Port the `-publish`/`-relink` flow into tui-base's rendertapes copy.

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
