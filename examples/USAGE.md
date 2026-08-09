# snap_input — snap's examples as one scriptable tool

`snap_input` bundles every snap example as a subcommand of one small binary.
It's a demo of each `snap` component — but also a scriptable input tool: run
`snap_input <command>`, the terminal UI takes over the screen, and the value
the user picks is written to **stdout** when they confirm. Nothing else
touches stdout, so any scripting language can call it and read the result
directly.

```bash
date=$(snap_input datepicker)    # -> 2026-07-12
when=$(snap_input timepicker)    # -> 08:30:45
svc=$(snap_input table)          # -> api
dir=$(snap_input pickers)        # -> /home/me/projects/snap
```

- The interactive UI renders on **stderr**, so it never pollutes captured
  stdout.
- Canceling (quit without choosing) prints **nothing** and exits **1** — check
  the exit code before trusting an empty result.
- Every command shows the same snap status bar with its key bindings at the
  bottom; pass `--no-help` to hide it if you only want the component itself.
- Mouse and keyboard both work identically in every command.
- `--output pretty|values|json|yaml|xml` selects the result format
  (pretty is the default; `values` is the old bare-lines behavior).
- `snap_input --version` prints the release tag this binary was built from —
  useful for confirming which build a downloaded archive contains.
- `snap_input help` lists all commands.

## Tour mode

Running `snap_input` with **no command** tours every example in one program
instead of printing usage:

| key | action |
| --- | --- |
| `tab` / `shift+tab` | next / previous page (a page that is capturing text, like `forms`, keeps tab for itself) |
| `alt+→` / `alt+←` | next / previous page, always available |
| `ctrl+b` | show or hide the page strip |
| `ctrl+t` | cycle the theme |
| `q` | end the tour and print what was confirmed |

Confirming on a page records its value and leaves the tour running, so one
pass can collect several. On exit each visited page's result is printed,
namespaced by page and in visit order:

```console
$ snap_input
datepicker.date     2026-07-12
table.service       api
```

Single-command runs are unchanged: they confirm and exit immediately, and
their keys are **not** namespaced, so `date=$(snap_input datepicker)` still
returns a bare date.

## All commands

One archive per OS/architecture contains the single `snap_input` binary and
this file.

| command | what it demos |
| --- | --- |
| `cellcanvas` | Whole-cell canvas and gradient helpers for animated truecolor effects. |
| `charts` | Sparklines, horizontal bars, pie, and sankey charts, stretch-to-fit sized. |
| `datepicker` | Calendar date picker: click-to-confirm days, month/year focus, keyboard/wheel paging. |
| `dependencies` | Build info and dependency reader (status info modal). |
| `forms` | Parser-backed form validation: required fields, durations, ISO dates, list splitting. |
| `linechart` | Braille line chart showing rolling streams, compact terminal-cell rendering. |
| `menu` | Right-click context menu with keyboard parity and terminal-edge clamping. |
| `navigation` | Tabs, Sidebar, and MinimalTopNav behind one navigator contract. |
| `pickers` | Drive-aware directory picker and path-editing interactions. |
| `pills` | Segmented pills, shape variants, and breadcrumb styling helpers. |
| `scrollbar` | Scrollbar presets with click/drag mapping through `OffsetAt`. |
| `status` | Status bar surfaces plus notification toast/history flows. |
| `table` | Sortable/filterable table with row activation, keyboard/mouse support. |
| `timepicker` | `HH:MM(:SS)` time field: per-column dropdowns, type-ahead, validation. |

Full source, the Go library these commands are built from, and rendered demo
GIFs live at <https://github.com/jarvisfriends/snap>. Each command's source is
`examples/<command>/`, and its VHS tape is `examples/<command>.tape`. Gifs are
build artifacts rather than committed files: rendering writes
`dist/<command>.gif` locally, and the published copy the docs link to is
recorded in `examples/<command>.gif.url`.
