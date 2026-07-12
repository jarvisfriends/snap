# VHS demo for snap/charts (model wiring) — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/charts/demo.gif
Set FontSize 28
Set Width 1400
Set Height 700
Set TypingSpeed 60ms
Type "./examples/charts/demo-bin"
Enter
# let the live stream animate all five charts
Sleep 8s
Type "q"
Sleep 1s
