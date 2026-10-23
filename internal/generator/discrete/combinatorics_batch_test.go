package discrete

import (
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func TestDiscreteCombinatoricsBatch_AllConceptsRegistered(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	concepts := []string{
		"discrete.combinatorics.counting_principle",
		"discrete.combinatorics.factorial",
		"discrete.combinatorics.multinomial",
		"discrete.combinatorics.circular",
		"discrete.combinatorics.derangement",
		"discrete.combinatorics.catalan",
	}

	for _, id := range concepts {
		gen, err := reg.Get(id)
		if err != nil || gen == nil {
			t.Fatalf("concept %s is not registered: %v", id, err)
		}
	}
}

func TestDiscreteCombinatoricsBatch_HelperMath(t *testing.T) {
	// Factorial
	expectedFactorials := []int{1, 1, 2, 6, 24, 120, 720, 5040}
	for n, expected := range expectedFactorials {
		if got := factorial(n); got != expected {
			t.Errorf("factorial(%d) = %d; want %d", n, got, expected)
		}
	}

	// Derangements (subfactorials)
	expectedDerangements := []int{1, 0, 1, 2, 9, 44, 265}
	for n, expected := range expectedDerangements {
		if got := derangement(n); got != expected {
			t.Errorf("derangement(%d) = %d; want %d", n, got, expected)
		}
	}

	// Catalan numbers
	expectedCatalan := []int{1, 1, 2, 5, 14, 42}
	for n, expected := range expectedCatalan {
		if got := catalan(n); got != expected {
			t.Errorf("catalan(%d) = %d; want %d", n, got, expected)
		}
	}

	// Multinomial
	if got := multinomialCoeff(5, []int{2, 2, 1}); got != 30 {
		t.Errorf("multinomialCoeff(5, [2, 2, 1]) = %d; want 30", got)
	}
	if got := multinomialCoeff(6, []int{3, 2, 1}); got != 60 {
		t.Errorf("multinomialCoeff(6, [3, 2, 1]) = %d; want 60", got)
	}
}

func TestDiscreteCombinatoricsBatch_SamplingAndValidation(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	concepts := []string{
		"discrete.combinatorics.counting_principle",
		"discrete.combinatorics.factorial",
		"discrete.combinatorics.multinomial",
		"discrete.combinatorics.circular",
		"discrete.combinatorics.derangement",
		"discrete.combinatorics.catalan",
	}

	difficulties := []float64{0.1, 0.5, 0.9}

	for _, id := range concepts {
		gen, err := reg.Get(id)
		if err != nil {
			t.Fatalf("generator not found for %s: %v", id, err)
		}

		templatesSeen := make(map[int]bool)

		for _, diff := range difficulties {
			for iter := 0; iter < 20; iter++ {
				ctx := generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(iter*1000 + int(diff*100)),
				}

				p := gen.Generate(ctx)

				if p.Question == "" {
					t.Errorf("[%s] question must not be empty", id)
				}

				val, err := strconv.Atoi(p.Answer)
				if err != nil {
					t.Errorf("[%s] answer %q is not a valid integer: %v", id, p.Answer, err)
				}
				if val < 0 {
					t.Errorf("[%s] answer must be non-negative, got %d", id, val)
				}

				if p.Facts == nil {
					t.Errorf("[%s] expected facts map, got nil", id)
				} else {
					tmplStr, ok := p.Facts["template"]
					if ok {
						if tmpl, err := strconv.Atoi(tmplStr); err == nil {
							templatesSeen[tmpl] = true
						}
					}
					ansFact, ok := p.Facts["answer"]
					if !ok || ansFact != p.Answer {
						t.Errorf("[%s] facts answer mismatch: got %v, want %s", id, ansFact, p.Answer)
					}
				}

				// Scaffolding check
				for _, subgoal := range []string{"Subgoal 1:", "Subgoal 2:", "Subgoal 3:"} {
					if !strings.Contains(p.Explanation, subgoal) {
						t.Errorf("[%s] explanation missing %q: %s", id, subgoal, p.Explanation)
					}
				}
			}
		}

		if len(templatesSeen) < 3 {
			t.Errorf("[%s] expected multiple templates seen, got only %d", id, len(templatesSeen))
		}
	}
}
