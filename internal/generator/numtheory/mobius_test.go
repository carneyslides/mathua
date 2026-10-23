package numtheory

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func TestComputeMobius_KnownValues(t *testing.T) {
	cases := []struct {
		n          int
		wantMu     int
		squareFree bool
		k          int
	}{
		{1, 1, true, 0},
		{2, -1, true, 1},
		{3, -1, true, 1},
		{4, 0, false, 1},
		{5, -1, true, 1},
		{6, 1, true, 2},
		{7, -1, true, 1},
		{8, 0, false, 1},
		{9, 0, false, 1},
		{10, 1, true, 2},
		{12, 0, false, 2},
		{30, -1, true, 3},  // 2 * 3 * 5
		{42, -1, true, 3},  // 2 * 3 * 7
		{60, 0, false, 3},  // 2^2 * 3 * 5
		{105, -1, true, 3}, // 3 * 5 * 7
		{210, 1, true, 4},  // 2 * 3 * 5 * 7
	}

	for _, c := range cases {
		mu, factors, sf, k := computeMobius(c.n)
		if mu != c.wantMu {
			t.Errorf("computeMobius(%d) mu = %d, want %d (factors: %s)", c.n, mu, c.wantMu, factors)
		}
		if sf != c.squareFree {
			t.Errorf("computeMobius(%d) squareFree = %v, want %v", c.n, sf, c.squareFree)
		}
		if k != c.k {
			t.Errorf("computeMobius(%d) primeCount = %d, want %d", c.n, k, c.k)
		}
	}
}

func TestMobiusGen_FactsAndOutputs(t *testing.T) {
	gen := &mobiusGen{}
	seenOutputs := make(map[string]int)

	for i := 0; i < 200; i++ {
		d := rand.Float64()
		p := gen.Generate(generator.GeneratorContext{Difficulty: d})

		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Fatalf("empty fields at difficulty %f: %+v", d, p)
		}

		if p.Answer != "-1" && p.Answer != "0" && p.Answer != "1" {
			t.Errorf("unexpected Answer %q, want -1, 0, or 1", p.Answer)
		}
		seenOutputs[p.Answer]++

		nStr, ok := p.Facts["n"]
		if !ok {
			t.Fatalf("missing Fact 'n'")
		}
		n, err := strconv.Atoi(nStr)
		if err != nil {
			t.Fatalf("invalid Fact 'n': %v", err)
		}

		mu, _, sf, k := computeMobius(n)
		if strconv.Itoa(mu) != p.Answer {
			t.Errorf("computed mu(%d) = %d != problem answer %s", n, mu, p.Answer)
		}
		if strconv.FormatBool(sf) != p.Facts["square_free"] {
			t.Errorf("square_free mismatch for n=%d", n)
		}
		if strconv.Itoa(k) != p.Facts["prime_count"] {
			t.Errorf("prime_count mismatch for n=%d", n)
		}
	}

	// Verify that all 3 outcomes (-1, 0, 1) are produced.
	for _, ans := range []string{"-1", "0", "1"} {
		if seenOutputs[ans] == 0 {
			t.Errorf("never generated answer %q across 200 samples", ans)
		}
	}
}
