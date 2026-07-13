# VHS demo for examples/navigation — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/navigation/demo.gif
Set FontSize 28
Set Width 1400
Set Height 800
Set TypingSpeed 60ms
Type "./examples/navigation/demo-bin"
Enter
Sleep 2s
# sidebar first: arrows walk the pages
Down
Sleep 700ms
Down
Sleep 700ms
# swap navigator styles at runtime — same pages, same active index
Type "n"
Sleep 1500ms
Right
Sleep 700ms
Right
Sleep 700ms
Type "n"
Sleep 1500ms
Left
Sleep 700ms
# back to the sidebar
Type "n"
Sleep 1200ms
Up
Sleep 700ms
# Enter picks the page: its ID is all the program prints
Enter
Sleep 1500ms
