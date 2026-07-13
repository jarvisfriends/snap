# VHS demo for examples/navigation — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/navigation/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 560
Set TypingSpeed 60ms
Hide
Type "./examples/navigation/demo-bin"
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
