# VHS demo for snap/timepicker (TimeFieldModel) — regenerate from the repo
# root with:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/timepicker.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input timepicker"
Enter
Sleep 3s
Show
# type 14 into hours
Type "14"
Sleep 800ms
# move to minutes, open its dropdown
Tab
Sleep 500ms
Space
Sleep 1s
# mouse: the wheel scrolls the dropdown
ScrollDown 2
Sleep 800ms
ScrollUp 1
Sleep 600ms
Down
Sleep 400ms
Enter
Sleep 800ms
# enter confirms the field: the time is written to stdout
Enter
Sleep 2500ms
Type "Ctrl+C"
