# Snap

![snap — ready-to-snap Bubble Tea components](assets/banner.svg)

**Jarvis Friends Snap** — ready-to-use, production-minded
[Bubble Tea v2](https://github.com/charmbracelet/bubbletea) components
("snaps"): navigation, tables, pickers, calendars, charts, and status
surfaces with first-class keyboard **and** mouse support.

Every snap is theme-free with injected style hooks, so it drops into any
Charm-stack app and adopts that app's look. Where a snap has multiple
implementations (navigation styles, scrollbar presets, pill shapes), it
exposes the choice through a small interface or preset list an app can
surface to its users at runtime.
[tui-base](https://github.com/jarvisfriends/tui-base) — an application
framework built on these components — is one such consumer; the sibling
[inspector](https://github.com/jarvisfriends/inspector) is another.

## Components

| Folder | What it is |
|---|---|
| `charts/` | Sparklines (block + directional braille), horizontal bars, pie/sankey, and braille line charts (plotted via [ntcharts](https://github.com/NimbleMarkets/ntcharts) primitives) — each also wrapped as a tea model with ID-routed data messages and stretch-to-fill sizing — plus a whole-cell `CellCanvas` with color `Gradient`s. For axes, bar charts, heatmaps, or candlesticks, use ntcharts directly |
| `datepicker/` | Calendar date picker: click-to-confirm days, header month/year focus, keyboard/wheel paging |
| `dependencies/` | Build-info / dependency reader for about views |
| `forms/` | Input parsing helpers for text fields (required, duration, ISO date, list splitting) with field-naming errors |
| `gate/` | Feature-gate registry with env overrides, for settings-surfaced flags |
| `geom/` | Rect/point geometry helpers for hit-testing and layout |
| `keys/` | Common key-binding map shared by snaps and apps, rebindable at runtime |
| `layout/` | Lipgloss-frame arithmetic: content origin, inner size, render-in-box |
| `logging/` | Reserved — not yet implemented |
| `menu/` | Right-click context menu (mouse + keyboard, terminal-clamped) |
| `navigation/` | Tabs, Sidebar, and minimal-top navigators behind one navigator contract, swappable at runtime |
| `notifications/` | Notification manager: severity, TTL, actions, progress, persistence |
| `osc/` | Taskbar/tab progress via OSC 9;4 (Windows Terminal, ConEmu, iTerm2) |
| `page/` | Shared page base (sizing + colors) for full-page components |
| `pickers/` | Drive-aware directory picker and multi-path editor with per-row pickers |
| `rendercheck/` | Test helpers: goldens, border/viewport integrity, layout-math and code-standard checks |
| `scrollbar/` | Vertical scrollbar with three presets, offset clamping, and click/drag-to-scroll mapping (`OffsetAt`) |
| `status/` | Status bar with interactive regions, info modal, notification toast/history surfaces |
| `styles/` | The shared style contract: semantic `AppStyle` palette, derived lipgloss styles, presets, YAML themes, and the pill/breadcrumb helpers |
| `table/` | Sortable, filterable data table (3-state header sort, live filter, row activation) |
| `timepicker/` | `HH:MM(:SS)` time field with per-column dropdowns, type-ahead, and validation |
| `uifx/` | Input plumbing: `MouseHandlers` dispatch, named hit `Zones`, effect tiers |
| `winterm/` | Windows default-terminal detection/repair (registry delegation values) |

The three navigation styles live side by side because they satisfy the same
navigator contract; an app can swap between them at runtime.

## Gallery

Every demo below lives in `examples/` and is a VHS tape rendered in the
official vhs container — regenerate them all with
`go -C tools/rendertapes run .` (Docker or Podman; the tool cross-compiles
each example, runs every `*.tape` in parallel, and drops the gifs next to
their tapes).

The examples double as script-friendly input tools: each one renders its
TUI on stderr and writes **only the value the user chose** to stdout — no
labels, no prefixes — exiting 1 when cancelled, so a shell can use them
directly:

```bash
date=$(go run ./examples/datepicker)   # → 2026-07-12
when=$(go run ./examples/timepicker)   # → 08:30:45
svc=$(go run ./examples/table)         # → api
```

Every example shows the same snap status bar with its key bindings; pass
`--no-help` to hide it (e.g. when embedding an example in a script).

### Date picker

![datepicker demo](examples/datepicker/demo.gif)

Calendar with click-to-highlight / click-again-to-confirm days, header
month/year focus, and paging: PgUp/PgDn months, Shift+PgUp/PgDn years, the
wheel over the title pages the unit under the pointer.

### Time picker

![timepicker demo](examples/timepicker/demo.gif)

Two (or three, with `ShowSeconds`) colon-separated columns editing a
`time.Time`'s clock: digits type ahead, Space/click opens a value dropdown,
the wheel spins the focused column and hops columns horizontally.

### Charts

![charts demo](examples/charts/demo.gif)

The chart models live-streaming ID-routed data: two sparklines, a braille
pie (thin slices fold into "Other" with a legend), a sankey, and an hbar,
all stretching into the space the window split gives them.

### Line chart

![linechart demo](examples/linechart/demo.gif)

The braille line chart model streaming two rolling series: 2x4 dots per
terminal cell, per-cell color blending where series overlap, stretch-to-fill
sizing.

### Cell canvas

![cellcanvas demo](examples/cellcanvas/demo.gif)

`charts.CellCanvas` + `charts.Gradient`: an animated truecolor plasma using
`▀` half-blocks (two pixels per cell — foreground top, background bottom)
with batched escapes that re-emit colors only when they change.

### Pickers

![pickers demo](examples/pickers/demo.gif)

Drive-aware directory picker: keyboard and wheel walk the tree (wheel left
= parent, right = open), Space selects, Ctrl+S picks the browsed folder.

### Context menu

![menu demo](examples/menu/demo.gif)

Right-click (or keyboard) pop-up menu at the pointer: disabled items are
skipped, hover and wheel move the cursor, clicking outside dismisses,
edges clamp to the terminal.

### Scrollbar

![scrollbar demo](examples/scrollbar/demo.gif)

The three presets over one scrolling pane — Smooth (sub-cell eighth-block
glide), Line (thin default), Classic (retro blocks). Clicking or dragging on
a bar jumps the view there (`scrollbar.OffsetAt`).

### Table

![table demo](examples/table/demo.gif)

Sortable, filterable data table: header clicks (or `s`) cycle the 3-state
column sort, `/` filters live, Enter or double-click opens a row, wheel
scrolls the selection.

### Dependencies / info modal

![dependencies demo](examples/dependencies/demo.gif)

The `dependencies` build-info reader rendered by `status.InfoModal`: Go
version, OS, VCS revision, and a scrollable dependency list. The wheel
scrolls and clicks-outside dismiss through the modal's own `HandleMouse` —
hosts forward events, no hit-testing.

### Forms

![forms demo](examples/forms/demo.gif)

`snap/forms`' parse-and-validate helpers driving a live form: field-naming
errors appear as you type and flip to parsed values (`ParseRequired`,
`ParseDuration`, `ParseISODate`), and `SplitAndClean` turns a messy
comma-list into clean tags.

### Pills

![pills demo](examples/pills/demo.gif)

`styles.Pill` badges and color-divided `styles.SegmentedPill` runs in ten
selectable `PillShape`s, plus the same shapes driving a nav strip and
`styles.Breadcrumbs`. Six shapes — Circle, Triangle, Diagonal, Fade, Block,
Plain — are pure Unicode and render everywhere (they're what the gif shows);
Round, Arrow, Slant, and Flame use Powerline glyphs for terminals with a
[Nerd Font](https://www.nerdfonts.com/).

### Navigation

![navigation demo](examples/navigation/demo.gif)

All three navigator styles — Sidebar, Tabs, MinimalTopNav — behind the one
`Navigator` contract: `n` swaps styles at runtime (the active page carries
over), arrows/clicks/wheel move between pages, Enter prints the picked
page's ID.

### Status bar + notifications

![status demo](examples/status/demo.gif)

The status bar every example mounts, shown in full: page key help on the
left, live segments and a summary on the right, notification toasts by
severity, a progress notification filling as a fake download runs, and the
ctrl+n history panel.

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
  invokes the *root* view's `OnMouse` (absolute coordinates) and does **not**
  translate for children — a parent adjusts x/y itself and calls the child's
  `View().OnMouse`. Never forward mouse into a child's `Update` — the runtime
  hands the raw event to both the root `OnMouse` *and* `Update`, so two doors
  means every click processed twice.

### Effect tiers (`uifx.Level`)

| Tier | Feedback | Root mouse mode |
|---|---|---|
| `LevelMinimal` | interactions only — no hover/drag cosmetics, minimal redraw churn (thin links) | `CellMotion` |
| `LevelMedium` (default) | + wheel everywhere, drag tracking while a button is held | `CellMotion` |
| `LevelHigh` | + hover highlighting of the element under the pointer | `AllMotion` |

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
