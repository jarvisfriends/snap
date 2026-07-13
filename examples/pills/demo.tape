# VHS demo for snap/styles pills — regenerate from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/pills/demo.gif
# The bundled Nerd Font renders the Powerline shapes (Round/Arrow/Slant/
# Flame) so the full ten-shape cycle shows in the gif.
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Width 1400
Set Height 760
Set TypingSpeed 60ms
Type "./examples/pills/demo-bin"
Enter
Sleep 2s
# walk the pure-Unicode shapes first: circle -> triangle -> diagonal ->
# fade -> block -> plain
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
Right
Sleep 1000ms
# keep going into the Nerd Font shapes: round -> arrow -> slant -> flame
Right
Sleep 1200ms
Right
Sleep 1200ms
Right
Sleep 1200ms
Right
Sleep 2000ms
# the wheel cycles too
ScrollUp 1
Sleep 800ms
ScrollDown 1
Sleep 1500ms
Type "q"
Sleep 800ms
