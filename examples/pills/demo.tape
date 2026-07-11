# VHS demo for snap/styles pills — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Output examples/pills/demo.gif
Set FontSize 28
Set Width 1400
Set Height 900
Set TypingSpeed 60ms
Type "./examples/pills/demo-bin"
Enter
Sleep 2s
# walk the shapes: round -> arrow -> slant -> flame -> block -> plain
Right
Sleep 1200ms
Right
Sleep 1200ms
Right
Sleep 1200ms
Right
Sleep 1200ms
Right
Sleep 1200ms
# wheel back up to round
ScrollUp 1
Sleep 700ms
ScrollUp 1
Sleep 700ms
ScrollUp 1
Sleep 700ms
ScrollUp 1
Sleep 700ms
ScrollUp 1
Sleep 1500ms
Type "q"
Sleep 800ms
