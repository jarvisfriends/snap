# Architecture

Snap is a Bubble Tea v2 component library ("snaps") for building terminal UIs. This document describes the system at the
level the
OpenSSF Baseline asks for: the actors involved, the actions they can take, and every external interface of
the released software.

## Actors

- **Host application** — a Go program that imports snap packages and composes them into a Bubble Tea model.
- **Terminal user** — interacts through the keyboard and mouse; input reaches components as Bubble Tea messages.
- **Terminal emulator** — receives rendered frames (ANSI) and OSC sequences from the host application.

## Actions and data flow

The host application drives a Bubble Tea event loop: terminal input arrives as messages, the model updates,
and a new frame is rendered to the terminal. This library sits inside that loop — it does not spawn its own
event sources beyond those documented below, and it holds no global mutable state that outlives the model.

## External interfaces

- The public Go API of each component package (`charts`, `datepicker`, `forms`, `menu`, `navigation`,
  `pickers`, `scrollbar`, `status`, `styles`, `table`, `timepicker`, ...). See the
  [Go reference](https://pkg.go.dev/github.com/jarvisfriends/snap).
- Released example binaries: read the terminal, render UI to **stderr**, print the selected value to
  **stdout**, and exit 1 on cancel (see `examples/USAGE.md`). They accept no network input.
- No network listeners, no configuration files, no environment-variable contracts beyond standard
  terminal detection (`TERM`, `COLORTERM`, ...).

## Security-relevant surfaces

- **Rendered content.** Snaps render strings supplied by the host application. Content that originates from
  untrusted sources (file names, network data) could carry ANSI escape sequences; components sanitize or
  width-measure input where they render it, and the threat model below tracks this surface.
- **Filesystem access.** The picker components (`pickers/`) list directories and drives read-only on behalf of
  the host application. They never write, delete, or follow symlinks outside the paths the host passes in.
- **Terminal control.** `osc/` emits OSC 52 clipboard and related sequences only when the host application
  explicitly invokes it.

See [threat-model.md](threat-model.md) for the corresponding threat analysis and mitigations.
