package complex

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

var convertedBatchIDs = []string{
	"complex.basics.parts",
	"complex.basics.powers",
	"complex.basics.reciprocal",
	"complex.basics.geometry",
	"complex.basics.arg_arithmetic",
	"complex.basics.equations",
	"complex.adv.geometric_series",
	"complex.adv.mobius_fixed",
	"complex.adv.residue_simple",
	"complex.adv.cross_ratio",
}

func TestComplexBatchConversions_CognitiveScaffolding(t *testing.T) {
	reg := batchRegistry(t)
	for _, id := range convertedBatchIDs {
		for i := 0; i < 30; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(20000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			if !strings.Contains(p.Explanation, "Subgoal 1: State governing formula") {
				t.Fatalf("%s: missing Subgoal 1 in explanation: %s", id, p.Explanation)
			}
			if !strings.Contains(p.Explanation, "Subgoal 2: Substitute parameters") {
				t.Fatalf("%s: missing Subgoal 2 in explanation: %s", id, p.Explanation)
			}
			if !strings.Contains(p.Explanation, "Subgoal 3: Evaluate and simplify") {
				t.Fatalf("%s: missing Subgoal 3 in explanation: %s", id, p.Explanation)
			}
		}
	}
}

func TestComplexBatchConversions_DifficultyScaling(t *testing.T) {
	reg := batchRegistry(t)
	lowP, err := reg.GenerateContext("complex.basics.powers", generator.GeneratorContext{
		Difficulty: 0.0,
		Seed:       42,
	})
	if err != nil {
		t.Fatalf("powers low difficulty: %v", err)
	}
	highP, err := reg.GenerateContext("complex.basics.powers", generator.GeneratorContext{
		Difficulty: 1.0,
		Seed:       42,
	})
	if err != nil {
		t.Fatalf("powers high difficulty: %v", err)
	}
	if lowP.Question == "" || highP.Question == "" {
		t.Fatal("empty power questions generated")
	}

	for _, id := range convertedBatchIDs {
		gen, err := reg.Get(id)
		if err != nil {
			t.Fatalf("generator not found: %s: %v", id, err)
		}
		for i := 0; i < 50; i++ {
			d := rand.Float64()
			prob := gen.Generate(generator.GeneratorContext{Difficulty: d})
			if prob.Question == "" || prob.Answer == "" || prob.Explanation == "" {
				t.Fatalf("%s: empty problem output at difficulty %.2f", id, d)
			}
			if prob.Facts == nil || prob.Facts["answer"] != prob.Answer {
				t.Fatalf("%s: fact answer missing or mismatch: got %q, want %q", id, prob.Facts["answer"], prob.Answer)
			}
		}
	}
}

func TestComplexBatchConversions_ParameterTransparency(t *testing.T) {
	reg := batchRegistry(t)
	checkedIDs := []string{
		"complex.basics.parts",
		"complex.basics.powers",
		"complex.basics.reciprocal",
		"complex.basics.geometry",
		"complex.basics.arg_arithmetic",
		"complex.basics.equations",
		"complex.adv.geometric_series",
		"complex.adv.residue_simple",
		"complex.adv.cross_ratio",
	}

	for _, id := range checkedIDs {
		for i := 0; i < 60; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(30000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			for key, raw := range p.Facts {
				if key == "answer" || key == "value" || key == "den" || key == "real_num" || key == "imag_num" || key == "re" || key == "im" || key == "real" || key == "imag" || key == "numerator" || key == "denominator" || key == "dx" || key == "dy" || key == "sum" || key == "x" || key == "y" {
					continue
				}
				value, err := strconv.Atoi(raw)
				if err != nil {
					continue
				}
				absVal := value
				if absVal < 0 {
					absVal = -absVal
				}
				pattern := regexp.MustCompile(`(^|\D)` + regexp.QuoteMeta(strconv.Itoa(absVal)) + `(\D|$)`)
				if !pattern.MatchString(p.Question) {
					t.Fatalf("%s: names parameter %s = %d, but question does not print it: %s",
						id, key, value, p.Question)
				}
			}
		}
	}
}

func TestComplexBatchConversions_ZeroOxfordComma(t *testing.T) {
	reg := batchRegistry(t)
	oxfordComma := regexp.MustCompile(`,\s+and\s+`)
	for _, id := range convertedBatchIDs {
		for i := 0; i < 20; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(40000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			if oxfordComma.MatchString(p.Explanation) {
				t.Fatalf("%s: explanation contains Oxford comma: %s", id, p.Explanation)
			}
			if oxfordComma.MatchString(p.Question) {
				t.Fatalf("%s: question contains Oxford comma: %s", id, p.Question)
			}
		}
	}
}
