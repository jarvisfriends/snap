// Copyright (c) 2026 Jarvis Friends contributors
// SPDX-License-Identifier: MIT

// Command rendertapes renders every component demo tape to a gif by running
// the official VHS container: rendering inside Docker produces consistent
// output across hosts, and sidesteps a Windows-native vhs/ttyd hang.
//
// Tapes run in parallel on a worker pool sized to the CPU count, so large
// tape sets finish fast without making the host unusable. Docker/Podman
// CLI commands do the container work so this tool stays dependency-light.
//
// Gifs are build artifacts written to dist/, not committed sources (see
// .gitignore), so the markdown that shows them has to point somewhere that
// outlives a clone: -publish uploads each gif to Charm hosting, records the
// URL in <name>.gif.url beside the tape, and repoints the docs. -relink
// replays those recorded URLs into the docs without rendering or a network
// round trip.
//
// -verbose streams the container's own output and echoes every command. The
// container is otherwise silent unless it fails, which hides the reason a
// render produced a bad gif rather than no gif at all.
//
// Usage, from the snap repo root:
//
//	go -C tools/rendertapes run . [-image ghcr.io/charmbracelet/vhs] [-workers N]
//	go -C tools/rendertapes run . -verbose   # stream container output live
//	go -C tools/rendertapes run . -publish   # render, upload, repoint docs
//	go -C tools/rendertapes run . -relink    # repoint docs from <gif>.url only
//
// This is a standalone module so tool-only dependencies never enter snap's
// library graph.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// verbose mirrors the -verbose flag. Container and toolchain output is
// captured either way; this decides whether it is shown when the command
// succeeds, instead of only being folded into an error message.
var verbose bool

// vlogf logs only under -verbose.
func vlogf(format string, args ...any) {
	if verbose {
		log.Printf(format, args...)
	}
}

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatalf("rendertapes: %v", err)
	}
}

// run holds the real body so deferred cleanup (demo binaries, the Docker
// client) executes on error paths — log.Fatal in main would skip defers.
func run() error {
	var (
		imageRef = flag.String("image", "ghcr.io/charmbracelet/vhs", "VHS container image")
		workers  = flag.Int("workers", runtime.NumCPU(), "parallel renders (default: CPU count)")
		repoRoot = flag.String("root", "", "snap repo root (default: two levels up from this tool)")
		publish  = flag.Bool("publish", false, "vhs-publish each rendered gif and point markdown at the hosted URL")
		relink   = flag.Bool("relink", false, "point markdown at the URLs a previous -publish recorded, without rendering")
	)
	flag.BoolVar(&verbose, "verbose", false, "stream container/toolchain output and echo every command")
	flag.BoolVar(&verbose, "v", false, "shorthand for -verbose")
	flag.Parse()

	root := *repoRoot
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		root = filepath.Dir(filepath.Dir(wd)) // tools/rendertapes -> repo root
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	root = absRoot

	tapes, err := findTapes(root)
	if err != nil {
		return fmt.Errorf("scan for *.tape under %s: %w", root, err)
	}
	if len(tapes) == 0 {
		return fmt.Errorf("no *.tape files found under %s", root)
	}

	// -relink is a docs-only repair: no Go toolchain, no container, no
	// network. It exists because a fresh clone has no gifs at all, so the
	// markdown can drift back to dead local paths without anyone re-rendering.
	if *relink {
		return relinkDocs(root, tapes)
	}

	if buildErr := buildDemoBinaries(root); buildErr != nil {
		return buildErr
	}
	defer cleanDemoBinaries(root)

	ctx := context.Background()

	if err := ensureImage(ctx, *imageRef); err != nil {
		return fmt.Errorf("pull %s: %w", *imageRef, err)
	}

	jobs := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex
	failures := 0

	for range max(*workers, 1) {
		wg.Go(func() {
			for tape := range jobs {
				rel, _ := filepath.Rel(root, tape)
				rel = filepath.ToSlash(rel)
				log.Printf("==> %s", rel)
				if err := renderTape(ctx, *imageRef, root, rel); err != nil {
					mu.Lock()
					failures++
					mu.Unlock()
					log.Printf("FAIL %s: %v", rel, err)
					continue
				}
				log.Printf("ok   %s", rel)
			}
		})
	}
	for _, t := range tapes {
		jobs <- t
	}
	close(jobs)
	wg.Wait()

	if failures > 0 {
		return fmt.Errorf("%d of %d tape(s) failed", failures, len(tapes))
	}
	log.Printf("rendertapes: %d gif(s) rendered", len(tapes))
	if *publish {
		return publishGifs(ctx, *imageRef, root, tapes)
	}
	return nil
}

// gifLink ties one tape's gif to where markdown should point at it. rel is
// the repo-relative path docs used before anything was published; prev is the
// URL the last publish recorded (empty on a first run). Rewrites accept
// either as the "from" side, so a re-publish still finds links that no longer
// mention the local path.
type gifLink struct{ rel, url, prev string }

// publishGifs uploads every rendered gif to Charm's hosting via `vhs publish`
// (gifs stay out of git — see .gitignore — keeping clones small), records
// each URL in <name>.gif.url beside the tape, and repoints markdown links.
func publishGifs(ctx context.Context, imageRef, root string, tapes []string) error {
	gifs, err := tapeGifs(root, tapes)
	if err != nil {
		return err
	}
	links := make([]gifLink, 0, len(gifs))
	for _, tg := range gifs {
		abs := filepath.Join(root, filepath.FromSlash(tg.gif))
		if _, statErr := os.Stat(abs); statErr != nil {
			return fmt.Errorf("publish %s: gif was not rendered: %w", tg.gif, statErr)
		}
		// Read the old URL before overwriting it — it is the only handle on
		// links that a previous run already rewrote away from the local path.
		urlFile := urlPath(root, tg)
		prev := recordedURL(urlFile)

		stream, flush := streamFor("publish " + tg.gif)
		out, err := runContainerOutput(ctx, stream,
			"run", "--rm", "-v", root+":/vhs", imageRef, "publish", tg.gif)
		flush()
		if err != nil {
			return fmt.Errorf("vhs publish %s: %w\n%s", tg.gif, err, strings.TrimSpace(out))
		}
		url := publishedURL(out)
		if url == "" {
			return fmt.Errorf("vhs publish %s: no gif URL in output:\n%s",
				tg.gif, strings.TrimSpace(out))
		}
		if err := os.WriteFile(urlFile, []byte(url+"\n"), 0o600); err != nil {
			return err
		}
		links = append(links, gifLink{rel: tg.gif, url: url, prev: prev})
		log.Printf("published %s -> %s", tg.gif, url)
	}
	return rewriteDocLinks(root, links)
}

// publishedURL extracts the hosted gif URL from `vhs publish` output, which
// offers the same link four ways — markdown, an <img> tag, a badge anchor,
// and finally the bare URL:
//
//	![Made with VHS](https://vhs.charm.sh/vhs-3ojYKOtC.gif)
//	<img src="https://vhs.charm.sh/vhs-3ojYKOtC.gif" alt="Made with VHS">
//	<img src="https://stuff.charm.sh/vhs/badge.svg">
//	https://vhs.charm.sh/vhs-3ojYKOtC.gif
//
// Only the last is a bare whitespace-delimited field, so requiring both the
// https:// prefix and a .gif suffix picks it out and cannot match the badge
// or the plain https://vhs.charm.sh anchor. Returns "" if the format changes.
func publishedURL(out string) string {
	for field := range strings.FieldsSeq(out) {
		if strings.HasPrefix(field, "https://") && strings.HasSuffix(field, ".gif") {
			return field
		}
	}
	return ""
}

// urlPath is where a tape's published URL is recorded: the gif's name with a
// .url suffix, beside the tape rather than beside the gif. Gifs land in dist/,
// which .gitignore excludes and `goreleaser release --clean` deletes, so a URL
// recorded there could never be committed and would not survive a release.
func urlPath(root string, tg tapeGif) string {
	dir := filepath.Dir(filepath.FromSlash(tg.tape))
	return filepath.Join(root, dir, path.Base(tg.gif)+".url")
}

// relinkDocs repoints markdown at the URLs a previous -publish recorded in
// <gif>.url, without rendering or uploading anything. Every tape must have a
// recorded URL: a partial relink would leave some images pointing at gifs that
// no clone contains, which is the exact breakage this mode exists to fix.
//
// This repairs links that still name the local gif path. Moving docs from one
// hosted URL to another is -publish's job, since only the publish step knows
// the URL being replaced.
func relinkDocs(root string, tapes []string) error {
	gifs, err := tapeGifs(root, tapes)
	if err != nil {
		return err
	}
	links := make([]gifLink, 0, len(gifs))
	var missing []string
	for _, tg := range gifs {
		url := recordedURL(urlPath(root, tg))
		if url == "" {
			missing = append(missing, tg.gif)
			continue
		}
		links = append(links, gifLink{rel: tg.gif, url: url})
	}
	if len(missing) > 0 {
		return fmt.Errorf("no <gif>.url recorded for %d of %d tape(s) — run with -publish first:\n  %s",
			len(missing), len(gifs), strings.Join(missing, "\n  "))
	}
	return rewriteDocLinks(root, links)
}

// recordedURL reads the hosted URL a previous publish wrote to urlFile,
// returning "" when there is none yet.
func recordedURL(urlFile string) string {
	//nolint:gosec // url paths are derived from tape paths inside the repo.
	b, err := os.ReadFile(urlFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// rewriteDocLinks points markdown image references at the hosted URLs,
// accepting either the local gif path or a previously published URL as the
// source so the rewrite is repeatable rather than one-shot.
func rewriteDocLinks(root string, links []gifLink) error {
	docs, err := findDocs(root)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		//nolint:gosec // doc paths come from the walk of root in findDocs.
		b, err := os.ReadFile(doc)
		if err != nil {
			continue
		}
		s := string(b)
		for _, l := range links {
			s = strings.ReplaceAll(s, "("+l.rel+")", "("+l.url+")")
			if l.prev != "" && l.prev != l.url {
				s = strings.ReplaceAll(s, "("+l.prev+")", "("+l.url+")")
			}
		}
		if s == string(b) {
			continue
		}
		//nolint:gosec // doc lives inside the repo root by construction.
		if err := os.WriteFile(doc, []byte(s), 0o600); err != nil {
			return err
		}
		log.Printf("updated links in %s", filepath.ToSlash(mustRel(root, doc)))
	}
	return nil
}

// findDocs walks the repo for every markdown file. Any doc may show a demo,
// so the sweep is repo-wide rather than a fixed list of directories.
func findDocs(root string) ([]string, error) {
	var docs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			docs = append(docs, path)
		}
		return nil
	})
	return docs, err
}

// tapeGif pairs a tape with a gif it writes. Both are repo-relative and
// slash-separated. The tape is kept because it, not the gif, determines where
// the published URL is recorded — see urlPath.
type tapeGif struct{ tape, gif string }

// tapeGifs returns the gif each tape writes, read from that tape's own
// `Output` directive. Deriving the path from the tape rather than globbing a
// fixed layout keeps publishing correct wherever a tape drops its gif (today:
// dist/), and covers a tape that emits more than one.
func tapeGifs(root string, tapes []string) ([]tapeGif, error) {
	var gifs []tapeGif
	seen := map[string]bool{}
	for _, tape := range tapes {
		//nolint:gosec // tape paths come from the walk of root in findTapes.
		b, err := os.ReadFile(tape)
		if err != nil {
			return nil, err
		}
		relTape := filepath.ToSlash(mustRel(root, tape))
		found := false
		for line := range strings.Lines(string(b)) {
			rest, ok := strings.CutPrefix(strings.TrimSpace(line), "Output ")
			if !ok {
				continue
			}
			// VHS tapes may also emit mp4/webm/frame directories; only gifs
			// are published and referenced from the docs.
			out := filepath.ToSlash(strings.Trim(strings.TrimSpace(rest), `"'`))
			if !strings.HasSuffix(out, ".gif") {
				continue
			}
			// Output paths are relative to the container mount, i.e. the repo
			// root; anything escaping it would be published from outside.
			out = strings.TrimPrefix(out, "./")
			if filepath.IsAbs(out) || out == ".." || strings.HasPrefix(out, "../") {
				return nil, fmt.Errorf("%s: Output %q escapes the repo root", relTape, out)
			}
			found = true
			if !seen[out] {
				seen[out] = true
				gifs = append(gifs, tapeGif{tape: relTape, gif: out})
			}
		}
		if !found {
			return nil, fmt.Errorf("%s: no `Output <path>.gif` directive", relTape)
		}
	}
	return gifs, nil
}

// findTapes walks the repo for every *.tape file (any name, any depth — the
// tapes live in examples/, one per subcommand, but nothing here depends on
// that), skipping VCS and tool directories.
func findTapes(root string) ([]string, error) {
	var tapes []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "tools", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".tape") {
			tapes = append(tapes, path)
		}
		return nil
	})
	return tapes, err
}

// ensureImage pulls the VHS image when it is not already present. A failing
// inspect is the normal "not pulled yet" signal, so only the pull is narrated.
func ensureImage(ctx context.Context, ref string) error {
	if err := runContainer(ctx, "image", "inspect", ref); err == nil {
		vlogf("image already present: %s", ref)
		return nil
	}
	log.Printf("pulling %s", ref)
	stream, flush := streamFor("pull")
	defer flush()
	out, err := runContainerOutput(ctx, stream, "pull", ref)
	if err != nil {
		return fmt.Errorf("%w\n%s", err, strings.TrimSpace(out))
	}
	return nil
}

// renderTape runs one tape in its own container with the repo mounted at
// /vhs (the image's working directory), mirroring
// `docker run --rm -v <root>:/vhs ghcr.io/charmbracelet/vhs <tape>`.
func renderTape(ctx context.Context, imageRef, root, relTape string) error {
	// The VHS container's terminal advertises TERM=xterm-256color, so
	// lipgloss downsamples true-color backgrounds. COLORTERM=truecolor
	// upgrades the profile so SGR colors match the page while honoring NO_COLOR.
	stream, flush := streamFor(relTape)
	defer flush()
	out, err := runContainerOutput(
		ctx, stream,
		"run", "--rm",
		"-e", "COLORTERM=truecolor",
		"-v", root+":/vhs",
		imageRef,
		relTape,
	)
	if err != nil {
		return fmt.Errorf("vhs container failed: %w\n%s", err, strings.TrimSpace(out))
	}
	return nil
}

// containerRuntime picks docker or podman once, probing each with `version`
// so a CLI whose daemon is down loses to one that works. Both accept the same
// arguments for the subset used here.
//
// This used to be decided per call, by running docker and silently retrying
// podman: every invocation on a podman-only host paid a doomed docker exec,
// and when both failed the caller saw docker's error even though podman was
// the runtime that would have run — masking the real failure.
var containerRuntime = sync.OnceValues(func() (string, error) {
	var tried []string
	for _, name := range []string{"docker", "podman"} {
		out, err := exec.Command(name, "version").CombinedOutput()
		if err == nil {
			vlogf("container runtime: %s", name)
			return name, nil
		}
		tried = append(tried, fmt.Sprintf("%s: %v: %s", name, err, firstLine(out)))
	}
	return "", fmt.Errorf("no working container runtime (need docker or podman):\n  %s",
		strings.Join(tried, "\n  "))
})

// firstLine reduces a command's output to its first non-empty line, enough to
// identify why a runtime probe failed without pasting a usage dump.
func firstLine(b []byte) string {
	for line := range strings.Lines(string(b)) {
		if s := strings.TrimSpace(line); s != "" {
			return s
		}
	}
	return "no output"
}

// runContainer runs the container CLI and discards its output unless -verbose
// or an error asks for it.
func runContainer(ctx context.Context, args ...string) error {
	_, err := runContainerOutput(ctx, nil, args...)
	return err
}

// runContainerOutput runs the container CLI and returns its combined output.
// When stream is non-nil the output is also forwarded there as it arrives, so
// a slow or wedged render reports progress instead of going quiet until it
// finishes.
func runContainerOutput(ctx context.Context, stream io.Writer, args ...string) (string, error) {
	rt, err := containerRuntime()
	if err != nil {
		return "", err
	}
	vlogf("$ %s %s", rt, strings.Join(args, " "))

	var buf bytes.Buffer
	var w io.Writer = &buf
	if stream != nil {
		w = io.MultiWriter(&buf, stream)
	}
	cmd := exec.CommandContext(ctx, rt, args...)
	cmd.Stdout = w
	cmd.Stderr = w
	return buf.String(), cmd.Run()
}

// prefixWriter tags each line it forwards with a label so the live output of
// tapes rendering in parallel stays attributable. Partial writes are buffered
// until their newline arrives, and flush emits whatever the command left
// without a trailing newline.
type prefixWriter struct {
	prefix string
	mu     sync.Mutex
	buf    []byte
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		w.emit(string(w.buf[:i]))
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

func (w *prefixWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.buf) > 0 {
		w.emit(string(w.buf))
		w.buf = nil
	}
}

// emit writes one line; the caller holds the mutex. log.Printf is itself
// goroutine-safe, so lines from separate renders interleave but never tear.
func (w *prefixWriter) emit(line string) {
	if s := strings.TrimRight(line, "\r"); strings.TrimSpace(s) != "" {
		log.Printf("[%s] %s", w.prefix, s)
	}
}

// streamFor returns a live-output sink under -verbose, plus the flush that
// emits any unterminated trailing line. Quiet runs get a nil sink, which
// runContainerOutput treats as capture-only.
func streamFor(label string) (sink io.Writer, flush func()) {
	if !verbose {
		return nil, func() {}
	}
	pw := &prefixWriter{prefix: label}
	return pw, pw.flush
}

// buildDemoBinaries cross-compiles the single snap_input example binary for
// linux/amd64 into dist/snap_input — the vhs container has no Go
// toolchain, so the tapes run the prebuilt binary with the example name as
// its subcommand (`snap_input timepicker`).
func buildDemoBinaries(root string) error {
	out := filepath.Join(root, "dist", "snap_input")
	cmd := exec.Command("go", "build", "-o", out, "./examples/snap_input")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	vlogf("$ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o %s ./examples/snap_input",
		filepath.ToSlash(mustRel(root, out)))
	outb, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build examples/snap_input: %w\n%s", err, outb)
	}
	if s := strings.TrimSpace(string(outb)); s != "" {
		vlogf("[build] %s", s)
	}
	log.Printf("built %s", filepath.ToSlash(mustRel(root, out)))
	return nil
}

// cleanDemoBinaries removes the cross-compiled demo binary after rendering.
func cleanDemoBinaries(root string) {
	_ = os.Remove(filepath.Join(root, "dist", "snap_input"))
}

func mustRel(base, target string) string {
	r, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return r
}
