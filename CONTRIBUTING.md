# Contributing to snap

Thanks for helping improve snap. This document covers the practical bar for changes.

## Requirements

- Go **1.26.5 or newer**.
- `golangci-lint` v2, `shellcheck`, `actionlint`, and `markdownlint-cli2` for the full local gate.

## Workflow

1. Branch from `main`.
2. Make the change with tests. Bug fixes need regression coverage.
3. Run the full local gate before pushing:

   ```bash
   bash tools/local_verify.sh
   ```

4. Open a PR against `main`. CI must pass on Linux and Windows.

## Code conventions

- Charm v2 imports only.
- Keep every interactive surface keyboard- and mouse-complete.
- Prefer extending existing snaps or shared helpers over introducing near-duplicate components.
- Examples should stay script-friendly when practical: render UI to stderr and final selected values to stdout.