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
# move the selection with keyboard + wheel
Down
Sleep 700ms
Down
Sleep 700ms
ScrollDown 1
Sleep 700ms
ScrollUp 1
Sleep 700ms
# 3-state column sort: s cycles asc -> desc -> next column (header shows the arrow)
Type "s"
Sleep 900ms
Type "s"
Sleep 900ms
Type "s"
Sleep 900ms
# live filter: / then type — rows narrow as you go, esc clears
Type "/"
Sleep 600ms
Type "us"
Sleep 1200ms
Type "-east"
Sleep 1200ms
Escape
Sleep 800ms
# enter (or double-click) opens the selected row's details
Enter
Sleep 1200ms
Type "q"
Sleep 800ms
