# VHS demo for snap/forms — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/forms/demo.gif
Set FontSize 28
Set Width 1400
Set Height 850
Set TypingSpeed 100ms
Type "./examples/forms/demo-bin"
Enter
Sleep 2s
# required field: error until something is typed
Sleep 800ms
Type "ship the release"
Sleep 1s
Tab
# duration: watch the error flip to a parsed value as the input completes
Type "7h"
Sleep 900ms
Type "30m"
Sleep 1200ms
Tab
# ISO date: partial input shows the friendly format hint
Type "2026-07"
Sleep 1s
Type "-14"
Sleep 1200ms
Tab
# tags: SplitAndClean trims the mess into chips
Type "go,  tui ,, release"
Sleep 2s
Escape
Sleep 800ms
