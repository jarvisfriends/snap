#!/usr/bin/env bash
set -uo pipefail

# Local verification — mirrors what CI enforces, but never stops early: every
# gate runs, results are collected, and the summary at the end reports all of
# them at once (so one drift finding doesn't hide a test failure behind it).
#
# Gates come in two severities:
#   FAIL — real breakage (build, tests, vet, lint). Non-zero exit.
#   WARN — drift that CI enforces but that shouldn't block local iteration
#          (gofmt/tidy/go fix drift, missing optional tools). Zero exit.
# CI runs the same checks individually and fails hard on all of them.

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}" || exit 1

FAILURES=()
WARNINGS=()

# run_gate <severity> <label> <command...> — runs the command, records the
# outcome, always returns success so the script continues.
run_gate() {
  local severity="$1" label="$2"
  shift 2
  echo "==> ${label}"
  if "$@"; then
    return 0
  fi
  if [[ "${severity}" == FAIL ]]; then
    FAILURES+=("${label}")
  else
    WARNINGS+=("${label}")
  fi
  return 0
}

warn_missing() {
  WARNINGS+=("$1 not installed — skipped locally (CI still enforces it): $2")
}

# ─── Go checks ────────────────────────────────────────────────────────────────
GO_FILE_COUNT=$(git ls-files -co --exclude-standard '*.go' | wc -l)
if [[ -f go.mod && ${GO_FILE_COUNT} -gt 0 ]]; then
  gofmt_check() {
    local unformatted
    mapfile -t go_files < <(git ls-files '*.go')
    [[ ${#go_files[@]} -eq 0 ]] && return 0
    unformatted=$(gofmt -l "${go_files[@]}" 2>/dev/null || true)
    if [[ -n "${unformatted}" ]]; then
      echo "gofmt required for:"
      echo "${unformatted}"
      return 1
    fi
  }
  run_gate WARN "gofmt (drift check)" gofmt_check

  if command -v golangci-lint >/dev/null 2>&1; then
    ver=$(golangci-lint --version 2>&1 || true)
    if [[ $ver =~ ([0-9]+)\.[0-9]+\.[0-9]+ && ${BASH_REMATCH[1]} == 1 ]]; then
      FAILURES+=("golangci-lint v1 detected — install v2: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest")
    else
      # Plain shell resolution, not `env`: on Windows dev boxes a stray Linux
      # ELF of the same name earlier in PATH breaks `env`-style invocation.
      lint_goos() { GOOS="$1" GOARCH=amd64 golangci-lint run ./...; }
      for target_os in windows linux; do
        run_gate FAIL "golangci-lint (GOOS=${target_os})" lint_goos "${target_os}"
      done
      if [[ -f tools/rendertapes/go.mod ]]; then
        lint_rendertapes() { (cd tools/rendertapes && golangci-lint run ./...); }
        run_gate FAIL "golangci-lint (tools/rendertapes)" lint_rendertapes
      fi
    fi
  else
    warn_missing golangci-lint "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
  fi
fi

# ─── shellcheck ───────────────────────────────────────────────────────────────
if command -v shellcheck >/dev/null 2>&1; then
  mapfile -t SH_FILES < <(git ls-files '*.sh' '.githooks/*' 'tools/pre-commit' | sort -u)
  if [[ ${#SH_FILES[@]} -gt 0 ]]; then
    run_gate FAIL "shellcheck" shellcheck "${SH_FILES[@]}"
  fi
else
  warn_missing shellcheck "choco install shellcheck || scoop install shellcheck"
fi

# ─── markdownlint ─────────────────────────────────────────────────────────────
mapfile -t MD_FILES < <(git ls-files '*.md')
if [[ ${#MD_FILES[@]} -gt 0 ]]; then
  if command -v markdownlint-cli2 >/dev/null 2>&1; then
    run_gate FAIL "markdownlint-cli2" markdownlint-cli2 "${MD_FILES[@]}"
  elif command -v npx >/dev/null 2>&1; then
    run_gate FAIL "markdownlint-cli2 (via npx)" npx --yes markdownlint-cli2 "${MD_FILES[@]}"
  else
    warn_missing markdownlint-cli2 "npm install -g markdownlint-cli2"
  fi
fi

# ─── actionlint (GitHub workflows only) ──────────────────────────────────────
if [[ -d .github/workflows ]]; then
  if command -v actionlint >/dev/null 2>&1; then
    run_gate FAIL "actionlint" actionlint
  else
    warn_missing actionlint "go install github.com/rhysd/actionlint/cmd/actionlint@latest"
  fi
fi

# ─── Go module + build + tests ───────────────────────────────────────────────
if [[ -f go.mod && ${GO_FILE_COUNT} -gt 0 ]]; then
  run_gate FAIL "go mod verify" go mod verify

  tidy_check() {
    go mod tidy
    if ! git diff --exit-code -- go.mod go.sum; then
      echo "go mod tidy changed files — commit the result before pushing"
      return 1
    fi
  }
  run_gate WARN "go mod tidy (drift check)" tidy_check

  gofix_check() {
    if ! go fix -diff ./...; then
      echo "go fix found suggestions. Run: go fix ./..., review, and commit."
      return 1
    fi
  }
  run_gate WARN "go fix (drift check)" gofix_check

  run_gate FAIL "go vet" go vet ./...
  run_gate FAIL "go build" go build ./...
  if [[ -f tools/rendertapes/go.mod ]]; then
    run_gate FAIL "go build (tools/rendertapes)" go -C tools/rendertapes build ./...
  fi
  run_gate FAIL "go test -race" go test -race ./...
fi

# ─── Summary ─────────────────────────────────────────────────────────────────
echo
if [[ ${#WARNINGS[@]} -gt 0 ]]; then
  echo "Warnings (CI enforces these; they do not fail local verification):"
  printf '  WARN: %s\n' "${WARNINGS[@]}"
fi
if [[ ${#FAILURES[@]} -gt 0 ]]; then
  echo "Failures:"
  printf '  FAIL: %s\n' "${FAILURES[@]}"
  echo "Local verification FAILED (${#FAILURES[@]} gate(s))."
  exit 1
fi
echo "Local verification passed."
