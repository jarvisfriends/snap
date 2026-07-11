# VHS demo for snap/menu — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/menu/demo.gif
Set FontSize 28
Set Width 1400
Set Height 800
Set TypingSpeed 60ms
Type "./examples/menu/demo-bin"
Enter
Sleep 2s
# open the context menu from the keyboard, walk it, choose
Type "m"
Sleep 1200ms
Down
Sleep 700ms
Down
Sleep 700ms
Enter
Sleep 1200ms
# again — the wheel moves the cursor too
Type "m"
Sleep 1s
ScrollDown 1
Sleep 700ms
Enter
Sleep 1200ms
Type "q"
Sleep 800ms
