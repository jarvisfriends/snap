# VHS demo for snap/charts' CellCanvas + Gradient — regenerate from the repo
# root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/cellcanvas/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Width 1400
Set Height 850
Set TypingSpeed 60ms
Type "./examples/cellcanvas/demo-bin"
Enter
# let the plasma drift through the full gradient ramp
Sleep 8s
Type "q"
Sleep 800ms
