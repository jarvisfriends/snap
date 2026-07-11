# VHS demo for snap/charts' CellCanvas + Gradient — regenerate from the repo
# root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/cellcanvas/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 850
Set TypingSpeed 60ms
Hide
Type "./examples/cellcanvas/demo-bin"
Enter
Sleep 1s
Show
# plasma drifts through the full gradient ramp
Sleep 8s
