#!/usr/bin/env bash
# Render every component demo tape to a gif.
#
# Run this under WSL/Linux (vhs + ttyd + ffmpeg installed there):
#   cd /mnt/e/code/home/go/src/github.com/jarvisfriends/snap
#   bash tools/render_tapes.sh
#
# Windows-native vhs currently hangs: ttyd starts, the headless Chrome
# connects, ttyd exits early (--once + first-disconnect), and vhs waits
# forever for the terminal canvas. Tracked in tui-base ROADMAP SP-15.
# Gifs are build artifacts — regenerate on change, never diff-gate them.
set -euo pipefail

cd "$(dirname "$0")/.."

for tape in */demo.tape; do
  echo "==> vhs $tape"
  vhs "$tape"
done

echo "done: $(find . -maxdepth 2 -name 'demo.gif' | wc -l) gif(s)"
