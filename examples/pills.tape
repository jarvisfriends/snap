# VHS demo for snap/styles pills — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/pills.gif
# The bundled Nerd Font renders the Powerline shapes (Round/Arrow/Slant)
# so the full nine-shape cycle shows in the gif.
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input pills"
Enter
Sleep 2s
Show
# walk every shape: the Nerd Font powerline set then the pure-Unicode set
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
# the wheel cycles too
ScrollUp 1
Sleep 800ms
ScrollDown 1
Sleep 800ms
# wrap back to the first shape so the loop is seamless
Right
Sleep 1600ms
Type "Ctrl+C"
