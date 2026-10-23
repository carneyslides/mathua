package linalg

import (
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/chuma-beep/mathua/internal/generator"
)

var batchDecompConceptIDs = []string{
	"linalg.vector.concept",
	"linalg.vector.dot",
	"linalg.vec.span",
	"linalg.vec.basis",
	"linalg.vector.parametric",
	"linalg.inner_product",
	"linalg.decomp.lu",
	"linalg.decomp.qr",
	"linalg.decomp.svd",
	"linalg.decomp.jordan",
	"linalg.subspace",
}

func decompTestRegistry(t *testing.T) *generator.Registry {
	t.Helper()
	reg := generator.NewRegistry()
	Register(reg)
	return reg
}

func TestBatchDecompositionsCognitiveScaffolding(t *testing.T) {
	reg := decompTestRegistry(t)
	for _, id := range batchDecompConceptIDs {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 30; i++ {
				diff := float64(i) / 29.0
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(404000 + i),
				})
				if err != nil {
					t.Fatalf("concept %s failed generating at difficulty %.2f: %v", id, diff, err)
				}
				if !strings.Contains(p.Explanation, "Subgoal 1:") {
					t.Errorf("concept %s missing Subgoal 1 at diff %.2f: %s", id, diff, p.Explanation)
				}
				if !strings.Contains(p.Explanation, "Subgoal 2:") {
					t.Errorf("concept %s missing Subgoal 2 at diff %.2f: %s", id, diff, p.Explanation)
				}
				if !strings.Contains(p.Explanation, "Subgoal 3:") {
					t.Errorf("concept %s missing Subgoal 3 at diff %.2f: %s", id, diff, p.Explanation)
				}
			}
		})
	}
}

func TestBatchDecompositionsReadOffAvoidance(t *testing.T) {
	reg := decompTestRegistry(t)
	for _, id := range batchDecompConceptIDs {
		if id == "linalg.vec.span" || id == "linalg.vector.parametric" {
			// span outputs 1 or 2 which can legitimately occur in coordinate numbers,
			// and parametric outputs explicit coordinate equation strings.
			continue
		}
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 40; i++ {
				diff := float64(i) / 39.0
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(505000 + i),
				})
				if err != nil {
					t.Fatalf("concept %s generation error: %v", id, err)
				}
				ansStr := strings.TrimSpace(p.Answer)
				tokens := strings.FieldsFunc(p.Question, func(r rune) bool {
					return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-'
				})
				for _, tok := range tokens {
					if tok == ansStr {
						t.Errorf("concept %s read-off violation: answer %s matches question token in %s", id, ansStr, p.Question)
					}
				}
			}
		})
	}
}

func TestBatchDecompositionsFactsIntegrity(t *testing.T) {
	reg := decompTestRegistry(t)
	for _, id := range batchDecompConceptIDs {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 25; i++ {
				diff := float64(i) / 24.0
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(606000 + i),
				})
				if err != nil {
					t.Fatalf("generation error: %v", err)
				}
				if len(p.Facts) == 0 {
					t.Fatalf("concept %s generated empty Facts", id)
				}
				if p.Facts["answer"] != p.Answer {
					t.Errorf("concept %s Facts[answer] = %q != Answer %q", id, p.Facts["answer"], p.Answer)
				}
				for k, v := range p.Facts {
					if k == "answer" || k == "dim" || k == "nonzero" {
						continue
					}
					// Integer facts should appear as numerals or sub-expressions in the question
					if _, err := strconv.Atoi(v); err == nil {
						if !strings.Contains(p.Question, v) {
							t.Errorf("concept %s fact key %s with value %s not in question: %s", id, k, v, p.Question)
						}
					}
				}
			}
		})
	}
}

func TestBatchDecompositionsDiversity(t *testing.T) {
	reg := decompTestRegistry(t)
	for _, id := range batchDecompConceptIDs {
		t.Run(id, func(t *testing.T) {
			seen := map[string]bool{}
			answers := map[string]bool{}
			for i := 0; i < 60; i++ {
				diff := float64(i%10) / 9.0
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(707000 + i),
				})
				if err != nil {
					t.Fatalf("generation failed: %v", err)
				}
				seen[p.Question] = true
				answers[p.Answer] = true
			}
			if len(seen) < 5 {
				t.Errorf("concept %s pool too thin: only %d distinct questions", id, len(seen))
			}
			if len(answers) < 2 {
				t.Errorf("concept %s answers too thin: only %d distinct answers", id, len(answers))
			}
		})
	}
}
