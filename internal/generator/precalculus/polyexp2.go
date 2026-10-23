package precalculus

// The second precalculus block: polynomial theory, the oblique case, logarithms and interest,
// composition, and the distance formula.
//
// The first block covered what a zero is and how to find one. These are what the zeros tell you: a
// zero's multiplicity (how the graph meets the axis), Descartes' rule (how many positive ones to
// expect), the conjugate-root theorem (why the non-real ones come in pairs), and the oblique
// asymptote a rational function has when its numerator's degree is exactly one higher. The rest fill
// gaps the first block left: solving a logarithmic equation, compounding interest, composing two
// functions, and the distance between two points.
//
// Every answer is an integer by construction, as in the first block.
import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

// smallInt lives in polyexp.go; this file uses it rather than defining a second copy.

// zerosMultiplicityGen asks for the multiplicity of one factor of a factored polynomial.
type zerosMultiplicityGen struct{}

func (g *zerosMultiplicityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := smallInt(-4, 4)
	b := smallInt(-4, 4)
	if b == a {
		b = a + 1
	}
	k := 2 + rand.Intn(2+int(ctx.Difficulty*2)) // 2..5
	m := 1 + rand.Intn(2)
	return generator.Problem{
		Question: fmt.Sprintf("The polynomial \\(p(x) = (x - %d)^{%d}(x - %d)^{%d}\\) has a zero at \\(x = %d\\). What is its multiplicity?", a, k, b, m, a),
		Answer:   fmt.Sprintf("%d", k),
		Explanation: fmt.Sprintf("The multiplicity of a zero is the exponent on its factor, which is %d. The graph meets the axis in the flat way an even multiplicity gives, rather than passing through it.",
			k),
		Facts: map[string]string{"multiplicity": fmt.Sprintf("%d", k)},
	}
}

// descartesGen counts the sign changes in a polynomial's coefficients, which bounds its positive
// real roots.
type descartesGen struct{}

func (g *descartesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	n := 3 + rand.Intn(1+int(ctx.Difficulty*2)) // 3 or 4 coefficients
	coeffs := make([]int, n)
	for i := range coeffs {
		c := smallInt(1, 6)
		if rand.Intn(2) == 0 {
			c = -c
		}
		coeffs[i] = c
	}
	changes := 0
	for i := 1; i < n; i++ {
		if (coeffs[i] > 0) != (coeffs[i-1] > 0) {
			changes++
		}
	}
	var poly []string
	for i, c := range coeffs {
		deg := n - 1 - i
		switch {
		case deg == 0:
			poly = append(poly, fmt.Sprintf("%d", c))
		case deg == 1:
			poly = append(poly, fmt.Sprintf("%dx", c))
		default:
			poly = append(poly, fmt.Sprintf("%dx^%d", c, deg))
		}
	}
	// Join the terms rather than accumulating into a string: appending inside a loop is quadratic in
	// the number of terms, which internal/lint flags, and there is no reason to pay it here.
	terms := make([]string, 0, len(poly))
	terms = append(terms, poly[0])
	for _, p := range poly[1:] {
		if p[0] == '-' {
			terms = append(terms, "- "+p[1:])
		} else {
			terms = append(terms, "+ "+p)
		}
	}
	joined := strings.Join(terms, " ")
	return generator.Problem{
		Question: fmt.Sprintf("Descartes' rule of signs counts the sign changes in the coefficients of \\(p(x)\\). For \\(p(x) = %s\\), how many sign changes are there?", joined),
		Answer:   fmt.Sprintf("%d", changes),
		Explanation: fmt.Sprintf("Reading the coefficients in order and counting where the sign flips gives %d changes, so the polynomial has at most %d positive real roots, and the count is reduced by an even number from there.",
			changes, changes),
		Facts: map[string]string{"changes": fmt.Sprintf("%d", changes)},
	}
}

// complexConjugateGen uses the conjugate-root theorem to recover a coefficient.
type complexConjugateGen struct{}

func (g *complexConjugateGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := smallInt(1, 3+int(ctx.Difficulty*3))
	b := smallInt(1, 3+int(ctx.Difficulty*3))
	// roots a + bi and a - bi give x^2 - 2a x + (a^2 + b^2).
	return generator.Problem{
		Question: fmt.Sprintf("A polynomial with real coefficients has \\(%d + %di\\) as a root. It also has the conjugate as a root. What is the coefficient of \\(x\\) in the quadratic factor \\(x^2 + Bx + C\\) built from that pair?", a, b),
		Answer:   fmt.Sprintf("%d", -2*a),
		Explanation: fmt.Sprintf("The quadratic with those two roots is (x - (%d + %di))(x - (%d - %di)) = x^2 - %dx + %d, so the coefficient of x is %d.",
			a, b, a, b, 2*a, a*a+b*b, -2*a),
		Facts: map[string]string{"B": fmt.Sprintf("%d", -2*a)},
	}
}

// obliqueAsymptoteGen asks for the slope of an oblique asymptote, which is the quotient's leading
// coefficient.
type obliqueAsymptoteGen struct{}

func (g *obliqueAsymptoteGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := smallInt(1, 2+int(ctx.Difficulty*3))
	b := smallInt(-4, 4)
	c := smallInt(-5, 5)
	d := smallInt(1, 3+int(ctx.Difficulty*3))
	return generator.Problem{
		Question: fmt.Sprintf("The function \\(f(x) = \\frac{%dx^2 + %dx + %d}{x + %d}\\) has an oblique asymptote. What is its slope?", a, b, c, d),
		Answer:   fmt.Sprintf("%d", a),
		Explanation: fmt.Sprintf("Dividing gives a quotient that is linear with leading coefficient %d, and the remainder vanishes for large x, so the oblique asymptote is \\(y = %dx + \\text{a constant}\\) and its slope is %d.",
			a, a, a),
		Facts: map[string]string{"slope": fmt.Sprintf("%d", a)},
	}
}

// logEquationGen solves a logarithmic equation whose argument is an exact power of the base.
type logEquationGen struct{}

func (g *logEquationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	basePool := []int{2, 3, 5, 10}
	if tier == 1 {
		basePool = []int{2, 3, 4, 5, 6, 10}
	} else if tier == 2 {
		basePool = []int{2, 3, 4, 5, 6, 7, 8, 10}
	}
	var base, k, v int
	var q, ex string
	for {
		base = basePool[rand.Intn(len(basePool))]
		kMin := 2
		kMax := 3 + tier
		if base >= 8 {
			kMax = 3
		}
		k = smallInt(kMin, kMax)
		v = 1
		for i := 0; i < k; i++ {
			v *= base
		}
		ansStr := strconv.Itoa(v)
		q = fmt.Sprintf("Solve \\(\\log_{%d} x = %d\\) for \\(x\\).", base, k)
		if v == base || v == k || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("By logarithmic definition, the logarithmic equation \\(\\log_b(x) = k\\) converts directly to exponential form \\(x = b^k\\). Identify base \\(b = %d\\) and exponent \\(k = %d\\), rewriting the equation as \\(x = %d^{%d}\\). Compute the power to find \\(x\\): \\(%d^{%d} = %d\\). Thus \\(x = %d\\).",
			base, k, base, k, base, k, v, v)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(v),
		Explanation: ex,
		Facts: map[string]string{
			"base":     strconv.Itoa(base),
			"exponent": strconv.Itoa(k),
			"x":        strconv.Itoa(v),
			"answer":   strconv.Itoa(v),
		},
	}
}

// compoundInterestGen uses a rate whose factor is a simple fraction, and a principal divisible by
// the denominator to that power, so the amount stays whole.
type compoundInterestGen struct{}

func (g *compoundInterestGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	rates := []struct {
		percent, num, den int
	}{{100, 2, 1}, {50, 3, 2}, {25, 5, 4}}
	r := rates[rand.Intn(len(rates))]
	t := 1 + rand.Intn(2+int(ctx.Difficulty*2))
	base := 1
	for i := 0; i < t; i++ {
		base *= r.den
	}
	principal := base * smallInt(1, 4+int(ctx.Difficulty*4))
	total := principal
	for i := 0; i < t; i++ {
		total = total / r.den * r.num
	}
	return generator.Problem{
		Question: fmt.Sprintf("%d dollars is invested at %d%% interest compounded once a year. How much is it worth after %d year(s)?", principal, r.percent, t),
		Answer:   fmt.Sprintf("%d", total),
		Explanation: fmt.Sprintf("Each year multiplies the balance by %d/%d, so after %d year(s) the amount is %d * (%d/%d)^%d = %d.",
			r.num, r.den, t, principal, r.num, r.den, t, total),
		Facts: map[string]string{"amount": fmt.Sprintf("%d", total)},
	}
}

// compositionValueGen evaluates a composition at a point.
type compositionValueGen struct{}

func (g *compositionValueGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := smallInt(1, 3+int(ctx.Difficulty*3))
	b := smallInt(-5, 5)
	c := smallInt(-3, 3)
	d := smallInt(-5, 5)
	k := smallInt(-3, 3+int(ctx.Difficulty*3))
	v := a*(c*k+d) + b
	return generator.Problem{
		Question: fmt.Sprintf("For \\(f(x) = %dx + %d\\) and \\(g(x) = %dx + %d\\), evaluate \\((f \\circ g)(%d)\\).", a, b, c, d, k),
		Answer:   fmt.Sprintf("%d", v),
		Explanation: fmt.Sprintf("Compose from the inside out: g(%d) = %d*%d + %d = %d, then f of that is %d*%d + %d = %d.",
			k, c, k, d, c*k+d, a, c*k+d, b, v),
		Facts: map[string]string{"value": fmt.Sprintf("%d", v)},
	}
}

// distanceGen uses a Pythagorean offset so the distance comes out whole.
type distanceGen struct{}

func (g *distanceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	triples := [][3]int{{3, 4, 5}, {6, 8, 10}, {5, 12, 13}, {8, 15, 17}}
	tr := triples[rand.Intn(len(triples))]
	x1 := smallInt(-5, 5)
	y1 := smallInt(-5, 5)
	sx, sy := 1, 1
	if rand.Intn(2) == 0 {
		sx = -1
	}
	if rand.Intn(2) == 0 {
		sy = -1
	}
	x2, y2 := x1+sx*tr[0], y1+sy*tr[1]
	return generator.Problem{
		Question: fmt.Sprintf("Find the distance between \\((%d,%d)\\) and \\((%d,%d)\\).", x1, y1, x2, y2),
		Answer:   fmt.Sprintf("%d", tr[2]),
		Explanation: fmt.Sprintf("The distance formula is the Pythagorean theorem on the coordinate differences: the horizontal gap is %d and the vertical gap is %d, so the distance is sqrt(%d + %d) = %d.",
			tr[0], tr[1], tr[0]*tr[0], tr[1]*tr[1], tr[2]),
		Facts: map[string]string{"distance": fmt.Sprintf("%d", tr[2])},
	}
}
