# VHS demo for examples/status — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/status/demo.gif
Set FontSize 28
Set Width 1400
Set Height 800
Set TypingSpeed 60ms
Type "./examples/status/demo-bin"
Enter
Sleep 2s
# raise one toast of each severity
Type "i"
Sleep 1200ms
Type "w"
Sleep 1200ms
Type "e"
Sleep 1500ms
# a progress notification fills as the fake download runs
Type "p"
Sleep 3500ms
# the history panel keeps everything, newest first
Ctrl+n
Sleep 1500ms
Down
Sleep 700ms
Down
Sleep 700ms
Ctrl+n
Sleep 1s
Type "q"
Sleep 800ms
