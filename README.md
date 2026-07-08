# jarvis-bubbles

Extended Bubble Tea components extracted from
[tui-base](https://github.com/jarvisfriends/tui-base) (ROADMAP items
X-1…X-4, shape decided in Q-10): one shared repo, organized into category
folders that grow over time. Each component ships with a VHS `.tape` demo
showing why you'd pick it over the alternatives.

## Layout

| Folder | Component | Source (tui-base) | Status |
|---|---|---|---|
| `navigation/tabs/` | Tab-bar navigator | `navigation` | placeholder |
| `navigation/sidebar/` | Sidebar navigator | `navigation` | placeholder |
| `navigation/minimal-top/` | Slim top-nav navigator | `navigation` | placeholder |
| `inspector/` | Runtime inspector overlay (`bubbleinspector`) | `pages/inspector` | placeholder |
| `logging/` | UI-bound logger with subscriber fan-out (`bubblelog`) | `logging` | placeholder |
| `status/` | Status bar with segments + notifications | `status` | placeholder |

The three navigation styles live side by side so an app can swap between
them; they share the folder because they satisfy the same navigator contract.

## Extraction contract

- A component moves here only when its tui-base deps are down to
  `bubbletea/v2` + `lipgloss/v2` (+ small shared helpers that move with it).
- The open question (tui-base ROADMAP Q-20) is dependency direction: whether
  tui-base imports these packages back (requires this repo to be public) or
  the extraction is a copy. No code moves until that's decided.
- Every component folder gets `tools/demo.tape` (VHS) and its own README.

## Development

`bash tools/local_verify.sh` is the gate (same as every other repo:
gofmt, golangci-lint on windows+linux, shellcheck, markdownlint, go vet,
`go test -race`).
