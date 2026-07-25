# Snap example tools

This binary is one of snap's example programs, distributed as a standalone
release asset. It's a demo of a `snap` component — but it's also a small,
scriptable input tool: run it, the terminal UI takes over the screen, and the
value the user picks is written to **stdout** when they confirm. Nothing else
touches stdout, so any scripting language can call the binary and read its
result directly.

```bash
date=$(./datepicker)      # -> 2026-07-12
when=$(./timepicker)      # -> 08:30:45
svc=$(./table)            # -> api
dir=$(./pickers)          # -> /home/me/projects/snap
```

- The interactive UI renders on **stderr**, so it never pollutes captured
  stdout.
- Canceling (quit without choosing) prints **nothing** and exits **1** — check
  the exit code before trusting an empty result.
- Every tool shows the same snap status bar with its key bindings at the
  bottom; pass `--no-help` to hide it if you only want the component itself.
- Mouse and keyboard both work identically in every tool.

## All example tools in this release

Each one ships as its own archive per OS/architecture (this file is included
in all of them), so you only download the tool you need.

| binary | what it demos |
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

Full source, the Go library these tools are built from, and rendered demo
GIFs live at <https://github.com/jarvisfriends/snap>.
