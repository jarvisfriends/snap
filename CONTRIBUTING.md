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

4. Sign off every commit (Developer Certificate of Origin —
   <https://developercertificate.org/>): `git commit --signoff`, or
   `git rebase --signoff main` to sign off a whole branch. CI checks this on
   every PR.
5. Open a PR against `main`. CI must pass on Linux and Windows.

## Code conventions

- Charm v2 imports only.
- Keep every interactive surface keyboard- and mouse-complete.
- Prefer extending existing snaps or shared helpers over introducing near-duplicate components.
- Examples should stay script-friendly when practical: render UI to stderr and final selected values to stdout.
## Testing policy

Tests are required, not optional:

- Every change that adds or modifies functionality must include tests exercising the new behavior.
- Every bug fix must include a regression test that fails without the fix.
- CI enforces a **90% statement-coverage gate** on the library packages (examples, `cmd/`, and tools are
  demo/wiring code and excluded); PRs that drop below it fail.

## Project policies

Roles, code review rules, and the access policy live in [GOVERNANCE.md](GOVERNANCE.md); security reporting
and vulnerability handling in [SECURITY.md](SECURITY.md); dependency rules in
[docs/dependency-policy.md](docs/dependency-policy.md).
