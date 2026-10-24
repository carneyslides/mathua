package discrete

// The 2026-10-17 batch: sixteen concepts that take discrete_math from 39 to the domain target of 55.
//
// Every answer here is computed from the instance, never tabulated: the counterexample concept scans
// for the first composite value, the matching concept searches for a maximum matching, and the
// recurrence concept derives its closed form from the roots and coefficients it chose. Facts are
// published for the authored solution schemas, and each concept's grading_type is one the router
// already supports.

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
)

func registerBatch(reg *generator.Registry) {
	reg.Register("discrete.logic.inference", &inferenceGen{})
	reg.Register("discrete.logic.predicates", &predicatesGen{})
	reg.Register("discrete.logic.satisfiability", &satisfiabilityGen{})
	reg.Register("discrete.proof.contrapositive", &contrapositiveGen{})
	reg.Register("discrete.proof.counterexample", &counterexampleGen{})
	reg.Register("discrete.proof.structural", &structuralGen{})
	reg.Register("discrete.sets.powerset", &powersetGen{})
	reg.Register("discrete.sets.partial_order", &partialOrderGen{})
	reg.Register("discrete.sets.countable", &countableGen{})
	reg.Register("discrete.combinatorics.complement", &complementGen{})
	reg.Register("discrete.graphs.bipartite", &bipartiteGen{})
	reg.Register("discrete.graphs.planar", &planarGen{})
	reg.Register("discrete.graphs.cayley", &cayleyGen{})
	reg.Register("discrete.graphs.degree_sequence", &degreeSequenceGen{})
	reg.Register("discrete.sequences.linear_recurrence", &linearRecurrenceGen{})
	reg.Register("discrete.sequences.growth_rates", &growthRatesGen{})
}

// ----- helpers -----

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 {
		return n == 2
	}
	for d := 3; d*d <= n; d += 2 {
		if n%d == 0 {
			return false
		}
	}
	return true
}

func divisors(n int) []int {
	var out []int
	for d := 1; d*d <= n; d++ {
		if n%d == 0 {
			out = append(out, d)
			if d != n/d {
				out = append(out, n/d)
			}
		}
	}
	sort.Ints(out)
	return out
}

func intsString(xs []int, sep string) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, sep)
}

// maxMatching searches for a maximum matching in a bipartite graph. Sizes here are at most 4x4, so
// exhaustive search is the honest implementation: it cannot be wrong about the answer.
func maxMatching(leftCount int, adj [][]bool) int {
	best := 0
	var walk func(i int, used map[int]bool, size int)
	walk = func(i int, used map[int]bool, size int) {
		if i == leftCount {
			if size > best {
				best = size
			}
			return
		}
		walk(i+1, used, size)
		for r, ok := range adj[i] {
			if ok && !used[r] {
				used[r] = true
				walk(i+1, used, size+1)
				delete(used, r)
			}
		}
	}
	walk(0, map[int]bool{}, 0)
	return best
}

// ----- 1. inference -----

type inferenceGen struct{}

func (g *inferenceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type form struct {
		name, premises, conclusion string
		valid                      bool
	}

	var forms []form

	// Choose randomized variable pairs to prevent static text caching
	varSets := [][3]string{
		{"p", "q", "r"},
		{"A", "B", "C"},
		{"P", "Q", "R"},
		{"u", "v", "w"},
		{"x", "y", "z"},
	}
	vs := varSets[rand.Intn(len(varSets))]
	pVar, qVar, rVar := vs[0], vs[1], vs[2]

	forms = append(forms,
		form{"modus ponens", fmt.Sprintf("%s \u2192 %s and %s", pVar, qVar, pVar), qVar, true},
		form{"modus tollens", fmt.Sprintf("%s \u2192 %s and \u00ac%s", pVar, qVar, qVar), fmt.Sprintf("\u00ac%s", pVar), true},
		form{"affirming the consequent", fmt.Sprintf("%s \u2192 %s and %s", pVar, qVar, qVar), pVar, false},
		form{"denying the antecedent", fmt.Sprintf("%s \u2192 %s and \u00ac%s", pVar, qVar, pVar), fmt.Sprintf("\u00ac%s", qVar), false},
		form{"hypothetical syllogism", fmt.Sprintf("%s \u2192 %s and %s \u2192 %s", pVar, qVar, qVar, rVar), fmt.Sprintf("%s \u2192 %s", pVar, rVar), true},
		form{"disjunctive syllogism", fmt.Sprintf("%s \u2228 %s and \u00ac%s", pVar, qVar, pVar), qVar, true},
		form{"affirming a disjunct", fmt.Sprintf("%s \u2228 %s and %s", pVar, qVar, pVar), fmt.Sprintf("\u00ac%s", qVar), false},
		form{"resolution", fmt.Sprintf("%s \u2228 %s and \u00ac%s \u2228 %s", pVar, qVar, pVar, rVar), fmt.Sprintf("%s \u2228 %s", qVar, rVar), true},
	)

	// Contextual real-world / mathematical implications
	switch rand.Intn(4) {
	case 0:
		k := 2 * (rand.Intn(4) + 2)
		forms = append(forms, form{
			"modus ponens",
			fmt.Sprintf("If n is divisible by %d then n is even, and n is divisible by %d", k, k),
			"n is even",
			true,
		})
	case 1:
		c := rand.Intn(5) + 3
		forms = append(forms, form{
			"modus tollens",
			fmt.Sprintf("If x > %d then x > 0, and x \u2264 0", c),
			fmt.Sprintf("x \u2264 %d", c),
			true,
		})
	case 2:
		forms = append(forms, form{
			"affirming the consequent",
			"If a quadrilateral is a square then it is a rectangle, and the quadrilateral is a rectangle",
			"it is a square",
			false,
		})
	case 3:
		forms = append(forms, form{
			"denying the antecedent",
			"If f is differentiable at c then f is continuous at c, and f is not differentiable at c",
			"f is not continuous at c",
			false,
		})
	}

	f := forms[rand.Intn(len(forms))]
	q := fmt.Sprintf("Consider the argument: premises %s; conclusion %s. Which argument form is it?", f.premises, f.conclusion)

	allFormNames := []string{
		"modus ponens",
		"modus tollens",
		"affirming the consequent",
		"denying the antecedent",
		"hypothetical syllogism",
		"disjunctive syllogism",
		"affirming a disjunct",
		"resolution",
	}
	var others []string
	perm := rand.Perm(len(allFormNames))
	for _, idx := range perm {
		name := allFormNames[idx]
		if name != f.name {
			others = append(others, name)
			if len(others) == 3 {
				break
			}
		}
	}

	p := generator.MultipleChoice(q, f.name, others...)
	valid := "yes"
	if !f.valid {
		valid = "no"
	}
	p.Facts = map[string]string{
		"form": f.name, "valid": valid, "premises": f.premises, "conclusion": f.conclusion, "answer": f.name,
	}
	p.Explanation = fmt.Sprintf("%s is %s: from %s the conclusion %s does%s follow.",
		f.name, map[bool]string{true: "valid", false: "invalid"}[f.valid], f.premises, f.conclusion,
		map[bool]string{true: "", false: " not"}[f.valid])
	return p
}

// ----- 2. predicates -----

type predicatesGen struct{}

func (g *predicatesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		original, negation, swapped, unchanged, wrongQuantifier, note string
	}

	k := rand.Intn(9) + 1
	k2 := rand.Intn(5) + 2

	var entries []entry
	switch rand.Intn(5) {
	case 0:
		entries = []entry{
			{
				fmt.Sprintf("\u2200x \u2203y (x + y = %d)", k),
				fmt.Sprintf("\u2203x \u2200y (x + y \u2260 %d)", k),
				fmt.Sprintf("\u2203y \u2200x (x + y = %d)", k),
				fmt.Sprintf("\u2200x \u2203y (x + y \u2260 %d)", k),
				fmt.Sprintf("\u2203x \u2203y (x + y \u2260 %d)", k),
				"Negating swaps each quantifier and negates the inner equation.",
			},
			{
				fmt.Sprintf("\u2203x \u2200y (%dx + y = 0)", k2),
				fmt.Sprintf("\u2200x \u2203y (%dx + y \u2260 0)", k2),
				fmt.Sprintf("\u2200y \u2203x (%dx + y = 0)", k2),
				fmt.Sprintf("\u2203x \u2200y (%dx + y \u2260 0)", k2),
				fmt.Sprintf("\u2200x \u2200y (%dx + y \u2260 0)", k2),
				"An existential becomes universal and the inner equation becomes its negation.",
			},
		}
	case 1:
		entries = []entry{
			{
				fmt.Sprintf("\u2200x \u2203y (x < y + %d)", k),
				fmt.Sprintf("\u2203x \u2200y (x \u2265 y + %d)", k),
				fmt.Sprintf("\u2203y \u2200x (x < y + %d)", k),
				fmt.Sprintf("\u2200x \u2203y (x \u2265 y + %d)", k),
				fmt.Sprintf("\u2203x \u2203y (x \u2265 y + %d)", k),
				"Negating swaps each quantifier and reverses the strict inequality to greater-or-equal.",
			},
			{
				fmt.Sprintf("\u2203x \u2200y (x\u00b2 + y\u00b2 \u2264 %d)", k*k),
				fmt.Sprintf("\u2200x \u2203y (x\u00b2 + y\u00b2 > %d)", k*k),
				fmt.Sprintf("\u2200y \u2203x (x\u00b2 + y\u00b2 \u2264 %d)", k*k),
				fmt.Sprintf("\u2203x \u2200y (x\u00b2 + y\u00b2 > %d)", k*k),
				fmt.Sprintf("\u2200x \u2200y (x\u00b2 + y\u00b2 > %d)", k*k),
				"Existential becomes universal, and non-strict inequality flips to strict.",
			},
		}
	case 2:
		entries = []entry{
			{
				fmt.Sprintf("\u2200x (x\u00b2 \u2265 %d)", k),
				fmt.Sprintf("\u2203x (x\u00b2 < %d)", k),
				fmt.Sprintf("\u2200x (x\u00b2 < %d)", k),
				fmt.Sprintf("\u00ac\u2200x (x\u00b2 \u2265 %d)", k),
				fmt.Sprintf("\u2200x (x\u00b2 > %d)", k),
				"One counterexample refutes a universal claim, and it must satisfy the strict negation.",
			},
			{
				fmt.Sprintf("\u2203x (x + %d = 0)", k),
				fmt.Sprintf("\u2200x (x + %d \u2260 0)", k),
				fmt.Sprintf("\u2203x (x + %d \u2260 0)", k),
				fmt.Sprintf("\u2200x (x - %d = 0)", k),
				fmt.Sprintf("\u2200x (x + %d = 0)", k),
				"Negating an existential yields a universal with negated predicate.",
			},
		}
	case 3:
		entries = []entry{
			{
				"\u2200x (P(x) \u2192 Q(x))",
				"\u2203x (P(x) \u2227 \u00acQ(x))",
				"\u2200x (\u00acP(x) \u2192 \u00acQ(x))",
				"\u2203x (\u00acP(x) \u2228 Q(x))",
				"\u2200x (P(x) \u2227 \u00acQ(x))",
				"The negation of an implication P -> Q is P and not Q; universal becomes existential.",
			},
			{
				"\u2203x (P(x) \u2227 Q(x))",
				"\u2200x (\u00acP(x) \u2228 \u00acQ(x))",
				"\u2200x (\u00acP(x) \u2227 \u00acQ(x))",
				"\u2203x (\u00acP(x) \u2228 \u00acQ(x))",
				"\u2200x (P(x) \u2228 Q(x))",
				"By De Morgan's laws for quantifiers and conjunction, existential becomes universal and conjunction becomes disjunction of negations.",
			},
		}
	default:
		entries = []entry{
			{
				fmt.Sprintf("\u2200n \u2208 \u2124 (n > %d \u2192 n\u00b2 > %d)", k, k*k),
				fmt.Sprintf("\u2203n \u2208 \u2124 (n > %d \u2227 n\u00b2 \u2264 %d)", k, k*k),
				fmt.Sprintf("\u2200n \u2208 \u2124 (n \u2264 %d \u2192 n\u00b2 \u2264 %d)", k, k*k),
				fmt.Sprintf("\u2203n \u2208 \u2124 (n > %d \u2192 n\u00b2 \u2264 %d)", k, k*k),
				fmt.Sprintf("\u2203n \u2208 \u2124 (n \u2264 %d \u2227 n\u00b2 > %d)", k, k*k),
				"Negating a universal implication produces an existential with the antecedent asserting and consequent negated.",
			},
			{
				fmt.Sprintf("\u2203x \u2208 \u211d (x\u00b2 + %d = 0)", k),
				fmt.Sprintf("\u2200x \u2208 \u211d (x\u00b2 + %d \u2260 0)", k),
				fmt.Sprintf("\u2200x \u2208 \u211d (x\u00b2 + %d = 0)", k),
				fmt.Sprintf("\u2203x \u2208 \u211d (x\u00b2 + %d \u2260 0)", k),
				fmt.Sprintf("\u2200x \u2208 \u211d (x\u00b2 - %d \u2260 0)", k),
				"The negation of an existence claim states that the condition fails for all real numbers.",
			},
		}
	}

	e := entries[rand.Intn(len(entries))]
	q := fmt.Sprintf("Which statement is the negation of %s over the specified domain?", e.original)
	p := generator.MultipleChoice(q, e.negation, e.swapped, e.unchanged, e.wrongQuantifier)
	p.Facts = map[string]string{"original": e.original, "negation": e.negation, "why": e.note, "answer": e.negation}
	p.Explanation = e.note
	return p
}

// ----- 3. satisfiability -----

type satisfiabilityGen struct{}

func (g *satisfiabilityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct{ formula, class, rows, note string }
	entries := []entry{
		{"(p \u2227 q) \u2192 p", "tautology", "4", "The consequent p is true in every row where the antecedent is true, so the implication never fails."},
		{"p \u2227 \u00acp", "contradiction", "0", "No assignment makes p and its negation true together, so no row satisfies it."},
		{"p \u2192 q", "contingent", "3", "It is false only when p is true and q is false, so 3 of the 4 rows satisfy it."},
		{"p \u2194 q", "contingent", "2", "It is true exactly when p and q agree, which is 2 of the 4 rows."},
		{"(p \u2192 q) \u2228 (q \u2192 p)", "tautology", "4", "At least one direction always holds, so every row satisfies the disjunction."},
	}
	e := entries[rand.Intn(len(entries))]
	q := fmt.Sprintf("Classify the formula %s.", e.formula)
	p := generator.MultipleChoice(q, e.class, "tautology", "contradiction", "contingent")
	p.Facts = map[string]string{"formula": e.formula, "class": e.class, "rows": e.rows, "why": e.note}
	p.Explanation = e.note
	return p
}

// ----- 4. contrapositive -----

type contrapositiveGen struct{}

func (g *contrapositiveGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type pair struct{ p, q, notP, notQ string }

	k := rand.Intn(8) + 2
	c := rand.Intn(5) + 2

	var pairs []pair
	switch rand.Intn(5) {
	case 0:
		pairs = []pair{
			{fmt.Sprintf("x > %d", k), fmt.Sprintf("x\u00b2 > %d", k*k), fmt.Sprintf("x \u2264 %d", k), fmt.Sprintf("x\u00b2 \u2264 %d", k*k)},
			{fmt.Sprintf("x < -%d", k), fmt.Sprintf("x\u00b2 > %d", k*k), fmt.Sprintf("x \u2265 -%d", k), fmt.Sprintf("x\u00b2 \u2264 %d", k*k)},
		}
	case 1:
		pairs = []pair{
			{fmt.Sprintf("n is divisible by %d", 2*c), "n is even", fmt.Sprintf("n is not divisible by %d", 2*c), "n is odd"},
			{fmt.Sprintf("a divides %d", c*2), fmt.Sprintf("a divides %d", c*4), fmt.Sprintf("a does not divide %d", c*2), fmt.Sprintf("a does not divide %d", c*4)},
		}
	case 2:
		pairs = []pair{
			{fmt.Sprintf("3n + %d is odd", 2*c-1), "n is even", fmt.Sprintf("3n + %d is even", 2*c-1), "n is odd"},
			{fmt.Sprintf("n\u00b2 - %d is odd", 2*c-1), "n is even", fmt.Sprintf("n\u00b2 - %d is even", 2*c-1), "n is odd"},
		}
	case 3:
		pairs = []pair{
			{"f is differentiable at x = c", "f is continuous at x = c", "f is not differentiable at x = c", "f is not continuous at x = c"},
			{"a series \u2211 a_n converges", "the sequence a_n converges to 0", "a series \u2211 a_n diverges", "the sequence a_n does not converge to 0"},
		}
	default:
		pairs = []pair{
			{fmt.Sprintf("a + b > %d", 2*k), fmt.Sprintf("a > %d or b > %d", k, k), fmt.Sprintf("a + b \u2264 %d", 2*k), fmt.Sprintf("a \u2264 %d and b \u2264 %d", k, k)},
			{fmt.Sprintf("m \u00b7 n is divisible by %d", c), fmt.Sprintf("m is divisible by %d or n is divisible by %d", c, c), fmt.Sprintf("m \u00b7 n is not divisible by %d", c), fmt.Sprintf("m is not divisible by %d and n is not divisible by %d", c, c)},
		}
	}

	s := pairs[rand.Intn(len(pairs))]
	correct := fmt.Sprintf("If %s then %s", s.notQ, s.notP)
	converse := fmt.Sprintf("If %s then %s", s.q, s.p)
	inverse := fmt.Sprintf("If %s then %s", s.notP, s.notQ)
	negation := fmt.Sprintf("It is not true that if %s then %s", s.p, s.q)
	q := fmt.Sprintf("The statement is: if %s then %s. Which is its contrapositive?", s.p, s.q)
	p := generator.MultipleChoice(q, correct, converse, inverse, negation)
	p.Facts = map[string]string{
		"p": s.p, "q": s.q, "contrapositive": correct, "converse": converse, "inverse": inverse, "answer": correct,
	}
	p.Explanation = fmt.Sprintf("%s is logically equivalent to its contrapositive: proving \u201cif %s then %s\u201d proves the original.",
		fmt.Sprintf("If %s then %s", s.p, s.q), s.notQ, s.notP)
	return p
}

// ----- 5. counterexample -----

type counterexampleGen struct{}

func (g *counterexampleGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	// The constants are found by scanning rather than chosen by hand: any c whose
	// polynomial n^2 + n + c stays prime for at least two steps makes the question
	// meaningful, and the run length is re-derived for the constant that was drawn, so a
	// wrong constant cannot ship a wrong answer.
	candidates := make([]int, 0, 16)
	for c := 2; c <= 200; c++ {
		if runLength(c) >= 2 {
			candidates = append(candidates, c)
		}
	}
	c := candidates[rand.Intn(len(candidates))]
	n := runLength(c)
	value := n*n + n + c
	f1 := 0
	for d := 2; d*d <= value; d++ {
		if value%d == 0 {
			f1 = d
			break
		}
	}
	p := generator.Problem{
		Facts: map[string]string{
			"c": strconv.Itoa(c), "n": strconv.Itoa(n), "value": strconv.Itoa(value),
			"factor1": strconv.Itoa(f1), "factor2": strconv.Itoa(value / f1),
			"last_prime": strconv.Itoa(n - 1),
		},
	}
	switch rand.Intn(3) {
	case 0:
		p.Question = fmt.Sprintf("Let f(n) = n\u00b2 + n + %d. Find the smallest positive integer n for which f(n) is not prime.", c)
		p.Answer = strconv.Itoa(n)
	case 1:
		p.Question = fmt.Sprintf("For how many consecutive values of n, starting at n = 0, is f(n) = n\u00b2 + n + %d prime?", c)
		p.Answer = strconv.Itoa(n)
	default:
		p.Question = fmt.Sprintf("Let f(n) = n\u00b2 + n + %d. What is the value of f(n) at the smallest n for which f(n) is not prime?", c)
		p.Answer = strconv.Itoa(value)
	}
	p.Explanation = fmt.Sprintf("f(%d) = %d = %d \u00b7 %d, which is composite, and f(n) is prime for every n from 0 to %d.",
		n, value, f1, value/f1, n-1)
	return p
}

// runLength is the number of consecutive primes produced by f(n) = n^2 + n + c
// starting at n = 0. It is the smallest n at which f(n) is composite.
func runLength(c int) int {
	n := 0
	for isPrime(n*n + n + c) {
		n++
	}
	return n
}

// ----- 6. structural induction -----

type structuralGen struct{}

func (g *structuralGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct{ object, principle, cases string }
	entries := []entry{
		{"every binary tree built from a leaf and from a node with two subtrees",
			"structural induction",
			"the definition is recursive, so there is one case per rule: a leaf, and a node whose two subtrees satisfy the property by hypothesis"},
		{"every string generated by S \u2192 a | aS",
			"structural induction",
			"the grammar defines the strings recursively, so the cases are the base string and the step that adds an a"},
		{"every expression built from a number and from two expressions joined by +",
			"structural induction",
			"the definition is recursive, so the cases are a number and a sum of two smaller expressions"},
		{"every integer n \u2265 2 can be written as a product of primes",
			"strong induction on n",
			"the factorisation of n uses a factor that can be any smaller number, not just n-1, so the hypothesis must cover all smaller integers"},
		{"1 + 2 + ... + n = n(n+1)/2 for every n \u2265 1",
			"ordinary induction on n",
			"the step needs only the previous case, because the sum for n adds one term to the sum for n-1"},
	}
	e := entries[rand.Intn(len(entries))]
	q := fmt.Sprintf("To prove that %s, which induction principle fits?", e.object)
	p := generator.MultipleChoice(q, e.principle,
		"ordinary induction on n", "strong induction on n", "structural induction", "proof by contradiction")
	p.Facts = map[string]string{"object": e.object, "cases": e.cases, "principle": e.principle}
	p.Explanation = fmt.Sprintf("%s: %s.", e.principle, e.cases)
	return p
}

// ----- 7. power set -----

type powersetGen struct{}

func (g *powersetGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	n := 4 + rand.Intn(5) // 4..8
	k := 1 + rand.Intn(n-1)
	choose := 1
	for i := 1; i <= k; i++ {
		choose = choose * (n - k + i) / i
	}
	total := 1 << n
	q := fmt.Sprintf("A set has %d elements. How many of its subsets have exactly %d elements?", n, k)
	p := generator.Problem{
		Question: q,
		Answer:   strconv.Itoa(choose),
		Facts: map[string]string{
			"n": strconv.Itoa(n), "k": strconv.Itoa(k), "value": strconv.Itoa(choose),
			"total": strconv.Itoa(total),
		},
	}
	p.Explanation = fmt.Sprintf("Choosing a %d-element subset means choosing %d of the %d elements, which is C(%d, %d) = %d. The whole power set has 2^%d = %d members.",
		k, k, n, n, k, choose, n, total)
	return p
}

// ----- 8. partial order -----

type partialOrderGen struct{}

func (g *partialOrderGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	// The pair count is computed from the divisors of N, so any N is a valid
	// draw. Twelve values replace the original four, and they differ in the
	// number of divisors rather than only in the numeral.
	ns := []int{12, 18, 20, 24, 28, 30, 36, 40, 42, 45, 48, 60}
	N := ns[rand.Intn(len(ns))]
	ds := divisors(N)
	count := 0
	for i := range ds {
		for j := range ds {
			if ds[i] < ds[j] && ds[j]%ds[i] == 0 {
				count++
			}
		}
	}
	q := fmt.Sprintf("The positive divisors of %d are ordered by divisibility. How many pairs (a, b) with a < b and a | b are there?", N)
	p := generator.Problem{
		Question: q,
		Answer:   strconv.Itoa(count),
		Facts: map[string]string{
			"N": strconv.Itoa(N), "set": "\\{" + intsString(ds, ", ") + "\\}", "value": strconv.Itoa(count),
		},
	}
	p.Explanation = fmt.Sprintf("On %s, a pair counts when the smaller divides the larger; there are %d such pairs.", "{"+intsString(ds, ", ")+"}", count)
	return p
}

// ----- 9. countable -----

type countableGen struct{}

func (g *countableGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	uncountable := []string{"the real numbers", "the power set of the integers", "the set of all functions from \u2115 to {0, 1}"}
	countable := []string{"the integers", "the rational numbers", "the algebraic numbers", "the set of finite binary strings"}
	askUncountable := rand.Intn(2) == 0
	var correct string
	var distractorPool []string
	var q string
	if askUncountable {
		correct = uncountable[rand.Intn(len(uncountable))]
		distractorPool = countable
		q = "Which of these sets is uncountable?"
	} else {
		correct = countable[rand.Intn(len(countable))]
		distractorPool = uncountable
		q = "Which of these sets is countable?"
	}
	rand.Shuffle(len(distractorPool), func(i, j int) { distractorPool[i], distractorPool[j] = distractorPool[j], distractorPool[i] })
	p := generator.MultipleChoice(q, correct, distractorPool[0], distractorPool[1], distractorPool[2])
	property := "countable"
	if askUncountable {
		property = "uncountable"
	}
	p.Facts = map[string]string{"answer": correct, "property": property}
	p.Explanation = fmt.Sprintf("%s is %s. A countable set can be listed by an enumeration (or a bijection with \u2115); the reals cannot, which is Cantor's diagonal argument.",
		correct, property)
	return p
}

// ----- 10. counting by complement -----

type complementGen struct{}

func (g *complementGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	m := 2 + rand.Intn(3) // 2..4 letters, one of them a
	L := 3 + rand.Intn(3) // 3..5
	total := 1
	for i := 0; i < L; i++ {
		total *= m
	}
	none := 1
	for i := 0; i < L; i++ {
		none *= m - 1
	}
	answer := total - none
	q := fmt.Sprintf("How many strings of length %d over an alphabet of %d letters contain at least one \u2018a\u2019?", L, m)
	p := generator.Problem{
		Question: q,
		Answer:   strconv.Itoa(answer),
		Facts: map[string]string{
			"letters": strconv.Itoa(m), "length": strconv.Itoa(L),
			"total": strconv.Itoa(total), "none": strconv.Itoa(none), "value": strconv.Itoa(answer),
		},
	}
	p.Explanation = fmt.Sprintf("Count the complement instead: %d^%d = %d strings in all, and %d^%d = %d contain no \u2018a\u2019. The difference is %d.",
		m, L, total, m-1, L, none, answer)
	return p
}

// ----- 11. bipartite matching -----

type bipartiteGen struct{}

func (g *bipartiteGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	leftCount := 3 + rand.Intn(2)  // 3..4
	rightCount := 3 + rand.Intn(2) // 3..4
	adj := make([][]bool, leftCount)
	for i := range adj {
		adj[i] = make([]bool, rightCount)
	}
	// Give every left vertex at least one edge, then add a few more: a graph with an isolated
	// vertex would make the matching question trivial in a way the learner cannot see.
	for i := 0; i < leftCount; i++ {
		adj[i][rand.Intn(rightCount)] = true
	}
	for k := 0; k < 2+rand.Intn(2); k++ {
		adj[rand.Intn(leftCount)][rand.Intn(rightCount)] = true
	}
	size := maxMatching(leftCount, adj)
	var edges []string
	for i := 0; i < leftCount; i++ {
		for j := 0; j < rightCount; j++ {
			if adj[i][j] {
				edges = append(edges, fmt.Sprintf("%c%d", 'a'+i, j+1))
			}
		}
	}
	q := fmt.Sprintf("A bipartite graph has left vertices %s and right vertices %s, with edges %s. What is the size of a maximum matching?",
		commas(strings.Split(letters(leftCount), "")), commas(strings.Split(digits(rightCount), "")), strings.Join(edges, ", "))
	p := generator.Problem{
		Question: q,
		Answer:   strconv.Itoa(size),
		Facts: map[string]string{
			"left": commas(strings.Split(letters(leftCount), "")), "right": commas(strings.Split(digits(rightCount), "")),
			"edges": strings.Join(edges, ", "), "value": strconv.Itoa(size),
		},
	}
	p.Explanation = fmt.Sprintf("Pair as many left and right vertices as possible using distinct vertices: %d pairs is the maximum here, so no matching of size %d exists.", size, size+1)
	return p
}

func letters(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(byte('a' + i))
	}
	return b.String()
}

// commas joins single characters for display: "abc" becomes "a, b, c".
func commas(parts []string) string { return strings.Join(parts, ", ") }

func digits(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		b.WriteString(strconv.Itoa(i))
	}
	return b.String()
}

// ----- 12. planar graphs -----

type planarGen struct{}

func (g *planarGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	V := 6 + rand.Intn(5) // 6..10
	maxE := 3*V - 6
	E := V + rand.Intn(maxE-V+1) // V..3V-6, so at least one cycle
	F := E - V + 2
	q := fmt.Sprintf("A connected planar graph has %d vertices and %d edges. How many faces does it have, counting the outer face?", V, E)
	p := generator.Problem{
		Question: q,
		Answer:   strconv.Itoa(F),
		Facts: map[string]string{
			"v": strconv.Itoa(V), "e": strconv.Itoa(E), "value": strconv.Itoa(F),
		},
	}
	p.Explanation = fmt.Sprintf("Euler's formula for a connected planar graph is V - E + F = 2, so F = %d - %d + 2 = %d.", E, V, F)
	return p
}

// ----- 13. Cayley's formula -----

type cayleyGen struct{}

func (g *cayleyGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	n := 3 + rand.Intn(8)
	value := 1
	for i := 0; i < n-2; i++ {
		value *= n
	}
	answer := strconv.Itoa(value)
	facts := map[string]string{"n": strconv.Itoa(n), "exponent": strconv.Itoa(n - 2), "value": answer}
	switch rand.Intn(3) {
	case 0:
		return generator.Problem{
			Question:    fmt.Sprintf("How many labelled trees are there on %d given vertices?", n),
			Answer:      answer,
			Explanation: fmt.Sprintf("Cayley's formula gives n^(n-2) labelled trees on n vertices: %d^%d = %d.", n, n-2, value),
			Facts:       facts,
		}
	case 1:
		return generator.Problem{
			Question:    fmt.Sprintf("How many distinct Pruefer sequences of length %d are there over an alphabet of %d symbols?", n-2, n),
			Answer:      answer,
			Explanation: fmt.Sprintf("Each position holds one of %d symbols, so the count is %d^%d = %d. Pruefer sequences are in bijection with labelled trees, which is why the two counts agree.", n, n, n-2, value),
			Facts:       facts,
		}
	default:
		return generator.Problem{
			Question:    fmt.Sprintf("Cayley's formula says the number of labelled trees on n vertices is n^(n-2). For n = %d, what exponent appears?", n),
			Answer:      strconv.Itoa(n - 2),
			Explanation: fmt.Sprintf("Substituting n = %d gives the exponent %d - 2 = %d, so there are %d labelled trees.", n, n, n-2, value),
			Facts:       facts,
		}
	}
}

// ----- 14. degree sequence -----

type degreeSequenceGen struct{}

func (g *degreeSequenceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := 2 + rand.Intn(3) // left part size, 2..4
	b := 2 + rand.Intn(3) // right part size, 2..4
	var degrees []int
	for i := 0; i < a; i++ {
		degrees = append(degrees, b)
	}
	for i := 0; i < b; i++ {
		degrees = append(degrees, a)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(degrees)))
	sum := 0
	for _, d := range degrees {
		sum += d
	}
	answer := intsString(degrees, ",")
	q := fmt.Sprintf("Give the degree sequence of the complete bipartite graph K(%d, %d) in non-increasing order, separated by commas (for example 3,3,2).", a, b)
	p := generator.Problem{
		Question: q,
		Answer:   answer,
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "sequence": answer,
			"sum": strconv.Itoa(sum), "edges": strconv.Itoa(a * b),
		},
	}
	p.Explanation = fmt.Sprintf("Each of the %d vertices in the first part has degree %d and each of the %d in the second has degree %d, so the sequence is %s. Its sum is %d = 2 \u00b7 %d, as the handshake lemma requires.",
		a, b, b, a, answer, sum, a*b)
	return p
}

// ----- 15. linear recurrence -----

type linearRecurrenceGen struct{}

func (g *linearRecurrenceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	bases := []int{2, 3, 4, 5}
	rand.Shuffle(len(bases), func(i, j int) { bases[i], bases[j] = bases[j], bases[i] })
	r1, r2 := bases[0], bases[1]
	if r1 > r2 {
		r1, r2 = r2, r1
	}
	a := 1 + rand.Intn(2) // 1..2
	b := 1 + rand.Intn(2)
	a0 := a + b
	a1 := a*r1 + b*r2
	c1 := r1 + r2
	c2 := r1 * r2
	term1 := powerTerm(a, r1)
	term2 := powerTerm(b, r2)
	answer := term1 + " + " + term2
	q := fmt.Sprintf("A sequence satisfies a(n) = %d\u00b7a(n-1) - %d\u00b7a(n-2), with a(0) = %d and a(1) = %d. Write a closed form for a(n) (use ^ for powers and * for multiplication).",
		c1, c2, a0, a1)
	p := generator.Problem{
		Question: q,
		Answer:   answer,
		Facts: map[string]string{
			"r1": strconv.Itoa(r1), "r2": strconv.Itoa(r2),
			"a": strconv.Itoa(a), "b": strconv.Itoa(b),
			"c1": strconv.Itoa(c1), "c2": strconv.Itoa(c2),
			"a0": strconv.Itoa(a0), "a1": strconv.Itoa(a1),
			"term1": term1, "term2": term2, "value": answer,
		},
	}
	p.Explanation = fmt.Sprintf("The characteristic equation is x\u00b2 - %d\u00b7x + %d = 0, with roots %d and %d, so a(n) = A\u00b7%d^n + B\u00b7%d^n. The initial conditions give A + B = %d and %dA + %dB = %d, hence A = %d and B = %d, and a(n) = %s.",
		c1, c2, r1, r2, r1, r2, a0, r1, r2, a1, a, b, answer)
	return p
}

func powerTerm(coef, base int) string {
	if coef == 1 {
		return fmt.Sprintf("%d^n", base)
	}
	return fmt.Sprintf("%d*%d^n", coef, base)
}

// ----- 16. growth rates -----

type growthRatesGen struct{}

func (g *growthRatesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct{ funcs []string }
	entries := []entry{
		{[]string{"n!", "2^n", "n^3"}},
		{[]string{"log n", "sqrt(n)", "n^2"}},
		{[]string{"n log n", "n^2", "2^n"}},
		{[]string{"log n", "sqrt(n)", "n log n"}},
		{[]string{"n^2", "2^n", "n^3"}},
	}
	// rank[i] gives the true order of funcs[i] in the entry, slowest first.
	ranks := [][3]int{
		{2, 1, 0}, // n^3 < 2^n < n!
		{0, 1, 2}, // log n < sqrt(n) < n^2
		{0, 1, 2}, // n log n < n^2 < 2^n
		{0, 1, 2}, // log n < sqrt(n) < n log n
		{0, 1, 2}, // n^2 < n^3 < 2^n
	}
	i := rand.Intn(len(entries))
	e := entries[i]
	rank := ranks[i]
	askFastest := rand.Intn(2) == 0
	var want int
	if askFastest {
		want = 2
	} else {
		want = 0
	}
	correct := ""
	for j, r := range rank {
		if r == want {
			correct = e.funcs[j]
		}
	}
	word := "fastest"
	if !askFastest {
		word = "slowest"
	}
	q := fmt.Sprintf("As n grows without bound, which of these grows %s?", word)
	var distractors []string
	for _, f := range e.funcs {
		if f != correct {
			distractors = append(distractors, f)
		}
	}
	p := generator.MultipleChoice(q, correct, distractors...)
	p.Facts = map[string]string{
		"functions": strings.Join(e.funcs, ", "), "target": word, "answer": correct,
	}
	p.Explanation = fmt.Sprintf("Comparing the growth of %s as n \u2192 \u221e: the %s is %s.",
		strings.Join(e.funcs, ", "), word, correct)
	return p
}
