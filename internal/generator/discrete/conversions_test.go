package discrete

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func TestDiscreteConversionsFuzz(t *testing.T) {
	gens := map[string]generator.Generator{
		"discrete.combinatorics.stars_bars":     &starsBarsConvertedGen{},
		"discrete.combinatorics.pigeonhole":      &pigeonholeConvertedGen{},
		"discrete.sets.inclusion_exclusion":     &inclusionExclusionConvertedGen{},
	}

	for id, gen := range gens {
		t.Run(id, func(t *testing.T) {
			answers := make(map[string]bool)
			for i := 0; i < 150; i++ {
				diff := rand.Float64()
				p := gen.Generate(generator.GeneratorContext{Difficulty: diff})
				if p.Question == "" || p.Answer == "" || p.Explanation == "" {
					t.Fatalf("empty field at difficulty=%.2f: %+v", diff, p)
				}
				val, err := strconv.Atoi(p.Answer)
				if err != nil {
					t.Fatalf("answer %q not an integer at difficulty=%.2f: %v", p.Answer, diff, err)
				}
				if val < 0 {
					t.Fatalf("answer %d must be non-negative: %+v", val, p)
				}
				answers[p.Answer] = true
				if p.Facts == nil || p.Facts["answer"] != p.Answer {
					t.Fatalf("missing or inconsistent answer fact: %+v", p.Facts)
				}
			}
			if len(answers) < 4 {
				t.Errorf("generator %s produced only %d distinct answers over 150 samples", id, len(answers))
			}
		})
	}
}

func TestStarsBarsCorrectness(t *testing.T) {
	// nCr formula check
	cases := []struct {
		n, k, want int
	}{
		{4, 1, 4},
		{4, 2, 6},
		{5, 2, 10},
		{6, 3, 20},
		{7, 4, 35},
	}
	for _, c := range cases {
		got := nCr(c.n, c.k)
		if got != c.want {
			t.Errorf("nCr(%d, %d) = %d, want %d", c.n, c.k, got, c.want)
		}
	}

	gen := &starsBarsConvertedGen{}
	for i := 0; i < 50; i++ {
		p := gen.Generate(generator.GeneratorContext{Difficulty: 0.5})
		n, _ := strconv.Atoi(p.Facts["n"])
		k, _ := strconv.Atoi(p.Facts["k"])
		posOnly := p.Facts["positive_only"] == "true"
		ans, _ := strconv.Atoi(p.Answer)

		var expected int
		if posOnly {
			expected = nCr(n-1, k-1)
		} else {
			expected = nCr(n+k-1, k-1)
		}
		if ans != expected {
			t.Errorf("stars_bars mismatch: n=%d k=%d posOnly=%v got=%d want=%d", n, k, posOnly, ans, expected)
		}
	}
}

func TestPigeonholeCorrectness(t *testing.T) {
	gen := &pigeonholeConvertedGen{}
	for i := 0; i < 50; i++ {
		p := gen.Generate(generator.GeneratorContext{Difficulty: 0.5})
		ans, _ := strconv.Atoi(p.Answer)
		if p.Facts["holes"] != "" && p.Facts["target"] != "" {
			k, _ := strconv.Atoi(p.Facts["holes"])
			m, _ := strconv.Atoi(p.Facts["target"])
			expected := k*(m-1) + 1
			if ans != expected {
				t.Errorf("pigeonhole target mismatch: k=%d m=%d got=%d want=%d", k, m, ans, expected)
			}
		} else if p.Facts["holes"] != "" && p.Facts["items"] != "" {
			k, _ := strconv.Atoi(p.Facts["holes"])
			n, _ := strconv.Atoi(p.Facts["items"])
			expected := (n + k - 1) / k
			if ans != expected {
				t.Errorf("pigeonhole ceil mismatch: n=%d k=%d got=%d want=%d", n, k, ans, expected)
			}
		} else if p.Facts["modulus"] != "" {
			mod, _ := strconv.Atoi(p.Facts["modulus"])
			expected := mod + 1
			if ans != expected {
				t.Errorf("pigeonhole modulus mismatch: mod=%d got=%d want=%d", mod, ans, expected)
			}
		}
	}
}

func TestInclusionExclusionCorrectness(t *testing.T) {
	gen := &inclusionExclusionConvertedGen{}
	for i := 0; i < 50; i++ {
		p := gen.Generate(generator.GeneratorContext{Difficulty: 0.5})
		ans, _ := strconv.Atoi(p.Answer)
		if p.Facts["setA"] != "" && p.Facts["setB"] != "" && p.Facts["intersection"] != "" {
			a, _ := strconv.Atoi(p.Facts["setA"])
			b, _ := strconv.Atoi(p.Facts["setB"])
			both, _ := strconv.Atoi(p.Facts["intersection"])
			expected := a + b - both
			if ans != expected {
				t.Errorf("PIE union mismatch: a=%d b=%d both=%d got=%d want=%d", a, b, both, ans, expected)
			}
		} else if p.Facts["setA"] != "" && p.Facts["setB"] != "" && p.Facts["union"] != "" {
			a, _ := strconv.Atoi(p.Facts["setA"])
			b, _ := strconv.Atoi(p.Facts["setB"])
			union, _ := strconv.Atoi(p.Facts["union"])
			expected := a + b - union
			if ans != expected {
				t.Errorf("PIE intersection mismatch: a=%d b=%d union=%d got=%d want=%d", a, b, union, ans, expected)
			}
		} else if p.Facts["setA"] != "" && p.Facts["intersection"] != "" {
			a, _ := strconv.Atoi(p.Facts["setA"])
			both, _ := strconv.Atoi(p.Facts["intersection"])
			expected := a - both
			if ans != expected {
				t.Errorf("PIE difference mismatch: a=%d both=%d got=%d want=%d", a, both, ans, expected)
			}
		} else if p.Facts["N"] != "" {
			N, _ := strconv.Atoi(p.Facts["N"])
			pVal, _ := strconv.Atoi(p.Facts["p"])
			qVal, _ := strconv.Atoi(p.Facts["q"])
			expected := N - (N / pVal) - (N / qVal) + (N / (pVal * qVal))
			if ans != expected {
				t.Errorf("PIE neither mismatch: N=%d p=%d q=%d got=%d want=%d", N, pVal, qVal, ans, expected)
			}
		}
	}
}
