# VHS demo for snap/timepicker (TimeFieldModel) — regenerate from the repo
# root with:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output timepicker/demo.gif
Set FontSize 36
Set Width 1200
Set Height 760
Set TypingSpeed 60ms
Type "./examples/timepicker/demo-bin"
Enter
Sleep 3s
# type 14 into hours
Type "14"
Sleep 800ms
# move to minutes, open its dropdown
Tab
Sleep 500ms
Space
Sleep 1s
# mouse: the wheel scrolls the dropdown (vhs >= 0.11)
ScrollDown 2
Sleep 800ms
ScrollUp 1
Sleep 600ms
Down
Sleep 400ms
Enter
Sleep 800ms
# finish the field
Enter
Sleep 1200ms
