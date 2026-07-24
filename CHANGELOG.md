# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the
project adheres to semantic versioning (breaking changes allowed before v1.0).

The full, authoritative release history — including every tagged version — is
on the [GitHub releases page](https://github.com/jarvisfriends/snap/releases).

## [Unreleased]

### Changed

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
