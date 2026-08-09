// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package main

import (
	"path/filepath"
	"testing"
)

// vhsPublishOutput is the tail of a real successful `vhs publish` run. The
// hosted link appears four times in three shapes, so the parser has to pick
// the bare one and ignore the badge.
const vhsPublishOutput = `Creating ./dist/timepicker.gif...
Publishing ./dist/timepicker.gif...
Done!

  Share your GIF with Markdown:
  ![Made with VHS](https://vhs.charm.sh/vhs-3ojYKOtCdS5WoeeQwaqizs.gif)

  Or HTML (with badge):
  <img src="https://vhs.charm.sh/vhs-3ojYKOtCdS5WoeeQwaqizs.gif" alt="Made with VHS">
  <a href="https://vhs.charm.sh">
    <img src="https://stuff.charm.sh/vhs/badge.svg">
  </a>

  Or link to it:
  https://vhs.charm.sh/vhs-3ojYKOtCdS5WoeeQwaqizs.gif
`

func TestPublishedURL(t *testing.T) {
	t.Parallel()

	const want = "https://vhs.charm.sh/vhs-3ojYKOtCdS5WoeeQwaqizs.gif"
	if got := publishedURL(vhsPublishOutput); got != want {
		t.Errorf("publishedURL() = %q; want %q", got, want)
	}
}

// TestPublishedURLRejectsNonGif pins the two near-misses in that output: the
// badge svg and the bare hosting root behind the badge anchor.
func TestPublishedURLRejectsNonGif(t *testing.T) {
	t.Parallel()

	for name, out := range map[string]string{
		"badge only":   `<img src="https://stuff.charm.sh/vhs/badge.svg">`,
		"anchor only":  `https://vhs.charm.sh`,
		"upload error": "Creating ./dist/timepicker.gif...\nEOF\n",
		"empty":        "",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := publishedURL(out); got != "" {
				t.Errorf("publishedURL(%q) = %q; want \"\"", out, got)
			}
		})
	}
}

// TestURLPath: the recorded URL lands beside the tape, named after the gif,
// never in dist/ where goreleaser --clean would delete it.
func TestURLPath(t *testing.T) {
	t.Parallel()

	root := filepath.FromSlash("/repo")
	got := urlPath(root, tapeGif{tape: "examples/timepicker.tape", gif: "dist/timepicker.gif"})
	want := filepath.Join(root, "examples", "timepicker.gif.url")
	if got != want {
		t.Errorf("urlPath() = %q; want %q", got, want)
	}
}
