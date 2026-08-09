# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the
project adheres to semantic versioning (breaking changes allowed before v1.0).

The full, authoritative release history — including every tagged version — is
on the [GitHub releases page](https://github.com/jarvisfriends/snap/releases).

## [Unreleased]

### Added

- Running `snap_input` with no command tours every example in one program:
  tab/shift+tab (or alt+←/→) change page, ctrl+b shows the page strip, ctrl+t
  cycles the theme, and q prints each visited page's confirmed value,
  namespaced by page. Single-command runs are unchanged — they confirm and
  exit, with un-namespaced keys.
- `styles.PickerStyles(*AppStyle)` and `styles.ApplyTimeFieldTheme(...)` map an
  app palette onto `pickers` and `timepicker`, replacing the fixed ANSI colors
  those components used to hardcode.
- `charts.HBarGradient` renders a horizontal bar as a two-color gradient;
  sparklines gained gradient rendering and stretch-to-fit for short series.

### Changed

- **Breaking (examples):** every example is now a subcommand of a single
  `examples/snap_input` binary rather than its own `main` package. Where you
  ran `go run ./examples/datepicker`, now run
  `go run ./examples/snap_input datepicker`. Releases ship one `snap_input`
  archive per OS/architecture in place of the previous fourteen per-example
  archives.
- Example selections can be formatted with `--output pretty|values|json|yaml|xml`.
- Demo GIFs are no longer committed. Tapes live at `examples/<command>.tape`
  and render to `dist/`. `tools/rendertapes -publish` renders each tape,
  uploads it to Charm hosting, records the URL in `examples/<name>.gif.url`,
  and repoints the markdown gallery; `.github/workflows/demos.yml` keeps that
  current. `-relink` replays recorded URLs with no Docker and no network, and
  `-verbose` streams the container's own output plus every command it runs.
- `keys.AppKeyMap.BindingDefs()` now returns only the app-global, mutually
  exclusive shortcuts intended for a key-rebinding settings UI. The
  component-level widget keys (Sort, Filter, Open, Cancel, Save, Delete,
  Submit, Open Detail) moved to the new `ComponentBindingDefs()` so a flat
  conflict check no longer reports their intentional contextual overlaps
  (e.g. `ctrl+c` on both Quit and Cancel) as errors.

## [0.2.3]

- Continuous-integration hardening: all GitHub Actions pinned to commit SHAs,
  an MIT `LICENSE`, an OpenSSF Scorecard workflow, CodeQL scanning,
  `govulncheck`, dependency review, a native Go fuzz target, and a Codecov
  coverage upload.

## [0.2.0]

- Table internals reworked; the table's key bindings are configured through
  `WithKeyMap`/`SetKeyMap` and the keymap field is no longer exported.

Earlier releases predate this changelog; see the GitHub releases page for their
notes.

[Unreleased]: https://github.com/jarvisfriends/snap/compare/v0.2.3...HEAD
[0.2.3]: https://github.com/jarvisfriends/snap/releases/tag/v0.2.3
[0.2.0]: https://github.com/jarvisfriends/snap/releases/tag/v0.2.0
