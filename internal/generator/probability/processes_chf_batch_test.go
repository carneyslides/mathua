package probability

import (
	"regexp"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
)

var probProcessesBatchIDs = []string{
	"prob.chf.moments",
	"prob.process.markov_chain",
	"prob.process.transition_matrix",
	"prob.process.stationary",
	"prob.process.random_walk",
}

func TestProbProcessesDynamicDifficulty(t *testing.T) {
	reg := probRegistry(t)
	for _, id := range probProcessesBatchIDs {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				diff := float64(i) / 49.0
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(77000 + i),
				})
				if err != nil {
					t.Fatalf("%s at diff %.2f: %v", id, diff, err)
				}
				if strings.TrimSpace(p.Question) == "" {
					t.Fatalf("%s at diff %.2f: empty question", id, diff)
				}
				if strings.TrimSpace(p.Answer) == "" {
					t.Fatalf("%s at diff %.2f: empty answer", id, diff)
				}
				if strings.TrimSpace(p.Explanation) == "" {
					t.Fatalf("%s at diff %.2f: empty explanation", id, diff)
				}
			}
		})
	}
}

func TestProbProcessesCognitiveScaffolding(t *testing.T) {
	reg := probRegistry(t)
	for _, id := range probProcessesBatchIDs {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 30; i++ {
				p := probSample(t, reg, id, i)
				exp := p.Explanation
				if !strings.Contains(exp, "Subgoal 1: State governing formula") {
					t.Fatalf("%s: missing Subgoal 1 in explanation: %q", id, exp)
				}
				if !strings.Contains(exp, "Subgoal 2: Substitute parameters") {
					t.Fatalf("%s: missing Subgoal 2 in explanation: %q", id, exp)
				}
				if !strings.Contains(exp, "Subgoal 3: Evaluate and simplify") {
					t.Fatalf("%s: missing Subgoal 3 in explanation: %q", id, exp)
				}
			}
		})
	}
}

func TestProbProcessesZeroOxfordCommas(t *testing.T) {
	reg := probRegistry(t)
	oxfordPattern := regexp.MustCompile(`, (?:and|or)\b`)
	for _, id := range probProcessesBatchIDs {
		for i := 0; i < 40; i++ {
			p := probSample(t, reg, id, i)
			if oxfordPattern.MatchString(p.Question) {
				t.Fatalf("%s: Oxford comma detected in question: %q", id, p.Question)
			}
			if oxfordPattern.MatchString(p.Explanation) {
				t.Fatalf("%s: Oxford comma detected in explanation: %q", id, p.Explanation)
			}
		}
	}
}

func TestProbProcessesPrintedParameters(t *testing.T) {
	reg := probRegistry(t)
	for _, id := range probProcessesBatchIDs {
		for i := 0; i < 100; i++ {
			p := probSample(t, reg, id, i)
			for name, value := range p.Facts {
				if name == "answer" || value == "" {
					continue
				}
				if !strings.Contains(p.Question, value) {
					t.Fatalf("%s names %s = %s in Facts but question does not print it: %q", id, name, value, p.Question)
				}
			}
		}
	}
}

func TestProbProcessesAnswersGradeWithNumericRouter(t *testing.T) {
	reg := probRegistry(t)
	router := grader.NewRouter()
	defer router.Close()

	for _, id := range probProcessesBatchIDs {
		for i := 0; i < 40; i++ {
			p := probSample(t, reg, id, i)
			res := router.Grade(grader.GradingNumeric, p.Answer, p.Answer)
			if res.Unavailable {
				t.Skipf("%s: grader unavailable", id)
			}
			if !res.Correct {
				t.Fatalf("%s: self-answer %q failed grading: %v", id, p.Answer, res)
			}
		}
	}
}
