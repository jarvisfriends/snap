# VHS demo for snap/scrollbar — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/scrollbar.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input scrollbar"
Enter
Sleep 2s
Show
# glide down in sub-cell steps: watch the smooth bar move
ScrollDown 1
Sleep 350ms
ScrollDown 1
Sleep 350ms
ScrollDown 1
Sleep 350ms
ScrollDown 1
Sleep 350ms
ScrollDown 1
Sleep 350ms
PageDown
Sleep 900ms
PageDown
Sleep 900ms
# climb all the way back to the top so the loop is seamless
PageUp
Sleep 700ms
PageUp
Sleep 700ms
PageUp
Sleep 1400ms
Type "Ctrl+C"
