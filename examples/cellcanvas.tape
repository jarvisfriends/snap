# VHS demo for snap/charts' CellCanvas + Gradient — regenerate from the repo
# root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Theme "Catppuccin Macchiato"
Set Framerate 20
Require ./dist/snap_input
Output ./dist/cellcanvas.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input cellcanvas"
Enter
Sleep 1s
Show
# plasma drifts through the full gradient ramp
Sleep 6s
Type "Ctrl+C"
