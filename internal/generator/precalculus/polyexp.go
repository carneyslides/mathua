package precalculus

// Polynomials, rational functions and the exponential family.
//
// The domain had conics, parametric and polar curves, sequences, vectors and a couple of special
// functions, but nothing on what a polynomial's zeros are or how they can be found, nothing on a
// rational function's asymptotes, and no exponential or logarithmic concept at all -- which is the
// bridge into calculus that the domain is named for. These are those three blocks.
//
// Every answer is an integer by construction: the polynomial evaluations are integers, the count of
// rational-root candidates is a product of divisor counts, the exponential and logarithmic equations
// use exact powers, and the circle's radius comes from a Pythagorean triple.
import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"unicode"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func containsWord(s, word string) bool {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.'
	})
	for _, f := range fields {
		if f == word {
			return true
		}
	}
	return false
}

// chooseSmall is the binomial coefficient, computed multiplicatively so it stays exact.
func chooseSmall(n, r int) int {
	if r < 0 || r > n {
		return 0
	}
	if r > n-r {
		r = n - r
	}
	c := 1
	for i := 0; i < r; i++ {
		c = c * (n - i) / (i + 1)
	}
	return c
}

// divisors counts the positive divisors of n, which is what the rational-root theorem multiplies.
func divisors(n int) int {
	if n < 0 {
		n = -n
	}
	if n == 0 {
		return 0
	}
	count := 0
	for i := 1; i <= n; i++ {
		if n%i == 0 {
			count++
		}
	}
	return count
}

func smallInt(lo, hi int) int {
	if hi < lo {
		hi = lo
	}
	return lo + rand.Intn(hi-lo+1)
}

// remainderTheoremGen evaluates a cubic at a point, which is what the remainder theorem asks for.
type remainderTheoremGen struct{}

func (g *remainderTheoremGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, c, r, v int
	var q, ex string
	for {
		rangeCoeff := 3 + tier*4
		a = 0
		for a == 0 {
			a = rand.Intn(2*rangeCoeff+1) - rangeCoeff
		}
		b = 0
		for b == 0 {
			b = rand.Intn(2*rangeCoeff+1) - rangeCoeff
		}
		c = 0
		for c == 0 {
			c = rand.Intn(2*rangeCoeff+1) - rangeCoeff
		}
		rangeR := 2 + tier
		r = 0
		for r == 0 {
			r = rand.Intn(2*rangeR+1) - rangeR
		}
		v = r*r*r + a*r*r + b*r + c
		ansStr := strconv.Itoa(v)
		q = fmt.Sprintf("What is the remainder when \\(p(x) = x^3 %+dx^2 %+dx %+d\\) is divided by \\(x - %d\\)?", a, b, c, r)
		if v == a || v == -a || v == b || v == -b || v == c || v == -c || v == r || v == -r || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("By the Polynomial Remainder Theorem, the remainder when dividing a polynomial \\(p(x)\\) by a linear factor \\(x - r\\) is equal to \\(p(r)\\). Set \\(x = %d\\) and substitute into \\(p(x)\\): \\(p(%d) = (%d)^3 + (%d)(%d)^2 + (%d)(%d) + (%d)\\). Compute each power and sum terms: \\(%d + (%d) + (%d) + (%d) = %d\\). The remainder is %d.",
			r, r, r, a, r, b, r, c, r*r*r, a*r*r, b*r, c, v, v)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(v),
		Explanation: ex,
		Facts: map[string]string{
			"a":         strconv.Itoa(a),
			"b":         strconv.Itoa(b),
			"c":         strconv.Itoa(c),
			"r":         strconv.Itoa(r),
			"remainder": strconv.Itoa(v),
			"answer":    strconv.Itoa(v),
		},
	}
}

// rationalRootsGen asks how many candidates the rational-root theorem produces, which is a product
// of two divisor counts rather than a list.
type rationalRootsGen struct{}

func (g *rationalRootsGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	leadMax := mathutil.ScaleInt(ctx.Difficulty, 3, 4+tier*4)
	constMax := mathutil.ScaleInt(ctx.Difficulty, 6, 8+tier*12)
	var lead, constant, n, divC, divL int
	var q, ex string
	for {
		lead = smallInt(2, leadMax)
		constant = smallInt(3, constMax)
		divC = divisors(constant)
		divL = divisors(lead)
		n = divC * divL
		ansStr := strconv.Itoa(n)
		q = fmt.Sprintf("A polynomial has leading coefficient %d and constant term %d, both non-zero. How many possible rational roots does the rational-root theorem give, before any are discarded?", lead, constant)
		if n == lead || n == constant || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("By the Rational Root Theorem, any candidate rational root in lowest terms \\(p/q\\) requires that \\(p\\) divides the constant term and \\(q\\) divides the leading coefficient. Find positive divisor counts for constant term %d (which has %d divisors) and leading coefficient %d (which has %d divisors). Multiply divisor counts to determine total possible magnitude pairings: \\(%d \\times %d = %d\\).",
			constant, divC, lead, divL, divC, divL, n)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(n),
		Explanation: ex,
		Facts: map[string]string{
			"lead":       strconv.Itoa(lead),
			"constant":   strconv.Itoa(constant),
			"candidates": strconv.Itoa(n),
			"answer":     strconv.Itoa(n),
		},
	}
}

// rationalAsymptoteGen asks for the vertical asymptote of a simple rational function.
type rationalAsymptoteGen struct{}

func (g *rationalAsymptoteGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := 1 + rand.Intn(3+int(ctx.Difficulty*3))
	b := smallInt(-4, 4+int(ctx.Difficulty*4))
	d := 1 + rand.Intn(2)
	e := d * smallInt(1, 4+int(ctx.Difficulty*4))
	return generator.Problem{
		Question: fmt.Sprintf("Where is the vertical asymptote of \\(f(x) = \\frac{%dx + %d}{%dx - %d}\\)?", a, b, d, e),
		Answer:   fmt.Sprintf("%d", e/d),
		Explanation: fmt.Sprintf("The vertical asymptote is where the denominator vanishes and the numerator does not: %dx - %d = 0 gives x = %d.",
			d, e, e/d),
		Facts: map[string]string{"x": fmt.Sprintf("%d", e/d)},
	}
}

// expEquationGen solves a base raised to an unknown power, with the right-hand side an exact power
// of that base.
type expEquationGen struct{}

func (g *expEquationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	basePool := []int{2, 3, 5, 10}
	if tier == 1 {
		basePool = []int{2, 3, 4, 5, 6, 10}
	} else if tier == 2 {
		basePool = []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
	}
	var base, k, v int
	var q, ex string
	for {
		base = basePool[rand.Intn(len(basePool))]
		kMin := 2
		kMax := 4 + tier
		if base >= 8 {
			kMax = 3
		}
		k = smallInt(kMin, kMax)
		if k == base {
			k++
		}
		v = 1
		for i := 0; i < k; i++ {
			v *= base
		}
		ansStr := strconv.Itoa(k)
		q = fmt.Sprintf("Solve \\(%d^x = %d\\) for \\(x\\).", base, v)
		if k == base || k == v || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("For exponential equations with identical bases \\(b^u = b^w\\) where \\(b > 0\\) and \\(b \\neq 1\\), the exponents must be equal: \\(u = w\\). Express the right-hand side %d as an exact power with base %d: \\(%d = %d^{%d}\\), transforming the equation to \\(%d^x = %d^{%d}\\). Equate the exponents to obtain the solution: \\(x = %d\\).",
			v, base, v, base, k, base, base, k, k)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(k),
		Explanation: ex,
		Facts: map[string]string{
			"base":   strconv.Itoa(base),
			"value":  strconv.Itoa(v),
			"x":      strconv.Itoa(k),
			"answer": strconv.Itoa(k),
		},
	}
}

// logEvaluateGen evaluates a logarithm whose argument is an exact power of the base.
type logEvaluateGen struct{}

func (g *logEvaluateGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	basePool := []int{2, 3, 5, 10}
	if tier == 1 {
		basePool = []int{2, 3, 4, 5, 6, 10}
	} else if tier == 2 {
		basePool = []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
	}
	var base, k, v int
	var q, ex string
	for {
		base = basePool[rand.Intn(len(basePool))]
		kMin := 2
		kMax := 4 + tier
		if base >= 8 {
			kMax = 3
		}
		k = smallInt(kMin, kMax)
		if k == base {
			k++
		}
		v = 1
		for i := 0; i < k; i++ {
			v *= base
		}
		ansStr := strconv.Itoa(k)
		q = fmt.Sprintf("Evaluate \\(\\log_{%d} %d\\).", base, v)
		if k == base || k == v || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("The logarithmic function is the inverse of exponentiation: \\(\\log_b(y) = x\\) if and only if \\(b^x = y\\), with property \\(\\log_b(b^k) = k\\). Rewrite argument %d as a power of base %d: \\(%d = %d^{%d}\\), yielding \\(\\log_{%d}(%d^{%d})\\). Apply the inverse identity \\(\\log_b(b^k) = k\\) to evaluate \\(\\log_{%d}(%d^{%d}) = %d\\).",
			v, base, v, base, k, base, base, k, base, base, k, k)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(k),
		Explanation: ex,
		Facts: map[string]string{
			"base":     strconv.Itoa(base),
			"argument": strconv.Itoa(v),
			"value":    strconv.Itoa(k),
			"answer":   strconv.Itoa(k),
		},
	}
}

// growthModelGen uses exact doubling or tripling, so the population stays whole.
type growthModelGen struct{}

func (g *growthModelGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	start := 1 + rand.Intn(4+int(ctx.Difficulty*6))
	hours := 1 + rand.Intn(2+int(ctx.Difficulty*3))
	periods := 1 + rand.Intn(3+int(ctx.Difficulty*3))
	factor := 2
	word := "doubles"
	if rand.Intn(2) == 0 {
		factor, word = 3, "triples"
	}
	v := start
	for i := 0; i < periods; i++ {
		v *= factor
	}
	return generator.Problem{
		Question: fmt.Sprintf("A culture starts with %d bacteria and %s every %d hours. How many are there after %d hours?", start, word, hours, hours*periods),
		Answer:   fmt.Sprintf("%d", v),
		Explanation: fmt.Sprintf("The population multiplies by %d once per %d hours, and %d hours is %d such periods, so it is %d * %d^%d = %d.",
			factor, hours, hours*periods, periods, start, factor, periods, v),
		Facts: map[string]string{"population": fmt.Sprintf("%d", v)},
	}
}

// inverseValueGen evaluates an inverse function at a point, which avoids asking for a symbolic
// inverse while testing that the inverse is what was found.
type inverseValueGen struct{}

func (g *inverseValueGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := 2 + rand.Intn(4)
	b := smallInt(-6, 6+int(ctx.Difficulty*6))
	k := smallInt(-3, 3+int(ctx.Difficulty*4))
	// f(x) = a x + b, so f^-1(y) = (y - b)/a; k is the offset so the value is whole.
	y := a*k + b
	return generator.Problem{
		Question: fmt.Sprintf("The function \\(f(x) = %dx + %d\\) is one-to-one, so it has an inverse. Evaluate \\(f^{-1}(%d)\\).", a, b, y),
		Answer:   fmt.Sprintf("%d", k),
		Explanation: fmt.Sprintf("Solve %dx + %d = %d: the inverse undoes the operations in reverse order, so x = (%d - %d)/%d = %d.",
			a, b, y, y, b, a, k),
		Facts: map[string]string{"value": fmt.Sprintf("%d", k)},
	}
}

// circleRadiusGen gives a circle in general form whose radius is whole, built from a Pythagorean
// triple and a second square.
type circleRadiusGen struct{}

func (g *circleRadiusGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	type triple struct{ p, q, r int }
	triples := []triple{{3, 4, 5}, {6, 8, 10}}
	if tier == 1 {
		triples = []triple{{3, 4, 5}, {6, 8, 10}, {5, 12, 13}, {8, 15, 17}}
	} else if tier == 2 {
		triples = []triple{{3, 4, 5}, {6, 8, 10}, {5, 12, 13}, {8, 15, 17}, {7, 24, 25}, {9, 12, 15}, {12, 16, 20}}
	}
	var t triple
	var radius, p, qCoord, f int
	var q, ex string
	for {
		t = triples[rand.Intn(len(triples))]
		p = t.p
		qCoord = t.q
		if rand.Intn(2) == 0 {
			p = -p
		}
		if rand.Intn(2) == 0 {
			qCoord = -qCoord
		}
		radMin := 2 + tier
		radMax := 4 + tier*4
		radius = smallInt(radMin, radMax)
		f = p*p + qCoord*qCoord - radius*radius
		if f == 0 {
			continue
		}
		ansStr := strconv.Itoa(radius)
		q = fmt.Sprintf("The circle \\(x^2 + y^2 %+dx %+dy %+d = 0\\) has radius \\(r\\). Find \\(r\\).", 2*p, 2*qCoord, f)
		if radius == 2*p || radius == -2*p || radius == 2*qCoord || radius == -2*qCoord || radius == f || radius == -f || containsWord(q, ansStr) {
			continue
		}
		h, k := -p, -qCoord
		ex = fmt.Sprintf("The standard equation of a circle is \\((x - h)^2 + (y - k)^2 = r^2\\). Expanding gives \\(x^2 + y^2 - 2hx - 2ky + (h^2 + k^2 - r^2) = 0\\), so \\(r = \\sqrt{h^2 + k^2 - F}\\). Identify center coordinates \\(h = %d\\), \\(k = %d\\) from linear terms and constant \\(F = %d\\): \\(r^2 = (%d)^2 + (%d)^2 - (%d)\\). Evaluate the squares and subtract constant: \\(r^2 = %d + %d - (%d) = %d\\). Taking the square root gives \\(r = \\sqrt{%d} = %d\\).",
			h, k, f, h, k, f, h*h, k*k, f, radius*radius, radius*radius, radius)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(radius),
		Explanation: ex,
		Facts: map[string]string{
			"coeff_x": strconv.Itoa(2 * p),
			"coeff_y": strconv.Itoa(2 * qCoord),
			"f":       strconv.Itoa(f),
			"radius":  strconv.Itoa(radius),
			"answer":  strconv.Itoa(radius),
		},
	}
}
