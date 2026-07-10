# VHS demo for snap/datepicker — regenerate from the repo root with:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Output datepicker/demo.gif
Set FontSize 36
Set Width 1200
Set Height 1000
Set TypingSpeed 60ms
Type "./examples/datepicker/demo-bin"
Enter
Sleep 3s
# keyboard: move the highlight
Right
Sleep 500ms
Right
Sleep 500ms
Down
Sleep 1s
# mouse: the wheel pages weeks (vhs >= 0.11)
ScrollDown 1
Sleep 800ms
ScrollDown 1
Sleep 800ms
ScrollUp 1
Sleep 800ms
Enter
Sleep 1500ms
