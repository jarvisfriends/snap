# TUI Codebase Reduction Analysis

> **Superseded in part.** This was an early external survey. Where it
> conflicts with `docs/examples-architecture.md`, that document wins — it
> records the decisions actually taken after reviewing the code. Notably:
> `table` is **kept** (snap/table already delegates to evertras/bubble-table
> and owns the sorting, mouse, and theming btable lacks), and rebuilding
> `menu` on `bubbles/list` is **not recommended** (list has no disabled-item
> support). Treat the sections below as options considered, not a plan.

Based on an analysis of your core TUI libraries (`tui-base`, `snap`) and your expanding ecosystem of applications (`accel`, `brick-breaker`, `dash`, `inspector`, `multi`, `network-vis`, `snake`), here is a breakdown of what you can safely replace with standard open-source libraries to reduce your maintenance burden, and what you should keep as your own intellectual property.

## 1. Components to Replace with Open Source Alternatives

Many of the components you are currently maintaining in your `snap` and `tui-base` repos have mature, widely used equivalents in the Go ecosystem. Adopting these will significantly reduce the code you need to maintain across all your applications.

### From `jarvisfriends/snap` (UI Components)

*   **`table`**
    *   **Replacement**: [charmbracelet/bubbles/table](https://github.com/charmbracelet/bubbles/tree/master/table) or [evertras/bubble-table](https://github.com/Evertras/bubble-table).
    *   **Why**: The official Charm `bubbles/table` covers 90% of use cases. If you need advanced features like sorting, filtering, and pagination, `evertras/bubble-table` is the community standard and is heavily maintained.
*   **`forms`**
    *   **Replacement**: [charmbracelet/huh](https://github.com/charmbracelet/huh).
    *   **Why**: You already have `huh` in your `go.mod`. Any custom form logic should be fully migrated to `huh`, which is now the official and most robust way to handle forms in Bubble Tea.
*   **`menu` / `pickers`**
    *   **Replacement**: [charmbracelet/bubbles/list](https://github.com/charmbracelet/bubbles/tree/master/list) or `huh.Select`.
    *   **Why**: The standard `list` bubble handles filtering, pagination, and keyboard navigation out of the box.
*   **`charts`**
    *   **Replacement**: [NimbleMarkets/ntcharts](https://github.com/NimbleMarkets/ntcharts) or [guptarohit/asciigraph](https://github.com/guptarohit/asciigraph).
    *   **Why**: Charting in the terminal involves a lot of complex math for rendering braille characters. `ntcharts` integrates directly with Bubble Tea and handles time-series, bar charts, and sparklines.
*   **`scrollbar`**
    *   **Replacement**: [charmbracelet/bubbles/viewport](https://github.com/charmbracelet/bubbles/tree/master/viewport) and `bubbles/paginator`.
    *   **Why**: Rather than maintaining a custom scrollbar math implementation, standardizing around Charm's `viewport` handles overflowing content and scrolling natively.

### From `jarvisfriends/tui-base` (Core Infrastructure)

*   **`logging`**
    *   **Replacement**: The Go 1.21+ standard library [`log/slog`](https://pkg.go.dev/log/slog) combined with [`gopkg.in/natefinch/lumberjack.v2`](https://github.com/natefinch/lumberjack) for file rotation.
    *   **Why**: Your current `logging.go` is complex. You can replace this entirely by configuring `slog` to write to a `lumberjack.Logger` for disk rotation. For the UI subscriber, write a tiny custom `slog.Handler` that wraps the JSON handler and sends a `tea.Msg` to your application whenever a log above a certain level is recorded.

## 2. Components to Keep & Maintain

Some of your code is highly specific to how your applications are structured and doesn't have a 1-to-1 open-source replacement.

*   **`router` and `pages` (`tui-base`)**
    *   **Why to Keep**: There is **no official router** for Bubble Tea. Your implementation provides a structured way to register pages and swap them out across all your complex apps (like `dash` or `inspector`). Replacing this would require rewriting your entire architecture.
*   **`config` and `envpath` (`tui-base`)**
    *   **Why to Keep**: This package brilliantly bridges forms with live-bound pointers to dynamically generate settings pages. Standard config libraries like `viper` only handle reading/writing to disk; they do not generate Bubble Tea UIs.
*   **`theme` and `overlay` (`tui-base`)**
    *   **Why to Keep**: Managing consistent color palettes and modal overlays (like dialogs) globally across your apps is highly specific to your brand and user experience. Charm's `lipgloss` provides the styling, but `theme` provides your design system.
*   **`filewatch` (`tui-base`)**
    *   **Why to Keep (or simplify)**: You use a wrapper around `fsnotify` to emit `tea.Cmd`. Your `filewatch` package handles debouncing, which is tricky to get right. 
*   **`datepicker` and `timepicker` (`snap`)**
    *   **Why to Keep**: The open-source Bubble Tea ecosystem currently lacks a robust, standardized date/time picker (it is not yet natively supported in `huh`). Keep maintaining these until a mature open-source alternative emerges.

## Next Steps & Recommendation

To safely reduce your maintenance burden without breaking features across your app ecosystem:

1.  **Phase 1**: Delete your custom `table`, `menu`, and `forms` implementations in the `snap` repo and refactor your apps (`dash`, `inspector`, etc.) to use `bubbles/table`, `bubbles/list`, and `huh` exclusively.
2.  **Phase 2**: Rip out `tui-base/logging` and replace it with standard `log/slog` + `lumberjack`.
3.  **Phase 3**: Focus your internal development efforts exclusively on your unique intellectual property: `router`, `config`, `theme`, `datepicker`, and `timepicker`.

Let me know if you would like me to draft the implementation plan for Phase 1 or Phase 2, and we can begin systematically replacing these components!
