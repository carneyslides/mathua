package solutions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func load(t *testing.T, files map[string]string) *Loader {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		write(t, dir, name, body)
	}
	l, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return l
}

func facts() map[string]string {
	return map[string]string{"a_num": "3", "a_den": "4", "b_num": "1", "b_den": "2", "lcm": "4", "b_scaled": "2", "num": "5", "answer": "5/4"}
}

const fracSchema = `{
  "concept": "frac.add.diff",
  "setup": "You are adding $[[a_num]]/[[a_den]]$ and $[[b_num]]/[[b_den]]$.",
  "steps": [
    {"fact": "lcm", "say": "Put both over LCM([[a_den]], [[b_den]]) = [[lcm]]."},
    {"fact": "b_scaled", "say": "Then $[[b_num]]/[[b_den]]$ becomes $[[b_scaled]]/[[lcm]]$."}
  ],
  "answer": "Adding the tops gives [[num]]/[[lcm]]."
}`

func TestAssemble_InterpolatesEveryFact(t *testing.T) {
	l := load(t, map[string]string{"frac.add.diff.json": fracSchema})
	got := l.Assemble("frac.add.diff", facts(), "FALLBACK")
	if got == "FALLBACK" {
		t.Fatal("schema was skipped; it should have resolved")
	}
	for _, want := range []string{
		"You are adding $3/4$ and $1/2$.",
		"Put both over LCM(4, 2) = 4.",
		"Then $1/2$ becomes $2/4$.",
		"Adding the tops gives 5/4.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	// No placeholder may survive into what a learner reads.
	if strings.Contains(got, "[[") {
		t.Errorf("unsubstituted placeholder left in output:\n%s", got)
	}
}

func TestAssemble_SkipsRatherThanGuessing(t *testing.T) {
	cases := map[string]struct {
		files   map[string]string
		concept string
		f       map[string]string
		why     string
	}{
		"no schema for the concept": {
			files:   map[string]string{"other.json": strings.Replace(fracSchema, `"concept": "frac.add.diff"`, `"concept": "frac.sub.diff"`, 1)},
			concept: "frac.add.diff",
			f:       facts(),
			why:     "a concept without a schema must keep the generator's explanation",
		},
		"placeholder has no fact": {
			files:   map[string]string{"frac.add.diff.json": fracSchema},
			concept: "frac.add.diff",
			f:       map[string]string{"a_num": "3", "a_den": "4", "b_num": "1", "b_den": "2", "lcm": "4", "b_scaled": "2"},
			why:     "a missing num must not render an empty result",
		},
		"step fact was renamed in the generator": {
			files:   map[string]string{"frac.add.diff.json": fracSchema},
			concept: "frac.add.diff",
			f:       map[string]string{"a_num": "3", "a_den": "4", "b_num": "1", "b_den": "2", "num": "5", "answer": "5/4"},
			why:     "a step whose fact is gone has lost its footing",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			l := load(t, tc.files)
			if got := l.Assemble(tc.concept, tc.f, "FALLBACK"); got != "FALLBACK" {
				t.Errorf("expected the fallback, got:\n%s\n(%s)", got, tc.why)
			}
		})
	}
}

// One generator often asks two different questions — find the complement of an
// angle, or classify it — and a single schema covers both. Each question's
// steps must apply only to the question it came from, without the other
// question's steps leaking in or the schema being dropped.
func TestRender_SkipsStepsWhoseFactIsAbsent(t *testing.T) {
	s := &Schema{
		Setup: "An angle of [[given]] degrees.",
		Steps: []Step{
			{Fact: "missing", Say: "A right angle is 90, so the other is 90 - [[given]] = [[missing]]."},
			{Fact: "cls", Say: "[[given]] degrees falls in the range for [[cls]]."},
		},
		Answer: "So the answer is [[answer]].",
	}
	complement, missing := Render(s, map[string]string{"given": "40", "missing": "50", "answer": "50"})
	if len(missing) > 0 {
		t.Fatalf("unexpected missing: %v", missing)
	}
	if strings.Contains(complement, "falls in the range") {
		t.Errorf("the classification step leaked into the complement question:\n%s", complement)
	}
	classify, missing := Render(s, map[string]string{"given": "120", "cls": "obtuse", "answer": "obtuse"})
	if len(missing) > 0 {
		t.Fatalf("unexpected missing: %v", missing)
	}
	if strings.Contains(classify, "A right angle is 90") {
		t.Errorf("the complement step leaked into the classification question:\n%s", classify)
	}
	// The drift is still reported to the test even though the runtime copes.
	got := Missing(s, map[string]string{"given": "40", "answer": "50"})
	if len(got) != 2 || got[0] != "missing" || got[1] != "cls" {
		t.Errorf("Missing should report both absent step facts, got %v", got)
	}
}

func TestAssemble_NilLoaderIsTheFallback(t *testing.T) {
	var l *Loader
	if got := l.Assemble("anything", facts(), "FALLBACK"); got != "FALLBACK" {
		t.Errorf("a nil loader must fall back, got %q", got)
	}
}

// LaTeX braces are everywhere in this corpus, so they cannot be the
// placeholder delimiter. A schema full of \frac must survive untouched.
func TestRender_LeavesLatexBracesAlone(t *testing.T) {
	s := &Schema{
		Steps:  []Step{{Fact: "m", Say: `The matrix $\begin{pmatrix} 1 \\ 0 \end{pmatrix}$ has [[m]] rows.`}},
		Answer: "So [[answer]].",
	}
	got, missing := Render(s, map[string]string{"m": "2", "answer": "done"})
	if len(missing) > 0 {
		t.Fatalf("unexpected missing: %v", missing)
	}
	if !strings.Contains(got, `\begin{pmatrix} 1 \\ 0 \end{pmatrix}`) {
		t.Errorf("LaTeX was mangled:\n%s", got)
	}
}

func TestLoad_RejectsUnusableSchemas(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "good.json", fracSchema)
	write(t, dir, "no-answer.json", `{"concept":"x","setup":"hi"}`)
	write(t, dir, "broken.json", `{"concept":`)
	l, err := Load(dir)
	if err == nil {
		t.Fatal("expected an error naming the unusable schemas")
	}
	for _, want := range []string{"no-answer.json", "broken.json"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should name %s: %v", want, err)
		}
	}
	// A bad file must not take the good ones down with it.
	if _, ok := l.Schema("frac.add.diff"); !ok {
		t.Error("a usable schema should still load alongside a broken one")
	}
}

func TestLoad_MissingDirectoryIsEmpty(t *testing.T) {
	l, err := Load(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("a missing directory must not be an error: %v", err)
	}
	if len(l.Concepts()) != 0 {
		t.Errorf("expected no concepts, got %v", l.Concepts())
	}
}

// The row-echelon form a systems question is graded against is written
// [[a,b,c],[0,d,e]]. An unrestricted [[...]] placeholder would eat that, so
// contents must look like a fact name to be treated as one.
func TestRender_LeavesMatrixNotationAlone(t *testing.T) {
	s := &Schema{
		Steps:  []Step{{Fact: "ans", Say: "The echelon form is [[2,1,7],[0,3,5]]."}},
		Answer: "So the answer is [[answer]].",
	}
	got, missing := Render(s, map[string]string{"ans": "ok", "answer": "[[4,3,6],[0,1,1]]"})
	if len(missing) > 0 {
		t.Fatalf("unexpected missing: %v", missing)
	}
	if !strings.Contains(got, "[[2,1,7],[0,3,5]]") {
		t.Errorf("matrix notation in a step was eaten:\n%s", got)
	}
	if !strings.Contains(got, "[[4,3,6],[0,1,1]]") {
		t.Errorf("matrix notation in the answer was eaten:\n%s", got)
	}
	// A capitalised name is not a fact name either, so it survives verbatim.
	// The leak test is what catches that typo, not the assembler — reporting it
	// here would mean reporting every matrix span as missing.
	got, missing = Render(&Schema{Answer: "[[Num]]"}, map[string]string{"Num": "7"})
	if len(missing) > 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
	if !strings.Contains(got, "[[Num]]") {
		t.Errorf("a capitalised name should be left literal, got %q", got)
	}
}

func TestUnresolved(t *testing.T) {
	// Case 1: Placeholders extracted cleanly
	text := "Compute [[first_var]] plus [[second_var]] to get result."
	unresolved := Unresolved(text)
	if len(unresolved) != 2 || unresolved[0] != "first_var" || unresolved[1] != "second_var" {
		t.Errorf("expected [first_var second_var], got %v", unresolved)
	}

	// Case 2: Matrix notation and unclosed brackets ignored
	matrixText := "Matrix [[1,2],[3,4]] alongside unclosed [[trailing"
	if got := Unresolved(matrixText); len(got) != 0 {
		t.Errorf("expected no placeholders in matrix text, got %v", got)
	}

	// Case 3: Plain text without brackets
	if got := Unresolved("plain text with no brackets"); len(got) != 0 {
		t.Errorf("expected empty slice for plain text, got %v", got)
	}
}

func TestLoader_NilAccessors(t *testing.T) {
	var l *Loader
	if c := l.Concepts(); c != nil {
		t.Errorf("expected nil concepts from nil loader, got %v", c)
	}
	if s, ok := l.Schema("any"); ok || s != nil {
		t.Errorf("expected nil schema from nil loader, got %v, %v", s, ok)
	}
}

func TestLoad_DuplicateAndInferredID(t *testing.T) {
	dir := t.TempDir()

	// Inferred concept ID from filename when concept field is omitted
	inferredSchema := `{"setup":"hello","answer":"world"}`
	write(t, dir, "inferred.concept.json", inferredSchema)

	// Duplicate concept schema
	dup1 := `{"concept":"shared.id","answer":"first"}`
	dup2 := `{"concept":"shared.id","answer":"second"}`
	write(t, dir, "dup1.json", dup1)
	write(t, dir, "dup2.json", dup2)

	l, err := Load(dir)
	if err == nil {
		t.Fatal("expected error on duplicate schema, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate schema for \"shared.id\"") {
		t.Errorf("expected duplicate error message, got %v", err)
	}

	// Inferred concept is still loaded in loader
	if s, ok := l.Schema("inferred.concept"); !ok || s == nil || s.Answer != "world" {
		t.Errorf("expected inferred.concept loaded, got %+v (ok=%v)", s, ok)
	}
}
func TestLoader_NilReceiverAndMissingCases(t *testing.T) {
	var l *Loader
	if got := l.Concepts(); got != nil {
		t.Errorf("nil Loader.Concepts() = %v, want nil", got)
	}
	if s, ok := l.Schema("any"); ok || s != nil {
		t.Errorf("nil Loader.Schema() = (%v, %v), want (nil, false)", s, ok)
	}
	if got := l.Assemble("any", nil, "fallback"); got != "fallback" {
		t.Errorf("nil Loader.Assemble() = %q, want 'fallback'", got)
	}

	valid := load(t, map[string]string{
		"concept_b.json": `{"concept":"concept_b","answer":"done_b"}`,
		"concept_a.json": `{"concept":"concept_a","answer":"done_a"}`,
	})
	concepts := valid.Concepts()
	if len(concepts) != 2 || concepts[0] != "concept_a" || concepts[1] != "concept_b" {
		t.Errorf("Concepts() = %v, want [concept_a concept_b]", concepts)
	}

	if got := valid.Assemble("nonexistent", nil, "fallback"); got != "fallback" {
		t.Errorf("nonexistent Assemble() = %q, want 'fallback'", got)
	}

	// Schema with no setup and multiple steps without fact conditions
	noSetupSchema := &Schema{
		Steps: []Step{
			{Say: "Step one"},
			{Say: "Step two"},
		},
		Answer: "Final answer",
	}
	rendered, missing := Render(noSetupSchema, map[string]string{})
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
	if !strings.Contains(rendered, "Step one\n\nStep two") {
		t.Errorf("rendered without setup =\n%s", rendered)
	}
}

func TestLoad_SubdirectoriesAndDuplicates(t *testing.T) {
	dir := t.TempDir()
	// Subdirectory must be skipped
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Non-json file must be skipped
	write(t, dir, "notes.txt", "some notes")

	// Schema without explicit "concept" falls back to filename stem
	write(t, dir, "implicit_id.json", `{"answer":"implicit answer"}`)

	// Two schemas declaring the same concept ID must yield duplicate error
	write(t, dir, "dup1.json", `{"concept":"dup_id","answer":"first"}`)
	write(t, dir, "dup2.json", `{"concept":"dup_id","answer":"second"}`)

	l, err := Load(dir)
	if err == nil {
		t.Fatal("expected error on duplicate concept schema")
	}
	if !strings.Contains(err.Error(), "duplicate schema for \"dup_id\"") {
		t.Errorf("error = %v, want duplicate mention", err)
	}

	// Implicit ID must have loaded properly
	if s, ok := l.Schema("implicit_id"); !ok || s.Answer != "implicit answer" {
		t.Errorf("expected implicit_id to load, got ok=%v s=%+v", ok, s)
	}
}

func TestLoad_ReadDirError(t *testing.T) {
	// Passing a regular file instead of a directory triggers os.ReadDir error
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("expected ReadDir error when path is not a directory")
	}
	if !strings.Contains(err.Error(), "read solutions dir") {
		t.Errorf("error = %v, want read solutions dir prefix", err)
	}
}

func TestLookup_EdgeCases(t *testing.T) {
	noted := []string{}
	noteFn := func(name string) {
		noted = append(noted, name)
	}

	// Unclosed placeholder bracket
	text := "Formula with [[unclosed bracket text"
	if got := lookup(text, nil, noteFn); got != text {
		t.Errorf("unclosed lookup = %q, want %q", got, text)
	}
	if len(noted) != 0 {
		t.Errorf("unexpected note: %v", noted)
	}
}
