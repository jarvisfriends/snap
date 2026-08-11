# VHS demo for snap/forms — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/forms.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input forms"
Enter
Sleep 2s
Show
# required field: error until something is typed
Sleep 800ms
Type "ship the release"
Sleep 1s
Tab
# duration: the error flips to a parsed value as the input completes
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
Sleep 1500ms
# submit: one form returns every value at once as YAML on stdout
Enter
Sleep 3s
Type "Ctrl+C"
