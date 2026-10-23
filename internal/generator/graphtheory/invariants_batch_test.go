package graphtheory

import (
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func TestGraphTheoryInvariantsBatch_AllConceptsRegistered(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	concepts := []string{
		"graph.complement",
		"graph.product",
		"graph.distance",
		"graph.cut_vertex",
		"graph.bridge",
		"graph.matching",
	}

	for _, id := range concepts {
		gen, err := reg.Get(id)
		if err != nil || gen == nil {
			t.Fatalf("concept %s is not registered: %v", id, err)
		}
	}
}

func TestGraphTheoryInvariantsBatch_ScaffoldingAndDifficultyScaling(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	concepts := []string{
		"graph.complement",
		"graph.product",
		"graph.distance",
		"graph.cut_vertex",
		"graph.bridge",
		"graph.matching",
	}

	difficulties := []float64{0.1, 0.5, 0.9}

	for _, id := range concepts {
		gen, err := reg.Get(id)
		if err != nil {
			t.Fatalf("generator not found for %s: %v", id, err)
		}

		for _, diff := range difficulties {
			for iter := 0; iter < 20; iter++ {
				ctx := generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(241000 + iter*100 + int(diff*100)),
				}

				p := gen.Generate(ctx)

				if strings.TrimSpace(p.Question) == "" {
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
	}
}
