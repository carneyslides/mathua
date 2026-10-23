package numtheory

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func TestComputeLegendreSymbol_KnownValues(t *testing.T) {
	cases := []struct {
		a          int
		p          int
		wantSymbol int
		wantRoot   int
	}{
		// Modulo 5 residues and non-residues
		{1, 5, 1, 1},
		{4, 5, 1, 2},
		{2, 5, -1, 0},
		{3, 5, -1, 0},
		{5, 5, 0, 0},
		{6, 5, 1, 1},
		{10, 5, 0, 0},

		// Modulo 7 residues (1, 2, 4) and non-residues (3, 5, 6)
		{1, 7, 1, 1},
		{2, 7, 1, 3}, // 3^2 = 9 = 2 mod 7
		{4, 7, 1, 2}, // 2^2 = 4 mod 7
		{3, 7, -1, 0},
		{5, 7, -1, 0},
		{6, 7, -1, 0},
		{7, 7, 0, 0},

		// Modulo 11 residues (1, 3, 4, 5, 9) and non-residues (2, 6, 7, 8, 10)
		{1, 11, 1, 1},
		{3, 11, 1, 5}, // 5^2 = 25 = 3 mod 11
		{4, 11, 1, 2}, // 2^2 = 4 mod 11
		{5, 11, 1, 4}, // 4^2 = 16 = 5 mod 11
		{9, 11, 1, 3}, // 3^2 = 9 mod 11
		{2, 11, -1, 0},
		{6, 11, -1, 0},
		{7, 11, -1, 0},
		{8, 11, -1, 0},
		{10, 11, -1, 0},
		{0, 11, 0, 0},
		{22, 11, 0, 0},

		// Modulus <= 2 edge cases
		{3, 2, 0, 0},
		{3, 1, 0, 0},
	}

	for _, c := range cases {
		sym, root := computeLegendreSymbol(c.a, c.p)
		if sym != c.wantSymbol {
			t.Errorf("computeLegendreSymbol(%d, %d) symbol = %d, want %d", c.a, c.p, sym, c.wantSymbol)
		}
		if root != c.wantRoot {
			t.Errorf("computeLegendreSymbol(%d, %d) root = %d, want %d", c.a, c.p, root, c.wantRoot)
		}
		if sym == 1 {
			r := ((c.a % c.p) + c.p) % c.p
			if (root*root)%c.p != r {
				t.Errorf("root verification failed for (%d/%d): root %d squared mod %d != %d", c.a, c.p, root, c.p, r)
			}
		}
	}
}

func TestLegendreGen_FactsAndOutputs(t *testing.T) {
	gen := &legendreGen{}
	seenSymbols := make(map[string]int)

	for i := 0; i < 200; i++ {
		d := rand.Float64()
		p := gen.Generate(generator.GeneratorContext{Difficulty: d})

		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Fatalf("empty fields at difficulty %f: %+v", d, p)
		}

		if p.Answer != "-1" && p.Answer != "0" && p.Answer != "1" {
			t.Fatalf("unexpected Answer %q, want -1, 0 or 1", p.Answer)
		}
		seenSymbols[p.Answer]++

		aStr, okA := p.Facts["a"]
		pStr, okP := p.Facts["p"]
		resStr, okRes := p.Facts["residue"]
		symStr, okSym := p.Facts["symbol"]
		ansStr, okAns := p.Facts["answer"]
		if !okA || !okP || !okRes || !okSym || !okAns {
			t.Fatalf("missing required fact in %+v", p.Facts)
		}

		a, errA := strconv.Atoi(aStr)
		prime, errP := strconv.Atoi(pStr)
		res, errRes := strconv.Atoi(resStr)
		sym, errSym := strconv.Atoi(symStr)
		if errA != nil || errP != nil || errRes != nil || errSym != nil {
			t.Fatalf("failed parsing facts: a=%v p=%v res=%v sym=%v", errA, errP, errRes, errSym)
		}

		if p.Answer != symStr || ansStr != symStr {
			t.Errorf("answer mismatch: Answer=%q ansFact=%q symFact=%q", p.Answer, ansStr, symStr)
		}

		wantSym, root := computeLegendreSymbol(a, prime)
		if sym != wantSym {
			t.Errorf("symbol mismatch for (%d/%d): got %d, want %d", a, prime, sym, wantSym)
		}

		expectedRes := ((a % prime) + prime) % prime
		if res != expectedRes {
			t.Errorf("residue mismatch: got %d, want %d", res, expectedRes)
		}

		if sym == 1 && (root*root)%prime != res {
			t.Errorf("witness root %d squared mod %d != %d", root, prime, res)
		}
	}

	// Verify that all 3 outcomes (-1, 0, 1) are generated across 200 samples.
	for _, expected := range []string{"-1", "0", "1"} {
		if seenSymbols[expected] == 0 {
			t.Errorf("never generated outcome %q across 200 samples", expected)
		}
	}
}
