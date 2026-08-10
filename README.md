# Snap

![snap — ready-to-snap Bubble Tea components](assets/banner.svg)

[![CI](https://github.com/jarvisfriends/snap/actions/workflows/ci.yml/badge.svg)](https://github.com/jarvisfriends/snap/actions/workflows/ci.yml)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/13784/badge)](https://www.bestpractices.dev/projects/13784)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/jarvisfriends/snap/badge)](https://scorecard.dev/viewer/?uri=github.com/jarvisfriends/snap)
[![Go Reference](https://pkg.go.dev/badge/github.com/jarvisfriends/snap.svg)](https://pkg.go.dev/github.com/jarvisfriends/snap)

**Jarvis Friends Snap** — ready-to-use, production-minded
[Bubble Tea v2](https://github.com/charmbracelet/bubbletea) components
("snaps"): navigation, tables, pickers, calendars, charts, and status
surfaces with first-class keyboard **and** mouse support.

Every snap is theme-free with injected style hooks, so it drops into any
Charm-stack app and adopts that app's look. Where a snap has multiple
implementations (navigation styles, scrollbar presets, pill shapes), it
exposes the choice through a small interface or preset list an app can
surface to its users at runtime.

## Components and gallery

Every demo below lives in `examples/` and is a VHS tape rendered in the
official vhs container. Regenerate them all with
`go -C tools/rendertapes run .` (Docker or Podman; the tool cross-compiles the
`snap_input` binary once, runs every `examples/*.tape` in parallel, and writes
the gifs to `dist/`). Add `-verbose` to stream the container's own output.

Gifs are build artifacts, not committed files, so the gallery below links to
copies hosted by Charm. `.github/workflows/demos.yml` keeps those links fresh:
it re-renders tapes on PRs that touch them, and on merge it publishes each gif
and opens a PR with the updated URLs. To do it by hand on a machine with
Docker or Podman, run `go -C tools/rendertapes run . -publish`.

Each hosted URL is recorded in `examples/<command>.gif.url` alongside the
sha256 of the bytes uploaded, so `-publish` only re-uploads gifs that actually
changed and `-relink` can repoint the docs with no network at all. Charm's
hosting rate-limits bursts and caps gif size, so uploads are batched with a
pause between them and retried with backoff; oversized gifs are reported
before anything is sent. Tune with `-publish-batch`, `-publish-pause`,
`-publish-attempts`, `-publish-backoff`, and `-max-gif-bytes`, or re-send an
unchanged gif with `-force-publish`.

The examples also work as script-friendly input tools: each one renders its
TUI on stderr and writes only the selected value to stdout, exiting 1 on
cancel.

```bash
date=$(go run ./examples/snap_input datepicker)   # -> 2026-07-12
when=$(go run ./examples/snap_input timepicker)   # -> 08:30:45
svc=$(go run ./examples/snap_input table)         # -> api
```

Run it with no command to tour every example in one program instead: tab and
shift+tab (or alt+←/→) change page, ctrl+b shows the page strip, ctrl+t cycles
the theme, and q ends the tour, printing each visited page's confirmed value.

Every example shows the same snap status bar with key bindings; pass
`--no-help` to hide it.

Every example is a subcommand of one prebuilt binary, shipped as a signed
per-OS/arch archive (`snap_snap_input_<os>_<arch>`) on the
[Releases page](https://github.com/jarvisfriends/snap/releases) — no Go
toolchain required to use it as a scripting tool. See
[`examples/USAGE.md`](examples/USAGE.md) for the full calling convention.

### Pickers

These allow user input and when exited will output what the user picked in stdout.

#### Date

Calendar date picker with click-to-confirm days, header month/year focus,
and keyboard/wheel paging.

![Date picker demo](https://vhs.charm.sh/vhs-3ZJmDUnpesbrtxhVN5gSm1.gif)

#### Time

`HH:MM(:SS)` time field with per-column dropdowns, type-ahead, and
validation.

![Time picker demo](https://vhs.charm.sh/vhs-44aMOFdpDYucWehiOOY9GU.gif)

#### Directory

Drive-aware directory picker and related path-editing interactions.

![Pickers demo](https://vhs.charm.sh/vhs-1W3jutKoXmndJvu8jGZb3l.gif)

### Table

Sortable/filterable table with row activation and keyboard/mouse support.

![Table demo](https://vhs.charm.sh/vhs-4Jl4q5CFtxbhNLpyWGLdQB.gif)

### Pills and breadcrumbs

Segmented pills, shape variants, and breadcrumb styling helpers from the
shared style contract.

![Pills demo](https://vhs.charm.sh/vhs-6DJjCDpQBE6MO3kWSFTv6.gif)

### Forms helpers

Parser-backed form validation for required fields, durations, ISO dates,
and list splitting.

![Forms demo](https://vhs.charm.sh/vhs-6f4Uim9rDinkKcxvgyVlsR.gif)

#### Context menu

Right-click menu with keyboard parity and terminal-edge clamping.

![Context menu demo](https://vhs.charm.sh/vhs-15e5MHRX2uZyxTQqsEcLKk.gif)

### Full Screen helpers

#### Navigation

Tabs, Sidebar, and MinimalTopNav behind one navigator contract.

![Navigation demo](https://vhs.charm.sh/vhs-4bDiplJUoDynmfj87ArnmV.gif)

#### Status and notifications

Status bar surfaces plus notification toast/history flows.

![Status demo](https://vhs.charm.sh/vhs-5DT59S12JGsaD68YC42Cqs.gif)

#### Scrollbar

Scrollbar presets with click/drag mapping through `OffsetAt`.

![Scrollbar demo](https://vhs.charm.sh/vhs-3ReBGHQwqCCQX98jDL7bcO.gif)

#### Dependencies modal

Build info and dependency reader rendered through the status info modal.

![Dependencies demo](https://vhs.charm.sh/vhs-3Q60UYLi5BRuWl0nasFnRz.gif)

#### Charts

Sparklines, horizontal bars, pie, and sankey charts rendered as
ID-routed tea models with stretch-to-fit sizing.

![Charts demo](https://vhs.charm.sh/vhs-1Aqwbp6nSo9Gabqskw2x1X.gif)

#### Line chart

Braille line chart showing rolling streams with compact terminal-cell
rendering.

![Line chart demo](https://vhs.charm.sh/vhs-2TOyaoNCEsnG2ACAIwouNM.gif)

#### Cell canvas

Whole-cell canvas and gradient helpers for animated truecolor effects.

![Cell canvas demo](https://vhs.charm.sh/vhs-4mNTJD2hRfxCLE5RxbHWSr.gif)

### Supporting packages (no standalone GIF)

- `gate/`: feature-gate registry with env overrides for settings-exposed flags.
- `geom/`: rect/point geometry helpers for hit-testing and layout math.
- `keys/`: rebindable common key map shared by snaps and apps.
- `layout/`: lipgloss frame arithmetic helpers.
- `logging/`: reserved placeholder.
- `osc/`: OSC 9;4 taskbar/tab progress integration.
- `page/`: shared page base for sizing and colors.
- `rendercheck/`: golden/layout/code-standard test helpers.
- `uifx/`: mouse handlers, named zones, and effect tiers.
- `winterm/`: Windows default-terminal detection and repair helpers.

The three navigation styles live side by side because they satisfy the same
navigator contract; an app can swap between them at runtime.

## Design rules

- **Theme-free with style hooks.** Components take injected styles (the
  datepicker/timepicker pattern) instead of importing an app theme, so any
  Bubble Tea app can adopt them. Hosts map their live theme onto the hooks.
- **Keyboard and mouse.** Every interactive element works keyboard-only,
  mouse-only, and mixed.
- **Settings-ready interfaces.** Where multiple implementations exist (e.g.
  navigation), a snap exposes an interface so an app can offer the choice to
  users at runtime.
- Dependencies stay down to `charm.land/{bubbletea,bubbles,lipgloss}/v2` plus
  small helpers that move with the component. One deliberate exception:
  `charts` plots braille through
  [ntcharts](https://github.com/NimbleMarkets/ntcharts) rather than
  duplicating its canvas — snap only keeps the chart shapes ntcharts lacks.
- Every component folder eventually gets a VHS `.tape` demo and its own README.

## Development

`bash tools/local_verify.sh` is the gate: gofmt, golangci-lint on
windows+linux, shellcheck, markdownlint, go vet, `go test -race`, and a
dependency review (module-level vulnerability scan plus OpenSSF Scorecards
on direct dependencies).

For color-audit passes, force a loud demo background at runtime with
`SNAP_DEMO_DEBUG_BG=#ff0066` before running an example or rendering tapes.
Any unthemed background holes become obvious immediately.

The test suite also runs `rendercheck.CheckCodeStandards` over the whole
module: display text is measured and padded in terminal cells, never bytes.
Concretely — no `len()` on display strings (use `lipgloss.Width`), no
printf width-padding of string verbs like `%-9s` (use
`lipgloss.PlaceHorizontal` / `Style.Width`), no `strings.Join(rows, "\n")`
(use `lipgloss.JoinVertical`), and no space-run gaps concatenated for
alignment (use `PlaceHorizontal` or a `Width`/padded style).

Consumers pin tagged releases; for cross-repo development against an
application, use a `go.work` file locally and keep `replace` directives out
of committed `go.mod` files.

## Input contract (mouse + keyboard)

Every visual snap splits input by concern:

- **`OnMouse` owns the pointer.** Clicks, wheel (all four directions), drag,
  and hover are handled in `View().OnMouse` (dispatched by
  `uifx.MouseHandlers` to the component's handler methods) — never in
  `Update`. Keeping the two paths separate isolates pointer logic from state
  transitions and leaves room to process them independently later.
- **`Update` owns keys and messages.** Component `Update`s contain no
  `tea.MouseMsg` cases; a host that feeds one raw mouse anyway hits dead
  code, not a second handler.
- **Hit zones are named layers, not hand-kept rectangles.** Components build
  `uifx.Zones` from the same `lipgloss.NewLayer(content).ID(name)` blocks the
  frame is composed of, and handlers ask `zones.Hit(x, y)` which zone the
  pointer landed in — powered by lipgloss v2's `Compositor.Hit`, so zones
  track layout changes and resolve overlap by z-order (the timepicker package
  is the reference; the datepicker's uniform grid and the pickers' list rows
  still use direct arithmetic where that is simpler).
- **Parents translate and call the child's `OnMouse`.** Bubble Tea v2 only
  invokes the _root_ view's `OnMouse` (absolute coordinates) and does **not**
  translate for children — a parent adjusts x/y itself and calls the child's
  `View().OnMouse`. Never forward mouse into a child's `Update` — the runtime
  hands the raw event to both the root `OnMouse` _and_ `Update`, so two doors
  means every click processed twice.

### Effect tiers (`uifx.Level`)

| Tier                    | Feedback                                                                       | Root mouse mode |
| ----------------------- | ------------------------------------------------------------------------------ | --------------- |
| `LevelMinimal`          | interactions only — no hover/drag cosmetics, minimal redraw churn (thin links) | `CellMotion`    |
| `LevelMedium` (default) | + wheel everywhere, drag tracking while a button is held                       | `CellMotion`    |
| `LevelHigh`             | + hover highlighting of the element under the pointer                          | `AllMotion`     |

Set a component's `Effects` field and give your root view
`Effects.MouseMode()`. Hover is a motion-event firehose — that is why it is
opt-in.

### Testing input without false failures

Input tests assert **semantic state** (the highlighted day, the focused
column, the cursor row) after events aimed at the component's **own recorded
hit zones** — never hardcoded screen coordinates and never styled output
(styles vary by color profile; where rendering must be checked, an injected
`Transform` marker keeps it profile-independent). That keeps every failure a
real behavior change.

## Verifying releases

Release archives are checksummed, ship SPDX SBOMs, and `checksums.txt` is signed with keyless cosign by the
release workflow. See [docs/release-verification.md](docs/release-verification.md) for the two-command
verification.
