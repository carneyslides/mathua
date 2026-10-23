package calculus_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/generator/calculus"
)

var batchVectorConcepts = []string{
	"calc.vec.curl",
	"calc.vec.divergence",
	"calc.vec.conservative_test",
	"calc.vec.greens_theorem",
	"calc.multi.transform_jacobian",
	"calc.multi.double_integral_region",
	"calc.multi.saddle_test",
	"calc.multi.lagrange",
	"calc.vec.divergence_theorem",
	"calc.vec.flux_uniform",
	"calc.vec.potential_value",
	"calc.multi.coordinates_jacobian",
}

func batchVectorRegistry() *generator.Registry {
	reg := generator.NewRegistry()
	calculus.Register(reg)
	return reg
}

func containsWord(s, word string) bool {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.'
	})
	for _, f := range fields {
		if f == word {
			return true
		}
	}
	return false
}

func TestVectorTheoremsBatch_Generation(t *testing.T) {
	reg := batchVectorRegistry()
	difficulties := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, id := range batchVectorConcepts {
		t.Run(id, func(t *testing.T) {
			seenQuestions := make(map[string]bool)

			for _, diff := range difficulties {
				for iter := 0; iter < 10; iter++ {
					p, err := reg.GenerateContext(id, generator.GeneratorContext{
						Difficulty: diff,
					})
					if err != nil {
						t.Fatalf("concept %s at difficulty %.2f iteration %d failed: %v", id, diff, iter, err)
					}

					if p.Question == "" {
						t.Errorf("concept %s generated empty question", id)
					}
					if p.Answer == "" {
						t.Errorf("concept %s generated empty answer", id)
					}
					if p.Explanation == "" {
						t.Errorf("concept %s generated empty explanation", id)
					}
					if p.Facts == nil {
						t.Errorf("concept %s generated nil facts", id)
					}

					seenQuestions[p.Question] = true
				}
			}

			if len(seenQuestions) < 5 {
				t.Errorf("concept %s had insufficient variety (%d distinct questions across 50 runs)", id, len(seenQuestions))
			}
		})
	}
}

func TestVectorTheoremsBatch_NoReadOff(t *testing.T) {
	reg := batchVectorRegistry()

	for _, id := range batchVectorConcepts {
		t.Run(id, func(t *testing.T) {
			for iter := 0; iter < 30; iter++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: float64(iter%5) / 4.0,
				})
				if err != nil {
					t.Fatalf("concept %s generation failed: %v", id, err)
				}

				if containsWord(p.Question, p.Answer) {
					t.Errorf("concept %s printed answer %s directly in question: %s", id, p.Answer, p.Question)
				}
			}
		})
	}
}

func TestVectorTheoremsBatch_ParameterTransparency(t *testing.T) {
	reg := batchVectorRegistry()

	skippedFactKeys := map[string]bool{
		"answer":     true,
		"curl":       true,
		"divergence": true,
		"difference": true,
		"area":       true,
		"det":        true,
		"D":          true,
		"maximum":    true,
		"flux":       true,
		"value":      true,
		"factor":     true,
	}

	for _, id := range batchVectorConcepts {
		t.Run(id, func(t *testing.T) {
			for iter := 0; iter < 20; iter++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: float64(iter%5) / 4.0,
				})
				if err != nil {
					t.Fatalf("concept %s generation failed: %v", id, err)
				}

				for key, raw := range p.Facts {
					if skippedFactKeys[key] {
						continue
					}
					val, err := strconv.Atoi(raw)
					if err != nil {
						continue
					}
					absVal := val
					if absVal < 0 {
						absVal = -absVal
					}
					pattern := regexp.MustCompile(`(^|\D)` + regexp.QuoteMeta(strconv.Itoa(absVal)) + `(\D|$)`)
					if !pattern.MatchString(p.Question) {
						t.Errorf("concept %s: parameter %s=%s (abs %d) missing from question stem: %s", id, key, raw, absVal, p.Question)
					}
				}
			}
		})
	}
}

func TestVectorTheoremsBatch_CognitiveScaffolding(t *testing.T) {
	reg := batchVectorRegistry()

	for _, id := range batchVectorConcepts {
		t.Run(id, func(t *testing.T) {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: 0.5,
			})
			if err != nil {
				t.Fatalf("concept %s generation failed: %v", id, err)
			}

			sentences := strings.Split(p.Explanation, ".")
			if len(sentences) < 3 {
				t.Errorf("concept %s explanation has fewer than 3 derivation steps: %s", id, p.Explanation)
			}
		})
	}
}
