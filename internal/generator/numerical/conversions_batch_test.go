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

var convertedNumaNumericIDs = []string{
	"numa.errors.absolute",
	"numa.errors.cancellation",
	"numa.errors.conditioning",
	"numa.roots.bisection",
	"numa.roots.fixed_point",
	"numa.roots.newton",
	"numa.roots.secant",
	"numa.roots.order",
	"numa.roots.horner",
	"numa.interp.lagrange",
	"numa.interp.divided_differences",
}

func testNumaRegistry(t *testing.T) *generator.Registry {
	t.Helper()
	reg := generator.NewRegistry()
	Register(reg)
	return reg
}

func TestNumaBatchConversions_CognitiveScaffolding(t *testing.T) {
	reg := testNumaRegistry(t)
	for _, id := range convertedNumaNumericIDs {
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

func TestNumaBatchConversions_DifficultyScaling(t *testing.T) {
	reg := testNumaRegistry(t)
	g := grader.NewRouter()
	defer g.Close()
	for _, id := range convertedNumaNumericIDs {
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

func TestNumaBatchConversions_ParameterTransparency(t *testing.T) {
	reg := testNumaRegistry(t)
	checked := 0
	for _, id := range convertedNumaNumericIDs {
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
				if key == "coeffs" {
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
							t.Fatalf("%s: names coeff %d, but question does not print it: %s", id, val, p.Question)
						}
						checked++
					}
					continue
				}
				value, err := strconv.Atoi(raw)
				if err != nil {
					// Might be float like tol or c
					fVal, err := strconv.ParseFloat(raw, 64)
					if err != nil {
						continue
					}
					if !strings.Contains(p.Question, raw) && !strings.Contains(p.Question, strconv.FormatFloat(fVal, 'f', -1, 64)) {
						t.Fatalf("%s: names float parameter %s = %v, but question does not print it: %s",
							id, key, fVal, p.Question)
					}
					checked++
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

func TestNumaBatchConversions_ZeroOxfordComma(t *testing.T) {
	reg := testNumaRegistry(t)
	oxfordComma := regexp.MustCompile(`,\s+and\s+`)
	for _, id := range convertedNumaNumericIDs {
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

func TestNumaBatchAnswerIsNotReadableOffTheQuestion(t *testing.T) {
	reg := testNumaRegistry(t)
	for _, id := range convertedNumaNumericIDs {
		for i := 0; i < 500; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{
				Difficulty: rand.Float64(),
				Seed:       int64(40000 + i),
			})
			if err != nil {
				t.Fatalf("%s: generate failed: %v", id, err)
			}
			switch id {
			case "numa.errors.absolute":
				if p.Facts["true"] == p.Answer || p.Facts["approx"] == p.Answer {
					t.Fatalf("%s: answer %s matches input: %s", id, p.Answer, p.Question)
				}
			case "numa.errors.cancellation":
				if p.Facts["x"] == p.Answer || p.Facts["y"] == p.Answer {
					t.Fatalf("%s: answer %s matches input: %s", id, p.Answer, p.Question)
				}
			case "numa.roots.bisection":
				if p.Facts["width"] == p.Answer {
					t.Fatalf("%s: answer %s matches width: %s", id, p.Answer, p.Question)
				}
			case "numa.roots.newton":
				if p.Facts["c"] == p.Answer || p.Facts["x0"] == p.Answer {
					t.Fatalf("%s: answer %s matches c or x0: %s", id, p.Answer, p.Question)
				}
			case "numa.roots.secant":
				if p.Facts["c"] == p.Answer || p.Facts["x0"] == p.Answer || p.Facts["x1"] == p.Answer {
					t.Fatalf("%s: answer %s matches secant inputs: %s", id, p.Answer, p.Question)
				}
			case "numa.roots.order":
				if p.Facts["a"] == p.Answer || p.Facts["b"] == p.Answer || p.Facts["c"] == p.Answer {
					t.Fatalf("%s: answer %s matches order inputs: %s", id, p.Answer, p.Question)
				}
			case "numa.interp.lagrange":
				for _, key := range []string{"y0", "y1", "y2", "q"} {
					if p.Facts[key] == p.Answer {
						t.Fatalf("%s: answer %s matches fact %s: %s", id, p.Answer, key, p.Question)
					}
				}
			case "numa.interp.divided_differences":
				for _, key := range []string{"x0", "y0", "x1", "y1"} {
					if p.Facts[key] == p.Answer {
						t.Fatalf("%s: answer %s matches fact %s: %s", id, p.Answer, key, p.Question)
					}
				}
			}
		}
	}
}

func TestNumaBatchAnswersVary(t *testing.T) {
	reg := testNumaRegistry(t)
	thresholds := map[string]int{
		"numa.errors.absolute":            15,
		"numa.errors.cancellation":        15,
		"numa.errors.conditioning":        10,
		"numa.roots.bisection":            8,
		"numa.roots.fixed_point":          12,
		"numa.roots.newton":               20,
		"numa.roots.secant":               20,
		"numa.roots.order":                4,
		"numa.roots.horner":               20,
		"numa.interp.lagrange":            20,
		"numa.interp.divided_differences": 15,
	}
	for _, id := range convertedNumaNumericIDs {
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
