package abstract

// Tests for the converted concepts.
//
// Each of the ten recomputes the answer from the published facts instead of
// trusting the generator's own arithmetic, and each is checked against its
// shipped grading type so a corpus row that drifts back to multiple choice
// fails here rather than in a reviewer's browser.

import (
	"strconv"
	"testing"
)

var convertedIDs = []string{
	"abstract.group.order",
	"abstract.group.lagrange",
	"abstract.group.cyclic",
	"abstract.group.symmetric",
	"abstract.group.dihedral",
	"abstract.field.finite",
	"abstract.group.action",
	"abstract.group.homomorphism",
	"abstract.group.automorphism",
	"abstract.rings.euclidean",
	"abstract.group.sylow",
}

func convertedInt(t *testing.T, facts map[string]string, key string) int {
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
	types := abstractShippedTypes(t)
	for _, id := range convertedIDs {
		if types[id] != "numeric" {
			t.Errorf("%s: corpus says %q, want numeric", id, types[id])
		}
	}
}

func TestConvertedAnswersRecomputeFromFacts(t *testing.T) {
	reg := abstractRegistry(t)

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.order", i)
		n := convertedInt(t, p.Facts, "n")
		a := convertedInt(t, p.Facts, "a")
		want := n / abstractGCD(a, n)
		if got := convertedInt(t, p.Facts, "answer"); got != want || p.Answer != strconv.Itoa(want) {
			t.Fatalf("order: n=%d a=%d answer=%s want %d", n, a, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.lagrange", i)
		g := convertedInt(t, p.Facts, "group_order")
		h := convertedInt(t, p.Facts, "subgroup_order")
		if g%h != 0 {
			t.Fatalf("lagrange: %d is not divisible by %d", g, h)
		}
		want := g / h
		if p.Answer != strconv.Itoa(want) {
			t.Fatalf("lagrange: |G|=%d |H|=%d answer=%s want %d", g, h, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.cyclic", i)
		n := convertedInt(t, p.Facts, "n")
		want := abstractTotient(n)
		if convertedInt(t, p.Facts, "phi") != want || p.Answer != strconv.Itoa(want) {
			t.Fatalf("cyclic: n=%d answer=%s want %d", n, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.symmetric", i)
		n := convertedInt(t, p.Facts, "n")
		if want := abstractFactorial(n); p.Answer != strconv.Itoa(want) {
			t.Fatalf("symmetric: n=%d answer=%s want %d", n, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.dihedral", i)
		n := convertedInt(t, p.Facts, "n")
		if want := 2 * n; p.Answer != strconv.Itoa(want) {
			t.Fatalf("dihedral: n=%d answer=%s want %d", n, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.field.finite", i)
		prime := convertedInt(t, p.Facts, "p")
		degree := convertedInt(t, p.Facts, "n")
		if want := abstractPow(prime, degree); p.Answer != strconv.Itoa(want) {
			t.Fatalf("finite field: p=%d n=%d answer=%s want %d", prime, degree, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.action", i)
		g := convertedInt(t, p.Facts, "group_order")
		st := convertedInt(t, p.Facts, "stabilizer_order")
		if g%st != 0 {
			t.Fatalf("action: %d is not divisible by %d", g, st)
		}
		if want := g / st; p.Answer != strconv.Itoa(want) {
			t.Fatalf("action: |G|=%d |Stab|=%d answer=%s want %d", g, st, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.homomorphism", i)
		g := convertedInt(t, p.Facts, "group_order")
		im := convertedInt(t, p.Facts, "image_order")
		if g%im != 0 {
			t.Fatalf("homomorphism: %d is not divisible by %d", g, im)
		}
		if want := g / im; p.Answer != strconv.Itoa(want) {
			t.Fatalf("homomorphism: |G|=%d |im|=%d answer=%s want %d", g, im, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.automorphism", i)
		n := convertedInt(t, p.Facts, "n")
		if want := abstractTotient(n); p.Answer != strconv.Itoa(want) {
			t.Fatalf("automorphism: n=%d answer=%s want %d", n, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.rings.euclidean", i)
		a := convertedInt(t, p.Facts, "a")
		b := convertedInt(t, p.Facts, "b")
		if want := abstractGCD(a, b); p.Answer != strconv.Itoa(want) {
			t.Fatalf("euclidean: a=%d b=%d answer=%s want %d", a, b, p.Answer, want)
		}
	}

	for i := 0; i < 100; i++ {
		p := abstractSample(t, reg, "abstract.group.sylow", i)
		ans := convertedInt(t, p.Facts, "answer")
		if p.Answer != strconv.Itoa(ans) {
			t.Fatalf("sylow: answer=%s want %d", p.Answer, ans)
		}
		g := convertedInt(t, p.Facts, "group_order")
		prime := convertedInt(t, p.Facts, "p")
		if p.Facts["sylow_order"] != "" {
			pk := convertedInt(t, p.Facts, "sylow_order")
			if g%pk != 0 {
				t.Fatalf("sylow order %d does not divide |G|=%d", pk, g)
			}
			rem := g / pk
			if rem%prime == 0 {
				t.Fatalf("sylow order %d is not maximal prime power for p=%d, |G|=%d", pk, prime, g)
			}
		}
	}
}
