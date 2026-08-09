# VHS demo for snap/charts (model wiring) — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/charts.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input charts"
Enter
Sleep 1s
Show
# live stream animates all five charts as one continuous feed
Sleep 8s
Type "Ctrl+C"
