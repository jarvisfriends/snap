# VHS demo for snap/scrollbar — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/scrollbar/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Width 1400
Set Height 850
Set TypingSpeed 60ms
Type "./examples/scrollbar/demo-bin"
Enter
Sleep 2s
# glide down slowly: watch the smooth bar move in sub-cell steps
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
PageDown
Sleep 900ms
ScrollUp 1
Sleep 350ms
ScrollUp 1
Sleep 350ms
PageUp
Sleep 900ms
Type "q"
Sleep 800ms
