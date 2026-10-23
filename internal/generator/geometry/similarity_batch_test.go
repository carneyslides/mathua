package geometry_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/generator/geometry"
)

var similarityBatchConcepts = []string{
	"geo.triangle.inequality",
	"geo.similar.area_ratio",
	"geo.circle.tangent_length",
	"geo.solid.pyramid_volume",
	"geo.similar.ratio",
	"geo.similar.missing_side",
	"geo.transform.dilation",
	"geo.polygon.interior_angles",
	"geo.polygon.exterior_angle",
	"geo.circle.inscribed_angle",
	"geo.circle.chord",
	"geo.triangle.special_right",
}

func setupSimilarityRegistry() *generator.Registry {
	reg := generator.NewRegistry()
	geometry.Register(reg)
	return reg
}

// answerAppearsAsNumber checks whether answer appears as a standalone token in question.
func answerAppearsAsNumber(question, answer string) bool {
	fields := strings.FieldsFunc(question, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, field := range fields {
		if field == answer {
			return true
		}
	}
	return false
}

// TestSimilarityBatch_Scaffolding verifies 3-step cognitive derivations across all concepts.
func TestSimilarityBatch_Scaffolding(t *testing.T) {
	reg := setupSimilarityRegistry()
	for _, id := range similarityBatchConcepts {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 20; i++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: float64(i) / 19.0,
					Seed:       int64(3000 + i),
				})
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", id, err)
				}
				if !strings.Contains(p.Explanation, "Subgoal 1: State governing formula") {
					t.Fatalf("%s missing Subgoal 1: %s", id, p.Explanation)
				}
				if !strings.Contains(p.Explanation, "Subgoal 2: Substitute parameters") {
					t.Fatalf("%s missing Subgoal 2: %s", id, p.Explanation)
				}
				if !strings.Contains(p.Explanation, "Subgoal 3: Evaluate and simplify") {
					t.Fatalf("%s missing Subgoal 3: %s", id, p.Explanation)
				}
			}
		})
	}
}

// TestSimilarityBatch_ReadOffPrevention verifies answer does not appear as a standalone number in stem.
func TestSimilarityBatch_ReadOffPrevention(t *testing.T) {
	reg := setupSimilarityRegistry()
	for _, id := range similarityBatchConcepts {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 30; i++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: float64(i) / 29.0,
					Seed:       int64(4000 + i),
				})
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", id, err)
				}
				if answerAppearsAsNumber(p.Question, p.Answer) {
					t.Fatalf("%s stem reveals answer %q: %q", id, p.Answer, p.Question)
				}
			}
		})
	}
}

// TestSimilarityBatch_ParameterTransparency verifies facts appear in question stem or represent solution steps.
func TestSimilarityBatch_ParameterTransparency(t *testing.T) {
	reg := setupSimilarityRegistry()
	for _, id := range similarityBatchConcepts {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 20; i++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: float64(i) / 19.0,
					Seed:       int64(5000 + i),
				})
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", id, err)
				}
				for key, val := range p.Facts {
					// Answer and intermediate schema calculation facts do not need to appear in stem
					if key == "side" || key == "area_ratio" || key == "tangent" || key == "volume" ||
						key == "length" || key == "sum" || key == "angle" || key == "chord" ||
						key == "bound" || key == "difference" || val == p.Answer {
						continue
					}
					if !strings.Contains(p.Question, val) {
						t.Fatalf("%s: published fact %s=%q missing from question stem: %q", id, key, val, p.Question)
					}
				}
			}
		})
	}
}

// TestSimilarityBatch_Variety verifies rich problem space across difficulty spectrum.
func TestSimilarityBatch_Variety(t *testing.T) {
	reg := setupSimilarityRegistry()
	for _, id := range similarityBatchConcepts {
		t.Run(id, func(t *testing.T) {
			distinctQuestions := make(map[string]bool)
			distinctAnswers := make(map[string]bool)
			for i := 0; i < 100; i++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: rand.Float64(),
					Seed:       int64(6000 + i),
				})
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", id, err)
				}
				distinctQuestions[p.Question] = true
				distinctAnswers[p.Answer] = true
			}
			if len(distinctQuestions) < 12 {
				t.Fatalf("%s: expected at least 12 distinct questions, got %d", id, len(distinctQuestions))
			}
			if len(distinctAnswers) < 8 {
				t.Fatalf("%s: expected at least 8 distinct answers, got %d", id, len(distinctAnswers))
			}
		})
	}
}

// TestSimilarityBatch_Recompute verifies mathematical accuracy of each generator.
func TestSimilarityBatch_Recompute(t *testing.T) {
	reg := setupSimilarityRegistry()
	for _, id := range similarityBatchConcepts {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 30; i++ {
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: float64(i) / 29.0,
					Seed:       int64(7000 + i),
				})
				if err != nil {
					t.Fatalf("unexpected error for %s: %v", id, err)
				}
				ansInt, err := strconv.Atoi(p.Answer)
				if err != nil {
					t.Fatalf("%s answer %q is not integer: %v", id, p.Answer, err)
				}

				switch id {
				case "geo.triangle.inequality":
					a, _ := strconv.Atoi(p.Facts["given_a"])
					b, _ := strconv.Atoi(p.Facts["given_b"])
					if strings.Contains(p.Question, "largest whole number") {
						want := a + b - 1
						if ansInt != want {
							t.Fatalf("want largest %d, got %d", want, ansInt)
						}
					} else {
						diff := b - a
						if diff < 0 {
							diff = -diff
						}
						want := diff + 1
						if ansInt != want {
							t.Fatalf("want smallest %d, got %d", want, ansInt)
						}
					}

				case "geo.similar.area_ratio":
					k, _ := strconv.Atoi(p.Facts["factor"])
					if ansInt != k*k {
						t.Fatalf("want area ratio %d, got %d", k*k, ansInt)
					}

				case "geo.circle.tangent_length":
					r, _ := strconv.Atoi(p.Facts["radius"])
					d, _ := strconv.Atoi(p.Facts["distance"])
					if r*r+ansInt*ansInt != d*d {
						t.Fatalf("tangent Pythagorean failed: r=%d, t=%d, d=%d", r, ansInt, d)
					}

				case "geo.solid.pyramid_volume":
					base, _ := strconv.Atoi(p.Facts["base"])
					h, _ := strconv.Atoi(p.Facts["height"])
					if ansInt != base*h/3 {
						t.Fatalf("pyramid volume mismatch: base=%d, h=%d, ans=%d", base, h, ansInt)
					}

				case "geo.similar.ratio":
					k, _ := strconv.Atoi(p.Facts["factor"])
					side, _ := strconv.Atoi(p.Facts["side"])
					if ansInt != side || ansInt%k != 0 {
						t.Fatalf("similar ratio mismatch: k=%d, side=%d", k, ansInt)
					}

				case "geo.similar.missing_side":
					k, _ := strconv.Atoi(p.Facts["factor"])
					given, _ := strconv.Atoi(p.Facts["given"])
					if ansInt != k*given {
						t.Fatalf("missing side mismatch: want %d, got %d", k*given, ansInt)
					}

				case "geo.transform.dilation":
					factorStr := p.Facts["factor"]
					if strings.HasPrefix(factorStr, "1/") {
						k, _ := strconv.Atoi(strings.TrimPrefix(factorStr, "1/"))
						origStr := fmt.Sprintf("%d", ansInt*k)
						if !strings.Contains(p.Question, origStr) {
							t.Fatalf("dilation reduction missing original length %s: %s", origStr, p.Question)
						}
					} else {
						k, _ := strconv.Atoi(factorStr)
						origStr := fmt.Sprintf("%d", ansInt/k)
						if !strings.Contains(p.Question, origStr) {
							t.Fatalf("dilation enlargement missing original length %s: %s", origStr, p.Question)
						}
					}

				case "geo.polygon.interior_angles":
					n, _ := strconv.Atoi(p.Facts["sides"])
					want := (n - 2) * 180
					if ansInt != want {
						t.Fatalf("interior angles mismatch: want %d, got %d", want, ansInt)
					}

				case "geo.polygon.exterior_angle":
					n, _ := strconv.Atoi(p.Facts["sides"])
					want := 360 / n
					if ansInt != want {
						t.Fatalf("exterior angle mismatch: want %d, got %d", want, ansInt)
					}

				case "geo.circle.inscribed_angle":
					central, _ := strconv.Atoi(p.Facts["central"])
					if ansInt != central/2 {
						t.Fatalf("inscribed angle mismatch: central=%d, want %d, got %d", central, central/2, ansInt)
					}

				case "geo.circle.chord":
					d, _ := strconv.Atoi(p.Facts["distance"])
					r, _ := strconv.Atoi(p.Facts["radius"])
					half := ansInt / 2
					if d*d+half*half != r*r {
						t.Fatalf("chord right triangle failed: d=%d, half=%d, r=%d", d, half, r)
					}

				case "geo.triangle.special_right":
					short, _ := strconv.Atoi(p.Facts["short_leg"])
					hyp, _ := strconv.Atoi(p.Facts["hypotenuse"])
					if hyp != 2*short {
						t.Fatalf("special right triangle failed: short=%d, hyp=%d", short, hyp)
					}
				}
			}
		})
	}
}
