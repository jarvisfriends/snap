# VHS demo for snap/table — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Output examples/table/demo.gif
Set FontSize 28
Set Width 1400
Set Height 1100
Set TypingSpeed 60ms
Type "./examples/table/demo-bin"
Enter
Sleep 2s
# move the selection, sort by clicking through with keyboard + wheel
Down
Sleep 700ms
Down
Sleep 700ms
ScrollDown 1
Sleep 700ms
ScrollUp 1
Sleep 700ms
Enter
Sleep 1200ms
Type "q"
Sleep 800ms
