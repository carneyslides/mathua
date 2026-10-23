package calculus

// The numeric calculus concepts: the limit of a rational sequence, a recursive sequence's term,
// the sum of an arithmetic sequence, the mass of a lamina with uniform density, a definite integral
// by substitution, the mass from a linear density and a higher-order derivative at a point.
//
// Converted to parametric generators scaling with ctx.Difficulty using mathutil primitives
// and structured three-subgoal cognitive derivations in explanations.

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerNumeric(reg *generator.Registry) {
	reg.Register("calc.seq.limit_value", &sequenceLimitGen{})
	reg.Register("calc.seq.recursive_value", &recursiveValueGen{})
	reg.Register("calc.seq.arithmetic_sum", &arithmeticSumGen{})
	reg.Register("calc.multi.mass_lamina", &massLaminaGen{})
	reg.Register("calc.integral.substitution_value", &substitutionValueGen{})
	reg.Register("calc.integral.mass_density", &massDensityGen{})
	reg.Register("calc.deriv.higher_order_value", &higherOrderValueGen{})
}

// sequenceLimitGen compares the leading coefficients of a rational sequence.
type sequenceLimitGen struct{}

// squaredPoly writes q n^2 + l n + c without the leading 1 or a zero term, so a
// denominator of 1n^2 + 2n + 6 reads as n^2 + 2n + 6.
func squaredPoly(q, l, c int) string {
	out := fmt.Sprintf("%dn^2", q)
	if q == 1 {
		out = "n^2"
	}
	switch {
	case l == 1:
		out += " + n"
	case l == -1:
		out += " - n"
	case l > 1:
		out += fmt.Sprintf(" + %dn", l)
	case l < -1:
		out += fmt.Sprintf(" - %dn", -l)
	}
	if c > 0 {
		out += fmt.Sprintf(" + %d", c)
	} else if c < 0 {
		out += fmt.Sprintf(" - %d", -c)
	}
	return out
}

func (g *sequenceLimitGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxRatio := mathutil.ScaleInt(ctx.Difficulty, 3, 10)
	ratio := 1 + rand.Intn(maxRatio)
	maxD := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
	d := 1 + rand.Intn(maxD)
	for (ratio*d)%10 == 1 && ratio*d > 1 {
		ratio++
	}
	a := ratio * d
	b, c := rand.Intn(10), rand.Intn(10)
	e, f := rand.Intn(10), rand.Intn(10)
	ans := strconv.Itoa(ratio)

	q := fmt.Sprintf("Find \\(\\lim_{n \\to \\infty} \\frac{%s}{%s}\\).", squaredPoly(a, b, c), squaredPoly(d, e, f))
	exp := fmt.Sprintf("Subgoal 1: State governing formula. For a rational sequence a_n = P(n)/Q(n) where numerator and denominator have the same degree 2, the limit as n approaches infinity equals the ratio of leading coefficients: lim_{n → ∞} a_n = a_2 / b_2.\n\nSubgoal 2: Substitute parameters. Identify leading quadratic coefficients: the numerator %s has leading coefficient %d and the denominator %s has leading coefficient %d.\n\nSubgoal 3: Evaluate and simplify. Divide leading coefficients: lim_{n → ∞} a_n = %d / %d = %d.",
		squaredPoly(a, b, c), a, squaredPoly(d, e, f), d, a, d, ratio)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"answer": ans,
		},
	}
}

// recursiveValueGen iterates an affine recurrence relation.
type recursiveValueGen struct{}

func (g *recursiveValueGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxStart := mathutil.ScaleInt(ctx.Difficulty, 4, 12)
	start := 1 + rand.Intn(maxStart)
	maxR := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
	r := 2 + rand.Intn(maxR)
	maxC := mathutil.ScaleInt(ctx.Difficulty, 3, 8)
	c := 1 + rand.Intn(maxC)
	maxSteps := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
	steps := 2 + rand.Intn(maxSteps)
	target := steps + 1

	v := start
	for i := 0; i < steps; i++ {
		v = r*v + c
	}
	ans := strconv.Itoa(v)

	q := fmt.Sprintf("A sequence is defined by \\(a_1 = %d\\) and \\(a_{n+1} = %da_n %+d\\). Find \\(a_{%d}\\).", start, r, c, target)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. An affine recurrence relation a_{n+1} = r·a_n + c is solved by iteratively generating terms starting from initial term a_1.\n\nSubgoal 2: Substitute parameters. Begin with a_1 = %d and apply the recursive step a_{n+1} = %da_n %+d across %d iterations to reach term index %d.\n\nSubgoal 3: Evaluate and simplify. Evaluating the recurrence successive times yields a_%d = %d.",
		start, r, c, steps, target, target, v)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"start":  strconv.Itoa(start),
			"r":      strconv.Itoa(r),
			"target": strconv.Itoa(target),
			"answer": ans,
		},
	}
}

// arithmeticSumGen evaluates the sum of an arithmetic sequence.
type arithmeticSumGen struct{}

func (g *arithmeticSumGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxA := mathutil.ScaleInt(ctx.Difficulty, 5, 20)
	a := rand.Intn(2*maxA+1) - maxA
	maxD := mathutil.ScaleInt(ctx.Difficulty, 3, 10)
	d := 1 + rand.Intn(maxD)
	maxNHalf := mathutil.ScaleInt(ctx.Difficulty, 3, 8)
	nHalf := 1 + rand.Intn(maxNHalf)
	n := 2 * nHalf // even so n/2 is whole
	sum := n * (2*a + (n-1)*d) / 2
	ans := strconv.Itoa(sum)

	q := fmt.Sprintf("Find the sum of the arithmetic sequence with first term %d, common difference %d and %d terms.", a, d, n)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The sum of the first n terms of an arithmetic progression is given by S_n = (n/2)·[2a + (n - 1)d].\n\nSubgoal 2: Substitute parameters. Substitute initial term a = %d, common difference d = %d and term count n = %d into the summation formula: S_%d = (%d/2)·[2(%d) + (%d - 1)(%d)].\n\nSubgoal 3: Evaluate and simplify. Evaluate inside brackets: 2(%d) + (%d)(%d) = %d + %d = %d. Multiplying by n/2 = %d gives S_%d = %d · %d = %d.",
		a, d, n, n, n, a, n, d, a, n-1, d, 2*a, (n-1)*d, 2*a+(n-1)*d, nHalf, n, nHalf, 2*a+(n-1)*d, sum)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"first_term": strconv.Itoa(a),
			"difference": strconv.Itoa(d),
			"terms":      strconv.Itoa(n),
			"answer":     ans,
		},
	}
}

// massLaminaGen calculates the mass of a rectangular lamina with uniform density.
type massLaminaGen struct{}

func (g *massLaminaGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxRho := mathutil.ScaleInt(ctx.Difficulty, 4, 12)
	rho := 1 + rand.Intn(maxRho)
	maxW := mathutil.ScaleInt(ctx.Difficulty, 4, 15)
	w := 2 + rand.Intn(maxW)
	maxH := mathutil.ScaleInt(ctx.Difficulty, 4, 15)
	h := 2 + rand.Intn(maxH)
	mass := rho * w * h
	ans := strconv.Itoa(mass)

	q := fmt.Sprintf("A rectangular lamina is %d units by %d units and has uniform density %d per unit area. Find its mass.", w, h, rho)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. For a planar lamina with uniform area density ρ over region R, total mass equals M = ρ · Area(R) = ρ · (width · height).\n\nSubgoal 2: Substitute parameters. Substitute width %d, height %d and density %d into the total mass formula.\n\nSubgoal 3: Evaluate and simplify. Compute area %d · %d = %d and multiply by density: M = %d · %d = %d units.",
		w, h, rho, w, h, w*h, rho, w*h, mass)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"width":   strconv.Itoa(w),
			"height":  strconv.Itoa(h),
			"density": strconv.Itoa(rho),
			"answer":  ans,
		},
	}
}

// substitutionValueGen evaluates a definite integral using u-substitution.
type substitutionValueGen struct{}

func (g *substitutionValueGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxA := mathutil.ScaleInt(ctx.Difficulty, 3, 8)
	aHalf := 1 + rand.Intn(maxA)
	a := 2 * aHalf // even so ((a^2+1)^2 - 1)/2 is whole
	uTop := a*a + 1
	v := (uTop*uTop - 1) / 2
	ans := strconv.Itoa(v)

	q := fmt.Sprintf("Evaluate \\(\\int_0^{%d} 2x(x^2 + 1)\\,dx\\).", a)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. By u-substitution, an integral ∫_α^β g'(x) f(g(x)) dx transforms into ∫_{g(α)}^{g(β)} f(u) du using substitution u = g(x) and differential du = g'(x) dx.\n\nSubgoal 2: Substitute parameters. Let u = x^2 + 1 so du = 2x dx. When x = 0, u = 0^2 + 1 = 1; when x = %d, u = (%d)^2 + 1 = %d. The integral becomes ∫_1^{%d} u du.\n\nSubgoal 3: Evaluate and simplify. Integrate to obtain [u^2 / 2]_1^{%d} = (%d^2 - 1^2) / 2 = (%d - 1) / 2 = %d.",
		a, a, uTop, uTop, uTop, uTop, uTop*uTop, v)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"answer": ans,
		},
	}
}

// massDensityGen integrates a linear density function along a one-dimensional rod.
type massDensityGen struct{}

func (g *massDensityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxK := mathutil.ScaleInt(ctx.Difficulty, 3, 10)
	kHalf := 1 + rand.Intn(maxK)
	k := 2 * kHalf // even so k L^2 / 2 is whole
	maxL := mathutil.ScaleInt(ctx.Difficulty, 4, 15)
	l := 1 + rand.Intn(maxL)
	v := kHalf * l * l
	ans := strconv.Itoa(v)

	q := fmt.Sprintf("A rod of length %d has linear density \\(\\rho(x) = %dx\\) at a distance \\(x\\) from one end. Find its mass.", l, k)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The total mass of a straight rod of length L with variable linear density ρ(x) is given by M = ∫_0^L ρ(x) dx.\n\nSubgoal 2: Substitute parameters. Insert linear density function ρ(x) = %dx and upper integration limit L = %d: M = ∫_0^{%d} %dx dx.\n\nSubgoal 3: Evaluate and simplify. Find the antiderivative and evaluate at limits: M = [%d x^2 / 2]_0^{%d} = %d(%d)^2 = %d(%d) = %d.",
		k, l, l, k, k, l, kHalf, l, kHalf, l*l, v)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"length": strconv.Itoa(l),
			"k":      strconv.Itoa(k),
			"answer": ans,
		},
	}
}

// higherOrderValueGen evaluates a higher-order derivative of a monomial power function at a point.
type higherOrderValueGen struct{}

func (g *higherOrderValueGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxN := mathutil.ScaleInt(ctx.Difficulty, 3, 6)
	n := 3 + rand.Intn(maxN-2)
	k := 1 + rand.Intn(n-1)
	maxA := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
	a := 1 + rand.Intn(maxA)

	coef := 1
	for i := 0; i < k; i++ {
		coef *= (n - i)
	}
	pw := 1
	for i := 0; i < n-k; i++ {
		pw *= a
	}
	v := coef * pw
	ans := strconv.Itoa(v)

	q := fmt.Sprintf("For \\(f(x) = x^{%d}\\), find \\(f^{(%d)}(%d)\\).", n, k, a)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. Repeated differentiation using the power rule gives the k-th derivative of f(x) = x^n as f^{(k)}(x) = [n! / (n - k)!] · x^{n - k} for order 1 ≤ k ≤ n.\n\nSubgoal 2: Substitute parameters. For degree n = %d and derivative order k = %d, falling factorial coefficient is %d! / (%d - %d)! = %d. The derivative is f^{(%d)}(x) = %d x^{%d}.\n\nSubgoal 3: Evaluate and simplify. Evaluate derivative at x = %d: f^{(%d)}(%d) = %d · (%d)^{%d} = %d · %d = %d.",
		n, k, n, n, k, coef, k, coef, n-k, a, k, a, coef, a, n-k, coef, pw, v)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"n":      strconv.Itoa(n),
			"k":      strconv.Itoa(k),
			"a":      strconv.Itoa(a),
			"answer": ans,
		},
	}
}
