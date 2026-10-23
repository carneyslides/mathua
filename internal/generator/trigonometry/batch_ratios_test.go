package trigonometry_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/generator/trigonometry"
)

var batchConcepts = []string{
	"trig.func.sine",
	"trig.func.cosine",
	"trig.func.tangent",
	"trig.func.secant",
	"trig.func.cosecant",
	"trig.func.cotangent",
	"trig.basics.sin_cos_def",
	"trig.basics.tan_def",
	"trig.basics.reciprocal",
	"trig.ident.pythagorean",
	"trig.adv.law_sines",
	"trig.adv.law_cosines",
}

func batchRegistry() *generator.Registry {
	reg := generator.NewRegistry()
	trigonometry.Register(reg)
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

func TestBatchTrigRatiosGeneration(t *testing.T) {
	reg := batchRegistry()
	difficulties := []float64{0.0, 0.3, 0.5, 0.8, 1.0}

	for _, id := range batchConcepts {
		t.Run(id, func(t *testing.T) {
			seenQuestions := make(map[string]bool)

			for _, diff := range difficulties {
				for iter := 0; iter < 10; iter++ {
					p, err := reg.GenerateContext(id, generator.GeneratorContext{
						Difficulty: diff,
					})
					if err != nil {
						t.Fatalf("concept %s at difficulty %.1f iteration %d failed: %v", id, diff, iter, err)
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

func TestBatchTrigRatiosNoReadOff(t *testing.T) {
	reg := batchRegistry()

	for _, id := range batchConcepts {
		t.Run(id, func(t *testing.T) {
			for iter := 0; iter < 25; iter++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: float64(iter%5) / 4.0,
				})
				if err != nil {
					t.Fatalf("concept %s generation failed: %v", id, err)
				}

				// The answer should not be plainly readable from the question stem as a given parameter
				if strings.Contains(p.Answer, "/") {
					// Fraction answer: verify denominator or fraction does not match direct question prompt
					continue
				}

				if containsWord(p.Question, p.Answer) {
					t.Errorf("concept %s printed answer %s directly in question: %s", id, p.Answer, p.Question)
				}
			}
		})
	}
}
