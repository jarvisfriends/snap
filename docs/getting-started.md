# Getting started with snap

## Install

```bash
go get github.com/jarvisfriends/snap
```

Requires Go 1.26+ and [Bubble Tea v2](https://github.com/charmbracelet/bubbletea)
(`charm.land/bubbletea/v2`) — snap components are `tea.Model`s that drop into
any Bubble Tea v2 program.

## Use a component

Every snap is theme-free: it exposes an injected style hook instead of
hardcoding colors, so it adopts whatever palette your app is already using.
A minimal wiring looks like:

```go
import (
    "github.com/jarvisfriends/snap/table"
    "github.com/jarvisfriends/snap/styles"
)

t := table.New(columns)
t.SetRows(rows)
t.SetColors(styles.Active()) // or your own *styles.AppStyle
```

From there `t` is a normal `tea.Model` — mount it in your `Update`/`View`
like any other Bubble Tea component. See the
[component gallery](../README.md#components-and-gallery) in the main README
for what each package provides (navigation, pickers, charts, forms
validation, status bar, and more).

## Try it without writing code

Every component is also a subcommand of `snap_input` — one small scriptable
tool, not just a demo. `go run` fetches and runs it directly, no clone
required:

```bash
go run github.com/jarvisfriends/snap/examples/snap_input@latest datepicker   # renders the picker, prints your choice
```

Pin a specific release instead of `@latest` by using its tag, e.g.
`@v0.2.8`. If you already have the repo cloned locally, drop the module path
and version instead: `go run ./examples/snap_input datepicker`.

You can also download a prebuilt, signed binary for your OS from the
[Releases page](https://github.com/jarvisfriends/snap/releases) — one archive
per OS/architecture holds the single `snap_input` binary. See
[`examples/USAGE.md`](../examples/USAGE.md) for the full calling convention
and a table of every subcommand.

## Where to go next

- [README.md](../README.md) — full component gallery with demo GIFs.
- [CONTRIBUTING.md](../CONTRIBUTING.md) — development setup, test/lint bar.
- [CHANGELOG.md](../CHANGELOG.md) — release history.
- [pkg.go.dev](https://pkg.go.dev/github.com/jarvisfriends/snap) — API reference.
