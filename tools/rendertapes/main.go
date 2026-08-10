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
// Publishing is paced, because Charm's hosting accepts a short burst and then
// refuses (the client prints a bare "EOF") and rejects gifs past a size cap.
// Neither limit is documented, so -publish-batch/-publish-pause/-max-gif-bytes
// expose them as flags. Oversized gifs are caught before anything is uploaded,
// refusals are retried with doubling backoff, and a gif whose bytes match the
// hash recorded next to its URL is skipped entirely — after the first run most
// publishes upload nothing.
//
// Usage, from the snap repo root:
//
//	go -C tools/rendertapes run . [-image ghcr.io/charmbracelet/vhs] [-workers N]
//	go -C tools/rendertapes run . -verbose   # stream container output live
//	go -C tools/rendertapes run . -publish   # render, upload changed, repoint docs
//	go -C tools/rendertapes run . -relink    # repoint docs from <gif>.url only
//
// This is a standalone module so tool-only dependencies never enter snap's
// library graph.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"time"
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

		// Charm's hosting neither documents its rate limit nor its size cap,
		// so these are tunable observations rather than a known contract.
		batch    = flag.Int("publish-batch", 4, "uploads before pausing (0 disables pacing)")
		pause    = flag.Duration("publish-pause", time.Minute, "wait between upload batches")
		attempts = flag.Int("publish-attempts", 4, "tries per gif, including the first")
		backoff  = flag.Duration("publish-backoff", 20*time.Second, "wait before the first retry; doubles each retry")
		maxGif   = flag.Int64("max-gif-bytes", 10<<20, "reject gifs larger than this before uploading (0 disables)")
		force    = flag.Bool("force-publish", false, "re-upload gifs whose bytes are unchanged since their recorded URL")
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
		return publishGifs(ctx, *imageRef, root, tapes, publishPolicy{
			batch:    *batch,
			pause:    *pause,
			attempts: *attempts,
			backoff:  *backoff,
			maxBytes: *maxGif,
			force:    *force,
		})
	}
	return nil
}

// gifLink ties one tape's gif to where markdown should point at it. rel is
// the repo-relative path docs used before anything was published; prev is the
// URL the last publish recorded (empty on a first run). Rewrites accept
// either as the "from" side, so a re-publish still finds links that no longer
// mention the local path.
type gifLink struct{ rel, url, prev string }

// publishPolicy bounds how hard the publisher leans on Charm's hosting. The
// service accepts a short burst and then starts refusing (the client prints a
// bare "EOF" and exits non-zero), and it rejects gifs past a size ceiling.
// Neither limit is documented, so every knob is a flag: these defaults match
// what the endpoint was observed to allow, not a published contract.
type publishPolicy struct {
	batch    int           // uploads before pausing
	pause    time.Duration // wait once a batch is done
	attempts int           // total tries per gif, including the first
	backoff  time.Duration // wait before retry 1; doubles each retry
	maxBytes int64         // per-gif ceiling the endpoint enforces
	force    bool          // re-upload even when the gif is byte-identical
}

// publishGifs uploads every rendered gif to Charm's hosting via `vhs publish`
// (gifs stay out of git — see .gitignore — keeping clones small), records
// each URL in <name>.gif.url beside the tape, and repoints markdown links.
//
// Uploads are paced rather than fired off back to back, and gifs whose bytes
// have not changed since their recorded upload are skipped: the rate limit is
// the scarce resource here, so the cheapest request is the one not made.
func publishGifs(ctx context.Context, imageRef, root string, tapes []string, pol publishPolicy) error {
	gifs, err := tapeGifs(root, tapes)
	if err != nil {
		return err
	}

	pending, links, err := planPublish(root, gifs, pol)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		log.Printf("all %d gif(s) already published and unchanged", len(gifs))
		return rewriteDocLinks(root, links)
	}
	log.Printf("publishing %d of %d gif(s); %d unchanged",
		len(pending), len(gifs), len(gifs)-len(pending))

	for i, job := range pending {
		// Pace between batches, not before the first upload.
		if i > 0 && pol.batch > 0 && i%pol.batch == 0 {
			log.Printf("uploaded %d; pausing %s before the next batch", i, pol.pause)
			if err := sleepCtx(ctx, pol.pause); err != nil {
				return err
			}
		}
		url, err := publishOne(ctx, imageRef, root, job.tg.gif, pol)
		if err != nil {
			return fmt.Errorf("%w\n\n%d of %d uploaded; recorded URLs are on disk, so "+
				"re-running resumes from here", err, i, len(pending))
		}
		if err := writeRecord(job.urlFile, url, job.hash); err != nil {
			return err
		}
		links = append(links, gifLink{rel: job.tg.gif, url: url, prev: job.prev})
		log.Printf("published %s -> %s", job.tg.gif, url)
	}
	return rewriteDocLinks(root, links)
}

// publishJob is one gif that needs uploading, with the record fields resolved
// up front so the upload loop only does I/O that can fail transiently.
type publishJob struct {
	tg      tapeGif
	urlFile string
	hash    string // sha256 of the gif about to be uploaded
	prev    string // URL the last upload recorded, "" on a first run
}

// planPublish splits the gifs into those needing an upload and those already
// published unchanged, and validates every one before a single byte is sent —
// a missing or oversized gif midway through would otherwise burn rate limit
// and leave the docs half rewritten.
func planPublish(root string, gifs []tapeGif, pol publishPolicy) (pending []publishJob, links []gifLink, err error) {
	var missing, oversized []string
	for _, tg := range gifs {
		abs := filepath.Join(root, filepath.FromSlash(tg.gif))
		info, statErr := os.Stat(abs)
		if statErr != nil {
			missing = append(missing, tg.gif)
			continue
		}
		if pol.maxBytes > 0 && info.Size() > pol.maxBytes {
			oversized = append(oversized, fmt.Sprintf("%s (%.2f MiB, limit %.2f MiB)",
				tg.gif, mib(info.Size()), mib(pol.maxBytes)))
			continue
		}
		hash, hashErr := fileHash(abs)
		if hashErr != nil {
			return nil, nil, hashErr
		}
		urlFile := urlPath(root, tg)
		prevURL, prevHash := readRecord(urlFile)

		// Unchanged and already hosted: keep the URL, skip the upload, but
		// still feed the link into the rewrite so a doc that lost it recovers.
		if !pol.force && prevURL != "" && prevHash == hash {
			links = append(links, gifLink{rel: tg.gif, url: prevURL, prev: prevURL})
			continue
		}
		pending = append(pending, publishJob{tg: tg, urlFile: urlFile, hash: hash, prev: prevURL})
	}
	if len(missing) > 0 {
		return nil, nil, fmt.Errorf("gif(s) not rendered — run without -publish first:\n  %s",
			strings.Join(missing, "\n  "))
	}
	if len(oversized) > 0 {
		return nil, nil, fmt.Errorf("gif(s) over the hosting size limit; shrink the tape "+
			"(smaller Set Width/Height/FontSize, fewer or shorter Sleep steps):\n  %s",
			strings.Join(oversized, "\n  "))
	}
	return pending, links, nil
}

// publishOne uploads a single gif, retrying on the transient refusals the
// endpoint answers a burst with. A run that produces no URL is treated as a
// failure even when the client exits 0, because that is how a refused upload
// presents: the "EOF" line is printed where the link should be.
func publishOne(ctx context.Context, imageRef, root, gif string, pol publishPolicy) (string, error) {
	wait := pol.backoff
	var last error
	for attempt := 1; attempt <= max(pol.attempts, 1); attempt++ {
		if attempt > 1 {
			log.Printf("retry %d/%d for %s in %s (%v)", attempt-1, pol.attempts-1, gif, wait, last)
			if err := sleepCtx(ctx, wait); err != nil {
				return "", err
			}
			wait *= 2
		}
		stream, flush := streamFor("publish " + gif)
		out, err := runContainerOutput(ctx, stream,
			"run", "--rm", "-v", root+":/vhs", imageRef, "publish", gif)
		flush()

		if url := publishedURL(out); url != "" {
			return url, nil
		}
		if err != nil {
			last = fmt.Errorf("vhs publish %s: %w: %s", gif, err, firstLine([]byte(out)))
		} else {
			last = fmt.Errorf("vhs publish %s: no gif URL in output: %s", gif, firstLine([]byte(out)))
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
	}
	return "", fmt.Errorf("%w (gave up after %d attempts)", last, max(pol.attempts, 1))
}

// sleepCtx waits for d unless the context is canceled first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func mib(n int64) float64 { return float64(n) / (1 << 20) }

// fileHash is the gif's sha256, recorded alongside its URL so an unchanged
// gif can be recognized and skipped on the next publish.
func fileHash(gif string) (string, error) {
	f, err := os.Open(filepath.Clean(gif))
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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
		url, _ := readRecord(urlPath(root, tg))
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

// readRecord parses a <name>.gif.url file. Line one is the hosted URL and is
// the only line docs or a human need; an optional `sha256:<hex>` line records
// which bytes were uploaded, letting the next publish skip an unchanged gif.
// Both are "" when the file is absent, so a first run just publishes.
func readRecord(urlFile string) (url, hash string) {
	//nolint:gosec // url paths are derived from tape paths inside the repo.
	b, err := os.ReadFile(urlFile)
	if err != nil {
		return "", ""
	}
	for line := range strings.Lines(string(b)) {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
		case strings.HasPrefix(line, "sha256:"):
			hash = strings.TrimPrefix(line, "sha256:")
		case url == "":
			url = line
		}
	}
	return url, hash
}

// writeRecord stores the hosted URL and the hash of the bytes behind it.
func writeRecord(urlFile, url, hash string) error {
	return os.WriteFile(urlFile, []byte(url+"\nsha256:"+hash+"\n"), 0o600)
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
