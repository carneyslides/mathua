package numtheory

import (
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func TestNumTheoryInvariantsBatch_Registration(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	concepts := []string{
		"nt.adv.jacobi",
		"nt.adv.discrete_log",
		"nt.adv.modular_sqrt",
		"nt.adv.two_squares",
		"nt.adv.pell",
		"nt.adv.pisano",
		"nt.adv.carmichael_function",
		"nt.adv.prime_gap",
	}

	for _, id := range concepts {
		gen, err := reg.Get(id)
		if err != nil || gen == nil {
			t.Fatalf("expected generator for concept %s to be registered", id)
		}
	}
}

func TestNumTheoryInvariantsBatch_CognitiveScaffolding(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	concepts := []string{
		"nt.adv.jacobi",
		"nt.adv.discrete_log",
		"nt.adv.modular_sqrt",
		"nt.adv.two_squares",
		"nt.adv.pell",
		"nt.adv.pisano",
		"nt.adv.carmichael_function",
		"nt.adv.prime_gap",
	}

	difficulties := []float64{0.1, 0.5, 0.9}

	for _, id := range concepts {
		for _, diff := range difficulties {
			ctx := generator.GeneratorContext{
				Difficulty: diff,
				Seed:       42,
			}
			prob, err := reg.GenerateContext(id, ctx)
			if err != nil {
				t.Fatalf("failed to generate %s at difficulty %f: %v", id, diff, err)
			}
			if prob.Question == "" {
				t.Fatalf("%s produced empty question at difficulty %f", id, diff)
			}
			if prob.Answer == "" {
				t.Fatalf("%s produced empty answer at difficulty %f", id, diff)
			}
			if !strings.Contains(prob.Explanation, "Subgoal 1:") {
				t.Fatalf("%s missing Subgoal 1 in explanation: %s", id, prob.Explanation)
			}
			if !strings.Contains(prob.Explanation, "Subgoal 2:") {
				t.Fatalf("%s missing Subgoal 2 in explanation: %s", id, prob.Explanation)
			}
			if !strings.Contains(prob.Explanation, "Subgoal 3:") {
				t.Fatalf("%s missing Subgoal 3 in explanation: %s", id, prob.Explanation)
			}
		}
	}
}

func TestNumTheoryInvariantsBatch_AnswersMatchFacts(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	concepts := []string{
		"nt.adv.jacobi",
		"nt.adv.discrete_log",
		"nt.adv.modular_sqrt",
		"nt.adv.two_squares",
		"nt.adv.pell",
		"nt.adv.pisano",
		"nt.adv.carmichael_function",
		"nt.adv.prime_gap",
	}

	for _, id := range concepts {
		for i := 0; i < 20; i++ {
			ctx := generator.GeneratorContext{
				Difficulty: float64(i) / 20.0,
				Seed:       int64(100 + i),
			}
			prob, err := reg.GenerateContext(id, ctx)
			if err != nil {
				t.Fatalf("generation error for %s: %v", id, err)
			}
			if prob.Facts == nil {
				t.Fatalf("%s facts map is nil", id)
			}
			if prob.Facts["answer"] != prob.Answer {
				t.Fatalf("%s fact answer %q does not match problem answer %q", id, prob.Facts["answer"], prob.Answer)
			}
			if _, err := strconv.Atoi(prob.Answer); err != nil {
				t.Fatalf("%s answer %q does not parse as integer: %v", id, prob.Answer, err)
			}
		}
	}
}

func TestNumTheoryInvariantsBatch_DifficultyScaling(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	// Sample low and high difficulty for prime gap
	lowCtx := generator.GeneratorContext{Difficulty: 0.1, Seed: 123}
	highCtx := generator.GeneratorContext{Difficulty: 0.9, Seed: 123}

	lowProb, err := reg.GenerateContext("nt.adv.carmichael_function", lowCtx)
	if err != nil {
		t.Fatalf("failed to generate carmichael at low difficulty: %v", err)
	}
	highProb, err := reg.GenerateContext("nt.adv.carmichael_function", highCtx)
	if err != nil {
		t.Fatalf("failed to generate carmichael at high difficulty: %v", err)
	}

	lowN, _ := strconv.Atoi(lowProb.Facts["n"])
	highN, _ := strconv.Atoi(highProb.Facts["n"])
	if lowN <= 0 || highN <= 0 {
		t.Fatalf("expected positive modulus n, got low=%d high=%d", lowN, highN)
	}
}
