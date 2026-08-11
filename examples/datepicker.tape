# VHS demo for examples/datepicker — regenerate from the repo root with:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/datepicker.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input datepicker"
Enter
Sleep 3s
Show
# keyboard: move the highlight
Right
Sleep 500ms
Right
Sleep 500ms
Down
Sleep 1s
# page months and years: ]/[ months, }/{ years
Type "]"
Sleep 800ms
Type "]"
Sleep 800ms
Type "}"
Sleep 1s
Type "{"
Sleep 800ms
Type "["
Sleep 800ms
Type "["
Sleep 1s
# mouse: the wheel pages weeks
ScrollDown 1
Sleep 800ms
ScrollDown 1
Sleep 800ms
ScrollUp 1
Sleep 800ms
# enter confirms: the ISO date is written to stdout
Enter
Sleep 2500ms
Type "Ctrl+C"
