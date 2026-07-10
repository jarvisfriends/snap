# VHS demo for snap/timepicker (TimeFieldModel) — regenerate with:  vhs timepicker/demo.tape
# (run from the snap repo root; gif output is a build artifact, not diffable)
Output timepicker/demo.gif
Set FontSize 16
Set Width 700
Set Height 420
Set TypingSpeed 100ms
Type "go run ./examples/timefield"
Enter
Sleep 3s
# type 14 into hours
Type "14"
Sleep 800ms
# move to minutes, open its dropdown, pick a value
Tab
Sleep 500ms
Space
Sleep 1s
Down
Sleep 400ms
Down
Sleep 400ms
Enter
Sleep 800ms
# finish the field
Enter
Sleep 1200ms
