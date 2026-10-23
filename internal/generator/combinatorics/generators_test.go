package combinatorics

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
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

func TestRegistration(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	expected := []string{
		"discrete.combinatorics.counting_principle",
		"discrete.combinatorics.factorial",
		"discrete.combinatorics.multinomial",
		"discrete.combinatorics.circular",
		"discrete.combinatorics.derangement",
		"discrete.combinatorics.catalan",
		"discrete.combinatorics.vandermonde",
		"discrete.combinatorics.double_counting",
		"discrete.combinatorics.surjections",
		"discrete.combinatorics.stirling_second",
		"discrete.combinatorics.bell",
		"discrete.combinatorics.integer_partitions",
		"discrete.combinatorics.lattice_paths",
		"discrete.combinatorics.compositions",
	}

	for _, id := range expected {
		gen, err := reg.Get(id)
		if err != nil || gen == nil {
			t.Errorf("concept %q not registered: %v", id, err)
		}
	}
}

func TestCombinatoricsAlgorithms(t *testing.T) {
	// Factorials
	factCases := map[int]int{0: 1, 1: 1, 2: 2, 3: 6, 4: 24, 5: 120, 6: 720, 7: 5040}
	for n, want := range factCases {
		if got := fact(n); got != want {
			t.Errorf("fact(%d) = %d, want %d", n, got, want)
		}
	}

	// Choose
	chooseCases := []struct {
		n, r, want int
	}{
		{5, 0, 1},
		{5, 1, 5},
		{5, 2, 10},
		{5, 3, 10},
		{5, 5, 1},
		{5, 6, 0},
		{6, 3, 20},
		{10, 4, 210},
	}
	for _, c := range chooseCases {
		if got := choose(c.n, c.r); got != c.want {
			t.Errorf("choose(%d, %d) = %d, want %d", c.n, c.r, got, c.want)
		}
	}

	// Derangements
	derangeCases := map[int]int{0: 1, 1: 0, 2: 1, 3: 2, 4: 9, 5: 44, 6: 265}
	for n, want := range derangeCases {
		if got := derangements(n); got != want {
			t.Errorf("derangements(%d) = %d, want %d", n, got, want)
		}
	}

	// Catalan numbers
	catalanCases := map[int]int{0: 1, 1: 1, 2: 2, 3: 5, 4: 14, 5: 42, 6: 132, 7: 429}
	for n, want := range catalanCases {
		if got := catalan(n); got != want {
			t.Errorf("catalan(%d) = %d, want %d", n, got, want)
		}
	}

	// Stirling numbers of the second kind S(m, n)
	stirlingCases := []struct {
		m, n, want int
	}{
		{2, 1, 1},
		{3, 2, 3},
		{4, 2, 7},
		{4, 3, 6},
		{5, 2, 15},
		{5, 3, 25},
		{5, 5, 1},
		{5, 0, 0},
		{4, 5, 0},
	}
	for _, c := range stirlingCases {
		if got := stirling2(c.m, c.n); got != c.want {
			t.Errorf("stirling2(%d, %d) = %d, want %d", c.m, c.n, got, c.want)
		}
	}

	// Surjection counts: ontoCount(m, n) = n! * S(m, n)
	for _, c := range stirlingCases {
		if c.n > 0 && c.m >= c.n {
			wantOnto := fact(c.n) * c.want
			if got := ontoCount(c.m, c.n); got != wantOnto {
				t.Errorf("ontoCount(%d, %d) = %d, want %d", c.m, c.n, got, wantOnto)
			}
		}
	}

	// Integer powers
	if got := pow(2, 5); got != 32 {
		t.Errorf("pow(2, 5) = %d, want 32", got)
	}
	if got := pow(3, 3); got != 27 {
		t.Errorf("pow(3, 3) = %d, want 27", got)
	}
}

func TestFuzzAllConcepts(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	for _, id := range reg.Concepts() {
		gen, err := reg.Get(id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				diff := rand.Float64()
				p := gen.Generate(generator.GeneratorContext{Difficulty: diff})
				if p.Question == "" || p.Answer == "" || p.Explanation == "" {
					t.Fatalf("empty fields at difficulty=%.2f: q=%q a=%q e=%q", diff, p.Question, p.Answer, p.Explanation)
				}
				// All combinatorics answers are integers.
				if _, err := strconv.Atoi(p.Answer); err != nil {
					t.Errorf("answer %q does not parse as integer: %v", p.Answer, err)
				}
				if len(p.Facts) == 0 {
					t.Errorf("empty facts map for %s", id)
				}
			}
		})
	}
}

func TestPoolDiversity(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	for _, id := range reg.Concepts() {
		gen, err := reg.Get(id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		answers := make(map[string]bool)
		for i := 0; i < 150; i++ {
			p := gen.Generate(generator.GeneratorContext{Difficulty: 0.5})
			answers[p.Answer] = true
		}
		if len(answers) < 2 {
			t.Errorf("concept %s produced only %d distinct answer(s) across 150 samples", id, len(answers))
		}
	}
}

func fuzzNumericGen(t *testing.T, gen generator.Generator) {
	t.Helper()
	gr := grader.NewRouter()
	for i := 0; i < 100; i++ {
		d := rand.Float64()
		p := gen.Generate(generator.GeneratorContext{Difficulty: d})
		if p.Question == "" || p.Answer == "" || p.Explanation == "" {
			t.Fatalf("empty field at difficulty=%.2f: q=%q a=%q e=%q", d, p.Question, p.Answer, p.Explanation)
		}
		if _, err := strconv.Atoi(p.Answer); err != nil {
			t.Fatalf("answer %q is not an integer", p.Answer)
		}
		res := gr.Grade(grader.GradingNumeric, p.Answer, p.Answer)
		if !res.Correct {
			t.Fatalf("self-grade failed for question %q with answer %q: %s", p.Question, p.Answer, res.Feedback)
		}
	}
}

func TestCountingPrincipleGen(t *testing.T) { fuzzNumericGen(t, &countingPrincipleGen{}) }
func TestFactorialGen(t *testing.T)         { fuzzNumericGen(t, &factorialGen{}) }
func TestMultinomialGen(t *testing.T)       { fuzzNumericGen(t, &multinomialGen{}) }
func TestCircularGen(t *testing.T)          { fuzzNumericGen(t, &circularGen{}) }
func TestDerangementGen(t *testing.T)       { fuzzNumericGen(t, &derangementGen{}) }
func TestCatalanGen(t *testing.T)           { fuzzNumericGen(t, &catalanGen{}) }
func TestVandermondeGen(t *testing.T)       { fuzzNumericGen(t, &vandermondeGen{}) }
func TestDoubleCountingGen(t *testing.T)    { fuzzNumericGen(t, &doubleCountingGen{}) }
func TestSurjectionGen(t *testing.T)        { fuzzNumericGen(t, &surjectionGen{}) }
func TestStirlingSecondGen(t *testing.T)    { fuzzNumericGen(t, &stirlingSecondGen{}) }
func TestBellGen(t *testing.T)              { fuzzNumericGen(t, &bellGen{}) }
func TestIntegerPartitionGen(t *testing.T)  { fuzzNumericGen(t, &integerPartitionGen{}) }
func TestLatticePathGen(t *testing.T)       { fuzzNumericGen(t, &latticePathGen{}) }
func TestCompositionGen(t *testing.T)       { fuzzNumericGen(t, &compositionGen{}) }

func TestFuzz(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)
	gr := grader.NewRouter()

	for _, id := range reg.Concepts() {
		gen, err := reg.Get(id)
		if err != nil {
			t.Fatalf("missing generator for registered concept %s: %v", id, err)
		}
		t.Run(id, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				d := rand.Float64()
				p := gen.Generate(generator.GeneratorContext{Difficulty: d})
				if p.Question == "" || p.Answer == "" || p.Explanation == "" {
					t.Fatalf("empty field at difficulty=%.2f: q=%q a=%q", d, p.Question, p.Answer)
				}
				val, err := strconv.Atoi(p.Answer)
				if err != nil {
					t.Fatalf("answer %q is not an integer", p.Answer)
				}
				if val < 0 {
					t.Fatalf("combinatorics answer %d is negative", val)
				}
				res := gr.Grade(grader.GradingNumeric, p.Answer, p.Answer)
				if !res.Correct {
					t.Fatalf("self-grade failed for answer %q", p.Answer)
				}
			}
		})
	}
}

func TestDifficultyScaling(t *testing.T) {
	reg := generator.NewRegistry()
	Register(reg)

	for _, id := range reg.Concepts() {
		gen, _ := reg.Get(id)
		pLow := gen.Generate(generator.GeneratorContext{Difficulty: 0.0})
		pHigh := gen.Generate(generator.GeneratorContext{Difficulty: 1.0})
		if pLow.Question == "" || pHigh.Question == "" {
			t.Fatalf("empty question at boundary difficulty for %s", id)
		}
		if pLow.Answer == "" || pHigh.Answer == "" {
			t.Fatalf("empty answer at boundary difficulty for %s", id)
		}
	}
}

func TestCombinatoricsMathHelpers(t *testing.T) {
	t.Run("fact", func(t *testing.T) {
		expected := []int{1, 1, 2, 6, 24, 120, 720}
		for n, want := range expected {
			if got := fact(n); got != want {
				t.Errorf("fact(%d) = %d, want %d", n, got, want)
			}
		}
	})

	t.Run("derangements", func(t *testing.T) {
		expected := []int{1, 0, 1, 2, 9, 44, 265}
		for n, want := range expected {
			if got := derangements(n); got != want {
				t.Errorf("derangements(%d) = %d, want %d", n, got, want)
			}
		}
	})

	t.Run("catalan", func(t *testing.T) {
		expected := []int{1, 1, 2, 5, 14, 42, 132}
		for n, want := range expected {
			if got := catalan(n); got != want {
				t.Errorf("catalan(%d) = %d, want %d", n, got, want)
			}
		}
	})

	t.Run("choose", func(t *testing.T) {
		cases := []struct {
			n, r, want int
		}{
			{5, 0, 1},
			{5, 1, 5},
			{5, 2, 10},
			{5, 3, 10},
			{5, 5, 1},
			{5, -1, 0},
			{3, 5, 0},
		}
		for _, c := range cases {
			if got := choose(c.n, c.r); got != c.want {
				t.Errorf("choose(%d, %d) = %d, want %d", c.n, c.r, got, c.want)
			}
		}
	})

	t.Run("stirling2", func(t *testing.T) {
		if got := stirling2(4, 2); got != 7 {
			t.Errorf("stirling2(4, 2) = %d, want 7", got)
		}
		if got := stirling2(5, 3); got != 25 {
			t.Errorf("stirling2(5, 3) = %d, want 25", got)
		}
		if got := stirling2(5, 0); got != 0 {
			t.Errorf("stirling2(5, 0) = %d, want 0", got)
		}
	})

	t.Run("ontoCountMatchesStirlingTheorem", func(t *testing.T) {
		// Theorem: ontoCount(m, n) == n! * S(m, n)
		for m := 2; m <= 6; m++ {
			for n := 1; n <= m; n++ {
				want := fact(n) * stirling2(m, n)
				got := ontoCount(m, n)
				if got != want {
					t.Errorf("ontoCount(%d, %d) = %d, want %d (%d! * S(%d,%d))", m, n, got, want, n, m, n)
				}
			}
		}
	})

	t.Run("composition", func(t *testing.T) {
		for n := 3; n <= 10; n++ {
			for k := 1; k <= n; k++ {
				parts := composition(n, k)
				if len(parts) != k {
					t.Fatalf("composition(%d, %d) returned %d parts, want %d", n, k, len(parts), k)
				}
				sum := 0
				for _, p := range parts {
					if p < 1 {
						t.Fatalf("composition(%d, %d) contained non-positive part %d", n, k, p)
					}
					sum += p
				}
				if sum != n {
					t.Fatalf("composition(%d, %d) parts sum to %d, want %d", n, k, sum, n)
				}
			}
		}
	})

	t.Run("productChain", func(t *testing.T) {
		if got := productChain(1); got != "1" {
			t.Errorf("productChain(1) = %q, want \"1\"", got)
		}
		if got := productChain(4); got != "1 x 2 x 3 x 4" {
			t.Errorf("productChain(4) = %q, want \"1 x 2 x 3 x 4\"", got)
		}
	})
}
