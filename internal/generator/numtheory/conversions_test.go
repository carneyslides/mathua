package numtheory

// Tests for the converted number theory concepts.
//
// Each concept recomputes its answer from the published facts instead of trusting the
// generator's arithmetic, and each is checked against its shipped grading type so a
// corpus row that drifts back to multiple choice fails here.

import (
	"strconv"
	"testing"

	"github.com/chuma-beep/mathua/internal/grader"
)

var convertedIDs = []string{
	"nt.adv.legendre",
	"nt.adv.mobius",
	"nt.basics.order_mod",
}

func convFactInt(t *testing.T, facts map[string]string, key string) int {
	t.Helper()
	raw, ok := facts[key]
	if !ok {
		t.Fatalf("missing fact %q", key)
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		t.Fatalf("fact %q is not an integer: %q", key, raw)
	}
	return v
}

func TestConvertedConceptsAreNumericInTheCorpus(t *testing.T) {
	types := ntShippedTypes(t)
	for _, id := range convertedIDs {
		if types[id] != "numeric" {
			t.Errorf("%s: corpus says %q, want numeric", id, types[id])
		}
	}
}

func TestConvertedAnswersRecomputeFromFacts(t *testing.T) {
	reg := ntBatchRegistry(t)
	router := grader.NewRouter()

	for i := 0; i < 100; i++ {
		p := ntBatchSample(t, reg, "nt.adv.legendre", i)
		a := convFactInt(t, p.Facts, "a")
		prime := convFactInt(t, p.Facts, "p")
		want := legendreSymbol(a, prime)
		wantStr := strconv.Itoa(want)

		if p.Answer != wantStr {
			t.Fatalf("legendre: a=%d p=%d answer=%s want %s", a, prime, p.Answer, wantStr)
		}
		if got := convFactInt(t, p.Facts, "answer"); got != want {
			t.Fatalf("legendre fact answer: got %d want %d", got, want)
		}

		res := router.Grade(grader.GradingNumeric, p.Answer, p.Answer)
		if !res.Correct {
			t.Fatalf("legendre: correct answer %q rejected by grader: %v", p.Answer, res.Feedback)
		}
	}

	for i := 0; i < 100; i++ {
		p := ntBatchSample(t, reg, "nt.adv.mobius", i)
		n := convFactInt(t, p.Facts, "n")
		want := mobiusMu(n)
		wantStr := strconv.Itoa(want)

		if p.Answer != wantStr {
			t.Fatalf("mobius: n=%d answer=%s want %s", n, p.Answer, wantStr)
		}
		if got := convFactInt(t, p.Facts, "answer"); got != want {
			t.Fatalf("mobius fact answer: got %d want %d", got, want)
		}

		res := router.Grade(grader.GradingNumeric, p.Answer, p.Answer)
		if !res.Correct {
			t.Fatalf("mobius: correct answer %q rejected by grader: %v", p.Answer, res.Feedback)
		}
	}

	for i := 0; i < 100; i++ {
		p := ntBatchSample(t, reg, "nt.basics.order_mod", i)
		a := convFactInt(t, p.Facts, "a")
		m := convFactInt(t, p.Facts, "m")
		want := multiplicativeOrder(a, m)
		wantStr := strconv.Itoa(want)

		if p.Answer != wantStr {
			t.Fatalf("order_mod: a=%d m=%d answer=%s want %s", a, m, p.Answer, wantStr)
		}
		if got := convFactInt(t, p.Facts, "answer"); got != want {
			t.Fatalf("order_mod fact answer: got %d want %d", got, want)
		}

		res := router.Grade(grader.GradingNumeric, p.Answer, p.Answer)
		if !res.Correct {
			t.Fatalf("order_mod: correct answer %q rejected by grader: %v", p.Answer, res.Feedback)
		}
	}
}
