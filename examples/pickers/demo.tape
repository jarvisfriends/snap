# VHS demo for snap/pickers — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Output examples/pickers/demo.gif
Set FontSize 28
Set Width 1400
Set Height 1100
Set TypingSpeed 60ms
Type "./examples/pickers/demo-bin"
Enter
Sleep 2s
# walk the tree: down, open, back out, select
Down
Sleep 800ms
Down
Sleep 800ms
Enter
Sleep 1200ms
# wheel drives it too: down a row, right opens, left backs out
ScrollDown 1
Sleep 800ms
ScrollUp 1
Sleep 800ms
Backspace
Sleep 1s
Down
Sleep 600ms
Space
Sleep 1500ms
