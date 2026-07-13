# VHS demo for snap/styles pills — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/pills/demo.gif
# The bundled Nerd Font renders the Powerline shapes (Round/Arrow/Slant)
# so the full nine-shape cycle shows in the gif.
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 720
Set TypingSpeed 60ms
Hide
Type "./examples/pills/demo-bin"
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
