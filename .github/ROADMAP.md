
# Tasks to update

## Overall
- [x] ~~Height issue~~ Fixed 2026-07-10: demo roots render inline by default, pinning content to the prompt line — they now set AltScreen. (The empty gifs had a second cause: the vhs container has no Go toolchain, so in-tape `go build` failed; rendertapes now cross-compiles demo binaries on the host.)
- [x] Done 2026-07-10: `tools/rendertapes` (own module) renders all tapes through the official vhs container via the Docker/Podman Go client, worker pool = NumCPU.
- [x] Done 2026-07-10: both tapes showcase ScrollUp/ScrollDown alongside keyboard input.

## Date Picker


## Time Picker

- [ ] Add in an optional seconds value, make private or remove hour and minute, use time.Time to store and retrieve Hour Minute and Second
- [ ] 
