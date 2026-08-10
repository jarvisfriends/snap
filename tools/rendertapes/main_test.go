// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"
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

// newPublishRoot builds a repo-shaped temp dir: tapes under examples/, gifs
// under dist/, and returns the root plus a helper to write a URL record.
func newPublishRoot(t *testing.T) (root string, writeGif func(name string, size int), writeRec func(name, url, hash string)) {
	t.Helper()
	root = t.TempDir()
	for _, d := range []string{"examples", "dist"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	writeGif = func(name string, size int) {
		t.Helper()
		p := filepath.Join(root, "dist", name+".gif")
		if err := os.WriteFile(p, []byte(strings.Repeat("g", size)), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeRec = func(name, url, hash string) {
		t.Helper()
		if err := writeRecord(filepath.Join(root, "examples", name+".gif.url"), url, hash); err != nil {
			t.Fatal(err)
		}
	}
	return root, writeGif, writeRec
}

func tg(name string) tapeGif {
	return tapeGif{tape: "examples/" + name + ".tape", gif: "dist/" + name + ".gif"}
}

// TestPlanPublishSkipsUnchanged: a gif whose bytes match its recorded upload
// costs no request, but still feeds its URL into the doc rewrite so a link
// deleted from the markdown comes back.
func TestPlanPublishSkipsUnchanged(t *testing.T) {
	t.Parallel()

	root, writeGif, writeRec := newPublishRoot(t)
	writeGif("table", 128)
	hash, err := fileHash(filepath.Join(root, "dist", "table.gif"))
	if err != nil {
		t.Fatal(err)
	}
	writeRec("table", "https://vhs.charm.sh/vhs-abc.gif", hash)

	pending, links, err := planPublish(root, []tapeGif{tg("table")}, publishPolicy{maxBytes: 10 << 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Errorf("pending = %d; want 0 (gif is unchanged)", len(pending))
	}
	if len(links) != 1 || links[0].url != "https://vhs.charm.sh/vhs-abc.gif" {
		t.Errorf("links = %+v; want the recorded URL carried through", links)
	}
}

// TestPlanPublishForce re-uploads an unchanged gif when asked.
func TestPlanPublishForce(t *testing.T) {
	t.Parallel()

	root, writeGif, writeRec := newPublishRoot(t)
	writeGif("table", 128)
	hash, err := fileHash(filepath.Join(root, "dist", "table.gif"))
	if err != nil {
		t.Fatal(err)
	}
	writeRec("table", "https://vhs.charm.sh/vhs-abc.gif", hash)

	pending, _, err := planPublish(root, []tapeGif{tg("table")}, publishPolicy{maxBytes: 10 << 20, force: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 {
		t.Errorf("pending = %d; want 1 under -force-publish", len(pending))
	}
}

// TestPlanPublishChangedAndUnrecorded: a re-rendered gif and a never-published
// one both queue, and the re-rendered one keeps its old URL as the rewrite's
// "from" side so docs already pointing at the hosted copy still get updated.
func TestPlanPublishChangedAndUnrecorded(t *testing.T) {
	t.Parallel()

	root, writeGif, writeRec := newPublishRoot(t)
	writeGif("table", 128)
	writeGif("menu", 64)
	writeRec("table", "https://vhs.charm.sh/vhs-old.gif", "stalehash")

	pending, links, err := planPublish(root, []tapeGif{tg("table"), tg("menu")}, publishPolicy{maxBytes: 10 << 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatalf("pending = %d; want 2", len(pending))
	}
	if len(links) != 0 {
		t.Errorf("links = %+v; want none until the uploads land", links)
	}
	for _, j := range pending {
		switch j.tg.gif {
		case "dist/table.gif":
			if j.prev != "https://vhs.charm.sh/vhs-old.gif" {
				t.Errorf("table prev = %q; want the superseded URL", j.prev)
			}
		case "dist/menu.gif":
			if j.prev != "" {
				t.Errorf("menu prev = %q; want empty on a first publish", j.prev)
			}
		}
	}
}

// TestPlanPublishRejectsOversized names every offender at once, before any
// upload spends the rate limit.
func TestPlanPublishRejectsOversized(t *testing.T) {
	t.Parallel()

	root, writeGif, _ := newPublishRoot(t)
	writeGif("cellcanvas", 4096)
	writeGif("linechart", 4096)
	writeGif("menu", 16)

	_, _, err := planPublish(root, []tapeGif{tg("cellcanvas"), tg("linechart"), tg("menu")},
		publishPolicy{maxBytes: 1024})
	if err == nil {
		t.Fatal("oversized gifs should fail the plan")
	}
	for _, want := range []string{"cellcanvas", "linechart"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
	if strings.Contains(err.Error(), "menu") {
		t.Errorf("error %q should not name the in-limit gif", err)
	}
}

// TestPlanPublishRejectsUnrendered: publishing without rendering first is a
// setup mistake, not something to discover mid-upload.
func TestPlanPublishRejectsUnrendered(t *testing.T) {
	t.Parallel()

	root, _, _ := newPublishRoot(t)
	if _, _, err := planPublish(root, []tapeGif{tg("table")}, publishPolicy{maxBytes: 10 << 20}); err == nil {
		t.Fatal("a missing gif should fail the plan")
	}
}

func TestRecordRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := filepath.Join(dir, "table.gif.url")
	if err := writeRecord(p, "https://vhs.charm.sh/vhs-abc.gif", "deadbeef"); err != nil {
		t.Fatal(err)
	}
	url, hash := readRecord(p)
	if url != "https://vhs.charm.sh/vhs-abc.gif" || hash != "deadbeef" {
		t.Errorf("readRecord = (%q, %q); want the written pair", url, hash)
	}
}

// TestReadRecordURLOnly: a record written before hashes existed (or seeded by
// hand) still yields its URL, and the empty hash forces one re-upload.
func TestReadRecordURLOnly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := filepath.Join(dir, "table.gif.url")
	if err := os.WriteFile(p, []byte("https://vhs.charm.sh/vhs-abc.gif\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	url, hash := readRecord(p)
	if url != "https://vhs.charm.sh/vhs-abc.gif" {
		t.Errorf("url = %q; want the recorded URL", url)
	}
	if hash != "" {
		t.Errorf("hash = %q; want empty", hash)
	}
}

func TestReadRecordMissing(t *testing.T) {
	t.Parallel()

	url, hash := readRecord(filepath.Join(t.TempDir(), "nope.gif.url"))
	if url != "" || hash != "" {
		t.Errorf("readRecord(missing) = (%q, %q); want empty", url, hash)
	}
}
