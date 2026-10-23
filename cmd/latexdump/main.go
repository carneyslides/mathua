// latexdump walks data/lessons and prints every lesson body after LaTeX
// canonicalization — exactly what the API serves — as JSON. Used by the
// frontend corpus gate (web/next-app/test/latexCorpus.test.ts) so KaTeX is
// tested against production content rather than raw scrape dialects.
//
// With -spans it additionally prints every math region, as classified by
// latexnorm.Scan over the canonicalized body. That is what
// scripts/normalize-latex.mjs feeds to Compute Engine: it is the only way to
// know which dollars are math delimiters and which are currency, and running
// the parser over the raw scrape would report every price as a syntax error.
// Spans come from the canonicalized text, so they are the spans a learner's
// browser actually receives.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/chuma-beep/latexnorm"
	"github.com/chuma-beep/mathua/internal/latex"
)

// span is one math region of one lesson, with enough context to find it again
// in the file and enough of it to recognise in a report.
type span struct {
	File    string `json:"file"`
	Index   int    `json:"index"`
	Kind    string `json:"kind"`
	Opener  string `json:"opener"`
	Closer  string `json:"closer"`
	Content string `json:"content"`
	// Byte offset of the region in the canonicalized body, so a report can be
	// traced back to the served content rather than the file on disk.
	Offset int `json:"offset"`
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("latexdump", flag.ContinueOnError)
	flags.SetOutput(stderr)
	spansOnly := flags.Bool("spans", false, "emit math spans as JSONL instead of a JSON map of bodies")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	// Positional, as before: `latexdump [root]`.
	root := "data/lessons"
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}

	walked := 0
	found := 0
	out := map[string]string{}

	writeSpan := func(rel, content string) error {
		for i, r := range latexnorm.Scan(content) {
			if r.Kind == latexnorm.Prose {
				continue
			}
			found++
			enc := json.NewEncoder(stdout)
			if err := enc.Encode(span{
				File:    rel,
				Index:   i,
				Kind:    r.Kind.String(),
				Opener:  r.Opener,
				Closer:  r.Closer,
				Content: r.Content,
				Offset:  r.Start,
			}); err != nil {
				return err
			}
		}
		return nil
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		walked++
		content := latex.Canonicalize(string(raw), latex.ForSource(filepath.ToSlash(rel), string(raw)))
		if *spansOnly {
			return writeSpan(filepath.ToSlash(rel), content)
		}
		out[filepath.ToSlash(rel)] = content
		return nil
	})
	if err != nil {
		fmt.Fprintln(stderr, "latexdump:", err)
		return 1
	}

	if *spansOnly {
		fmt.Fprintf(stderr, "latexdump: %d files, %d math spans\n", walked, found)
		return 0
	}
	enc := json.NewEncoder(stdout)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(stderr, "latexdump encode:", err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
