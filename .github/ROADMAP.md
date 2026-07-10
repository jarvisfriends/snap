
# Tasks to update

## Overall
- [ ] When making the tapes a different size, they still seem to be treated as a limited height in our display... Whats going on? Lets fix
- [ ] Change render_tapes.sh to a brief go file in its own folder, Run each tape generation in its own goroutine so they complete faster when provided more resources instead of sequential. use the worker paradigm since we plan on the number of these getting large and don't want the host to get unusable. number of goroutines == num cores/threads is fine
- [ ] as of v0.11.0 of tape, it now supports ScrollDown ScrollUp for mouse input, lets remember to show off the mouse input to show how the mouse and keyboard are both supported

## Date Picker


## Time Picker

- [ ] Add in an optional seconds value, make private or remove hour and minute, use time.Time to store and retrieve Hour Minute and Second
- [ ] 
