package combinatorics

import (
	"math/rand"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
)

func fuzzGen(t *testing.T, gen generator.Generator) {
	t.Helper()
	for i := 0; i < 50; i++ {
		d := rand.Float64()
		p := gen.Generate(generator.GeneratorContext{Difficulty: d})
		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Fatalf("empty field at difficulty=%.2f: q=%q a=%q e=%q", d, p.Question, p.Answer, p.Explanation)
		}
	}
}

func TestFact(t *testing.T) {
	cases := []struct {
		n    int
		want int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 6},
		{4, 24},
		{5, 120},
	}
	for _, tc := range cases {
		if got := fact(tc.n); got != tc.want {
			t.Errorf("fact(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestDerangements(t *testing.T) {
	cases := []struct {
		n    int
		want int
	}{
		{0, 1},
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 9},
		{5, 44},
	}
	for _, tc := range cases {
		if got := derangements(tc.n); got != tc.want {
			t.Errorf("derangements(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestCatalan(t *testing.T) {
	cases := []struct {
		n    int
		want int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 5},
		{4, 14},
		{5, 42},
	}
	for _, tc := range cases {
		if got := catalan(tc.n); got != tc.want {
			t.Errorf("catalan(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestChoose(t *testing.T) {
	if got := choose(5, -1); got != 0 {
		t.Errorf("choose(5, -1) = %d, want 0", got)
	}
	if got := choose(5, 6); got != 0 {
		t.Errorf("choose(5, 6) = %d, want 0", got)
	}
	if got := choose(5, 0); got != 1 {
		t.Errorf("choose(5, 0) = %d, want 1", got)
	}
	if got := choose(5, 2); got != 10 {
		t.Errorf("choose(5, 2) = %d, want 10", got)
	}
	if got := choose(5, 3); got != 10 {
		t.Errorf("choose(5, 3) = %d, want 10", got)
	}
}

func TestPick(t *testing.T) {
	if got := pick(0); got != 0 {
		t.Errorf("pick(0) = %d, want 0", got)
	}
	if got := pick(-5); got != 0 {
		t.Errorf("pick(-5) = %d, want 0", got)
	}
}

func TestSpan(t *testing.T) {
	if got := span(0.0, 2, 5); got != 2 {
		t.Errorf("span(0.0) = %d, want 2", got)
	}
	if got := span(1.0, 2, 5); got != 7 {
		t.Errorf("span(1.0) = %d, want 7", got)
	}
}

func TestProductChain(t *testing.T) {
	if got := productChain(0); got != "1" {
		t.Errorf("productChain(0) = %q, want 1", got)
	}
	if got := productChain(1); got != "1" {
		t.Errorf("productChain(1) = %q, want 1", got)
	}
	if got := productChain(4); got != "1 x 2 x 3 x 4" {
		t.Errorf("productChain(4) = %q, want 1 x 2 x 3 x 4", got)
	}
}

func TestComposition(t *testing.T) {
	if got := composition(5, 0); got != nil {
		t.Errorf("composition(5, 0) = %v, want nil", got)
	}
	parts := composition(7, 3)
	if len(parts) != 3 {
		t.Fatalf("len(composition(7, 3)) = %d, want 3", len(parts))
	}
	sum := 0
	for _, p := range parts {
		if p < 1 {
			t.Errorf("part %d < 1", p)
		}
		sum += p
	}
	if sum != 7 {
		t.Errorf("sum = %d, want 7", sum)
	}
}

func TestGenerators(t *testing.T) {
	fuzzGen(t, &countingPrincipleGen{})
	fuzzGen(t, &factorialGen{})
	fuzzGen(t, &multinomialGen{})
	fuzzGen(t, &circularGen{})
	fuzzGen(t, &derangementGen{})
	fuzzGen(t, &catalanGen{})
}

func TestRegistry(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)
	for _, id := range reg.Concepts() {
		gen, err := reg.Get(id)
		if err != nil || gen == nil {
			t.Fatalf("generator not registered for %s: %v", id, err)
		}
		t.Run(id, func(t *testing.T) {
			fuzzGen(t, gen)
		})
	}
}
