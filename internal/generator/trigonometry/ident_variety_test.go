package trigonometry

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
)

func TestElevatedTrigVariety(t *testing.T) {
	rand.Seed(42)
	reg := generator.NewRegistry()
	Register(reg)
	router := grader.NewRouter()
	defer router.Close()

	conceptIDs := []string{
		"trig.ident.product_to_sum",
		"trig.ident.sum_to_product",
		"trig.ident.pythagorean_variants",
		"trig.eq.general_solution",
		"trig.ineq.advanced",
		"trig.hyperbolic.identities",
		"trig.hyperbolic.derivative",
	}

	for _, id := range conceptIDs {
		gen, err := reg.Get(id)
		if err != nil {
			t.Fatalf("concept %s not registered: %v", id, err)
		}

		stems := make(map[string]bool)
		for i := 0; i < 50; i++ {
			ctx := generator.GeneratorContext{Difficulty: 0.5}
			prob := gen.Generate(ctx)

			if strings.TrimSpace(prob.Question) == "" {
				t.Errorf("%s: empty question at draw %d", id, i)
			}
			if strings.TrimSpace(prob.Answer) == "" {
				t.Errorf("%s: empty answer at draw %d", id, i)
			}
			if strings.TrimSpace(prob.Explanation) == "" {
				t.Errorf("%s: empty explanation at draw %d", id, i)
			}

			opts := trigRaiseOptions(prob.Question)
			if len(opts) < 3 {
				t.Errorf("%s: expected at least 3 options, got %d", id, len(opts))
			}
			found := 0
			for _, o := range opts {
				if o == prob.Answer {
					found++
				}
			}
			if found != 1 {
				t.Errorf("%s: correct answer %q appeared %d times in options %v", id, prob.Answer, found, opts)
			}
			for _, o := range opts {
				if o == prob.Answer {
					continue
				}
				if router.Grade(grader.GradingMultipleChoice, prob.Answer, o).Correct {
					t.Errorf("%s: distractor %q was accepted as correct against answer %q", id, o, prob.Answer)
				}
			}

			stems[prob.Question] = true
		}

		if len(stems) < 8 {
			t.Errorf("%s: variety too low, got %d distinct stems across 50 draws", id, len(stems))
		}
	}
}
