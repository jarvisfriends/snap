# VHS demo for snap/charts (model wiring) — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/charts/charts.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./examples/snap_input/demo-bin charts"
Enter
Sleep 1s
Show
# live stream animates all five charts as one continuous feed
Sleep 8s
