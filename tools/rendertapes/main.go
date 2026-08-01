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
// Usage, from the snap repo root:
//
//	go -C tools/rendertapes run . [-image ghcr.io/charmbracelet/vhs] [-workers N]
//
// This is a standalone module so tool-only dependencies never enter snap's
// library graph.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

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
	)
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
		return publishGifs(ctx, *imageRef, root)
	}
	return nil
}

// publishGifs uploads every rendered gif to Charm's hosting via `vhs publish`
// (gifs stay out of git — see .gitignore — keeping clones small), records
// each URL in <gif>.url beside the tape, and rewrites markdown image links
// from the repo-relative gif path to the hosted URL.
func publishGifs(ctx context.Context, imageRef, root string) error {
	gifs, err := filepath.Glob(filepath.Join(root, "examples", "*", "*.gif"))
	if err != nil {
		return err
	}
	urls := map[string]string{}
	for _, gif := range gifs {
		rel := filepath.ToSlash(mustRel(root, gif))
		out, err := runDockerOutput(ctx,
			"run", "--rm", "-v", root+":/vhs", imageRef, "publish", rel)
		if err != nil {
			return fmt.Errorf("vhs publish %s: %w\n%s", rel, err, strings.TrimSpace(out))
		}
		url := ""
		for ln := range strings.FieldsSeq(out) {
			if strings.HasPrefix(ln, "https://") {
				url = ln
			}
		}
		if url == "" {
			return fmt.Errorf("vhs publish %s: no URL in output:\n%s", rel, strings.TrimSpace(out))
		}
		if err := os.WriteFile(gif+".url", []byte(url+"\n"), 0o600); err != nil {
			return err
		}
		urls[rel] = url
		log.Printf("published %s -> %s", rel, url)
	}
	return rewriteDocLinks(root, urls)
}

// rewriteDocLinks points markdown image references at the hosted URLs.
func rewriteDocLinks(root string, urls map[string]string) error {
	docs, _ := filepath.Glob(filepath.Join(root, "*.md"))
	more, _ := filepath.Glob(filepath.Join(root, "docs", "*.md"))
	docs = append(docs, more...)
	docs = append(docs, filepath.Join(root, "examples", "USAGE.md"))
	for _, doc := range docs {
		//nolint:gosec // doc paths come from our own repo glob above.
		b, err := os.ReadFile(doc)
		if err != nil {
			continue
		}
		s := string(b)
		for rel, url := range urls {
			s = strings.ReplaceAll(s, "("+rel+")", "("+url+")")
		}
		if s != string(b) {
			// Write only inside the repo root (the globs above guarantee it;
			// this check makes that explicit for taint analysis).
			if rel, relErr := filepath.Rel(root, doc); relErr != nil || strings.HasPrefix(rel, "..") {
				continue
			}
			//nolint:gosec // doc is validated to live inside the repo root.
			if err := os.WriteFile(doc, []byte(s), 0o600); err != nil {
				return err
			}
			log.Printf("updated links in %s", filepath.ToSlash(mustRel(root, doc)))
		}
	}
	return nil
}

// findTapes walks the repo for every *.tape file (any name, any depth —
// a package can ship several, e.g. charts/sparkline.tape next to
// charts/pie.tape), skipping VCS and tool directories.
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

// ensureImage pulls the VHS image when it is not already present.
func ensureImage(ctx context.Context, ref string) error {
	if err := runDocker(ctx, "image", "inspect", ref); err == nil {
		return nil
	}
	return runDocker(ctx, "pull", ref)
}

// renderTape runs one tape in its own container with the repo mounted at
// /vhs (the image's working directory), mirroring
// `docker run --rm -v <root>:/vhs ghcr.io/charmbracelet/vhs <tape>`.
func renderTape(ctx context.Context, imageRef, root, relTape string) error {
	// The VHS container's terminal advertises TERM=xterm-256color, so
	// lipgloss downsamples true-color backgrounds. COLORTERM=truecolor
	// upgrades the profile so SGR colors match the page while honoring NO_COLOR.
	out, err := runDockerOutput(
		ctx,
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

func runDocker(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		podman := exec.CommandContext(ctx, "podman", args...)
		podman.Stdout = io.Discard
		podman.Stderr = io.Discard
		if pErr := podman.Run(); pErr == nil {
			return nil
		}
		return err
	}
	return nil
}

func runDockerOutput(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), nil
	}
	podman := exec.CommandContext(ctx, "podman", args...)
	pout, pErr := podman.CombinedOutput()
	if pErr == nil {
		return string(pout), nil
	}
	if strings.TrimSpace(string(pout)) != "" {
		return string(pout), pErr
	}
	return string(out), err
}

// buildDemoBinaries cross-compiles the single snap_input example binary for
// linux/amd64 into examples/snap_input/demo-bin — the vhs container has no Go
// toolchain, so the tapes run the prebuilt binary with the example name as
// its subcommand (`demo-bin timepicker`).
func buildDemoBinaries(root string) error {
	out := filepath.Join(root, "examples", "snap_input", "demo-bin")
	cmd := exec.Command("go", "build", "-o", out, "./examples/snap_input")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	if outb, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build examples/snap_input: %w\n%s", err, outb)
	}
	log.Printf("built %s", filepath.ToSlash(mustRel(root, out)))
	return nil
}

// cleanDemoBinaries removes the cross-compiled demo binary after rendering.
func cleanDemoBinaries(root string) {
	_ = os.Remove(filepath.Join(root, "examples", "snap_input", "demo-bin"))
}

func mustRel(base, target string) string {
	r, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return r
}
