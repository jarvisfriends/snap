# VHS demo for snap/dependencies (via the status.InfoModal) — regenerate
# from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/dependencies.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input dependencies"
Enter
Sleep 2500ms
Show
# wheel + keys scroll the dependency list inside the modal
ScrollDown 2
Sleep 800ms
ScrollDown 2
Sleep 800ms
PageDown
Sleep 900ms
# scroll back to the top so the loop is seamless (the modal stays open)
PageUp
Sleep 800ms
ScrollUp 2
Sleep 800ms
ScrollUp 2
Sleep 1400ms
Type "Ctrl+C"
