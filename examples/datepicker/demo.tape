# VHS demo for examples/datepicker — regenerate from the repo root with:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/datepicker/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 36
Set Padding 14
Set Width 1200
Set Height 860
Set TypingSpeed 60ms
Hide
Type "./examples/datepicker/demo-bin"
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
