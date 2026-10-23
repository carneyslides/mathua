package probability

import (
	"regexp"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
)

var jointMomentsBatchIDs = []string{
	"prob.joint.marginal",
	"prob.joint.continuous",
	"prob.expect.conditional",
	"prob.expect.total",
	"prob.expect.total_variance",
	"prob.cov.covariance",
	"prob.cov.correlation",
	"prob.conv.delta",
}

func TestProbJointMomentsDynamicDifficulty(t *testing.T) {
	reg := probRegistry(t)
	for _, id := range jointMomentsBatchIDs {
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				diff := float64(i) / 49.0
				p, err := reg.GenerateContext(id, generator.GeneratorContext{
					Difficulty: diff,
					Seed:       int64(42000 + i),
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

func TestProbJointMomentsCognitiveScaffolding(t *testing.T) {
	reg := probRegistry(t)
	for _, id := range jointMomentsBatchIDs {
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

func TestProbJointMomentsZeroOxfordCommas(t *testing.T) {
	reg := probRegistry(t)
	oxfordPattern := regexp.MustCompile(`, (?:and|or)\b`)
	for _, id := range jointMomentsBatchIDs {
		for i := 0; i < 30; i++ {
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

func TestProbJointMomentsGradesWithNumericRouter(t *testing.T) {
	reg := probRegistry(t)
	router := grader.NewRouter()
	defer router.Close()

	for _, id := range jointMomentsBatchIDs {
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
