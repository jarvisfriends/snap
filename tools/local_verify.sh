#!/usr/bin/env bash
set -euo pipefail

# Local verification gate — mirrors what CI enforces. Sections skip themselves
# when they don't apply to this repo (no go.mod, no shell scripts, …) so the
# same script ships in every repo.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

# ─── Go checks ────────────────────────────────────────────────────────────────
GO_FILE_COUNT=$(git ls-files -co --exclude-standard '*.go' | wc -l)
if [[ -f go.mod && ${GO_FILE_COUNT} -gt 0 ]]; then
  echo "==> gofmt (check only)"
  mapfile -t GO_FILES < <(git ls-files '*.go')
  if [[ ${#GO_FILES[@]} -gt 0 ]]; then
    UNFORMATTED=$(gofmt -l "${GO_FILES[@]}" 2>/dev/null || true)
    if [[ -n "${UNFORMATTED}" ]]; then
      echo "ERROR: gofmt required for:"
      echo "${UNFORMATTED}"
      exit 1
    fi
  fi

  echo "==> golangci-lint"

  check_golangci_lint() {
    if ! command -v golangci-lint >/dev/null 2>&1; then
      echo "ERROR: golangci-lint not found. Install v2 with:"
      echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
      echo "Ensure \$GOBIN or \$GOPATH/bin is on your PATH."
      exit 1
    fi
    ver=$(golangci-lint --version 2>&1 || true)
    if [[ $ver =~ ([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
      major=${BASH_REMATCH[1]}
    elif [[ $ver =~ v([0-9]+) ]]; then
      major=${BASH_REMATCH[1]}
    else
      major=""
    fi
    if [[ "$major" == "1" ]]; then
      echo "ERROR: Detected golangci-lint v1: $ver"
      echo "Remove the old v1 installation and install v2 with:"
      echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
      echo "Then ensure \$GOBIN or \$GOPATH/bin is on your PATH."
      exit 1
    fi
  }

  check_golangci_lint
  for target_os in windows linux; do
    echo "==> golangci-lint (GOOS=${target_os})"
    GOOS="${target_os}" GOARCH=amd64 golangci-lint run ./...
  done
fi

# ─── shellcheck ───────────────────────────────────────────────────────────────
if command -v shellcheck >/dev/null 2>&1; then
  echo "==> shellcheck"
  mapfile -t SH_FILES < <(git ls-files '*.sh' '.githooks/*' 'tools/pre-commit' | sort -u)
  if [[ ${#SH_FILES[@]} -gt 0 ]]; then
    shellcheck "${SH_FILES[@]}"
  fi
else
  echo "WARN: shellcheck not found; skipping shell lint (CI still enforces this)."
  echo "  Windows: choco install shellcheck || scoop install shellcheck"
fi

# ─── markdownlint ─────────────────────────────────────────────────────────────
mapfile -t MD_FILES < <(git ls-files '*.md')
if [[ ${#MD_FILES[@]} -gt 0 ]]; then
  if command -v markdownlint-cli2 >/dev/null 2>&1; then
    echo "==> markdownlint-cli2"
    markdownlint-cli2 "${MD_FILES[@]}"
  elif command -v npx >/dev/null 2>&1; then
    echo "==> markdownlint-cli2 (via npx)"
    npx --yes markdownlint-cli2 "${MD_FILES[@]}"
  else
    echo "WARN: markdownlint-cli2 and npx not found; skipping markdown lint."
    echo "  npm install -g markdownlint-cli2"
  fi
fi

# ─── actionlint (GitHub workflows only) ──────────────────────────────────────
if [[ -d .github/workflows ]]; then
  if command -v actionlint >/dev/null 2>&1; then
    echo "==> actionlint"
    actionlint
  else
    echo "WARN: actionlint not found; skipping workflow lint (CI still enforces this)."
    echo "  go install github.com/rhysd/actionlint/cmd/actionlint@latest"
  fi
fi

# ─── Go module + build + tests ───────────────────────────────────────────────
if [[ -f go.mod && ${GO_FILE_COUNT} -gt 0 ]]; then
  echo "==> go mod verify"
  go mod verify

  echo "==> go mod tidy (drift check)"
  go mod tidy
  if ! git diff --exit-code -- go.mod go.sum; then
    echo "ERROR: go mod tidy changed files — commit the result before pushing"
    exit 1
  fi

  echo "==> go fix (drift check)"
  if ! go fix -diff ./...; then
    echo ""
    echo "ERROR: go fix found suggestions. Run: go fix ./..."
    echo "Then review the changes with 'git diff' and commit them."
    exit 1
  fi

  echo "==> go vet"
  go vet ./...

  echo "==> go test -race"
  go test -race ./...
fi

echo "Local verification passed."
