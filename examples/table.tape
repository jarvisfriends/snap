# VHS demo for snap/table — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/table.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input table"
Enter
Sleep 2s
Show
# move the selection with keyboard + wheel
Down
Sleep 700ms
Down
Sleep 700ms
ScrollDown 1
Sleep 700ms
ScrollUp 1
Sleep 700ms
# 3-state column sort: s cycles asc -> desc -> next column
Type "s"
Sleep 900ms
Type "s"
Sleep 900ms
Type "s"
Sleep 900ms
# live filter: rows narrow as you type, esc clears
Type "/"
Sleep 600ms
Type "us"
Sleep 1200ms
Type "-east"
Sleep 1200ms
Escape
Sleep 800ms
# enter opens the selected row; its key is written to stdout
Enter
Sleep 2500ms
Type "Ctrl+C"
