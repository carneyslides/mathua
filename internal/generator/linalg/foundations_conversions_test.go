package linalg

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
)

var convertedFoundationsIDs = []string{
	"linalg.independence.count",
	"linalg.rank.nullity",
	"linalg.matrix.inverse_2x2",
	"linalg.eigen.trace",
	"linalg.eigen.det_product",
	"linalg.orth.projection_coeff",
	"linalg.vec.norm_squared",
	"linalg.transform.det_scale",
}

func testFoundationsRegistry(t *testing.T) *generator.Registry {
	t.Helper()
	reg := generator.NewRegistry()
	Register(reg)
	return reg
}

func TestLinalgFoundations_CognitiveScaffolding(t *testing.T) {
	reg := testFoundationsRegistry(t)
	for _, id := range convertedFoundationsIDs {
		for i := 0; i < 30; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(10000 + i),
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

func TestLinalgFoundations_DifficultyScaling(t *testing.T) {
	reg := testFoundationsRegistry(t)
	for _, id := range convertedFoundationsIDs {
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
			if _, err := strconv.Atoi(prob.Answer); err != nil {
				t.Fatalf("%s: answer %q is not an integer", id, prob.Answer)
			}
		}
	}
}

func TestLinalgFoundations_ParameterTransparency(t *testing.T) {
	reg := testFoundationsRegistry(t)
	checked := 0
	for _, id := range convertedFoundationsIDs {
		for i := 0; i < 60; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(20000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			for key, raw := range p.Facts {
				if key == "answer" {
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
				checked++
			}
		}
	}
	if checked < 400 {
		t.Fatalf("only %d facts checked, expected at least 400", checked)
	}
}

func TestLinalgFoundations_ZeroOxfordComma(t *testing.T) {
	reg := testFoundationsRegistry(t)
	oxfordComma := regexp.MustCompile(`,\s+and\s+`)
	for _, id := range convertedFoundationsIDs {
		for i := 0; i < 20; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(30000 + i),
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

func TestLinalgFoundations_AnswersAreIntegers(t *testing.T) {
	reg := testFoundationsRegistry(t)
	for _, id := range convertedFoundationsIDs {
		for i := 0; i < 100; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(40000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			if _, err := strconv.Atoi(p.Answer); err != nil {
				t.Fatalf("%s: answer %q is not an integer: %s", id, p.Answer, p.Question)
			}
		}
	}
}

func TestLinalgFoundations_AnswersGradeWithShippedType(t *testing.T) {
	types := linalgShippedTypes(t)
	reg := testFoundationsRegistry(t)
	router := grader.NewRouter()
	defer router.Close()

	for _, id := range convertedFoundationsIDs {
		typ, ok := types[id]
		if !ok || typ == "" {
			t.Fatalf("%s: no grading_type in corpus", id)
		}
		for i := 0; i < 50; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(50000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			res := router.Grade(grader.GradingType(typ), p.Answer, p.Answer)
			if res.Unavailable {
				t.Skipf("%s: grader unavailable", id)
			}
			if !res.Correct {
				t.Fatalf("%s: answer %q does not grade as %s", id, p.Answer, typ)
			}
		}
	}
}

func TestLinalgFoundations_DoesNotPrintItsAnswer(t *testing.T) {
	reg := testFoundationsRegistry(t)
	const draws = 200
	const maxShare = 0.1
	for _, id := range convertedFoundationsIDs {
		printed := 0
		for i := 0; i < draws; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: rand.Float64(),
				Seed:       rand.Int63(),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			answer := p.Answer
			if strings.HasPrefix(answer, "-") {
				answer = answer[1:]
			}
			pattern := regexp.MustCompile(`(^|\D)` + regexp.QuoteMeta(answer) + `(\D|$)`)
			if pattern.MatchString(p.Question) {
				printed++
			}
		}
		if share := float64(printed) / draws; share > maxShare {
			t.Fatalf("%s: %d of %d draws print answer in stem (%.0f%%), limit %.0f%%",
				id, printed, draws, share*100, maxShare*100)
		}
	}
}

func TestLinalgFoundations_StaysVaried(t *testing.T) {
	reg := testFoundationsRegistry(t)
	const draws = 300
	for _, id := range convertedFoundationsIDs {
		questions := map[string]bool{}
		answers := map[string]bool{}
		for i := 0; i < draws; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: rand.Float64(),
				Seed:       rand.Int63(),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			questions[p.Question] = true
			answers[p.Answer] = true
		}
		minAnswers := 5
		if id == "linalg.independence.count" {
			minAnswers = 3
		}
		if len(questions) < 10 || len(answers) < minAnswers {
			t.Fatalf("%s: only %d distinct questions and %d distinct answers in %d draws",
				id, len(questions), len(answers), draws)
		}
	}
}
