package numerical

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
)

var convertedNumaBatch2NumericIDs = []string{
	"numa.interp.least_squares",
	"numa.interp.normal_equations",
	"numa.integration.trapezoid",
	"numa.integration.midpoint",
	"numa.integration.simpson",
	"numa.integration.composite",
	"numa.integration.gauss",
	"numa.linear.elimination",
	"numa.linear.lu",
	"numa.linear.jacobi",
	"numa.linear.power",
}

func testNumaBatch2Registry(t *testing.T) *generator.Registry {
	t.Helper()
	reg := generator.NewRegistry()
	Register(reg)
	return reg
}

func TestNumaBatch2Conversions_CognitiveScaffolding(t *testing.T) {
	reg := testNumaBatch2Registry(t)
	for _, id := range convertedNumaBatch2NumericIDs {
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

func TestNumaBatch2Conversions_DifficultyScaling(t *testing.T) {
	reg := testNumaBatch2Registry(t)
	g := grader.NewRouter()
	defer g.Close()
	for _, id := range convertedNumaBatch2NumericIDs {
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
			res := g.Grade(grader.GradingNumeric, prob.Answer, prob.Answer)
			if !res.Correct {
				t.Fatalf("%s: self-grading failed for answer %q: %s", id, prob.Answer, res.Feedback)
			}
		}
	}
}

func TestNumaBatch2Conversions_ParameterTransparency(t *testing.T) {
	reg := testNumaBatch2Registry(t)
	checked := 0
	for _, id := range convertedNumaBatch2NumericIDs {
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
				if key == "xs" || key == "ys" {
					for _, part := range strings.Split(raw, ",") {
						val, err := strconv.Atoi(part)
						if err != nil {
							continue
						}
						absVal := val
						if absVal < 0 {
							absVal = -absVal
						}
						pattern := regexp.MustCompile(`(^|\D)` + regexp.QuoteMeta(strconv.Itoa(absVal)) + `(\D|$)`)
						if !pattern.MatchString(p.Question) {
							t.Fatalf("%s: names %s %d, but question does not print it: %s", id, key, val, p.Question)
						}
						checked++
					}
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
		t.Fatalf("only %d facts checked", checked)
	}
}

func TestNumaBatch2Conversions_ZeroOxfordComma(t *testing.T) {
	reg := testNumaBatch2Registry(t)
	oxfordComma := regexp.MustCompile(`,\s+and\s+`)
	for _, id := range convertedNumaBatch2NumericIDs {
		for i := 0; i < 20; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: float64(i%10) / 9.0,
				Seed:       int64(30000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			if oxfordComma.MatchString(p.Question) {
				t.Fatalf("%s: question contains Oxford comma: %s", id, p.Question)
			}
			if oxfordComma.MatchString(p.Answer) {
				t.Fatalf("%s: answer contains Oxford comma: %s", id, p.Answer)
			}
			if oxfordComma.MatchString(p.Explanation) {
				t.Fatalf("%s: explanation contains Oxford comma: %s", id, p.Explanation)
			}
		}
	}
}

func TestNumaBatch2AnswerIsNotReadableOffTheQuestion(t *testing.T) {
	reg := testNumaBatch2Registry(t)
	for _, id := range convertedNumaBatch2NumericIDs {
		for i := 0; i < 500; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: rand.Float64(),
				Seed:       int64(40000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			switch id {
			case "numa.interp.least_squares", "numa.interp.normal_equations":
				for _, part := range strings.Split(p.Facts["xs"], ",") {
					if part == p.Answer {
						t.Fatalf("%s: answer %s matches x point: %s", id, p.Answer, p.Question)
					}
				}
				for _, part := range strings.Split(p.Facts["ys"], ",") {
					if part == p.Answer {
						t.Fatalf("%s: answer %s matches y point: %s", id, p.Answer, p.Question)
					}
				}
			case "numa.integration.trapezoid", "numa.integration.midpoint", "numa.integration.simpson":
				if p.Facts["a"] == p.Answer || p.Facts["b"] == p.Answer {
					t.Fatalf("%s: answer %s matches endpoint: %s", id, p.Answer, p.Question)
				}
			case "numa.integration.composite":
				if p.Facts["b"] == p.Answer || p.Facts["n"] == p.Answer {
					t.Fatalf("%s: answer %s matches fact b/n: %s", id, p.Answer, p.Question)
				}
			case "numa.integration.gauss":
				if p.Facts["k"] == p.Answer {
					t.Fatalf("%s: answer %s matches k: %s", id, p.Answer, p.Question)
				}
			case "numa.linear.elimination":
				for _, key := range []string{"a", "b", "c", "d", "e", "f"} {
					if p.Facts[key] == p.Answer {
						t.Fatalf("%s: answer %s matches fact %s: %s", id, p.Answer, key, p.Question)
					}
				}
			case "numa.linear.lu":
				for _, key := range []string{"a", "b", "c", "d"} {
					if p.Facts[key] == p.Answer {
						t.Fatalf("%s: answer %s matches fact %s: %s", id, p.Answer, key, p.Question)
					}
				}
			case "numa.linear.jacobi":
				for _, key := range []string{"a", "b", "e", "y0"} {
					if p.Facts[key] == p.Answer {
						t.Fatalf("%s: answer %s matches fact %s: %s", id, p.Answer, key, p.Question)
					}
				}
			case "numa.linear.power":
				for _, key := range []string{"a", "b", "d"} {
					if p.Facts[key] == p.Answer {
						t.Fatalf("%s: answer %s matches fact %s: %s", id, p.Answer, key, p.Question)
					}
				}
			}
		}
	}
}

func TestNumaBatch2AnswersVary(t *testing.T) {
	reg := testNumaBatch2Registry(t)
	thresholds := map[string]int{
		"numa.interp.least_squares":    15,
		"numa.interp.normal_equations": 15,
		"numa.integration.trapezoid":   15,
		"numa.integration.midpoint":    15,
		"numa.integration.simpson":     15,
		"numa.integration.composite":   15,
		"numa.integration.gauss":       5,
		"numa.linear.elimination":      15,
		"numa.linear.lu":               15,
		"numa.linear.jacobi":           15,
		"numa.linear.power":            15,
	}
	for _, id := range convertedNumaBatch2NumericIDs {
		seen := map[string]bool{}
		for i := 0; i < 200; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: 1.0,
				Seed:       int64(50000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			seen[p.Answer] = true
		}
		want := thresholds[id]
		if want == 0 {
			want = 15
		}
		if len(seen) < want {
			t.Fatalf("%s serves only %d distinct answers at full difficulty, want at least %d", id, len(seen), want)
		}
	}
}
