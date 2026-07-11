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
      # Full lint once with the native GOOS. Other platforms only re-lint
      # the packages whose sources actually diverge by OS (file suffixes or
      # //go:build tags) — everything else typechecks identically, so a
      # second full pass is pure duplication. Plain shell resolution, not
      # `env`: on Windows dev boxes a stray Linux ELF of the same name
      # earlier in PATH breaks `env`-style invocation.
      lint_goos() { GOOS="$1" GOARCH=amd64 golangci-lint run "${@:2}"; }
      native_goos="$(go env GOOS)"
      run_gate FAIL "golangci-lint (GOOS=${native_goos})" lint_goos "${native_goos}" ./...

      os_specific_pkgs() {
        {
          git ls-files '*_windows.go' '*_linux.go' '*_unix.go' '*_darwin.go' '*_other.go'
          git ls-files '*.go' | xargs grep -lE '^//go:build .*(windows|linux|darwin|unix)' 2>/dev/null
        } | xargs -r -n1 dirname | sort -u | sed 's|^|./|; s|$|/...|'
      }
      mapfile -t OS_PKGS < <(os_specific_pkgs)
      other_goos=linux
      [[ "${native_goos}" == linux ]] && other_goos=windows
      if [[ ${#OS_PKGS[@]} -gt 0 ]]; then
        run_gate FAIL "golangci-lint (GOOS=${other_goos}, OS-specific pkgs: ${OS_PKGS[*]})"           lint_goos "${other_goos}" "${OS_PKGS[@]}"
      fi
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

# ─── dependency review (CI parity: vulnerabilities + scorecards) ─────────────
# CI's dependency-review action scans every go.mod (nested tool modules
# included) for known vulnerabilities and OpenSSF Scorecards below the repo
# threshold on changed deps. Mirror it locally: module-level govulncheck per
# module (reachability-independent — vulnerable versions FAIL like CI), and a
# scorecard sweep over every module's direct deps (WARN). Kept in sync with
# tui-base tools/local_verify.sh.
SCORECARD_THRESHOLD="${SCORECARD_THRESHOLD:-3.0}"

# resolve_scorecard_repo maps a Go module path to the github.com/{owner}/{repo}
# slug the scorecard API understands: native GitHub paths and golang.org/x/*
# directly, other vanity hosts through their go-import meta tag (best effort).
resolve_scorecard_repo() {
  local mod="$1"
  case "$mod" in
    github.com/*)
      echo "$mod" | cut -d/ -f1-3
      ;;
    golang.org/x/*)
      echo "github.com/golang/$(echo "${mod#golang.org/x/}" | cut -d/ -f1)"
      ;;
    *)
      # Best effort: failures just produce an empty slug -> "skip" below.
      curl -fsS --max-time 10 "https://${mod}?go-get=1" 2>/dev/null \
        | sed -n 's|.*go-import[^>]*https://\(github\.com/[^" ]*\).*|\1|p' \
        | head -1 | sed 's|\.git$||' | cut -d/ -f1-3 || true
      ;;
  esac
  return 0
}

vulncheck_module() {
  # govulncheck loads the current-directory package even for module scans,
  # so run from the module's first package dir (module roots without .go
  # files — like snap's — error out otherwise).
  local dir="$1" pkgdir
  pkgdir=$( (cd "$dir" && go list -f '{{.Dir}}' ./... 2>/dev/null | head -1) || true)
  [[ -z "$pkgdir" ]] && pkgdir="$dir"
  (cd "$pkgdir" && govulncheck -scan module)
}

if command -v govulncheck >/dev/null 2>&1; then
  while IFS= read -r modfile; do
    moddir=$(dirname "$modfile")
    run_gate FAIL "govulncheck -scan module (${moddir})" vulncheck_module "$moddir"

    echo "==> OpenSSF Scorecards (${moddir} direct deps, threshold ${SCORECARD_THRESHOLD})"
    while IFS= read -r dep; do
      repo=$(resolve_scorecard_repo "$dep")
      if [[ -z "$repo" ]]; then
        echo "    skip  $dep (no GitHub mapping for scorecard)"
        continue
      fi
      # The aggregate "score" field precedes the per-check scores in the
      # response, so the first match is the overall scorecard.
      score=$(curl -fsS --max-time 10 "https://api.securityscorecards.dev/projects/${repo}" \
        2>/dev/null | grep -o '"score":-\?[0-9.]*' | head -1 | cut -d: -f2 || true)
      if [[ -z "$score" ]]; then
        echo "    skip  $dep (no scorecard data for $repo)"
        continue
      fi
      if awk -v s="$score" -v t="$SCORECARD_THRESHOLD" 'BEGIN{exit !(s<t)}'; then
        echo "    WARN  $dep scores $score (< $SCORECARD_THRESHOLD)"
        WARNINGS+=("scorecard: ${moddir} ${dep} scores ${score} (< ${SCORECARD_THRESHOLD}) — CI dependency review flags this")
      else
        echo "    ok    $dep ($score)"
      fi
    done < <(cd "$moddir" && awk '
      /^require \(/ {inreq=1; next}
      inreq && /^\)/ {inreq=0; next}
      inreq && !/\/\/ indirect/ && NF >= 2 {print $1}
      /^require [^(]/ && !/\/\/ indirect/ {print $2}
    ' go.mod)
  done < <(git ls-files | grep -E '(^|/)go\.mod$')
else
  warn_missing "govulncheck" "go install golang.org/x/vuln/cmd/govulncheck@latest"
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
