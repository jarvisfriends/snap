# VHS demo for snap/menu — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/menu.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input menu"
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
Type "Ctrl+C"
