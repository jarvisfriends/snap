# VHS demo for snap/pickers — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/pickers/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./examples/pickers/demo-bin"
Enter
Sleep 2s
Show
# walk the tree: down, open, back out
Down
Sleep 800ms
Down
Sleep 800ms
Enter
Sleep 1200ms
ScrollDown 1
Sleep 800ms
ScrollUp 1
Sleep 800ms
Backspace
Sleep 1s
Down
Sleep 600ms
# space selects: the chosen path is written to stdout
Space
Sleep 2500ms
