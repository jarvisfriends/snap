# VHS demo for snap/dependencies (via the status.InfoModal) — regenerate
# from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 40
Output examples/dependencies/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Padding 14
Set Width 1400
Set Height 860
Set TypingSpeed 60ms
Hide
Type "./examples/dependencies/demo-bin"
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
