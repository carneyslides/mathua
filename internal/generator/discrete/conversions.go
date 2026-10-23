package discrete

// Three multiple-choice discrete math concepts converted into computed numeric generators:
// stars and bars (discrete.combinatorics.stars_bars), the pigeonhole principle
// (discrete.combinatorics.pigeonhole), and the inclusion-exclusion principle
// (discrete.sets.inclusion_exclusion).
//
// These three concepts previously presented static option selections or yes/no questions.
// Each concept now algorithmically calculates exact combinatorial and set-theoretic answers
// from randomized parameters, publishes verified Facts, and satisfies the pool ratchet.

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
)

func nCr(n, k int) int {
	if k < 0 || k > n {
		return 0
	}
	if k == 0 || k == n {
		return 1
	}
	if k > n/2 {
		k = n - k
	}
	res := 1
	for i := 1; i <= k; i++ {
		res = res * (n - i + 1) / i
	}
	return res
}

// ----- stars and bars (discrete.combinatorics.stars_bars) -----

type starsBarsConvertedGen struct{}

func (g *starsBarsConvertedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)

	k := 2 + rand.Intn(3) // 2..4 bins/variables
	if ctx.Difficulty >= 0.7 {
		k = 3 + rand.Intn(3) // 3..5 bins/variables
	}

	var n, ans int
	var q, exp string
	posOnly := false

	switch mode {
	case 0:
		n = 3 + rand.Intn(int(3+ctx.Difficulty*6))
		ans = nCr(n+k-1, k-1)
		parts := make([]string, k)
		for i := 0; i < k; i++ {
			parts[i] = fmt.Sprintf("x_%d", i+1)
		}
		varSum := strings.Join(parts, " + ")
		q = fmt.Sprintf("How many non-negative integer solutions are there to \\(%s = %d\\)? (where each \\(x_i \\ge 0\\))", varSum, n)
		exp = fmt.Sprintf("Using stars and bars for non-negative integers, the number of solutions is \\(\\binom{n + k - 1}{k - 1} = \\binom{%d + %d - 1}{%d - 1} = \\binom{%d}{%d} = %d\\).",
			n, k, k, n+k-1, k-1, ans)

	case 1:
		posOnly = true
		n = k + 1 + rand.Intn(int(2+ctx.Difficulty*5))
		ans = nCr(n-1, k-1)
		parts := make([]string, k)
		for i := 0; i < k; i++ {
			parts[i] = fmt.Sprintf("x_%d", i+1)
		}
		varSum := strings.Join(parts, " + ")
		q = fmt.Sprintf("How many positive integer solutions are there to \\(%s = %d\\)? (where each \\(x_i \\ge 1\\))", varSum, n)
		exp = fmt.Sprintf("Using stars and bars for positive integers (substituting \\(y_i = x_i - 1 \\ge 0\\)), the number of solutions is \\(\\binom{n - 1}{k - 1} = \\binom{%d - 1}{%d - 1} = \\binom{%d}{%d} = %d\\).",
			n, k, n-1, k-1, ans)

	case 2:
		n = 3 + rand.Intn(int(3+ctx.Difficulty*6))
		ans = nCr(n+k-1, k-1)
		q = fmt.Sprintf("How many ways can %d identical balls be distributed into %d distinct boxes (boxes may be empty)?", n, k)
		exp = fmt.Sprintf("Distributing %d identical objects into %d distinct bins with empty bins allowed corresponds to stars and bars with \\(n = %d\\) stars and \\(k - 1 = %d\\) bars: \\(\\binom{%d}{%d} = %d\\).",
			n, k, n, k-1, n+k-1, k-1, ans)

	case 3:
		posOnly = true
		n = k + 1 + rand.Intn(int(2+ctx.Difficulty*5))
		ans = nCr(n-1, k-1)
		q = fmt.Sprintf("How many ways can %d identical books be distributed among %d students so that every student receives at least one book?", n, k)
		exp = fmt.Sprintf("Giving 1 book to each of the %d students leaves %d books to distribute freely among %d students. By stars and bars, this is \\(\\binom{n - 1}{k - 1} = \\binom{%d - 1}{%d - 1} = \\binom{%d}{%d} = %d\\).",
			k, n-k, k, n, k, n-1, k-1, ans)
	}

	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(ans),
		Explanation: exp,
		Facts: map[string]string{
			"n":             strconv.Itoa(n),
			"k":             strconv.Itoa(k),
			"positive_only": strconv.FormatBool(posOnly),
			"answer":        strconv.Itoa(ans),
		},
	}
}

// ----- pigeonhole principle (discrete.combinatorics.pigeonhole) -----

type pigeonholeConvertedGen struct{}

func (g *pigeonholeConvertedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(3)

	var q, exp, ansStr string
	var facts map[string]string

	switch mode {
	case 0:
		k := 2 + rand.Intn(int(3+ctx.Difficulty*4))
		m := 2 + rand.Intn(int(2+ctx.Difficulty*3))
		ans := k*(m-1) + 1
		items := []string{"socks", "marbles", "cards", "tokens"}[rand.Intn(4)]
		q = fmt.Sprintf("A drawer contains %s of %d different colors. What is the minimum number of %s you must draw in the dark to guarantee at least %d of the same color?", items, k, items, m)
		exp = fmt.Sprintf("By the generalized pigeonhole principle, the worst-case scenario draws \\(%d \\times (%d - 1) = %d\\) %s without any color reaching %d. Drawing one more forces a color to have %d %s: \\(%d \\times (%d - 1) + 1 = %d\\).",
			k, m, k*(m-1), items, m, m, items, k, m, ans)
		ansStr = strconv.Itoa(ans)
		facts = map[string]string{
			"holes":  strconv.Itoa(k),
			"target": strconv.Itoa(m),
			"answer": ansStr,
		}

	case 1:
		k := 2 + rand.Intn(int(3+ctx.Difficulty*4))
		mult := 2 + rand.Intn(int(2+ctx.Difficulty*3))
		rem := 1 + rand.Intn(k-1)
		n := mult*k + rem
		ans := (n + k - 1) / k
		q = fmt.Sprintf("If %d pigeons are placed into %d pigeonholes, what is the minimum number of pigeons that the most occupied pigeonhole must contain?", n, k)
		exp = fmt.Sprintf("By the pigeonhole principle, if \\(n = %d\\) items are placed into \\(k = %d\\) boxes, at least one box must contain at least \\(\\lceil n / k \\rceil = \\lceil %d / %d \\rceil = %d\\) items.",
			n, k, n, k, ans)
		ansStr = strconv.Itoa(ans)
		facts = map[string]string{
			"items":  strconv.Itoa(n),
			"holes":  strconv.Itoa(k),
			"answer": ansStr,
		}

	case 2:
		mod := 3 + rand.Intn(int(4+ctx.Difficulty*8))
		ans := mod + 1
		q = fmt.Sprintf("How many integers must be chosen to guarantee that at least two of them have the same remainder when divided by %d?", mod)
		exp = fmt.Sprintf("There are %d possible remainders modulo %d (namely \\(0, 1, \\dots, %d\\)). By the pigeonhole principle, choosing \\(%d + 1 = %d\\) integers guarantees that at least two share the same remainder.",
			mod, mod, mod-1, mod, ans)
		ansStr = strconv.Itoa(ans)
		facts = map[string]string{
			"modulus": strconv.Itoa(mod),
			"answer":  ansStr,
		}
	}

	return generator.Problem{
		Question:    q,
		Answer:      ansStr,
		Explanation: exp,
		Facts:       facts,
	}
}

// ----- inclusion-exclusion principle (discrete.sets.inclusion_exclusion) -----

type inclusionExclusionConvertedGen struct{}

func (g *inclusionExclusionConvertedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)

	var q, exp, ansStr string
	var facts map[string]string

	switch mode {
	case 0:
		both := 2 + rand.Intn(int(3+ctx.Difficulty*5))
		onlyA := 3 + rand.Intn(int(4+ctx.Difficulty*8))
		onlyB := 3 + rand.Intn(int(4+ctx.Difficulty*8))
		a := onlyA + both
		b := onlyB + both
		union := a + b - both
		q = fmt.Sprintf("If \\(|A| = %d\\), \\(|B| = %d\\), and \\(|A \\cap B| = %d\\), find \\(|A \\cup B|\\).", a, b, both)
		exp = fmt.Sprintf("By the principle of inclusion-exclusion for two sets, \\(|A \\cup B| = |A| + |B| - |A \\cap B| = %d + %d - %d = %d\\).", a, b, both, union)
		ansStr = strconv.Itoa(union)
		facts = map[string]string{"setA": strconv.Itoa(a), "setB": strconv.Itoa(b), "intersection": strconv.Itoa(both), "answer": ansStr}

	case 1:
		both := 2 + rand.Intn(int(3+ctx.Difficulty*5))
		onlyA := 3 + rand.Intn(int(4+ctx.Difficulty*8))
		onlyB := 3 + rand.Intn(int(4+ctx.Difficulty*8))
		a := onlyA + both
		b := onlyB + both
		union := a + b - both
		q = fmt.Sprintf("If \\(|A| = %d\\), \\(|B| = %d\\), and \\(|A \\cup B| = %d\\), find \\(|A \\cap B|\\).", a, b, union)
		exp = fmt.Sprintf("Rearranging the inclusion-exclusion formula \\(|A \\cup B| = |A| + |B| - |A \\cap B|\\) yields \\(|A \\cap B| = |A| + |B| - |A \\cup B| = %d + %d - %d = %d\\).", a, b, union, both)
		ansStr = strconv.Itoa(both)
		facts = map[string]string{"setA": strconv.Itoa(a), "setB": strconv.Itoa(b), "union": strconv.Itoa(union), "answer": ansStr}

	case 2:
		both := 2 + rand.Intn(int(3+ctx.Difficulty*5))
		onlyA := 3 + rand.Intn(int(4+ctx.Difficulty*8))
		a := onlyA + both
		q = fmt.Sprintf("If \\(|A| = %d\\) and \\(|A \\cap B| = %d\\), how many elements are in \\(A\\) but not in \\(B\\) (that is, \\(|A \\setminus B|\\))?", a, both)
		exp = fmt.Sprintf("The set difference \\(|A \\setminus B| = |A| - |A \\cap B| = %d - %d = %d\\).", a, both, onlyA)
		ansStr = strconv.Itoa(onlyA)
		facts = map[string]string{"setA": strconv.Itoa(a), "intersection": strconv.Itoa(both), "answer": ansStr}

	case 3:
		type pair struct{ p, q int }
		pairs := []pair{{2, 3}, {2, 5}, {3, 5}, {3, 7}, {2, 7}}
		chosen := pairs[rand.Intn(len(pairs))]
		p, qVal := chosen.p, chosen.q
		pq := p * qVal
		N := 20 + rand.Intn(int(20+ctx.Difficulty*40))
		multP := N / p
		multQ := N / qVal
		multBoth := N / pq
		ans := N - multP - multQ + multBoth
		q = fmt.Sprintf("How many positive integers less than or equal to %d are divisible by neither %d nor %d?", N, p, qVal)
		exp = fmt.Sprintf("Let \\(A\\) be integers \\(\\le %d\\) divisible by %d (\\(|A| = \\lfloor %d/%d \\rfloor = %d\\)), and \\(B\\) be integers divisible by %d (\\(|B| = \\lfloor %d/%d \\rfloor = %d\\)). Their intersection \\(A \\cap B\\) is divisible by %d (\\(|A \\cap B| = \\lfloor %d/%d \\rfloor = %d\\)). By inclusion-exclusion, the count divisible by neither is \\(%d - (%d + %d - %d) = %d\\).",
			N, p, N, p, multP, qVal, N, qVal, multQ, pq, N, pq, multBoth, N, multP, multQ, multBoth, ans)
		ansStr = strconv.Itoa(ans)
		facts = map[string]string{"N": strconv.Itoa(N), "p": strconv.Itoa(p), "q": strconv.Itoa(qVal), "answer": ansStr}
	}

	return generator.Problem{
		Question:    q,
		Answer:      ansStr,
		Explanation: exp,
		Facts:       facts,
	}
}
