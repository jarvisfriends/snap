# VHS demo for snap/menu — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/menu/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 800
Set TypingSpeed 60ms
Hide
Type "./examples/menu/demo-bin"
Enter
Sleep 2s
Show
# open the context menu, walk it, choose: the item ID goes to stdout
Type "m"
Sleep 1200ms
Down
Sleep 700ms
Down
Sleep 700ms
Enter
Sleep 2500ms
