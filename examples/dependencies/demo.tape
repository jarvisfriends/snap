# VHS demo for snap/dependencies (via the status.InfoModal) — regenerate
# from the repo root:
#   go -C tools/rendertapes run .
# (rendertapes cross-compiles the linux demo binary and runs this tape in the
# official vhs container; gifs are build artifacts, not diffable sources)
Set WindowBar Colorful
Set WindowBarSize 60
Output examples/dependencies/demo.gif
Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 28
Set Width 1400
Set Height 900
Set TypingSpeed 60ms
Type "./examples/dependencies/demo-bin"
Enter
Sleep 2500ms
# the wheel scrolls the dependency list (via InfoModal.HandleMouse)
ScrollDown 2
Sleep 800ms
ScrollDown 2
Sleep 800ms
# keyboard drives it too
PageDown
Sleep 900ms
PageUp
Sleep 900ms
ScrollUp 2
Sleep 800ms
# esc closes the modal, i reopens it
Escape
Sleep 1s
Type "i"
Sleep 1200ms
Escape
Sleep 600ms
Type "q"
Sleep 800ms
