# VHS demo for examples/status — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/status.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input status"
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
Type "Ctrl+C"
