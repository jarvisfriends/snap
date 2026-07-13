# VHS demo for snap/charts' braille line chart — regenerate from the repo
# root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/linechart/demo.gif
Set FontSize 28
Set Width 1400
Set Height 850
Set TypingSpeed 60ms
Type "./examples/linechart/demo-bin"
Enter
# let the two series sweep across the full window and blend where they cross
Sleep 10s
Type "q"
Sleep 800ms
