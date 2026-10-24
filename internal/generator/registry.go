package generator

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"

	"github.com/chuma-beep/mathua/internal/latex"
	"github.com/chuma-beep/mathua/internal/solutions"
)

type Registry struct {
	mu        sync.RWMutex
	gens      map[string]Generator
	solutions *solutions.Loader
	randMu    sync.Mutex
	// duplicates records every concept that a second Register call was refused for. Callers in
	// every generator package ignore Register's error, so a concept claimed by two packages used
	// to fail silently: one implementation became unreachable while its own tests kept passing
	// against the type directly. Two packages did exactly that with the six basic combinatorics
	// counts on 23 October 2026. Duplicates lets a test see the collision.
	duplicates []string
}

func NewRegistry() *Registry {
	return &Registry{
		gens: make(map[string]Generator),
	}
}

// SetSolutions attaches the corpus solution schemas. Optional: with none
// attached every problem keeps the generator's own Explanation, which is what
// the registry did before schemas existed.
func (r *Registry) SetSolutions(l *solutions.Loader) {
	r.mu.Lock()
	r.solutions = l
	r.mu.Unlock()
}

// explain assembles the explanation for a served problem: a corpus schema
// interpolated with the instance's facts where one exists, otherwise the
// generator's own Explanation. Interpolation happens before canonicalisation so
// that any LaTeX a fact contributes is normalised exactly like the question's.
func (r *Registry) explain(conceptID string, p Problem) string {
	r.mu.RLock()
	loader := r.solutions
	r.mu.RUnlock()
	if loader == nil {
		return latex.Canonicalize(p.Explanation, latex.Generators)
	}
	// Copy: the generator owns p.Facts and reuse of a Problem must not see
	// the always-present entries we add below.
	facts := make(map[string]string, len(p.Facts)+1)
	for k, v := range p.Facts {
		facts[k] = v
	}
	// The graded answer is a fact every schema may reference, so an authored
	// explanation never has to restate it in prose to close.
	facts["answer"] = p.Answer
	return latex.Canonicalize(loader.Assemble(conceptID, facts, p.Explanation), latex.Generators)
}

func (r *Registry) Register(conceptID string, gen Generator) error {
	if conceptID == "" {
		return fmt.Errorf("concept ID must not be empty")
	}
	if gen == nil {
		return fmt.Errorf("generator must not be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.gens[conceptID]; exists {
		r.duplicates = append(r.duplicates, conceptID)
		return fmt.Errorf("generator already registered for concept %q", conceptID)
	}
	// Self-check: if this is a GradedGenerator, make sure it can grade its own output.
	if gg, ok := gen.(GradedGenerator); ok {
		p := gg.Generate(GeneratorContext{Difficulty: 0.5})
		result := gg.Grade(p.Answer, p.Answer)
		if !result.Correct {
			return fmt.Errorf("generator for %q cannot grade its own answer (expected=%q): %s",
				conceptID, p.Answer, result.Feedback)
		}
		// Also verify it rejects a clearly wrong answer
		wrongResult := gg.Grade(p.Answer, "")
		if wrongResult.Correct {
			return fmt.Errorf("generator for %q accepted empty answer as correct", conceptID)
		}
	}
	r.gens[conceptID] = gen
	return nil
}

func (r *Registry) Get(conceptID string) (Generator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	gen, exists := r.gens[conceptID]
	if !exists {
		return nil, fmt.Errorf("no generator registered for concept %q", conceptID)
	}
	return gen, nil
}

func (r *Registry) Generate(conceptID string, difficulty float64) (Problem, error) {
	if difficulty < 0 {
		difficulty = 0
	}
	if difficulty > 1 {
		difficulty = 1
	}
	return r.GenerateContext(conceptID, GeneratorContext{Difficulty: difficulty})
}

func (r *Registry) GenerateContext(conceptID string, ctx GeneratorContext) (Problem, error) {
	r.mu.RLock()
	gen, exists := r.gens[conceptID]
	r.mu.RUnlock()
	if !exists {
		return Problem{}, fmt.Errorf("no generator registered for concept %q", conceptID)
	}
	// Seed global rand from ctx.Seed if provided (per-question seeded via hash of
	// StudentID+ConceptID+AttemptID). Use mutex only for the Seed call so
	// generation itself is not serialized.
	if ctx.Seed != 0 {
		r.randMu.Lock()
		rand.Seed(ctx.Seed)
		r.randMu.Unlock()
	}
	p := booleanChoices(gen.Generate(ctx))
	p.Question = latex.Canonicalize(p.Question, latex.Generators)
	p.Explanation = r.explain(conceptID, p)
	p.Difficulty = difficultyOf(ctx.Difficulty)
	p.KPOrdinal = ctx.KPOrdinal
	if len(p.ScaffoldSteps) == 0 {
		if sp, ok := gen.(ScaffoldProblem); ok {
			p.ScaffoldSteps = sp.ScaffoldSteps(ctx)
		}
	}
	for _, w := range latex.Validate(p.Question+"\n"+p.Explanation, latex.Generators) {
		fmt.Printf("latex warning in generator %q: %s\n", conceptID, w)
	}
	return p, nil
}

// difficultyOf normalises a generator difficulty into the pointer form Problem carries.
//
// A zero is reported as unknown rather than as difficulty 0: `Generate` clamps a negative
// to 0, and a question generated at "no difficulty in particular" has no difficulty to
// record. Everything from 0.3 up is a real value the engine chose.
func difficultyOf(d float64) *float64 {
	if d <= 0 {
		return nil
	}
	v := d
	return &v
}

func (r *Registry) BatchGenerate(conceptID string, count int, difficulty float64) ([]Problem, error) {
	return r.BatchGenerateContext(conceptID, count, GeneratorContext{Difficulty: difficulty})
}

func (r *Registry) BatchGenerateContext(conceptID string, count int, ctx GeneratorContext) ([]Problem, error) {
	if count < 1 {
		count = 1
	}
	if count > 20 {
		count = 20
	}
	r.mu.RLock()
	gen, exists := r.gens[conceptID]
	r.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("no generator registered for concept %q", conceptID)
	}
	if ctx.Seed != 0 {
		r.randMu.Lock()
		rand.Seed(ctx.Seed)
		r.randMu.Unlock()
	}
	problems := make([]Problem, 0, count)
	seen := make(map[string]bool)
	// Use higher multiplier to avoid starvation for low-entropy generators (e.g. small table-driven ML gens)
	maxAttempts := count * 10
	if maxAttempts < 30 {
		maxAttempts = 30
	}
	for i := 0; i < maxAttempts && len(problems) < count; i++ {
		p := gen.Generate(ctx)
		p.Question = latex.Canonicalize(p.Question, latex.Generators)
		p.Explanation = r.explain(conceptID, p)
		p.Difficulty = difficultyOf(ctx.Difficulty)
		p.KPOrdinal = ctx.KPOrdinal
		for _, w := range latex.Validate(p.Question+"\n"+p.Explanation, latex.Generators) {
			fmt.Printf("latex warning in generator %q (batch): %s\n", conceptID, w)
		}
		if !seen[p.Question] {
			seen[p.Question] = true
			problems = append(problems, p)
		}
	}
	return problems, nil
}

func (r *Registry) Has(conceptID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.gens[conceptID]
	return exists
}

func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.gens)
}

// Duplicates returns the concepts that were registered more than once, in the order the refused
// registrations arrived. It is empty for a registry built by hand from one package; a registry
// built by all.Register should be empty too, and a test there asserts it.
func (r *Registry) Duplicates() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.duplicates))
	copy(out, r.duplicates)
	return out
}

func (r *Registry) Concepts() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.gens))
	for id := range r.gens {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// booleanChoices turns a yes/no answer into tappable options.
//
// One hundred and eighteen concepts in this corpus answered "yes", "no", "true"
// or "false" with no option list, across topology, calculus, complex analysis,
// abstract algebra and machine learning. A maths field cannot take the word: it
// strips the spaces and reads the letters as variables, and 118 of them were
// reported with a learner typing "y" for yes (report 118). The rule belongs
// here rather than in each generator, because the fault is a property of the
// answer, not of any one question.
//
// The stem's own "(yes/no)" hint is removed, because the options now say it.
func booleanChoices(p Problem) Problem {
	word := strings.ToLower(strings.TrimSpace(p.Answer))
	other := ""
	switch word {
	case "yes":
		other = "no"
	case "no":
		other = "yes"
	case "true":
		other = "false"
	case "false":
		other = "true"
	default:
		return p
	}
	stem := strings.TrimSpace(p.Question)
	for _, hint := range []string{" (yes/no)", " (true/false)", " (enter yes or no)", " (enter true or false)"} {
		stem = strings.ReplaceAll(stem, hint, "")
	}
	if ChoiceOptions(p.Question) != nil {
		// The generator already supplied options; only the stale hint goes.
		p.Question = stem
		return p
	}
	q := MultipleChoice(stem, p.Answer, other)
	q.Explanation = p.Explanation
	q.Facts = p.Facts
	q.Difficulty = p.Difficulty
	return q
}
