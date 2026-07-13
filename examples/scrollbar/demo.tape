# VHS demo for snap/scrollbar — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/scrollbar/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 850
Set TypingSpeed 60ms
Hide
Type "./examples/scrollbar/demo-bin"
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
