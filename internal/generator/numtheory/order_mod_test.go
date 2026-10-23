package numtheory

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func TestComputeMultiplicativeOrder_KnownValues(t *testing.T) {
	cases := []struct {
		a          int
		m          int
		wantOrder  int
		wantPowers []int
	}{
		{2, 5, 4, []int{2, 4, 3, 1}},
		{2, 7, 3, []int{2, 4, 1}},
		{3, 7, 6, []int{3, 2, 6, 4, 5, 1}},
		{2, 9, 6, []int{2, 4, 8, 7, 5, 1}},
		{3, 8, 2, []int{3, 1}},
		{2, 11, 10, []int{2, 4, 8, 5, 10, 9, 7, 3, 6, 1}},
		{3, 10, 4, []int{3, 9, 7, 1}},
		{4, 13, 6, []int{4, 3, 12, 9, 10, 1}},
		// Non-coprime or non-positive modulus cases return 0 and nil.
		{2, 6, 0, nil},
		{3, 9, 0, nil},
		{2, 1, 0, nil},
	}

	for _, c := range cases {
		gotOrd, gotPowers := computeMultiplicativeOrder(c.a, c.m)
		if gotOrd != c.wantOrder {
			t.Errorf("computeMultiplicativeOrder(%d, %d) order = %d, want %d", c.a, c.m, gotOrd, c.wantOrder)
		}
		if !reflect.DeepEqual(gotPowers, c.wantPowers) {
			t.Errorf("computeMultiplicativeOrder(%d, %d) powers = %v, want %v", c.a, c.m, gotPowers, c.wantPowers)
		}
	}
}

func TestOrderModGen_FactsAndOutputs(t *testing.T) {
	gen := &orderModGen{}

	for i := 0; i < 200; i++ {
		d := rand.Float64()
		p := gen.Generate(generator.GeneratorContext{Difficulty: d})

		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Fatalf("empty fields at difficulty %f: %+v", d, p)
		}

		aStr, okA := p.Facts["a"]
		mStr, okM := p.Facts["m"]
		ordStr, okOrd := p.Facts["order"]
		ansStr, okAns := p.Facts["answer"]
		if !okA || !okM || !okOrd || !okAns {
			t.Fatalf("missing required fact in %+v", p.Facts)
		}

		a, errA := strconv.Atoi(aStr)
		m, errM := strconv.Atoi(mStr)
		ord, errOrd := strconv.Atoi(ordStr)
		if errA != nil || errM != nil || errOrd != nil {
			t.Fatalf("failed parsing facts to integer: a=%v m=%v ord=%v", errA, errM, errOrd)
		}

		if mathutil.GCD(a, m) != 1 {
			t.Errorf("base %d and modulus %d must be coprime", a, m)
		}
		if p.Answer != ordStr || ansStr != ordStr {
			t.Errorf("answer mismatch: Answer=%q ansFact=%q ordFact=%q", p.Answer, ansStr, ordStr)
		}

		wantOrd, powers := computeMultiplicativeOrder(a, m)
		if ord != wantOrd {
			t.Errorf("order mismatch for %d mod %d: got %d, want %d", a, m, ord, wantOrd)
		}
		if len(powers) != ord {
			t.Errorf("power count mismatch: len=%d, ord=%d", len(powers), ord)
		}

		// Verify minimal power condition: a^k = 1 mod m iff k = ord.
		for k := 1; k < ord; k++ {
			if powers[k-1] == 1 {
				t.Errorf("power %d^%d ≡ 1 mod %d before order %d", a, k, m, ord)
			}
		}
		if powers[ord-1] != 1 {
			t.Errorf("power %d^%d ≢ 1 mod %d at order %d", a, ord, m, ord)
		}
	}
}
