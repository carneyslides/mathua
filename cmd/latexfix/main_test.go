package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuma-beep/latexnorm"
	"github.com/chuma-beep/mathua/internal/latex"
)

// The first version of this tool stripped the `$…$` off every span it touched,
// turning rendered math into literal text across 33 files. It read like content
// loss and it was invisible in the pass/fail count, because the spans still
// "existed" — they were just prose now. These tests exist to make that impossible
// to reintroduce.

func TestRepairPreservesDelimiters(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"inline math keeps its dollars", `$f✓(x)$`, `$f(x)$`},
		// Removing the tick leaves a bare numeral, which must stop being math
		// altogether or the canonicalizer escapes it to \$42\$ (ADR-016).
		{"bare numeral loses its delimiters", `$42✓$`, `42`},
		{"repaired expression keeps its math", `$2 + 7 = 9✓$`, `$2 + 7 = 9$`},
		{"display math keeps its double dollars", "$$\\frac{1}{2}✓$$", "$$\\frac{1}{2}$$"},
		{"paren math keeps its backslashes", `\(f’(x)\)`, `\(f'(x)\)`},
		{"bracket math keeps its brackets", `\[f’(x)\]`, `\[f'(x)\]`},
		{"two spans on one line both survive", `$a✓$ and $b✓$`, `$a$ and $b$`},
		{"untouched text is byte-identical", `no math here at all`, `no math here at all`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _ := repair(c.in)
			if got != c.want {
				t.Errorf("repair(%q)\n got %q\nwant %q", c.in, got, c.want)
			}
		})
	}
}

// Prose inside math is prose. `\overset{\text{—}}{16}` means "16 repeats"; a blind
// dash rewrite would silently change what the lesson says.
func TestRepairLeavesTextGroupsAlone(t *testing.T) {
	in := `$\overset{\text{—}}{16} + f’(x)$`
	want := `$\overset{\text{—}}{16} + f'(x)$`
	got, _ := repair(in)
	if got != want {
		t.Errorf("repair(%q)\n got %q\nwant %q", in, got, want)
	}
}

func TestRepairLeavesNestedTextGroupsIntact(t *testing.T) {
	in := `$\text{a {b} ’c} + d’$`
	want := `$\text{a {b} ’c} + d'$`
	got, _ := repair(in)
	if got != want {
		t.Errorf("repair(%q)\n got %q\nwant %q", in, got, want)
	}
}

// The corpus is mostly English. A typographic apostrophe in prose is correct, and
// the whole reason this tool scans regions instead of doing find-and-replace.
func TestRepairDoesNotTouchProse(t *testing.T) {
	in := `John’s theorem states that \(x\) is nice. He said “yes” — twice.`
	got, changes := repair(in)
	if got != in {
		t.Errorf("prose was modified:\n got %q\nwant %q", got, in)
	}
	if len(changes) != 0 {
		t.Errorf("expected no changes in prose, got %d", len(changes))
	}
}

// Repair must not shift byte offsets, because the report quotes them.
func TestRepairIsStable(t *testing.T) {
	in := `Intro $a✓$ middle $b’$ end $c$`
	once, _ := repair(in)
	twice, changes := repair(once)
	if once != twice {
		t.Errorf("repair is not idempotent:\n once %q\ntwice %q", once, twice)
	}
	if len(changes) != 0 {
		t.Errorf("second pass still changed %d spans", len(changes))
	}
}

func TestTextGroupFindsNested(t *testing.T) {
	got := textGroup(`\text{a {b} c} x`)
	if len(got) != 1 {
		t.Fatalf("expected 1 group, got %d: %v", len(got), got)
	}
	if got[0][1] != len(`\text{a {b} c}`) {
		t.Errorf("group end = %d, want %d", got[0][1], len(`\text{a {b} c}`))
	}
}

// Every file must still scan into the same regions afterwards, or the repair
// silently restructured the document rather than editing it.
func TestRepairKeepsRegionCount(t *testing.T) {
	in := "Prose with $x$ and\n\n$$y$$\n\nmore \\(z\\) end $w✓$."
	out, _ := repair(in)
	before := len(latexnorm.Scan(in))
	after := len(latexnorm.Scan(out))
	if before != after {
		t.Errorf("region count changed: %d → %d", before, after)
	}
	if strings.Count(in, "$") != strings.Count(out, "$") {
		t.Errorf("dollar count changed: %d → %d", strings.Count(in, "$"), strings.Count(out, "$"))
	}
}

// The distinction that matters most: a derivative's apostrophe is a typo, a
// quotation mark is not. Rewriting the second to `'` renders primes and silently
// changes what the lesson teaches.
func TestRepairDistinguishesApostropheFromQuotation(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"derivative", `$f’(x) = 18x - 4$`, `$f'(x) = 18x - 4$`},
		{"subscripted derivative", `$D’(x_0)$`, `$D'(x_0)$`},
		{"after a closing brace", `$\sigma’(x)$`, `$\sigma'(x)$`},
		{"closing quotation mark is left alone", `{a < b ‘a\ \text{is less than}\ b’}`, `{a < b ‘a\ \text{is less than}\ b’}`},
		{"quote around a symbol still normalises to ASCII", `$‘x’$`, `$‘x'$`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, _ := repair(c.in)
			if got != c.want {
				t.Errorf("repair(%q)\n got %q\nwant %q", c.in, got, c.want)
			}
		})
	}
}

// The 14,936 doubled delimiters put every Algebrica inline formula in a *prose*
// region, so no region-based tool could see them. Un-escaping is the fix, and it
// is only safe because canonicalization erases the difference. That is the
// invariant, asserted directly rather than assumed.
func TestUnescapingDoesNotChangeWhatIsServed(t *testing.T) {
	cases := []string{
		`denoted by \\( f'(c) \\), is defined`,
		`\\[f'(c) = 5 \\]`,
		`mixed \\( a \\) and $b$ and $$c$$`,
		`no math at all, just prose with a backslash \\( misdirection`,
		`\begin{align} x &= 1 \\[2pt] y &= 2 \end{align}`,
	}
	for _, in := range cases {
		out, _ := repair(in)
		before := latex.Canonicalize(in, latex.ForSource("t.md", in))
		after := latex.Canonicalize(out, latex.ForSource("t.md", out))
		if before != after {
			t.Errorf("served output changed\n in: %q\nout: %q\nbefore: %q\n after: %q", in, out, before, after)
		}
	}
}

func TestUnescapeRewritesInlineDelimitersOnly(t *testing.T) {
	cases := []struct{ in, want string }{
		{`\\( x \\)`, `\( x \)`},
		{`\\[2pt] row break`, `\\[2pt] row break`}, // a row break, not a delimiter
		{`\\ alone`, `\\ alone`},
	}
	for _, c := range cases {
		got, _ := unescapeDelimiters(c.in)
		if got != c.want {
			t.Errorf("unescapeDelimiters(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Un-escaping is what makes the math visible, so a doubled-delimiter formula must
// come out repaired rather than merely re-spelled.
func TestUnescapingExposesFormulasToTheFixers(t *testing.T) {
	in := `denoted by \\( f’(c) \\), is defined`
	want := `denoted by \( f'(c) \), is defined`
	got, changes := repair(in)
	if got != want {
		t.Errorf("repair(%q)\n got %q\nwant %q", in, got, want)
	}
	if len(changes) == 0 {
		t.Error("expected changes to be reported")
	}
}

func TestUnescapeDistinguishesDisplayFromRowBreak(t *testing.T) {
	cases := []struct{ in, want string }{
		{`\\[ x = 1 \\]`, `\[ x = 1 \]`},
		{`\begin{align} x &= 1 \\[2pt] y &= 2 \end{align}`, `\begin{align} x &= 1 \\[2pt] y &= 2 \end{align}`},
		{`\begin{align} a \\[0.5em] b \end{align}`, `\begin{align} a \\[0.5em] b \end{align}`},
		{`\begin{align} a \\[] b \end{align}`, `\begin{align} a \\[] b \end{align}`},
		// `\\[1em]` is a row break with 1em of space, not a display opener, even
		// though more content follows. The corpus has 1,157 of these.
		{`a \\[1em] b \\[2pt] c`, `a \\[1em] b \\[2pt] c`},
	}
	for _, c := range cases {
		got, _ := unescapeDelimiters(c.in)
		if got != c.want {
			t.Errorf("unescapeDelimiters(%q)\n got %q\nwant %q", c.in, got, c.want)
		}
	}
}

func TestRun_DryRunAndWrite(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "lesson.md")
	content := "# Lesson\n\nHere is a formula $f✓(x)$ in prose.\n"
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	// Dry run mode (default)
	var stdout, stderr bytes.Buffer
	code := run([]string{dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("dry run exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "repaired math span") {
		t.Errorf("expected 'repaired math span' in stdout, got:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "dry run (pass -w to write)") {
		t.Errorf("expected dry run notice in stderr, got:\n%s", stderr.String())
	}

	// File on disk must remain unchanged in dry run
	data, _ := os.ReadFile(filePath)
	if string(data) != content {
		t.Fatalf("dry run altered file content")
	}

	// In-place write mode (-w)
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"-w", dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("write exit code = %d, want 0\nstderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "written") {
		t.Errorf("expected 'written' in stderr, got:\n%s", stderr.String())
	}

	// File on disk must now be updated
	data, _ = os.ReadFile(filePath)
	expected := "# Lesson\n\nHere is a formula $f(x)$ in prose.\n"
	if string(data) != expected {
		t.Fatalf("file content = %q, want %q", string(data), expected)
	}
}

func TestRun_FlagAndPathErrors(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// Invalid flag returns code 2
	code := run([]string{"-invalid-flag-xyz"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("invalid flag code = %d, want 2", code)
	}

	// Nonexistent root path returns code 1
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"/nonexistent/directory/path/surely"}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("nonexistent path code = %d, want 1", code)
	}
}

func TestSnippet_LengthBounds(t *testing.T) {
	short := "short text"
	if got := snippet(short, 50); got != short {
		t.Errorf("snippet short = %q, want %q", got, short)
	}

	long := "a very long sentence with many words that exceeds the maximum length"
	got := snippet(long, 20)
	if !strings.Contains(got, "…") {
		t.Errorf("snippet long = %q, want ellipsis", got)
	}
}
