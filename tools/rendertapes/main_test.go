// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTape drops a tape into <root>/examples and returns its absolute path,
// matching what findTapes hands to tapeGifs.
func writeTape(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := filepath.Join(root, "examples")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name+".tape")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestTapeGifsReadsOutput: the gif path comes from the tape's own Output
// directive, so moving where gifs land needs no change here. The "./" prefix
// the tapes use is normalised away so the path matches the repo-relative form
// everything else speaks.
func TestTapeGifsReadsOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	tape := writeTape(t, root, "table", "Set Width 1300\nOutput ./dist/table.gif\nType \"hi\"\n")

	gifs, err := tapeGifs(root, []string{tape})
	if err != nil {
		t.Fatal(err)
	}
	if len(gifs) != 1 {
		t.Fatalf("got %d gifs; want 1", len(gifs))
	}
	if gifs[0].gif != "dist/table.gif" {
		t.Errorf("gif = %q; want %q", gifs[0].gif, "dist/table.gif")
	}
	if gifs[0].tape != "examples/table.tape" {
		t.Errorf("tape = %q; want %q", gifs[0].tape, "examples/table.tape")
	}
}

// TestTapeGifsIgnoresNonGifOutput: VHS can also emit mp4/webm/frame dirs, and
// only gifs are shown in the docs.
func TestTapeGifsIgnoresNonGifOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	tape := writeTape(t, root, "table", "Output ./dist/table.mp4\nOutput ./dist/table.gif\n")

	gifs, err := tapeGifs(root, []string{tape})
	if err != nil {
		t.Fatal(err)
	}
	if len(gifs) != 1 || gifs[0].gif != "dist/table.gif" {
		t.Errorf("gifs = %+v; want only the gif Output", gifs)
	}
}

// TestTapeGifsRequiresOutput: a tape with no gif Output renders nothing the
// gallery can show, which is a mistake worth catching before a render.
func TestTapeGifsRequiresOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	tape := writeTape(t, root, "table", "Set Width 1300\nType \"hi\"\n")

	if _, err := tapeGifs(root, []string{tape}); err == nil {
		t.Fatal("a tape with no gif Output should fail")
	}
}

// TestTapeGifsRejectsEscapingOutput: Output is relative to the container
// mount, i.e. the repo root, so a path climbing out of it would write
// somewhere the repo does not own.
func TestTapeGifsRejectsEscapingOutput(t *testing.T) {
	t.Parallel()

	for name, out := range map[string]string{
		"parent":   "../evil.gif",
		"absolute": "/tmp/evil.gif",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			tape := writeTape(t, root, "table", "Output "+out+"\n")
			_, err := tapeGifs(root, []string{tape})
			if err == nil {
				t.Fatalf("Output %q should be rejected", out)
			}
			if !strings.Contains(err.Error(), "escapes the repo root") {
				t.Errorf("error = %v; want an escape complaint", err)
			}
		})
	}
}

// TestTapeGifsDedupes: several tapes writing the same gif yield one entry, so
// a release upload does not attach the same asset twice.
func TestTapeGifsDedupes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	a := writeTape(t, root, "a", "Output ./dist/shared.gif\n")
	b := writeTape(t, root, "b", "Output ./dist/shared.gif\n")

	gifs, err := tapeGifs(root, []string{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if len(gifs) != 1 {
		t.Errorf("gifs = %+v; want one entry for the shared output", gifs)
	}
}

// TestReportGifSizesFailsOnMissingGif: a tape that declared an Output but
// wrote nothing is a silent-failure case worth surfacing.
func TestReportGifSizesFailsOnMissingGif(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	tape := writeTape(t, root, "table", "Output ./dist/table.gif\n")

	if err := reportGifSizes(root, []string{tape}, 0); err == nil {
		t.Fatal("a declared-but-unwritten gif should fail")
	}
}

func TestReportGifSizesPasses(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	tape := writeTape(t, root, "table", "Output ./dist/table.gif\n")
	if err := os.MkdirAll(filepath.Join(root, "dist"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "dist", "table.gif"), []byte("gif"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := reportGifSizes(root, []string{tape}, 5<<20); err != nil {
		t.Errorf("reportGifSizes() = %v; want nil", err)
	}
}
