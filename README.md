# Snap

**Jarvis Friends Snap** — ready-to-use, production-minded Bubble Tea v2
components ("snaps") extracted from
[tui-base](https://github.com/jarvisfriends/tui-base): navigation, tables,
pickers, and calendar items with first-class keyboard **and** mouse support.

Snap is the single source of truth for these components: tui-base imports them
back rather than redefining them (tui-base ROADMAP Q-20, answered 2026-07-09).
The sibling [inspector](https://github.com/jarvisfriends/inspector) repo holds
the runtime debugger for any Charm-based app.

## Layout

| Folder | Component | Source (tui-base) | Status |
|---|---|---|---|
| `keys/` | Common key-binding map shared by snaps and apps | `keys` | **moved 2026-07-09** |
| `geom/` | Rect/point geometry helpers for hit-testing and layout | `geom` | **moved 2026-07-09** |
| `datepicker/` | Calendar date picker (formerly `bubble-datepicker`) | `datepicker` | **moved 2026-07-09** |
| `timepicker/` | Time picker | `timepicker` | **moved 2026-07-09** — UX redesign tracked in tui-base ROADMAP SP-8 |
| `dependencies/` | Build-info / dependency reader for about views | `common/dependencies.go` | **moved 2026-07-09** |
| `table/` | Data table with selection + scrolling | `table` | pending style-hook decoupling (SP-6) |
| `navigation/tabs/` | Tab-bar navigator | `navigation` | placeholder (SP-5) |
| `navigation/sidebar/` | Sidebar navigator | `navigation` | placeholder (SP-5) |
| `navigation/minimal-top/` | Slim top-nav navigator | `navigation` | placeholder (SP-5) |
| `pickers/` | Drive-aware DirPicker + MultiFileEditor (multi-path rows, per-row pickers) | `pages/settings` | **moved 2026-07-09** — style hooks + huh-theme/collapse-path injection |
| `logging/` | UI-bound logger with subscriber fan-out | `logging` | placeholder — shape depends on the zap decision (tui-base SP-10) |
| `status/` | Status bar with segments + notifications | `status` | placeholder (X-4) |

The three navigation styles live side by side because they satisfy the same
navigator contract; an app can swap between them at runtime.

## Design rules

- **Theme-free with style hooks.** Components take injected styles (the
  datepicker/timepicker pattern) instead of importing an app theme, so any
  Bubble Tea app can adopt them. tui-base maps its live theme onto the hooks.
- **Keyboard and mouse.** Every interactive element works keyboard-only,
  mouse-only, and mixed.
- **Settings-ready interfaces.** Where multiple implementations exist (e.g.
  navigation), a snap exposes an interface so an app can offer the choice to
  users at runtime (tui-base surfaces this in its settings page).
- Dependencies stay down to `charm.land/{bubbletea,bubbles,lipgloss}/v2` plus
  small helpers that move with the component.
- Every component folder eventually gets a VHS `.tape` demo and its own README.

## Development

`bash tools/local_verify.sh` is the gate (same as every other repo: gofmt,
golangci-lint on windows+linux, shellcheck, markdownlint, go vet,
`go test -race`).

Cross-repo development against tui-base uses a `go.work` file (see tui-base's
go.work recipe in `docs/migration-from-bubbletea.md`); tui-base's `go.mod` only
ever references tagged snap releases.
