# VHS demo for examples/status — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/status/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 800
Set TypingSpeed 60ms
Hide
Type "./examples/status/demo-bin"
Enter
Sleep 2s
Show
# raise one toast of each severity
Type "iwe"
Sleep 1200ms
# a progress notification fills as the fake download runs
Type "p"
Sleep 2500ms
# the history panel keeps everything, newest first
Ctrl+n
Sleep 2500ms
Down
Sleep 700ms
Down
Sleep 1500ms
# dismiss everything, then close: back to the empty bar for a seamless loop
Type "d"
Sleep 1000ms
Ctrl+n
Sleep 1500ms
