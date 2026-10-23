package numtheory

// The 2026-10-18 number-theory batch: ten concepts that take the domain from
// 32 to the target of 42.
//
// The domain already covers divisibility, gcd, congruences, Fermat, Euler phi,
// Wilson, Legendre, quadratic reciprocity, orders, primitive roots, Mobius, the
// prime number theorem, zeta, Chebyshev, elliptic curves, continued fractions,
// class numbers and RSA. Several obvious additions were already taught and were
// not repeated: modular inverses in the Bezout lesson, Euler's criterion in the
// Legendre lesson, perfect numbers and Mersenne primes in the divisor-sum
// lesson, twin primes and Brun's sieve in the sieve lesson, and Gaussian
// integers in the class-number lesson.
//
// The batch fills real gaps instead: the Jacobi symbol, discrete logarithms,
// square roots modulo a prime, sums of two squares, Pell's equation, Pisano
// periods, the Carmichael function, the Miller-Rabin test, prime gaps and
// Bertrand's postulate.
//
// Eight concepts are numeric and two are multiple choice. Every numeric
// generator publishes the facts its answer follows from, so the tests recompute
// the answer from those facts. The wiring is scoped: only the new concepts plus
// the course's advertised RSA and Diophantine topics are targeted, so the
// introductory course does not become a catalogue of zeta functions, elliptic
// curves and class numbers.

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func computeJacobi(a, n int) int {
	if n <= 0 || n%2 == 0 {
		return 0
	}
	a %= n
	if a < 0 {
		a += n
	}
	result := 1
	for a != 0 {
		for a%2 == 0 {
			a /= 2
			if n%8 == 3 || n%8 == 5 {
				result = -result
			}
		}
		a, n = n, a
		if a%4 == 3 && n%4 == 3 {
			result = -result
		}
		a %= n
	}
	if n == 1 {
		return result
	}
	return 0
}

func computeGCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func computeLCM(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return (a / computeGCD(a, b)) * b
}

func computeCarmichael(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	m := n
	for p := 2; p*p <= m; p++ {
		if m%p != 0 {
			continue
		}
		e := 0
		pe := 1
		for m%p == 0 {
			m /= p
			e++
			pe *= p
		}
		var part int
		if p == 2 && e >= 3 {
			part = 1 << (e - 2)
		} else {
			part = (pe / p) * (p - 1)
		}
		result = computeLCM(result, part)
	}
	if m > 1 {
		result = computeLCM(result, m-1)
	}
	return result
}

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

func computeNextPrime(n int) int {
	for c := n + 1; ; c++ {
		if isPrime(c) {
			return c
		}
	}
}

func powMod(base, exp, mod int) int {
	res := 1
	base %= mod
	if base < 0 {
		base += mod
	}
	for exp > 0 {
		if exp&1 == 1 {
			res = (res * base) % mod
		}
		base = (base * base) % mod
		exp >>= 1
	}
	return res
}

func findPrimitiveRoot(p int) int {
	if p == 2 {
		return 1
	}
	phi := p - 1
	n := phi
	var factors []int
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			factors = append(factors, d)
			for n%d == 0 {
				n /= d
			}
		}
	}
	if n > 1 {
		factors = append(factors, n)
	}
	for g := 2; g < p; g++ {
		ok := true
		for _, q := range factors {
			if powMod(g, phi/q, p) == 1 {
				ok = false
				break
			}
		}
		if ok {
			return g
		}
	}
	return 2
}

func computePisano(n int) int {
	if n <= 1 {
		return 1
	}
	a, b := 0, 1
	period := 0
	for {
		a, b = b, (a+b)%n
		period++
		if a == 0 && b == 1 {
			return period
		}
	}
}

func registerBatch(reg *generator.Registry) {
	reg.Register("nt.adv.jacobi", &jacobiGen{})
	reg.Register("nt.adv.discrete_log", &discreteLogGen{})
	reg.Register("nt.adv.modular_sqrt", &modularSqrtGen{})
	reg.Register("nt.adv.two_squares", &twoSquaresGen{})
	reg.Register("nt.adv.pell", &pellGen{})
	reg.Register("nt.adv.pisano", &pisanoGen{})
	reg.Register("nt.adv.carmichael_function", &carmichaelFunctionGen{})
	reg.Register("nt.adv.miller_rabin", &millerRabinGen{})
	reg.Register("nt.adv.prime_gap", &primeGapGen{})
	reg.Register("nt.adv.bertrand", &bertrandGen{})
}

// ---------------------------------------------------------------------------
// 1. The Jacobi symbol (numeric)
// ---------------------------------------------------------------------------

type jacobiGen struct{}

func (g *jacobiGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	minN := mathutil.ScaleInt(ctx.Difficulty, 9, 35)
	maxN := mathutil.ScaleInt(ctx.Difficulty, 25, 105)
	if maxN <= minN {
		maxN = minN + 6
	}
	k := rand.Intn((maxN-minN)/2 + 1)
	n := minN + 2*k
	if n%2 == 0 {
		n++
	}

	var a int
	for attempts := 0; attempts < 100; attempts++ {
		a = rand.Intn(2*n) - n
		if a != 0 && computeGCD(mathutil.Abs(a), n) == 1 {
			break
		}
	}
	if a == 0 || computeGCD(mathutil.Abs(a), n) != 1 {
		a = 2
		if computeGCD(a, n) != 1 {
			a = -1
		}
	}

	ans := computeJacobi(a, n)
	ansStr := strconv.Itoa(ans)
	question := fmt.Sprintf("Compute the Jacobi symbol (a/n) for a = %d and n = %d. (enter 1 or -1)", a, n)
	explanation := fmt.Sprintf("Subgoal 1: State governing formula. The Jacobi symbol (a/n) extends the Legendre symbol multiplicatively across prime factors of odd modulus n.\nSubgoal 2: Substitute parameters. Substitute a = %d and n = %d into quadratic reciprocity and reduction identities.\nSubgoal 3: Evaluate and simplify. Successive reduction yields (a/n) = %s.", a, n, ansStr)

	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts:       map[string]string{"a": strconv.Itoa(a), "n": strconv.Itoa(n), "answer": ansStr},
	}
}

// ---------------------------------------------------------------------------
// 2. Discrete logarithms (numeric)
// ---------------------------------------------------------------------------

type discreteLogGen struct{}

var dlPrimesLow = []int{7, 11, 13, 17, 19, 23}
var dlPrimesMid = []int{29, 31, 37, 41}
var dlPrimesHigh = []int{43, 47, 53, 59}

func (g *discreteLogGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var primePool []int
	switch tier {
	case 0:
		primePool = dlPrimesLow
	case 1:
		primePool = dlPrimesMid
	default:
		primePool = dlPrimesHigh
	}
	p := primePool[rand.Intn(len(primePool))]

	baseRoot := findPrimitiveRoot(p)
	var roots []int
	for k := 1; k < p-1; k++ {
		if computeGCD(k, p-1) == 1 {
			roots = append(roots, powMod(baseRoot, k, p))
		}
	}
	if len(roots) == 0 {
		roots = []int{baseRoot}
	}
	gElem := roots[rand.Intn(len(roots))]

	x := 1 + rand.Intn(p-1)
	h := powMod(gElem, x, p)

	ansStr := strconv.Itoa(x)
	question := fmt.Sprintf("Modulo %d, the element %d is a primitive root. Find x with %d^x = %d (mod %d), where 1 <= x <= %d. (enter a number)", p, gElem, gElem, h, p, p-1)
	explanation := fmt.Sprintf("Subgoal 1: State governing formula. The discrete logarithm x satisfies g^x = h (mod p) for primitive root g modulo prime p.\nSubgoal 2: Substitute parameters. Set generator base %d, target value %d and prime modulus %d.\nSubgoal 3: Evaluate and simplify. Successive modular exponentiation gives %d^%d = %d (mod %d), so x = %s.", gElem, h, p, gElem, x, h, p, ansStr)

	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts: map[string]string{
			"p": strconv.Itoa(p), "g": strconv.Itoa(gElem), "h": strconv.Itoa(h), "answer": ansStr,
		},
	}
}

// ---------------------------------------------------------------------------
// 3. Square roots modulo a prime (numeric)
// ---------------------------------------------------------------------------

type modularSqrtGen struct{}

var modSqrtPrimesLow = []int{7, 11, 13, 17, 19}
var modSqrtPrimesMid = []int{23, 29, 31, 37, 41}
var modSqrtPrimesHigh = []int{43, 47, 53, 59, 61, 67, 71}

func (g *modularSqrtGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var primePool []int
	switch tier {
	case 0:
		primePool = modSqrtPrimesLow
	case 1:
		primePool = modSqrtPrimesMid
	default:
		primePool = modSqrtPrimesHigh
	}
	p := primePool[rand.Intn(len(primePool))]

	maxR := (p - 1) / 2
	r := 1 + rand.Intn(maxR)
	a := (r * r) % p

	ansStr := strconv.Itoa(r)
	question := fmt.Sprintf("Solve x^2 = %d (mod %d) for the smaller root x in 1..%d. (enter a number)", a, p, p-1)
	explanation := fmt.Sprintf("Subgoal 1: State governing formula. The quadratic congruence x^2 = a (mod p) has two roots r and p - r in 1..(p-1).\nSubgoal 2: Substitute parameters. Substitute residue a = %d and prime modulus p = %d.\nSubgoal 3: Evaluate and simplify. The square roots are %d and %d, with smaller root x = %s.", a, p, r, p-r, ansStr)

	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts:       map[string]string{"a": strconv.Itoa(a), "p": strconv.Itoa(p), "answer": ansStr},
	}
}

// ---------------------------------------------------------------------------
// 4. Sums of two squares (numeric)
// ---------------------------------------------------------------------------

type twoSquaresGen struct{}

func (g *twoSquaresGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	minN := mathutil.ScaleInt(ctx.Difficulty, 3, 25)
	maxN := mathutil.ScaleInt(ctx.Difficulty, 15, 80)
	if maxN <= minN {
		maxN = minN + 10
	}
	n := minN + rand.Intn(maxN-minN+1)

	count := 0
	for x := -n; x <= n; x++ {
		if x*x > n {
			continue
		}
		for y := -n; y <= n; y++ {
			if x*x+y*y == n {
				count++
			}
		}
	}

	ansStr := strconv.Itoa(count)
	question := fmt.Sprintf("How many integer pairs (x, y), counting signs and order, satisfy x^2 + y^2 = %d? (enter a number)", n)
	explanation := fmt.Sprintf("Subgoal 1: State governing formula. By Jacobi's two-square theorem, r_2(n) = 4(d_1(n) - d_3(n)) counts ordered signed coordinate pairs (x, y) with x^2 + y^2 = n.\nSubgoal 2: Substitute parameters. Substitute n = %d and test integer coordinate lattice points.\nSubgoal 3: Evaluate and simplify. Enumerating valid signed pairs yields %s solutions.", n, ansStr)

	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts:       map[string]string{"n": strconv.Itoa(n), "answer": ansStr},
	}
}

// ---------------------------------------------------------------------------
// 5. Pell's equation (numeric)
// ---------------------------------------------------------------------------

type pellGen struct{}

type pellCase struct {
	d int
	x int
	y int
}

var pellCasesLow = []pellCase{
	{d: 2, x: 3, y: 2},
	{d: 3, x: 2, y: 1},
	{d: 5, x: 9, y: 4},
	{d: 6, x: 5, y: 2},
	{d: 7, x: 8, y: 3},
	{d: 8, x: 3, y: 1},
	{d: 11, x: 10, y: 3},
	{d: 12, x: 7, y: 2},
	{d: 15, x: 4, y: 1},
	{d: 20, x: 9, y: 2},
	{d: 24, x: 5, y: 1},
	{d: 35, x: 6, y: 1},
	{d: 48, x: 7, y: 1},
}

var pellCasesMid = []pellCase{
	{d: 10, x: 19, y: 6},
	{d: 14, x: 15, y: 4},
	{d: 17, x: 33, y: 8},
	{d: 18, x: 17, y: 4},
	{d: 21, x: 55, y: 12},
	{d: 23, x: 24, y: 5},
	{d: 26, x: 51, y: 10},
	{d: 27, x: 26, y: 5},
	{d: 30, x: 11, y: 2},
	{d: 32, x: 17, y: 3},
	{d: 33, x: 23, y: 4},
	{d: 34, x: 35, y: 6},
	{d: 37, x: 73, y: 12},
	{d: 38, x: 37, y: 6},
	{d: 39, x: 25, y: 4},
	{d: 40, x: 19, y: 3},
	{d: 42, x: 13, y: 2},
	{d: 50, x: 99, y: 14},
	{d: 56, x: 15, y: 2},
	{d: 63, x: 8, y: 1},
}

var pellCasesHigh = []pellCase{
	{d: 13, x: 649, y: 180},
	{d: 19, x: 170, y: 39},
	{d: 22, x: 197, y: 42},
	{d: 28, x: 127, y: 24},
	{d: 31, x: 1520, y: 273},
	{d: 41, x: 2049, y: 320},
	{d: 43, x: 3482, y: 531},
	{d: 44, x: 199, y: 30},
	{d: 45, x: 161, y: 24},
	{d: 51, x: 50, y: 7},
	{d: 52, x: 649, y: 90},
	{d: 54, x: 485, y: 66},
	{d: 55, x: 89, y: 12},
	{d: 57, x: 151, y: 20},
	{d: 60, x: 31, y: 4},
	{d: 62, x: 63, y: 8},
	{d: 65, x: 129, y: 16},
	{d: 66, x: 65, y: 8},
	{d: 68, x: 33, y: 4},
	{d: 69, x: 7775, y: 936},
	{d: 70, x: 251, y: 30},
}

func (g *pellGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var pool []pellCase
	switch tier {
	case 0:
		pool = pellCasesLow
	case 1:
		pool = pellCasesMid
	default:
		pool = pellCasesHigh
	}
	c := pool[rand.Intn(len(pool))]

	ansStr := strconv.Itoa(c.x)
	question := fmt.Sprintf("Find the smallest positive integer x for which Pell's equation x^2 - %d*y^2 = 1 has a positive integer solution y. (enter a number)", c.d)
	explanation := fmt.Sprintf("Subgoal 1: State governing formula. Pell's equation x^2 - d*y^2 = 1 has fundamental solution (x_1, y_1) obtained from continued fraction convergents of sqrt(d).\nSubgoal 2: Substitute parameters. Substitute d = %d to evaluate x^2 - %d*y^2 = 1.\nSubgoal 3: Evaluate and simplify. Minimal positive integer solution is x = %s (with y = %d).", c.d, c.d, ansStr, c.y)

	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts:       map[string]string{"d": strconv.Itoa(c.d), "answer": ansStr},
	}
}

// ---------------------------------------------------------------------------
// 6. Pisano periods (numeric)
// ---------------------------------------------------------------------------

type pisanoGen struct{}

func (g *pisanoGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	minN := mathutil.ScaleInt(ctx.Difficulty, 2, 7)
	maxN := mathutil.ScaleInt(ctx.Difficulty, 6, 20)
	if maxN <= minN {
		maxN = minN + 2
	}
	n := minN + rand.Intn(maxN-minN+1)

	period := computePisano(n)
	ansStr := strconv.Itoa(period)

	question := fmt.Sprintf("The Fibonacci sequence starts F0 = 0, F1 = 1 and satisfies F(n) = F(n-1) + F(n-2). What is its Pisano period modulo %d, the smallest k > 0 with F(k) = 0 and F(k+1) = 1 mod %d? (enter a number)", n, n)
	explanation := fmt.Sprintf("Subgoal 1: State governing formula. The Pisano period pi(n) is the length of the repetition cycle of Fibonacci numbers modulo n.\nSubgoal 2: Substitute parameters. Iterate consecutive Fibonacci pairs modulo %d starting from (0, 1).\nSubgoal 3: Evaluate and simplify. The initial pair (0, 1) first repeats after %s steps, giving pi(%d) = %s.", n, ansStr, n, ansStr)

	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts:       map[string]string{"n": strconv.Itoa(n), "answer": ansStr},
	}
}

// ---------------------------------------------------------------------------
// 7. The Carmichael function (numeric)
// ---------------------------------------------------------------------------

type carmichaelFunctionGen struct{}

func (g *carmichaelFunctionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	minN := mathutil.ScaleInt(ctx.Difficulty, 8, 35)
	maxN := mathutil.ScaleInt(ctx.Difficulty, 30, 150)
	if maxN <= minN {
		maxN = minN + 5
	}
	n := minN + rand.Intn(maxN-minN+1)

	ans := computeCarmichael(n)
	ansStr := strconv.Itoa(ans)

	question := fmt.Sprintf("What is the Carmichael function lambda(%d), the exponent of the multiplicative group modulo %d? (enter a number)", n, n)
	explanation := fmt.Sprintf("Subgoal 1: State governing formula. The Carmichael function lambda(n) is the lcm of prime power components lambda(p^k), where lambda(2^k) = 2^(k-2) for k >= 3 and lambda(p^k) = phi(p^k) for odd primes.\nSubgoal 2: Substitute parameters. Factor n = %d into prime powers and compute the exponent of each component.\nSubgoal 3: Evaluate and simplify. Taking the least common multiple across all components gives lambda(%d) = %s.", n, n, ansStr)

	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts:       map[string]string{"n": strconv.Itoa(n), "answer": ansStr},
	}
}

// ---------------------------------------------------------------------------
// 8. The Miller-Rabin test
// ---------------------------------------------------------------------------

type millerRabinGen struct{}

func (g *millerRabinGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "The Miller-Rabin test is based on which congruence?",
			correct:     "a^(n-1) = 1 mod n together with the square-root-of-1 structure",
			distractors: []string{"n divides (n-1)! + 1", "a^2 = 1 mod n for every a", "n divides 2^n - 2"},
			reason:      "The test strengthens Fermat's congruence by examining the repeated square roots of 1 that appear on the way to a^(n-1).",
		},
		{
			question:    "A strong probable prime to base a...",
			correct:     "passes the Miller-Rabin test for that base",
			distractors: []string{"is necessarily prime", "fails Fermat's test", "is a Carmichael number"},
			reason:      "Passing one base is evidence of primality, not proof: a strong pseudoprime can pass a chosen base.",
		},
		{
			question:    "The Miller-Rabin error probability is reduced by...",
			correct:     "testing more independent bases",
			distractors: []string{"using a larger modulus only", "repeating the same base", "checking divisibility by 2"},
			reason:      "Each independent base cuts the false-positive probability, and a few dozen bases make the error astronomically small.",
		},
		{
			question:    "A composite that passes Miller-Rabin for a base is called...",
			correct:     "a strong pseudoprime to that base",
			distractors: []string{"a Carmichael prime", "a Mersenne prime", "a Fermat prime"},
			reason:      "Strong pseudoprimes are the composites that survive the strong test for a particular base.",
		},
		{
			question:    "Why does Miller-Rabin beat trial division for large n?",
			correct:     "it runs in polynomial time in log n",
			distractors: []string{"it factors n completely", "it proves primality with certainty", "it needs no modular arithmetic"},
			reason:      "Trial division is exponential in the number of digits; Miller-Rabin uses fast modular exponentiation and is polynomial in log n.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 9. Prime gaps and the next prime (numeric)
// ---------------------------------------------------------------------------

type primeGapGen struct{}

var smallPrimes = []int{
	2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47,
	53, 59, 61, 67, 71, 73, 79, 83, 89, 97, 101, 103, 107, 109,
	113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179,
	181, 191, 193, 197, 199, 211, 223, 227, 229, 233, 239, 241,
}

func (g *primeGapGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	kind := "gap"
	if rand.Intn(2) == 1 {
		kind = "next"
	}

	var pVal int
	var ans int
	var question string
	var explanation string

	if kind == "gap" {
		var primePool []int
		switch tier {
		case 0:
			primePool = smallPrimes[:15]
		case 1:
			primePool = smallPrimes[15:35]
		default:
			primePool = smallPrimes[25:]
		}
		pVal = primePool[rand.Intn(len(primePool))]
		next := computeNextPrime(pVal)
		ans = next - pVal
		ansStr := strconv.Itoa(ans)

		question = fmt.Sprintf("What is the gap between %d and the next larger prime? (enter a number)", pVal)
		explanation = fmt.Sprintf("Subgoal 1: State governing formula. A prime gap g = p_{n+1} - p_n measures the difference between consecutive primes.\nSubgoal 2: Substitute parameters. Identify the prime immediately following %d.\nSubgoal 3: Evaluate and simplify. The next prime after %d is %d, giving gap %d - %d = %s.", pVal, pVal, next, next, pVal, ansStr)
	} else {
		minVal := mathutil.ScaleInt(ctx.Difficulty, 10, 80)
		maxVal := mathutil.ScaleInt(ctx.Difficulty, 50, 250)
		if maxVal <= minVal {
			maxVal = minVal + 10
		}
		pVal = minVal + rand.Intn(maxVal-minVal+1)
		next := computeNextPrime(pVal)
		ans = next
		ansStr := strconv.Itoa(ans)

		question = fmt.Sprintf("What is the smallest prime greater than %d? (enter a number)", pVal)
		explanation = fmt.Sprintf("Subgoal 1: State governing formula. Find the minimal integer p > n that has no positive divisors other than 1 and itself.\nSubgoal 2: Substitute parameters. Test sequential integers strictly greater than %d for primality.\nSubgoal 3: Evaluate and simplify. The first prime reached above %d is %s.", pVal, pVal, ansStr)
	}

	ansStr := strconv.Itoa(ans)
	return generator.Problem{
		Question:    question,
		Answer:      ansStr,
		Explanation: explanation,
		Facts:       map[string]string{"kind": kind, "p": strconv.Itoa(pVal), "answer": ansStr},
	}
}

// ---------------------------------------------------------------------------
// 10. Bertrand's postulate
// ---------------------------------------------------------------------------

type bertrandGen struct{}

func (g *bertrandGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Bertrand's postulate says that for every integer n > 1 there is...",
			correct:     "a prime p with n < p < 2n",
			distractors: []string{"a prime p with 2n < p < 3n", "a perfect square between n and 2n", "a twin prime pair below n"},
			reason:      "The postulate guarantees at least one prime in every interval (n, 2n) with n > 1.",
		},
		{
			question:    "Which statement is equivalent to Bertrand's postulate?",
			correct:     "the prime counting function satisfies pi(2n) > pi(n) for n > 1",
			distractors: []string{"pi(n) grows like n/log n", "there are infinitely many twin primes", "every even number is a sum of two primes"},
			reason:      "A prime strictly between n and 2n is exactly what makes the count at 2n exceed the count at n.",
		},
		{
			question:    "Bertrand's postulate guarantees a prime between...",
			correct:     "100 and 200",
			distractors: []string{"200 and 300", "1000 and 1001", "1 and 2"},
			reason:      "Apply the postulate with n = 100: some prime lies strictly between 100 and 200.",
		},
		{
			question:    "Bertrand's postulate is weaker than which result?",
			correct:     "the prime number theorem",
			distractors: []string{"the fundamental theorem of arithmetic", "Euler's theorem", "the sieve of Eratosthenes"},
			reason:      "The prime number theorem gives the asymptotic density of the primes, from which the existence of a prime in (n, 2n) follows for large n.",
		},
		{
			question:    "Who first proved Bertrand's postulate?",
			correct:     "Chebyshev",
			distractors: []string{"Euler", "Fermat", "Gauss"},
			reason:      "Chebyshev proved the postulate in 1850; Erdős later gave an elementary proof.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}
