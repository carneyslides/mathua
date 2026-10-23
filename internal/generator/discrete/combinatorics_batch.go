package discrete

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerCombinatoricsBatch(reg *generator.Registry) {
	reg.Register("discrete.combinatorics.counting_principle", &countingPrincipleGen{})
	reg.Register("discrete.combinatorics.factorial", &factorialGen{})
	reg.Register("discrete.combinatorics.multinomial", &multinomialGen{})
	reg.Register("discrete.combinatorics.circular", &circularGen{})
	reg.Register("discrete.combinatorics.derangement", &derangementGen{})
	reg.Register("discrete.combinatorics.catalan", &catalanGen{})
}

// Factorial computation helper
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	res := 1
	for i := 2; i <= n; i++ {
		res *= i
	}
	return res
}

// Subfactorial (derangement) computation helper: D_n = (n - 1)(D_{n - 1} + D_{n - 2})
func derangement(n int) int {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return 0
	}
	d0 := 1
	d1 := 0
	var curr int
	for i := 2; i <= n; i++ {
		curr = (i - 1) * (d1 + d0)
		d0 = d1
		d1 = curr
	}
	return curr
}

// Catalan number computation helper: C_n = (2n)! / ((n + 1)! * n!) = nCr(2n, n) / (n + 1)
func catalan(n int) int {
	if n <= 1 {
		return 1
	}
	return nCr(2*n, n) / (n + 1)
}

// Multinomial coefficient helper: n! / (k_1! * k_2! * ... * k_m!)
func multinomialCoeff(n int, parts []int) int {
	num := factorial(n)
	den := 1
	for _, p := range parts {
		den *= factorial(p)
	}
	return num / den
}

// ----- 1. counting_principle (discrete.combinatorics.counting_principle) -----

type countingPrincipleGen struct{}

func (g *countingPrincipleGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	var prompt, expl string
	var ans int

	switch mode {
	case 0:
		// Product rule: 2 independent stages
		a := 3 + rand.Intn(3) + mathutil.ScaleInt(ctx.Difficulty, 0, 5)
		b := 2 + rand.Intn(3) + mathutil.ScaleInt(ctx.Difficulty, 0, 4)
		ans = a * b

		prompt = fmt.Sprintf("A college student selects one course from %d science electives and one course from %d mathematics electives. By the product rule of counting, how many distinct course pairings can be selected?", a, b)
		expl = fmt.Sprintf("Subgoal 1: State governing counting rule. According to the product rule of counting, if a first task can be performed in \\(|A|\\) ways and an independent second task can be performed in \\(|B|\\) ways, the sequence of two tasks can be performed in \\(|A \\times B| = |A| \\times |B|\\) ways.\nSubgoal 2: Substitute parameters. Here, there are %d science options and %d mathematics options.\nSubgoal 3: Evaluate and simplify. Multiply the independent choices: %d \\times %d = %d.", a, b, a, b, ans)

	case 1:
		// Product rule: 3 independent stages
		a := 3 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 4)
		b := 3 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 3)
		c := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 3)
		ans = a * b * c

		prompt = fmt.Sprintf("A security passcode is created by choosing one uppercase letter from %d available letters, followed by one digit from %d available digits and one symbol from %d special characters. How many distinct passcodes can be formed?", a, b, c)
		expl = fmt.Sprintf("Subgoal 1: State governing counting rule. When forming an outcome across three independent stages with \\(n_1\\), \\(n_2\\) and \\(n_3\\) options, the total number of distinct outcomes is given by the product rule: \\(N = n_1 \\times n_2 \\times n_3\\).\nSubgoal 2: Substitute parameters. The sequential stages offer %d letters, %d digits and %d symbols.\nSubgoal 3: Evaluate and simplify. Multiply the options: %d \\times %d \\times %d = %d.", a, b, c, a, b, c, ans)

	case 2:
		// Sum rule: disjoint choices
		a := 4 + rand.Intn(4) + mathutil.ScaleInt(ctx.Difficulty, 0, 6)
		b := 3 + rand.Intn(4) + mathutil.ScaleInt(ctx.Difficulty, 0, 5)
		ans = a + b

		prompt = fmt.Sprintf("A researcher chooses a research topic from either %d biology projects or %d chemistry projects, where no project belongs to both fields. By the sum rule of counting, how many total project options are available?", a, b)
		expl = fmt.Sprintf("Subgoal 1: State governing counting rule. By the sum rule of counting, if event \\(A\\) can occur in \\(|A|\\) ways and disjoint event \\(B\\) can occur in \\(|B|\\) ways with \\(A \\cap B = \\emptyset\\), then either event \\(A\\) or event \\(B\\) can occur in \\(|A \\cup B| = |A| + |B|\\) ways.\nSubgoal 2: Substitute parameters. The disjoint topic categories contain %d biology projects and %d chemistry projects.\nSubgoal 3: Evaluate and simplify. Add the disjoint counts: %d + %d = %d.", a, b, a, b, ans)

	case 3:
		// Combined sum and product rule: two distinct meal deals
		s1 := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2)
		d1 := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2)
		p2 := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2)
		s2 := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2)
		ans = s1*d1 + p2*s2

		prompt = fmt.Sprintf("A dining hall offers two combo menus. Combo 1 offers 1 sandwich from %d choices and 1 drink from %d choices. Combo 2 offers 1 pasta from %d choices and 1 dessert from %d choices. If a customer chooses exactly one combo meal, how many total meal combinations are possible?", s1, d1, p2, s2)
		expl = fmt.Sprintf("Subgoal 1: State governing counting rule. Compute the combinations within each independent combo via the product rule, then add the mutually exclusive combo totals using the sum rule: \\(N = (s_1 \\times d_1) + (p_2 \\times s_2)\\).\nSubgoal 2: Substitute parameters. Combo 1 provides %d \\times %d options and Combo 2 provides %d \\times %d options.\nSubgoal 3: Evaluate and simplify. Evaluate each product and add: (%d \\times %d) + (%d \\times %d) = %d + %d = %d.", s1, d1, p2, s2, s1, d1, p2, s2, s1*d1, p2*s2, ans)

	case 4:
		// Product rule without replacement: sequential distinct positions
		n := 5 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 4)
		ans = n * (n - 1) * (n - 2)

		prompt = fmt.Sprintf("A student council with %d members elects a president, a vice president and a secretary. If no person can hold more than one office, how many distinct officer slates can be selected?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing counting rule. By the sequential product rule without replacement, selecting 3 distinct positions from \\(n\\) candidates produces \\(n \\times (n - 1) \\times (n - 2)\\) possible slates.\nSubgoal 2: Substitute parameters. For \\(n = %d\\), there are %d choices for president, %d remaining choices for vice president and %d remaining choices for secretary.\nSubgoal 3: Evaluate and simplify. Multiply the sequential options: %d \\times %d \\times %d = %d.", n, n, n-1, n-2, n, n-1, n-2, ans)
	}

	return generator.Problem{
		Question:    prompt,
		Answer:      strconv.Itoa(ans),
		Explanation: expl,
		Facts: map[string]string{
			"answer":     strconv.Itoa(ans),
			"template":   strconv.Itoa(mode),
			"difficulty": fmt.Sprintf("%.2f", ctx.Difficulty),
		},
	}
}

// ----- 2. factorial (discrete.combinatorics.factorial) -----

type factorialGen struct{}

func (g *factorialGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	var prompt, expl string
	var ans int

	switch mode {
	case 0:
		// Direct factorial: shelf arrangement
		n := 3 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 3..6
		ans = factorial(n)

		parts := make([]string, n)
		for i := 0; i < n; i++ {
			parts[i] = strconv.Itoa(n - i)
		}
		prodStr := strings.Join(parts, " \\times ")

		prompt = fmt.Sprintf("How many distinct linear arrangements can be formed using %d distinct books on a single shelf?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The number of linear permutations of \\(n\\) distinct objects is given by the factorial \\(n! = n \\times (n - 1) \\times \\cdots \\times 1\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), write the factorial product: \\(%s\\).\nSubgoal 3: Evaluate and simplify. Multiply the terms: %s = %d.", n, prodStr, prodStr, ans)

	case 1:
		// Factorial quotient: n! / (n - k)!
		k := 2 + rand.Intn(2) // 2 or 3
		n := 4 + k + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 6..9
		denom := n - k
		ans = 1
		parts := make([]string, k)
		for i := 0; i < k; i++ {
			val := n - i
			ans *= val
			parts[i] = strconv.Itoa(val)
		}
		prodStr := strings.Join(parts, " \\times ")

		prompt = fmt.Sprintf("Evaluate the factorial quotient \\(\\frac{%d!}{%d!}\\).", n, denom)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. By the definition of factorials, expand the numerator and cancel the identical denominator factorial: \\(\\frac{n!}{m!} = n \\times (n - 1) \\times \\cdots \\times (m + 1)\\).\nSubgoal 2: Substitute parameters. For \\(\\frac{%d!}{%d!}\\), cancel \\(%d!\\) from numerator and denominator, leaving \\(%s\\).\nSubgoal 3: Evaluate and simplify. Multiply the remaining factors: %s = %d.", n, denom, denom, prodStr, prodStr, ans)

	case 2:
		// Factorial fraction: n! / (a! * b!) where a + b = n
		n := 5 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 5..8
		a := 2 + rand.Intn(n-3)
		b := n - a
		ans = nCr(n, a)

		prompt = fmt.Sprintf("Evaluate the factorial fraction \\(\\frac{%d!}{%d! \\cdot %d!}\\).", n, a, b)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Expand the larger denominator factorial from the numerator to simplify: \\(\\frac{n!}{a! \\cdot b!} = \\frac{n \\times (n - 1) \\times \\cdots \\times (a + 1)}{b!}\\).\nSubgoal 2: Substitute parameters. With \\(n = %d\\), \\(a = %d\\) and \\(b = %d\\), compute numerator %d! = %d and denominator %d! \\cdot %d! = %d.\nSubgoal 3: Evaluate and simplify. Divide the evaluated terms: %d / %d = %d.", n, a, b, n, factorial(n), a, b, factorial(a)*factorial(b), factorial(n), factorial(a)*factorial(b), ans)

	case 3:
		// Factorial algebraic simplification: (n! + (n - 1)!) / (n - 1)! = n + 1
		n := 4 + rand.Intn(3) + mathutil.ScaleInt(ctx.Difficulty, 0, 3) // 4..9
		ans = n + 1

		prompt = fmt.Sprintf("Evaluate the factorial expression \\(\\frac{%d! + %d!}{%d!}\\).", n, n-1, n-1)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Factor out \\((n - 1)!\\) from the numerator using the recurrence \\(n! = n \\times (n - 1)!\\): \\(\\frac{n! + (n - 1)!}{(n - 1)!} = \\frac{(n - 1)!(n + 1)}{(n - 1)!} = n + 1\\).\nSubgoal 2: Substitute parameters. Substitute \\(n = %d\\) into the simplified formula \\(n + 1\\): %d + 1.\nSubgoal 3: Evaluate and simplify. Add the terms: %d + 1 = %d.", n, n, n, ans)

	case 4:
		// Partial permutation: P(n, 3) = n! / (n - 3)!
		n := 5 + rand.Intn(3) + mathutil.ScaleInt(ctx.Difficulty, 0, 3) // 5..10
		ans = n * (n - 1) * (n - 2)

		prompt = fmt.Sprintf("Compute the number of permutations of %d items taken 3 at a time, defined by \\(P(%d, 3) = \\frac{%d!}{(%d - 3)!}\\).", n, n, n, n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The permutation formula for \\(r\\) ordered selections from \\(n\\) items is \\(P(n, r) = \\frac{n!}{(n - r)!} = n \\times (n - 1) \\times \\cdots \\times (n - r + 1)\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\) and \\(r = 3\\), the product is %d \\times %d \\times %d.\nSubgoal 3: Evaluate and simplify. Multiply the consecutive terms: %d \\times %d \\times %d = %d.", n, n, n-1, n-2, n, n-1, n-2, ans)
	}

	return generator.Problem{
		Question:    prompt,
		Answer:      strconv.Itoa(ans),
		Explanation: expl,
		Facts: map[string]string{
			"answer":     strconv.Itoa(ans),
			"template":   strconv.Itoa(mode),
			"difficulty": fmt.Sprintf("%.2f", ctx.Difficulty),
		},
	}
}

// ----- 3. multinomial (discrete.combinatorics.multinomial) -----

type multinomialGen struct{}

func (g *multinomialGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	var prompt, expl string
	var ans int

	switch mode {
	case 0:
		// Word anagrams with repeated characters
		type wordEntry struct {
			word   string
			counts []int
			desc   string
		}
		words := []wordEntry{
			{"LEVEL", []int{2, 2, 1}, "2 Ls, 2 Es and 1 V"},
			{"RADAR", []int{2, 2, 1}, "2 Rs, 2 As and 1 D"},
			{"BANANA", []int{3, 2, 1}, "3 As, 2 Ns and 1 B"},
			{"PEPPER", []int{3, 2, 1}, "3 Ps, 2 Es and 1 R"},
			{"CHEESE", []int{3, 1, 1, 1}, "3 Es, 1 C, 1 H and 1 S"},
			{"SUCCESS", []int{3, 2, 1, 1}, "3 Ss, 2 Cs, 1 U and 1 E"},
		}
		idx := rand.Intn(len(words))
		if ctx.Difficulty >= 0.7 {
			idx = 2 + rand.Intn(len(words)-2)
		} else if ctx.Difficulty <= 0.3 {
			idx = rand.Intn(2)
		}
		w := words[idx]
		n := len(w.word)
		ans = multinomialCoeff(n, w.counts)

		factTerms := make([]string, len(w.counts))
		for i, c := range w.counts {
			factTerms[i] = fmt.Sprintf("%d!", c)
		}
		denomStr := strings.Join(factTerms, " \\cdot ")

		prompt = fmt.Sprintf("How many distinct letter arrangements can be formed using all the letters in the word \"%s\"?", w.word)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The number of distinct permutations of a multiset of \\(n\\) items with duplicate frequencies \\(k_1, k_2, \\dots, k_m\\) is given by the multinomial coefficient \\(\\frac{n!}{k_1! \\cdot k_2! \\cdots k_m!}\\).\nSubgoal 2: Substitute parameters. The word \"%s\" has %d letters with character counts: %s. Write \\(\\frac{%d!}{%s}\\).\nSubgoal 3: Evaluate and simplify. Compute numerator %d! = %d and divide by the product of factorials to obtain: %d.", w.word, n, w.desc, n, denomStr, n, factorial(n), ans)

	case 1:
		// Colored flags in a vertical line
		r := 2 + rand.Intn(2)
		b := 2 + rand.Intn(2)
		gCount := 1 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 1)
		n := r + b + gCount
		ans = multinomialCoeff(n, []int{r, b, gCount})

		prompt = fmt.Sprintf("A signal mast displays %d flags in a vertical line: %d identical red flags, %d identical blue flags and %d identical green flags. How many distinct flag signals can be formed?", n, r, b, gCount)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The number of unique linear arrangements of \\(n\\) colored objects with repeated counts \\(n_1, n_2, n_3\\) is \\(\\frac{n!}{n_1! \\cdot n_2! \\cdot n_3!}\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\) total flags with %d red, %d blue and %d green, write \\(\\frac{%d!}{%d! \\cdot %d! \\cdot %d!}\\).\nSubgoal 3: Evaluate and simplify. Evaluate the factorials: \\(\\frac{%d}{%d} = %d\\).", n, r, b, gCount, n, r, b, gCount, factorial(n), factorial(r)*factorial(b)*factorial(gCount), ans)

	case 2:
		// Partitioning into labeled groups of sizes k1, k2, k3
		k1 := 2 + rand.Intn(2)
		k2 := 2 + rand.Intn(2)
		k3 := 2 + mathutil.ScaleInt(ctx.Difficulty, 0, 2)
		n := k1 + k2 + k3
		ans = multinomialCoeff(n, []int{k1, k2, k3})

		prompt = fmt.Sprintf("In how many ways can %d distinct students be divided into three distinct research groups of sizes %d, %d and %d respectively?", n, k1, k2, k3)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Partitioning \\(n\\) distinct items into labeled groups of sizes \\(k_1, k_2, k_3\\) (where \\(k_1 + k_2 + k_3 = n\\)) is given by \\(\\binom{n}{k_1, k_2, k_3} = \\frac{n!}{k_1! \\cdot k_2! \\cdot k_3!}\\).\nSubgoal 2: Substitute parameters. With \\(n = %d\\) and group sizes %d, %d and %d, write \\(\\frac{%d!}{%d! \\cdot %d! \\cdot %d!}\\).\nSubgoal 3: Evaluate and simplify. Compute the factorials and simplify: \\(\\frac{%d}{%d} = %d\\).", n, k1, k2, k3, n, k1, k2, k3, factorial(n), factorial(k1)*factorial(k2)*factorial(k3), ans)

	case 3:
		// Multinomial theorem coefficient: coeff of x^a y^b z^c in (x + y + z)^n
		a := 1 + rand.Intn(2)
		b := 1 + rand.Intn(2)
		c := 1 + mathutil.ScaleInt(ctx.Difficulty, 0, 2)
		n := a + b + c
		ans = multinomialCoeff(n, []int{a, b, c})

		prompt = fmt.Sprintf("Find the coefficient of \\(x^{%d} y^{%d} z^{%d}\\) in the algebraic expansion of \\((x + y + z)^{%d}\\).", a, b, c, n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. By the multinomial theorem, the coefficient of \\(x^a y^b z^c\\) in \\((x + y + z)^n\\) where \\(a + b + c = n\\) is given by \\(\\binom{n}{a, b, c} = \\frac{n!}{a! \\cdot b! \\cdot c!}\\).\nSubgoal 2: Substitute parameters. For powers \\(a = %d\\), \\(b = %d\\) and \\(c = %d\\) with total power \\(n = %d\\), write \\(\\frac{%d!}{%d! \\cdot %d! \\cdot %d!}\\).\nSubgoal 3: Evaluate and simplify. Evaluate the factorials and divide: \\(\\frac{%d}{%d} = %d\\).", a, b, c, n, n, a, b, c, factorial(n), factorial(a)*factorial(b)*factorial(c), ans)

	case 4:
		// 3D grid path walk
		x := 1 + rand.Intn(2)
		y := 1 + rand.Intn(2)
		z := 1 + mathutil.ScaleInt(ctx.Difficulty, 0, 2)
		n := x + y + z
		ans = multinomialCoeff(n, []int{x, y, z})

		prompt = fmt.Sprintf("How many grid paths from the origin \\((0, 0, 0)\\) to \\((%d, %d, %d)\\) exist in three-dimensional space if each step moves one unit in the positive \\(x\\), \\(y\\) or \\(z\\) direction?", x, y, z)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. A grid path in three dimensions requiring \\(x\\) steps in \\(X\\), \\(y\\) steps in \\(Y\\) and \\(z\\) steps in \\(Z\\) is an arrangement of a multiset of total length \\(n = x + y + z\\): \\(\\frac{(x + y + z)!}{x! \\cdot y! \\cdot z!}\\).\nSubgoal 2: Substitute parameters. With %d steps in \\(x\\), %d steps in \\(y\\) and %d steps in \\(z\\), total steps \\(n = %d\\).\nSubgoal 3: Evaluate and simplify. Calculate \\(\\frac{%d!}{%d! \\cdot %d! \\cdot %d!} = %d\\).", x, y, z, n, n, x, y, z, ans)
	}

	return generator.Problem{
		Question:    prompt,
		Answer:      strconv.Itoa(ans),
		Explanation: expl,
		Facts: map[string]string{
			"answer":     strconv.Itoa(ans),
			"template":   strconv.Itoa(mode),
			"difficulty": fmt.Sprintf("%.2f", ctx.Difficulty),
		},
	}
}

// ----- 4. circular (discrete.combinatorics.circular) -----

type circularGen struct{}

func (g *circularGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	var prompt, expl string
	var ans int

	switch mode {
	case 0:
		// Standard circular seating: (n - 1)!
		n := 4 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 4..7
		ans = factorial(n - 1)

		prompt = fmt.Sprintf("In how many distinct ways can %d people be seated around a circular table where arrangements that differ only by rotation are considered equivalent?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The number of circular permutations of \\(n\\) distinct objects where rotations are identical is \\(P_c(n) = (n - 1)!\\).\nSubgoal 2: Substitute parameters. Fixing one person to break rotational symmetry leaves \\((%d - 1)! = %d!\\) ways to arrange the remaining seated guests.\nSubgoal 3: Evaluate and simplify. Compute the factorial: %d! = %d.", n, n-1, n-1, ans)

	case 1:
		// Circular seating with 2 adjacent people: 2 * (n - 2)!
		n := 5 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 5..8
		ans = 2 * factorial(n-2)

		prompt = fmt.Sprintf("In how many ways can %d people sit around a circular table if two specific guests (Alice and Bob) must sit next to each other?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Treat the two adjacent guests as a single combined block. This leaves \\(n - 1\\) entities to seat in a circle in \\(((n - 1) - 1)! = (n - 2)!\\) ways. Within the block, the two guests can arrange themselves in \\(2! = 2\\) ways. Total: \\(2 \\times (n - 2)!\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), the formula gives \\(2 \\times (%d - 2)! = 2 \\times %d!\\).\nSubgoal 3: Evaluate and simplify. Multiply by the internal arrangements: 2 \\times %d = %d.", n, n, n-2, factorial(n-2), ans)

	case 2:
		// Circular seating with 2 people NOT adjacent: (n - 3) * (n - 2)!
		n := 5 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 5..8
		total := factorial(n - 1)
		adj := 2 * factorial(n-2)
		ans = total - adj

		prompt = fmt.Sprintf("In how many ways can %d people sit around a circular table if two specific guests must NOT sit next to each other?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Apply the complement principle: subtract the arrangements where the two guests are adjacent from the total circular arrangements: \\(N = (n - 1)! - 2(n - 2)!\\).\nSubgoal 2: Substitute parameters. Total circular arrangements equal \\((%d - 1)! = %d! = %d\\), and adjacent arrangements equal \\(2 \\times (%d - 2)! = 2 \\times %d = %d\\).\nSubgoal 3: Evaluate and simplify. Subtract to find the non-adjacent arrangements: %d - %d = %d.", n, n-1, total, n, factorial(n-2), adj, total, adj, ans)

	case 3:
		// Alternating seating: (n - 1)! * n!
		n := 3 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 1) // 3..5
		ans = factorial(n-1) * factorial(n)

		prompt = fmt.Sprintf("In how many ways can %d men and %d women be seated alternately around a circular table?", n, n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. First arrange the \\(n\\) men in a circle in \\((n - 1)!\\) ways. This defines \\(n\\) distinct intervening spaces for the \\(n\\) women, who can be arranged in \\(n!\\) ways. Total: \\((n - 1)! \\times n!\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), the product is \\((%d - 1)! \\times %d! = %d! \\times %d!\\).\nSubgoal 3: Evaluate and simplify. Multiply the factorials: %d \\times %d = %d.", n, n, n, n-1, n, factorial(n-1), factorial(n), ans)

	case 4:
		// Key ring / necklace with reflection symmetry: (n - 1)! / 2
		n := 4 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 4..7
		ans = factorial(n-1) / 2

		prompt = fmt.Sprintf("How many distinct key rings can be formed using %d distinct keys if both rotations and flipping the ring over are considered equivalent?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. When both rotational and reflectional symmetries are equivalent (as with a key ring or necklace), the number of distinct arrangements is given by \\(\\frac{(n - 1)!}{2}\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), compute \\(\\frac{(%d - 1)!}{2} = \\frac{%d!}{2}\\).\nSubgoal 3: Evaluate and simplify. Compute the factorial and divide by 2: %d / 2 = %d.", n, n, n-1, factorial(n-1), ans)
	}

	return generator.Problem{
		Question:    prompt,
		Answer:      strconv.Itoa(ans),
		Explanation: expl,
		Facts: map[string]string{
			"answer":     strconv.Itoa(ans),
			"template":   strconv.Itoa(mode),
			"difficulty": fmt.Sprintf("%.2f", ctx.Difficulty),
		},
	}
}

// ----- 5. derangement (discrete.combinatorics.derangement) -----

type derangementGen struct{}

func (g *derangementGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	var prompt, expl string
	var ans int

	switch mode {
	case 0:
		// Direct derangement / Secret Santa
		n := 3 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 3..6
		ans = derangement(n)

		prompt = fmt.Sprintf("In a gift exchange among %d participants, each person brings one gift and receives one gift. How many outcomes result in nobody receiving their own gift (a complete derangement \\(D_{%d}\\))?", n, n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. A derangement \\(D_n\\) is a permutation with no fixed points, given by the inclusion-exclusion formula \\(D_n = n! \\sum_{k=0}^n \\frac{(-1)^k}{k!}\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), evaluate the alternating subfactorial expansion.\nSubgoal 3: Evaluate and simplify. The derangement value is \\(D_{%d} = %d\\).", n, n, ans)

	case 1:
		// Exactly 1 fixed point: n * D_{n - 1}
		n := 4 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 1) // 4..6
		dSub := derangement(n - 1)
		ans = n * dSub

		prompt = fmt.Sprintf("How many permutations of \\(\\{1, 2, \\dots, %d\\}\\) leave exactly one element fixed in its original position?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. To have exactly one fixed point, choose the fixed element in \\(\\binom{n}{1} = n\\) ways and derange the remaining \\(n - 1\\) elements in \\(D_{n-1}\\) ways: \\(N = n \\times D_{n-1}\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), there are %d choices for the fixed element and \\(D_{%d} = %d\\) derangements for the rest.\nSubgoal 3: Evaluate and simplify. Multiply the choices: %d \\times %d = %d.", n, n, n-1, dSub, n, dSub, ans)

	case 2:
		// Exactly 2 fixed points: nCr(n, 2) * D_{n - 2}
		n := 4 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 1) // 4..6
		comb := nCr(n, 2)
		dSub := derangement(n - 2)
		ans = comb * dSub

		prompt = fmt.Sprintf("How many permutations of \\(\\{1, 2, \\dots, %d\\}\\) leave exactly two elements fixed in their original positions?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Choose the 2 fixed positions in \\(\\binom{n}{2}\\) ways and derange the remaining \\(n - 2\\) elements in \\(D_{n-2}\\) ways: \\(N = \\binom{n}{2} \\times D_{n-2}\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), \\(\\binom{%d}{2} = %d\\) and \\(D_{%d} = %d\\).\nSubgoal 3: Evaluate and simplify. Multiply the combinations and derangements: %d \\times %d = %d.", n, n, comb, n-2, dSub, comb, dSub, ans)

	case 3:
		// At least one fixed point: n! - D_n
		n := 3 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 3..6
		total := factorial(n)
		dVal := derangement(n)
		ans = total - dVal

		prompt = fmt.Sprintf("How many permutations of %d elements leave at least one element fixed in its original position?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. By the complement principle, the number of permutations with at least one fixed point is the total permutations minus the complete derangements: \\(N = n! - D_n\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), total permutations equal \\(%d! = %d\\) and derangements equal \\(D_{%d} = %d\\).\nSubgoal 3: Evaluate and simplify. Subtract derangements from total: %d - %d = %d.", n, n, total, n, dVal, total, dVal, ans)

	case 4:
		// Derangement recurrence step: D_n = (n - 1)(D_{n - 1} + D_{n - 2})
		n := 4 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 1) // 4..6
		dPrev2 := derangement(n - 2)
		dPrev1 := derangement(n - 1)
		ans = (n - 1) * (dPrev1 + dPrev2)

		prompt = fmt.Sprintf("Using the derangement recurrence relation \\(D_n = (n - 1)(D_{n-1} + D_{n-2})\\), given that \\(D_{%d} = %d\\) and \\(D_{%d} = %d\\), compute the value of \\(D_{%d}\\).", n-2, dPrev2, n-1, dPrev1, n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The standard linear recurrence for derangements is \\(D_n = (n - 1)(D_{n-1} + D_{n-2})\\).\nSubgoal 2: Substitute parameters. With \\(n = %d\\), substitute \\(D_{%d} = %d\\) and \\(D_{%d} = %d\\): \\((%d - 1)(%d + %d)\\).\nSubgoal 3: Evaluate and simplify. Add within parentheses and multiply: %d \\times %d = %d.", n, n-1, dPrev1, n-2, dPrev2, n, dPrev1, dPrev2, n-1, dPrev1+dPrev2, ans)
	}

	return generator.Problem{
		Question:    prompt,
		Answer:      strconv.Itoa(ans),
		Explanation: expl,
		Facts: map[string]string{
			"answer":     strconv.Itoa(ans),
			"template":   strconv.Itoa(mode),
			"difficulty": fmt.Sprintf("%.2f", ctx.Difficulty),
		},
	}
}

// ----- 6. catalan (discrete.combinatorics.catalan) -----

type catalanGen struct{}

func (g *catalanGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	var prompt, expl string
	var ans int

	switch mode {
	case 0:
		// Direct Catalan number
		n := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 2..5
		ans = catalan(n)
		comb := nCr(2*n, n)

		prompt = fmt.Sprintf("Compute the %d-th Catalan number \\(C_{%d} = \\frac{1}{%d + 1}\\binom{%d}{%d}\\).", n, n, n, 2*n, n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The \\(n\\)-th Catalan number is defined by \\(C_n = \\frac{1}{n + 1}\\binom{2n}{n} = \\frac{(2n)!}{(n + 1)! \\cdot n!}\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\), write \\(\\frac{1}{%d}\\binom{%d}{%d} = \\frac{1}{%d} \\times %d\\).\nSubgoal 3: Evaluate and simplify. Divide by \\(n + 1\\): %d / %d = %d.", n, n+1, 2*n, n, n+1, comb, comb, n+1, ans)

	case 1:
		// Valid parentheses strings
		n := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 2..5
		ans = catalan(n)
		comb := nCr(2*n, n)

		prompt = fmt.Sprintf("How many validly matched strings of %d pairs of parentheses (such as \"((()))\" or \"()()()\") can be formed?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The number of valid expressions formed from \\(n\\) pairs of balanced parentheses corresponds bijectively to Dyck words counted by the \\(n\\)-th Catalan number: \\(C_n = \\frac{1}{n + 1}\\binom{2n}{n}\\).\nSubgoal 2: Substitute parameters. For %d pairs of parentheses, compute \\(C_{%d} = \\frac{1}{%d}\\binom{%d}{%d}\\).\nSubgoal 3: Evaluate and simplify. Evaluate the binomial coefficient \\(\\binom{%d}{%d} = %d\\) and divide: %d / %d = %d.", n, n, n+1, 2*n, n, 2*n, n, comb, comb, n+1, ans)

	case 2:
		// Monotonic lattice paths not crossing diagonal
		n := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 2..5
		ans = catalan(n)
		comb := nCr(2*n, n)

		prompt = fmt.Sprintf("How many monotonic grid paths from \\((0, 0)\\) to \\((%d, %d)\\) take only unit steps \\((+1, 0)\\) and \\((0, +1)\\) without ever passing strictly above the diagonal line \\(y = x\\)?", n, n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Monotonic grid paths from \\((0, 0)\\) to \\((n, n)\\) staying on or below the diagonal line \\(y = x\\) are counted by the Catalan number \\(C_n = \\frac{1}{n + 1}\\binom{2n}{n}\\).\nSubgoal 2: Substitute parameters. For grid dimensions %d \\times %d, compute \\(C_{%d} = \\frac{1}{%d}\\binom{%d}{%d}\\).\nSubgoal 3: Evaluate and simplify. Calculate \\(\\binom{%d}{%d} = %d\\) and simplify: %d / %d = %d.", n, n, n, n+1, 2*n, n, 2*n, n, comb, comb, n+1, ans)

	case 3:
		// Full binary trees with n internal nodes
		n := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 2..5
		ans = catalan(n)
		comb := nCr(2*n, n)

		prompt = fmt.Sprintf("How many distinct rooted full binary trees have exactly %d internal vertices (nodes having exactly two child vertices)?", n)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. The number of distinct rooted binary trees with \\(n\\) internal nodes (or \\(n + 1\\) leaves) is given by the \\(n\\)-th Catalan number \\(C_n = \\frac{1}{n + 1}\\binom{2n}{n}\\).\nSubgoal 2: Substitute parameters. For \\(n = %d\\) internal vertices, write \\(C_{%d} = \\frac{1}{%d}\\binom{%d}{%d}\\).\nSubgoal 3: Evaluate and simplify. Evaluate the binomial term \\(\\binom{%d}{%d} = %d\\) and divide: %d / %d = %d.", n, n, n+1, 2*n, n, 2*n, n, comb, comb, n+1, ans)

	case 4:
		// Convex polygon triangulation
		// vertices k = n + 2, where n in 2..5 => k in 4..7
		n := 2 + rand.Intn(2) + mathutil.ScaleInt(ctx.Difficulty, 0, 2) // 2..5
		k := n + 2
		polyNames := map[int]string{
			4: "quadrilateral",
			5: "pentagon",
			6: "hexagon",
			7: "heptagon",
			8: "octagon",
		}
		ans = catalan(n)
		comb := nCr(2*n, n)

		prompt = fmt.Sprintf("In how many ways can a convex %s (with %d vertices) be triangulated into triangles by drawing non-intersecting internal diagonals?", polyNames[k], k)
		expl = fmt.Sprintf("Subgoal 1: State governing formula. Euler's polygon division theorem states that a convex polygon with \\(k\\) vertices can be partitioned into triangles by non-crossing diagonals in \\(C_{k-2}\\) ways, where \\(C_m = \\frac{1}{m + 1}\\binom{2m}{m}\\).\nSubgoal 2: Substitute parameters. For a %d-gon, the index is \\(m = %d - 2 = %d\\). Evaluate \\(C_{%d} = \\frac{1}{%d}\\binom{%d}{%d}\\).\nSubgoal 3: Evaluate and simplify. Compute \\(\\binom{%d}{%d} = %d\\) and divide: %d / %d = %d.", k, k, n, n, n+1, 2*n, n, 2*n, n, comb, comb, n+1, ans)
	}

	return generator.Problem{
		Question:    prompt,
		Answer:      strconv.Itoa(ans),
		Explanation: expl,
		Facts: map[string]string{
			"answer":     strconv.Itoa(ans),
			"template":   strconv.Itoa(mode),
			"difficulty": fmt.Sprintf("%.2f", ctx.Difficulty),
		},
	}
}
