# VHS demo for snap/pickers — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/pickers.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input pickers"
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
Type "Ctrl+C"
