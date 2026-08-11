# Example architecture & component alignment notes

## The shared shell (examples/internal/exui)

Every example now follows one structure: it builds its component, hands
`exui.NewChrome(bindings…)` its key hints, and forwards messages through
`chrome.Update` before its own logic. The shell owns everything an app host
would own — the status bar, the notifications manager behind the bell
(ctrl+n or click: full history panel with ↑/↓, enter dismiss, D dismiss-all),
the info modal behind ⓘ (ctrl+e: version + full dependency list via
`dependencies.ExpandedBuildInfo`), and a ctrl+d debug overlay (build,
runtime, term size, tint, notification counts). `chrome.Frame` composites
open surfaces and chains shell mouse handling (modal first, bar icon
regions, then the example's own handler).

The contract: examples supply **data** (`chrome.Notify(...)`,
`chrome.Manager().AddWithOptions(...)`, `SetSegment(name, fn)`); the shell
supplies behavior. In tui-base the same data arrives from APIs/OS/async
messages — the display path is identical. Overlays allocate nothing while
closed; the bar re-renders only via `refresh()` after actual state changes.

## One page, one host (examples/snap_input/tour.go)

Each example package exports `New() exui.Page` — a `tea.Model` plus
`Result() []exui.Field` — and no longer owns a `Run()`, a quit key, or its own
output plumbing. One host drives both `snap_input <name>` (a single page) and
`snap_input` (all of them), so there is no second code path to drift.

Three pieces make that work:

1. **Quit is shell-owned.** `Chrome.Update` claims `q`/`ctrl+c`, so no
   component or click can quit on its own. Pages whose component eats
   keystrokes — the table filter, a huh form, an open context menu — register
   `chrome.SetCapture(pred)`, and while it reports true only `ctrl+c` quits.
   Without that guard, typing "q" into a filter box would end the program.
2. **Confirming means different things per host.** `exui.Confirm()` returns
   `tea.Quit` for a single page (keeping `date=$(snap_input datepicker)` one
   action) and nil in a tour, where the value is recorded and the page stays
   up. Results are namespaced (`datepicker.date`) only when more than one page
   is in play.
3. **Pages are recreated on entry, not kept alive.** Components snapshot
   styles at construction, so recreation is what makes `ctrl+t` theme cycling
   reach the page you are looking at. Because a model is dropped on exit, the
   host copies `Result()` out first, and pages that allocate (the pickers temp
   tree) implement `exui.Cleaner` so every instance is torn down.

The tab strip stacks a `navigation.Tabs` above the page and shifts pointer
coordinates back into page space — Bubble Tea hands the root view absolute
coordinates and never translates for children (see the README input contract).

Two things the original design (ROADMAP, 2026-08-01) called differently, and
why the shipped version deviates:

- **Page nav is `alt+←/→` as well as `tab`/`shift+tab`.** Tab alone could not
  be the contract: the forms page is a huh form that needs tab for its own
  fields. The tour yields tab whenever the page reports capturing, so the
  alt chords are the always-available escape hatch. `pills` lost its
  tab/shift+tab shape binding to free the chord — arrows and the wheel still
  cycle shapes.
- **Single-page runs confirm and exit.** Inverting quit uniformly would have
  made `date=$(snap_input datepicker)` a two-keystroke operation, regressing
  the scripting contract that README, USAGE.md, architecture.md, and
  getting-started.md all document.

`styles.TintIDs()` was not needed — `styles.BuiltinTintIDs()` already existed.
The `ctrl+t` ring is the demo default plus the seven built-ins, deliberately
not the full registry (bubbletint registers 349 tints; browsing those is a
theme picker's job, not a cycle key's).

## Decisions from the review

**pickers**: fixed ANSI colors (245/212/252/240) replaced through a new
`styles.PickerStyles(c)` adapter (same pattern as `styles.ScrollbarStyles`).
The same defect still exists in `menu.DefaultStyles`, `datepicker`,
`timepicker`, and `charts/piechart_model.go:115` — each needs the same
five-line adapter in `styles/` when touched next.

**table chromeRows**: replaced with `headerRows` (bubble-table's single
header line while its borders/footer are disabled) + a measured
`lipgloss.Height(footer)` — a wrapping footer or future border can no longer
desync the page-size math.

**table vs btable (evertras/bubble-table v0.22.3)**: keep it. snap/table
already delegates rendering/pagination/filter to btable and owns only what
btable lacks: numeric-aware 3-state sorting (btable's SortByAsc/Desc sorts
raw strings), all mouse behavior (btable has none), row-activation msg,
width fitting, themed footer. charm's bubbles/v2 table is far smaller
(keyboard viewport only — no sort/filter/mouse), so switching would mean
rewriting, not deleting. No third-party Go table currently offers
sort+filter+mouse+theming on bubbletea v2; revisit if btable ships v2-native
mouse support.

**menu vs navigation vs bubbles**: menu is the outlier — it carries its own
KeyMap (duplicating keys.AppKeyMap.Up/Down/Select/Dismiss) and its own
Styles with hard-coded colors, and reimplements the item-cursor loop that
Sidebar already gets from bubbles/list. Wheel→cursor math exists 4× (menu
inline, navigation's horizontal/verticalWheelDelta, per-navigator call
sites) and cyclic `(i±1+n)%n` advance 6×. Alignment path, in order of
value: (1) menu adopts keys.AppKeyMap + a styles.MenuStyles adapter;
(2) menu/navigators share navigation's wheelDelta helpers and geom.Rect
hit-testing (menu already uses geom.Rect — navigators hand-roll ranges);
(3) Tabs' 46-line overflow window could become bubbles/paginator, but only
the index math — variable-width tabs stay custom. Rebuilding menu on
bubbles/list is possible (Sidebar proves the shape) but list has no
disabled-item support, so it buys little; not recommended now.

## Cross-cutting conflicts found (queued for the next major rev)

1. **InfoModal ignores injected colors** — it reads `styles.Active()` directly
   while everything else honors `SetColors(*AppStyle)`/`page.Base`. Fix: make
   it `styles.ColorAware` like the bar.
2. **`status/user_notifications` advertises actions it doesn't implement** —
   the history footer promises Enter/dismiss-all but only cursor moves exist;
   the exui shell papers over it. Fix: move that handling into the status
   package and emit `notifications.ActivateMsg` (currently dead).
3. **`NotifHistoryCursorDown(n)` leaks state** — callers must pass the item
   count. Fix: read it from the attached manager.
4. **Toasts are documented but unrendered** — `Manager.Visible()` has no
   renderer anywhere. Fix: render toasts in status, or drop the docs claim.
5. **`styles/status.go` legacy renderers** duplicate `status.RenderStyled`.
   Fix: delete in the major rev.
6. **menu's private KeyMap/Styles** vs keys/styles everywhere else (see
   above). Fix: adapters + AppKeyMap.
7. **Per-repo demo pipelines diverged** — inspector rendered at release time
   (dirty-tree failure), snap committed gifs, tui-base committed gifs with a
   stale GIF_COVERAGE doc. Now unified: gifs are never committed; `rendertapes`
   renders them into `dist/` and CI attaches them to a release, so markdown
   links at `releases/latest/download/<name>.gif` stay fixed while each tag
   keeps its own copy.
8. **Version pin drift** — stringer was pinned at v0.47.0 while go.mod moved
   to x/tools v0.48.0; Renovate annotations now bump pins, but the pin should
   be derived from go.mod in CI to make drift impossible.
