# VHS demo for examples/navigation — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles ./dist/snap_input and runs this tape in the
# official vhs container; gifs land in gitignored dist/). To run one tape by
# hand, build that binary first — the Require line below fails fast without it.
Set WindowBar Colorful
Set WindowBarSize 40
Set Framerate 30
Set Theme "Catppuccin Macchiato"
Require ./dist/snap_input
Output ./dist/navigation.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 5
Set Width 1300
Set Height 700
Set TypingSpeed 60ms
Hide
Type "./dist/snap_input navigation"
Enter
Sleep 2s
Show
# sidebar: walk to the next page
Down
Sleep 900ms
# swap the navigator style at runtime: same pages, same active index
Type "n"
Sleep 1500ms
Right
Sleep 900ms
Type "n"
Sleep 1500ms
Left
Sleep 900ms
# back to the sidebar navigator and first page for a seamless loop
Type "n"
Sleep 1200ms
Up
Sleep 1400ms
Type "Ctrl+C"
