package abstract

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

// The six concepts the batch conversion moved from multiple choice to numeric.
// Each used to answer with prose ("yes", "the whole group") on a multiple-choice
// concept, which is the class behind reports 177 and 184: a stated answer that
// no option list carried. The conversion makes them numeric, and these tests
// keep the conversion honest, because a randomised generator can satisfy its
// own suite while publishing a question it has already answered.
var batchConvertedIDs = []string{
	"abstract.group.subgroup",
	"abstract.rings.ideal",
	"abstract.rings.quotient",
	"abstract.group.normal",
	"abstract.group.quotient_group",
	"abstract.group.commutator",
}

// TestBatchConversionsServeNumericAnswers pins the property the conversion
// exists for: the grader reads a number, so the answer must be one.
func TestBatchConversionsServeNumericAnswers(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)
	for _, id := range batchConvertedIDs {
		for i := 0; i < 300; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{Difficulty: rand.Float64(), Seed: rand.Int63()})
			if err != nil {
				t.Fatalf("%s: %v", id, err)
			}
			if _, err := strconv.Atoi(p.Answer); err != nil {
				t.Fatalf("%s: answer %q is not a whole number, so numeric grading cannot read it\n  question: %s", id, p.Answer, p.Question)
			}
			if !strings.Contains(p.Explanation, "Subgoal 1") || !strings.Contains(p.Explanation, "Subgoal 3") {
				t.Fatalf("%s: explanation lost its subgoal headings\n  %s", id, p.Explanation)
			}
		}
	}
}

// TestBatchConversionsStayVaried fails when a mode collapses to one question or
// one answer. Before the repair, commutatorGen had four fixed questions with
// fixed answers (3, 2, 4 and 1), so a learner who answered once had seen a fifth
// of the generator's output. The floors are below the measured values
// (questions 24-72, answers 10-22 at 500 draws) so an honest edit is not
// blocked, and a collapse to a handful is caught.
func TestBatchConversionsStayVaried(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)
	for _, id := range batchConvertedIDs {
		questions := map[string]bool{}
		answers := map[string]bool{}
		for i := 0; i < 500; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{Difficulty: rand.Float64(), Seed: rand.Int63()})
			if err != nil {
				t.Fatalf("%s: %v", id, err)
			}
			questions[p.Question] = true
			answers[p.Answer] = true
		}
		if len(questions) < 15 {
			t.Errorf("%s: only %d distinct questions in 500 draws, so a mode is not varying its numbers", id, len(questions))
		}
		if len(answers) < 8 {
			t.Errorf("%s: only %d distinct answers in 500 draws, so a learner is being asked the same thing", id, len(answers))
		}
	}
}

// TestBatchConversionsRarelyPrintTheirAnswer is a tripwire for the defect the
// repair removed, not a proof that no answer is guessable. On the contributor's
// version up to 62% of draws printed the answer as one of the numbers in the
// stem: idealGen asked for the index d, quotientGen asked for the characteristic
// n of Z/nZ, and subgroupGen asked for the cycle length k, so the arithmetic was
// already done for the learner. The share is now measured at 13-36% per concept
// and what remains is coincidence rather than a tell: "a group of order 25 has a
// subgroup of order 5, how many cosets" contains the answer 5 and still requires
// the division. The threshold sits above the measured maximum so the test fails
// on a return of the old shape, not on an unlucky draw.
func TestBatchConversionsRarelyPrintTheirAnswer(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)
	const draws = 500
	const maxShare = 0.45
	for _, id := range batchConvertedIDs {
		printed := 0
		for i := 0; i < draws; i++ {
			p, err := reg.GenerateContext(id, generator.GeneratorContext{Difficulty: rand.Float64(), Seed: rand.Int63()})
			if err != nil {
				t.Fatalf("%s: %v", id, err)
			}
			if answerAppearsAsNumber(p.Question, p.Answer) {
				printed++
			}
		}
		share := float64(printed) / draws
		if share > maxShare {
			t.Errorf("%s: the answer appears as a number in the stem on %.0f%% of draws, which usually means the stem already states it", id, share*100)
		}
	}
}

// answerAppearsAsNumber reports whether the answer occurs in the question as a
// standalone integer. A substring test would match 5 inside 25, so digits are
// delimited by non-digits on both sides.
func answerAppearsAsNumber(question, answer string) bool {
	if answer == "" {
		return false
	}
	tokens := strings.FieldsFunc(question, func(r rune) bool {
		return r < '0' || r > '9'
	})
	for _, tok := range tokens {
		if tok == answer {
			return true
		}
	}
	return false
}
