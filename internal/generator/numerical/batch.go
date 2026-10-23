package numerical

// The first numerical-analysis batch: twenty concepts that open the domain with
// error and conditioning, root finding and the first half of interpolation.
//
// The domain is about the gap between an exact mathematical object and the
// finitely many digits a machine can hold and the finitely many steps a method
// can afford. This batch measures that gap (absolute and relative error,
// roundoff, cancellation, conditioning, stability, truncation), closes it by
// iteration (bisection, fixed-point, Newton, secant, false position and the
// orders those methods achieve) and starts reconstructing a function from
// samples (Lagrange interpolation, divided differences, the error formula,
// Chebyshev nodes and splines).
//
// Nine concepts are multiple choice and eleven are numeric. Every numeric
// generator publishes the parameters its answer follows from, so the tests
// recompute the answer rather than trusting the author's arithmetic.

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func decimal6(v float64) string {
	s := strconv.FormatFloat(v, 'f', 6, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-0" {
		s = "0"
	}
	return s
}

// ---------------------------------------------------------------------------
// Absolute and relative error (numeric)
// ---------------------------------------------------------------------------

type absRelErrorGen struct{}

func (g *absRelErrorGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxTrue := mathutil.ScaleInt(ctx.Difficulty, 500, 4000)
	tVal := (1 + rand.Intn(maxTrue/100)) * 100
	maxDelta := mathutil.ScaleInt(ctx.Difficulty, 4, 25)
	delta := 1 + rand.Intn(maxDelta)
	approx := tVal + delta
	if rand.Intn(2) == 1 {
		approx = tVal - delta
	}

	relErr := float64(delta) / float64(tVal)
	ans := decimal6(relErr)

	q := fmt.Sprintf("A quantity has true value %d and is approximated by %d. What is the relative error |approx − true|/|true|? (enter a decimal)", tVal, approx)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The relative error measures the magnitude of approximation error relative to the true quantity: \\(\\text{RelErr} = \\frac{|\\text{approx} - \\text{true}|}{|\\text{true}|}\\).\n\nSubgoal 2: Substitute parameters. Given true value %d and approximation %d, substitute into the formula: \\(\\text{RelErr} = \\frac{|%d - %d|}{|%d|}\\).\n\nSubgoal 3: Evaluate and simplify. Compute the absolute difference \\(|%d - %d| = %d\\) and divide by %d: \\(\\text{RelErr} = \\frac{%d}{%d} = %s\\).",
		tVal, approx, approx, tVal, tVal,
		approx, tVal, delta, tVal, delta, tVal, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"true":   strconv.Itoa(tVal),
			"approx": strconv.Itoa(approx),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// Floating-point roundoff and machine epsilon (multiple choice)
// ---------------------------------------------------------------------------

type roundoffGen struct{}

func (g *roundoffGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Machine epsilon is...",
			correct:     "the gap between one and the next representable floating-point number",
			distractors: []string{"the smallest positive number a machine can hold", "the largest exponent", "the number of decimal digits printed"},
			reason:      "It measures the relative spacing of the floats near one, which is the natural unit of roundoff.",
		},
		{
			question:    "The number nought point one cannot be stored exactly in binary floating point because...",
			correct:     "its binary expansion does not terminate, so it must be rounded",
			distractors: []string{"it is larger than one", "it is negative", "the machine runs out of memory"},
			reason:      "Only fractions whose denominator is a power of two terminate in binary.",
		},
		{
			question:    "Roundoff error is unavoidable in floating-point arithmetic because...",
			correct:     "the machine keeps only finitely many significant digits",
			distractors: []string{"algorithms are written badly", "the input data are always noisy", "mathematics is inexact"},
			reason:      "Every real number is rounded to the nearest member of a finite set.",
		},
		{
			question:    "A single rounding of a real number to the nearest float produces a relative error no larger than...",
			correct:     "about half of machine epsilon",
			distractors: []string{"one", "the square root of machine epsilon", "the number itself"},
			reason:      "Rounding to nearest can never be off by more than half the gap to the next float.",
		},
		{
			question:    "Floating-point addition is not associative, so...",
			correct:     "the order of a sum can change the computed result",
			distractors: []string{"addition is undefined", "the sum is always exact", "subtraction is impossible"},
			reason:      "Each intermediate result is rounded, so regrouping changes which roundings occur.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Catastrophic cancellation (numeric)
// ---------------------------------------------------------------------------

type cancellationGen struct{}

func (g *cancellationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxK := mathutil.ScaleInt(ctx.Difficulty, 10, 200)
	k := 5 + rand.Intn(maxK-4)
	d := 2 + rand.Intn(10)
	if ((k+1)*d)%2 != 0 {
		d = 2 * d
	}
	x := (k + 1) * d / 2
	y := x - d

	for iter := 0; iter < 5 && (k == x || k == y); iter++ {
		k += 10
		x = (k + 1) * d / 2
		y = x - d
	}

	ans := decimal6(float64(x+y) / float64(d))
	q := fmt.Sprintf("The subtraction x − y loses accuracy when x and y are close: the relative error of the difference is amplified by (|x| + |y|)/|x − y| relative to the inputs. For x = %d and y = %d, what is that amplification factor? (enter a decimal)", x, y)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. Subtraction of nearby quantities suffers from catastrophic cancellation, amplifying input errors by the factor \\(\\kappa = \\frac{|x| + |y|}{|x - y|}\\).\n\nSubgoal 2: Substitute parameters. Substitute \\(x = %d\\) and \\(y = %d\\) into the amplification ratio: \\(\\kappa = \\frac{|%d| + |%d|}{|%d - %d|}\\).\n\nSubgoal 3: Evaluate and simplify. Compute the numerator \\(%d + %d = %d\\) and denominator \\(|%d| = %d\\), then divide: \\(\\kappa = \\frac{%d}{%d} = %s\\).",
		x, y, x, y, x, y,
		x, y, x+y, x-y, d, x+y, d, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"x":      strconv.Itoa(x),
			"y":      strconv.Itoa(y),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// The condition number of a problem (numeric)
// ---------------------------------------------------------------------------

type conditioningGen struct{}

func (g *conditioningGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxN := mathutil.ScaleInt(ctx.Difficulty, 6, 25)
	n := 2 + rand.Intn(maxN-1)
	ans := decimal6(float64(n))

	q := fmt.Sprintf("The relative condition number of the function x raised to the power n at a point x is |x·f'(x)|/|f(x)|. For f(x) = x^%d, what is that condition number? (enter a decimal)", n)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The relative condition number of a differentiable function \\(f(x)\\) measures output sensitivity to input perturbation: \\(\\kappa(x) = \\frac{|x f'(x)|}{|f(x)|}\\).\n\nSubgoal 2: Substitute parameters. For monomial \\(f(x) = x^{%d}\\), compute derivative \\(f'(x) = %d x^{%d}\\) and substitute: \\(\\kappa(x) = \\frac{|x \\cdot %d x^{%d}|}{|x^{%d}|}\\).\n\nSubgoal 3: Evaluate and simplify. Multiply in the numerator to obtain \\(%d |x^{%d}|\\) and divide out \\(|x^{%d}|\\): \\(\\kappa(x) = %d\\).",
		n, n, n-1, n, n-1, n,
		n, n, n, n)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"n":      strconv.Itoa(n),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// Stability of an algorithm (multiple choice)
// ---------------------------------------------------------------------------

type stabilityGen struct{}

func (g *stabilityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "An algorithm is backward stable when...",
			correct:     "it returns the exact answer to a nearby problem",
			distractors: []string{"it is fast", "it uses little memory", "it returns the exact answer to the given problem"},
			reason:      "The rounding is pushed back into a small perturbation of the data, which the condition number then amplifies.",
		},
		{
			question:    "A stable algorithm applied to an ill-conditioned problem...",
			correct:     "can still return an inaccurate answer, because the problem amplifies the perturbation",
			distractors: []string{"always returns an accurate answer", "cannot be implemented", "has no rounding error"},
			reason:      "Stability controls the algorithm; the condition number controls the problem and the two multiply.",
		},
		{
			question:    "The condition number of a problem measures...",
			correct:     "how much a relative change in the data can change the relative answer",
			distractors: []string{"the speed of the algorithm", "the memory it uses", "the order of convergence"},
			reason:      "It is a property of the problem alone, independent of the method chosen to solve it.",
		},
		{
			question:    "A backward-stable algorithm guarantees that the computed result is the exact result of...",
			correct:     "a problem whose data differ from the given data by a small relative amount",
			distractors: []string{"the given problem exactly", "any problem of the same size", "the problem in infinite precision"},
			reason:      "That is the definition of backward stability.",
		},
		{
			question:    "For a well-conditioned problem with condition number near one, a backward-stable algorithm...",
			correct:     "has a relative error of the order of machine epsilon",
			distractors: []string{"has a relative error near one", "has no error at all", "diverges"},
			reason:      "The small backward perturbation passes through a problem that barely amplifies it.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Taylor truncation error (multiple choice)
// ---------------------------------------------------------------------------

type taylorErrorGen struct{}

func (g *taylorErrorGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "The truncation error of a numerical method is the error made by...",
			correct:     "replacing an exact process with a finite approximation",
			distractors: []string{"finite-precision arithmetic", "noisy input data", "a badly chosen algorithm"},
			reason:      "It is the price of cutting off a series or a limiting process and it is separate from roundoff.",
		},
		{
			question:    "For a method of order n the local truncation error behaves like...",
			correct:     "a constant times the step raised to the power n plus one",
			distractors: []string{"the step itself", "one over the step", "a constant, independent of the step"},
			reason:      "The leading term left after a Taylor expansion of order n is the next one.",
		},
		{
			question:    "Halving the step of a second-order method reduces its global error by about...",
			correct:     "a factor of four",
			distractors: []string{"a factor of two", "a factor of eight", "nothing"},
			reason:      "A second-order method has global error proportional to the square of the step.",
		},
		{
			question:    "Taylor's theorem with remainder is used in numerical analysis to...",
			correct:     "bound the truncation error of a difference formula",
			distractors: []string{"compute exact derivatives", "solve linear systems", "factor a matrix"},
			reason:      "The remainder term is exactly the error the formula makes.",
		},
		{
			question:    "The global error of a one-step method accumulates...",
			correct:     "over the whole interval, not just one step",
			distractors: []string{"over one step only", "over the number of digits", "only from the input data"},
			reason:      "The local errors of many steps add up, which is why the global order is one lower than the local order.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// The bisection method (numeric)
// ---------------------------------------------------------------------------

type bisectionGen struct{}

func (g *bisectionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	widths := []float64{1, 2, 4, 8, 16, 32}
	w := widths[rand.Intn(len(widths))]
	wStr := strconv.Itoa(int(w))

	maxK := mathutil.ScaleInt(ctx.Difficulty, 6, 16)
	k := 4 + rand.Intn(maxK-3)
	tol := w / math.Pow(2, float64(k))
	tolStr := strconv.FormatFloat(tol, 'f', -1, 64)

	steps := int(math.Ceil(math.Log2(w / tol)))
	ans := strconv.Itoa(steps)

	for iter := 0; iter < 5 && ans == wStr; iter++ {
		k++
		tol = w / math.Pow(2, float64(k))
		tolStr = strconv.FormatFloat(tol, 'f', -1, 64)
		steps = int(math.Ceil(math.Log2(w / tol)))
		ans = strconv.Itoa(steps)
	}

	q := fmt.Sprintf("Bisection halves the interval at each step. Starting from an interval of width %s, how many steps are needed to make the width at most %s? (enter a whole number)", wStr, tolStr)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The bisection method halves the search interval at each iteration, giving width \\(W_k = \\frac{W_0}{2^k}\\) after k steps. Requiring \\(W_k \\le \\epsilon\\) implies \\(2^k \\ge \\frac{W_0}{\\epsilon}\\), so \\(k = \\lceil \\log_2(W_0 / \\epsilon) \\rceil\\).\n\nSubgoal 2: Substitute parameters. Substitute initial width \\(W_0 = %s\\) and tolerance \\(\\epsilon = %s\\): \\(\\frac{W_0}{\\epsilon} = \\frac{%s}{%s} = %s\\).\n\nSubgoal 3: Evaluate and simplify. Take the base-two logarithm \\(\\log_2(%s) = %d\\), which establishes that exactly %d steps are required.",
		wStr, tolStr, wStr, tolStr, strconv.FormatFloat(w/tol, 'f', -1, 64),
		strconv.FormatFloat(w/tol, 'f', -1, 64), steps, steps)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"width":  wStr,
			"tol":    tolStr,
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// Fixed-point iteration (numeric)
// ---------------------------------------------------------------------------

type fixedPointGen struct{}

func (g *fixedPointGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxA := mathutil.ScaleInt(ctx.Difficulty, 3, 10)
	a := 1 + rand.Intn(maxA)
	bd := 2 + rand.Intn(3)
	bn := 1
	if bd > 2 && rand.Intn(2) == 1 {
		bn = bd - 1
	}
	x0 := rand.Intn(4)
	n := 2
	if ctx.Difficulty >= 0.5 {
		n = 3
	}

	x := float64(x0)
	b := float64(bn) / float64(bd)
	var steps []string
	cur := x
	for i := 0; i < n; i++ {
		cur = float64(a) + b*cur
		steps = append(steps, fmt.Sprintf("x_%d = %d + (%d/%d)(%s) = %s", i+1, a, bn, bd, decimal6(x), decimal6(cur)))
		x = cur
	}
	ans := decimal6(x)

	q := fmt.Sprintf("Fixed-point iteration applies x_{k+1} = a + b·x_k with a = %d, b = %d/%d, starting from x_0 = %d. What is x_%d after %d steps? (enter a decimal)",
		a, bn, bd, x0, n, n)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. Fixed-point iteration generates sequential estimates through the map \\(x_{k+1} = g(x_k)\\), where \\(g(x) = a + b x\\).\n\nSubgoal 2: Substitute parameters. Given affine map \\(g(x) = %d + \\frac{%d}{%d} x\\) and initial estimate \\(x_0 = %d\\), evaluate iterates iteratively.\n\nSubgoal 3: Evaluate and simplify. %s, concluding with \\(x_%d = %s\\).",
		a, bn, bd, x0, strings.Join(steps, ", "), n, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"bn":     strconv.Itoa(bn),
			"bd":     strconv.Itoa(bd),
			"x0":     strconv.Itoa(x0),
			"n":      strconv.Itoa(n),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// Newton's method (numeric)
// ---------------------------------------------------------------------------

type newtonGen struct{}

func (g *newtonGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxC := mathutil.ScaleInt(ctx.Difficulty, 8, 40)
	c := 2 + rand.Intn(maxC-1)
	maxX0 := mathutil.ScaleInt(ctx.Difficulty, 3, 10)
	x0 := 1 + rand.Intn(maxX0)

	for x0*x0 == c {
		c++
	}

	cStr := strconv.Itoa(c)
	x0Str := strconv.Itoa(x0)
	x1 := (float64(x0) + float64(c)/float64(x0)) / 2
	ans := decimal6(x1)

	for iter := 0; iter < 5 && (ans == cStr || ans == x0Str); iter++ {
		c += 3
		cStr = strconv.Itoa(c)
		x1 = (float64(x0) + float64(c)/float64(x0)) / 2
		ans = decimal6(x1)
	}

	q := fmt.Sprintf("One Newton step for f(x) = x² − c is x_1 = (x_0 + c/x_0)/2. For c = %s and x_0 = %s, what is x_1? (enter a decimal)", cStr, x0Str)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. Newton's method for \\(f(x) = x^2 - c\\) simplifies to the Babylonian averaging recurrence \\(x_1 = \\frac{1}{2} \\left(x_0 + \\frac{c}{x_0}\\right)\\).\n\nSubgoal 2: Substitute parameters. Substitute parameter \\(c = %s\\) and initial guess \\(x_0 = %s\\): \\(x_1 = \\frac{1}{2} \\left(%s + \\frac{%s}{%s}\\right)\\).\n\nSubgoal 3: Evaluate and simplify. Compute quotient \\(\\frac{%s}{%s} = %s\\), add %s and divide by 2: \\(x_1 = %s\\).",
		cStr, x0Str, x0Str, cStr, x0Str,
		cStr, x0Str, decimal6(float64(c)/float64(x0)), x0Str, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"c":      cStr,
			"x0":     x0Str,
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// The secant method (numeric)
// ---------------------------------------------------------------------------

type secantGen struct{}

func (g *secantGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxC := mathutil.ScaleInt(ctx.Difficulty, 6, 30)
	c := float64(2 + rand.Intn(maxC-1))
	x0 := float64(1 + rand.Intn(4))
	x1 := x0 + float64(1+rand.Intn(3))

	f := func(x float64) float64 { return x*x - c }
	for f(x1) == f(x0) {
		x1++
	}
	x2 := x1 - f(x1)*(x1-x0)/(f(x1)-f(x0))
	ans := decimal6(x2)

	for iter := 0; iter < 5 && (ans == decimal6(c) || ans == decimal6(x0) || ans == decimal6(x1)); iter++ {
		c += 2
		x2 = x1 - f(x1)*(x1-x0)/(f(x1)-f(x0))
		ans = decimal6(x2)
	}

	q := fmt.Sprintf("One step of the secant method for f(x) = x² − c uses x_2 = x_1 − f(x_1)(x_1 − x_0)/(f(x_1) − f(x_0)). For c = %s, x_0 = %s and x_1 = %s, what is x_2? (enter a decimal)",
		decimal6(c), decimal6(x0), decimal6(x1))
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The secant method replaces the analytical derivative with the difference quotient across the two previous points: \\(x_2 = x_1 - f(x_1) \\frac{x_1 - x_0}{f(x_1) - f(x_0)}\\).\n\nSubgoal 2: Substitute parameters. For \\(f(x) = x^2 - %s\\), evaluate \\(f(x_0) = f(%s) = %s\\) and \\(f(x_1) = f(%s) = %s\\). Substitute into the secant recurrence: \\(x_2 = %s - (%s) \\frac{%s - %s}{(%s) - (%s)}\\).\n\nSubgoal 3: Evaluate and simplify. Evaluate the finite difference quotient and subtract to obtain \\(x_2 = %s\\).",
		decimal6(c), decimal6(x0), decimal6(f(x0)), decimal6(x1), decimal6(f(x1)),
		decimal6(x1), decimal6(f(x1)), decimal6(x1), decimal6(x0), decimal6(f(x1)), decimal6(f(x0)),
		ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"c":      decimal6(c),
			"x0":     decimal6(x0),
			"x1":     decimal6(x1),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// The method of false position (multiple choice)
// ---------------------------------------------------------------------------

type falsePositionGen struct{}

func (g *falsePositionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "The method of false position keeps a bracket of the root and updates...",
			correct:     "one endpoint, by the intercept of the line joining the two endpoints",
			distractors: []string{"both endpoints by halving", "the midpoint of the bracket", "the derivative at the midpoint"},
			reason:      "It is the secant method constrained to keep the sign change inside the bracket.",
		},
		{
			question:    "False position usually converges...",
			correct:     "faster than bisection but slower than Newton",
			distractors: []string{"faster than Newton", "more slowly than bisection", "at exactly the bisection rate"},
			reason:      "It uses the function values to place the new endpoint, which beats halving but lacks the derivative's speed.",
		},
		{
			question:    "A known weakness of false position is that...",
			correct:     "one endpoint can stay fixed for many iterations when the function is very curved",
			distractors: []string{"it needs the derivative", "it cannot start without a bracket", "it always diverges"},
			reason:      "The interpolation can land very close to one side, so the other endpoint moves little.",
		},
		{
			question:    "Compared with the secant method, false position...",
			correct:     "keeps the root bracketed",
			distractors: []string{"does not bracket the root", "requires the derivative", "cannot converge superlinearly"},
			reason:      "Dropping the secant point that would leave the bracket is what distinguishes the two methods.",
		},
		{
			question:    "False position needs to begin with...",
			correct:     "two points where the function has opposite signs",
			distractors: []string{"one point and the derivative", "two points with the same sign", "no points at all"},
			reason:      "The sign change is the bracket that guarantees a root inside.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Order of convergence (numeric)
// ---------------------------------------------------------------------------

type orderGen struct{}

func (g *orderGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxP := mathutil.ScaleInt(ctx.Difficulty, 3, 5)
	p := 1 + rand.Intn(maxP)
	a := p + 1 + rand.Intn(3)
	diff := 1 + rand.Intn(2)
	b := a + diff
	c := b + p*diff

	ans := decimal6(float64(c-b) / float64(b-a))

	q := fmt.Sprintf("The errors of an iterative method satisfy e_{k+2} ≈ e_{k+1}^p / e_k^{p−1}. From three successive errors of 10^-%d, 10^-%d and 10^-%d, estimate the order p as log(e_{k+2}/e_{k+1}) divided by log(e_{k+1}/e_k). (enter a decimal)",
		a, b, c)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The order of convergence p satisfies \\(p = \\frac{\\log(e_{k+2}/e_{k+1})}{\\log(e_{k+1}/e_k)}\\). For error powers \\(e_i = 10^{-a_i}\\), this logarithmic ratio reduces directly to the exponent difference ratio: \\(p = \\frac{c - b}{b - a}\\).\n\nSubgoal 2: Substitute parameters. Substitute exponents \\(a = %d\\), \\(b = %d\\) and \\(c = %d\\): \\(p = \\frac{%d - %d}{%d - %d}\\).\n\nSubgoal 3: Evaluate and simplify. Compute numerator \\(%d - %d = %d\\) and denominator \\(%d - %d = %d\\), then divide: \\(p = \\frac{%d}{%d} = %s\\).",
		a, b, c, c, b, b, a,
		c, b, c-b, b, a, b-a, c-b, b-a, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"b":      strconv.Itoa(b),
			"c":      strconv.Itoa(c),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// Multiple roots and modified Newton (multiple choice)
// ---------------------------------------------------------------------------

type multipleRootGen struct{}

func (g *multipleRootGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "At a root of multiplicity m, the ordinary Newton iteration converges...",
			correct:     "linearly rather than quadratically",
			distractors: []string{"quadratically", "not at all", "cubically"},
			reason:      "The zero of the derivative at the root removes the quadratic term that normally drives the fast convergence in the root",
		},
		{
			question:    "The modified Newton iteration for a root of multiplicity m...",
			correct:     "multiplies the Newton step by m",
			distractors: []string{"divides the Newton step by m", "drops the derivative", "uses the second derivative only"},
			reason:      "Restoring the quadratic rate needs a step m times the ordinary one.",
		},
		{
			question:    "A function with a double root has...",
			correct:     "both the function and its derivative equal to zero at the root",
			distractors: []string{"only the function equal to zero", "only the derivative equal to zero", "neither equal to zero"},
			reason:      "Multiplicity counts how many derivatives vanish with the function.",
		},
		{
			question:    "Near a simple root, Newton's method converges...",
			correct:     "quadratically",
			distractors: []string{"linearly", "cubically", "not at all"},
			reason:      "A nonzero derivative at the root gives the doubling of correct digits each step.",
		},
		{
			question:    "Detecting a multiple root numerically is hard because...",
			correct:     "the function is flat near the root, so small errors blur the multiplicity",
			distractors: []string{"the derivative is infinite there", "the function is discontinuous", "there is no root"},
			reason:      "Near a repeated root the graph is tangent to the axis and the tiny function values are dominated by roundoff.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Bracketing versus open methods (multiple choice)
// ---------------------------------------------------------------------------

type bracketingGen struct{}

func (g *bracketingGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "A bracketing method needs...",
			correct:     "two points where the function has opposite signs",
			distractors: []string{"one point and the derivative", "two points with the same sign", "no starting points"},
			reason:      "The sign change guarantees a root inside the interval by continuity.",
		},
		{
			question:    "An open method needs...",
			correct:     "one or two starting guesses and may diverge",
			distractors: []string{"a sign change to begin", "an interval that contains the root", "a guaranteed bracket"},
			reason:      "Without a bracket there is no guarantee, so the method can wander or cycle.",
		},
		{
			question:    "Bisection is guaranteed to converge...",
			correct:     "whenever the initial interval brackets a root",
			distractors: []string{"for any starting points", "only for polynomials", "only when the derivative is known"},
			reason:      "Halving a bracket that contains a sign change keeps it containing one.",
		},
		{
			question:    "Newton's method can fail when...",
			correct:     "the derivative is zero or the iterate leaves the region of attraction",
			distractors: []string{"the function is continuous", "a bracket is given", "the root is simple"},
			reason:      "A vanishing derivative makes the step infinite and a distant iterate can converge elsewhere or diverge.",
		},
		{
			question:    "The secant method differs from false position in that...",
			correct:     "it does not keep the root bracketed",
			distractors: []string{"it needs the derivative", "it halves the interval", "it cannot converge superlinearly"},
			reason:      "The secant method uses the newest two points even if the bracket is lost, which speeds it up and weakens the guarantee.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Horner's method (numeric)
// ---------------------------------------------------------------------------

type hornerGen struct{}

func (g *hornerGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	deg := 2
	if ctx.Difficulty >= 0.5 {
		deg = 3
	}
	maxCoeff := mathutil.ScaleInt(ctx.Difficulty, 3, 10)
	coeffs := make([]int, deg+1)
	for i := range coeffs {
		coeffs[i] = rand.Intn(2*maxCoeff+1) - maxCoeff
	}
	if coeffs[deg] == 0 {
		coeffs[deg] = 1 + rand.Intn(maxCoeff)
	}
	maxX := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
	x := 2 + rand.Intn(maxX-1)

	val := 0
	var stepDesc []string
	for k := len(coeffs) - 1; k >= 0; k-- {
		prev := val
		val = val*x + coeffs[k]
		if k == len(coeffs)-1 {
			stepDesc = append(stepDesc, fmt.Sprintf("b_%d = %d", k, val))
		} else {
			stepDesc = append(stepDesc, fmt.Sprintf("b_%d = (%d)(%d) + (%d) = %d", k, prev, x, coeffs[k], val))
		}
	}
	ans := strconv.Itoa(val)

	parts := make([]string, len(coeffs))
	for i, a := range coeffs {
		parts[i] = strconv.Itoa(a)
	}

	q := fmt.Sprintf("Horner's method evaluates a polynomial by nested multiplication. For the polynomial with coefficients (constant first) %s evaluated at x = %d, what is the value p(x)? (enter a decimal)",
		strings.Join(parts, ", "), x)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. Horner's nested evaluation computes \\(p(x) = a_n x^n + \\dots + a_1 x + a_0\\) through intermediate values: \\(b_n = a_n\\) and \\(b_k = b_{k+1} x + a_k\\) for \\(k = n-1, \\dots, 0\\), where \\(p(x) = b_0\\).\n\nSubgoal 2: Substitute parameters. For coefficients \\((%s)\\) and evaluation point \\(x = %d\\), compute nested operations sequentially.\n\nSubgoal 3: Evaluate and simplify. %s, which yields final polynomial value \\(p(%d) = %s\\).",
		strings.Join(parts, ", "), x, strings.Join(stepDesc, ", "), x, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"coeffs": strings.Join(parts, ","),
			"x":      strconv.Itoa(x),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// Lagrange interpolation (numeric)
// ---------------------------------------------------------------------------

type lagrangeGen struct{}

func (g *lagrangeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxA := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
	a := 1 + rand.Intn(maxA)
	maxB := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
	b := rand.Intn(2*maxB+1) - maxB
	maxC := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
	c := rand.Intn(2*maxC+1) - maxC
	maxQ := mathutil.ScaleInt(ctx.Difficulty, 3, 6)
	qVal := 3 + rand.Intn(maxQ-2)

	y0 := c
	y1 := a + b + c
	y2 := 4*a + 2*b + c
	val := a*qVal*qVal + b*qVal + c
	ans := strconv.Itoa(val)

	for iter := 0; iter < 5 && (ans == strconv.Itoa(y0) || ans == strconv.Itoa(y1) || ans == strconv.Itoa(y2) || ans == strconv.Itoa(qVal)); iter++ {
		a++
		y1 = a + b + c
		y2 = 4*a + 2*b + c
		val = a*qVal*qVal + b*qVal + c
		ans = strconv.Itoa(val)
	}

	l0 := (qVal - 1) * (qVal - 2) / 2
	l1 := -qVal * (qVal - 2)
	l2 := qVal * (qVal - 1) / 2

	q := fmt.Sprintf("The Lagrange interpolant through the three points (0, %d), (1, %d) and (2, %d) is the quadratic through them. What is the interpolated value at x = %d? (enter a decimal)",
		y0, y1, y2, qVal)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The quadratic Lagrange interpolant through nodes 0, 1 and 2 is \\(P(x) = y_0 L_0(x) + y_1 L_1(x) + y_2 L_2(x)\\), with cardinal basis polynomials \\(L_0(x) = \\frac{(x-1)(x-2)}{2}\\), \\(L_1(x) = -x(x-2)\\) and \\(L_2(x) = \\frac{x(x-1)}{2}\\).\n\nSubgoal 2: Substitute parameters. At query point \\(x = %d\\), evaluate basis weights: \\(L_0(%d) = %d\\), \\(L_1(%d) = %d\\) and \\(L_2(%d) = %d\\).\n\nSubgoal 3: Evaluate and simplify. Multiply each basis weight by its nodal value: \\(P(%d) = (%d)(%d) + (%d)(%d) + (%d)(%d) = %s\\).",
		qVal, qVal, l0, qVal, l1, qVal, l2,
		qVal, y0, l0, y1, l1, y2, l2, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"y0":     strconv.Itoa(y0),
			"y1":     strconv.Itoa(y1),
			"y2":     strconv.Itoa(y2),
			"q":      strconv.Itoa(qVal),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// Newton's divided differences (numeric)
// ---------------------------------------------------------------------------

type dividedDiffGen struct{}

func (g *dividedDiffGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	maxX0 := mathutil.ScaleInt(ctx.Difficulty, 1, 5)
	x0 := rand.Intn(maxX0 + 1)
	maxDx := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
	dx := 1 + rand.Intn(maxDx)
	x1 := x0 + dx

	maxM := mathutil.ScaleInt(ctx.Difficulty, 4, 30)
	m := 2 + rand.Intn(maxM)
	for m == x0 || m == x1 {
		m++
	}
	maxY0 := mathutil.ScaleInt(ctx.Difficulty, 10, 30)
	y0 := 20 + rand.Intn(maxY0)
	for y0 == m {
		y0 += 5
	}
	y1 := y0 + m*dx
	for y1 == m {
		y0 += 5
		y1 = y0 + m*dx
	}
	ans := decimal6(float64(m))

	q := fmt.Sprintf("The first divided difference of the points (%d, %d) and (%d, %d) is (y_1 − y_0)/(x_1 − x_0). What is it? (enter a decimal)",
		x0, y0, x1, y1)
	exp := fmt.Sprintf("Subgoal 1: State governing formula. The first divided difference between two distinct points \\((x_0, y_0)\\) and \\((x_1, y_1)\\) measures the average rate of change: \\(f[x_0, x_1] = \\frac{y_1 - y_0}{x_1 - x_0}\\).\n\nSubgoal 2: Substitute parameters. Substitute coordinates \\((%d, %d)\\) and \\((%d, %d)\\): \\(f[x_0, x_1] = \\frac{%d - (%d)}{%d - %d}\\).\n\nSubgoal 3: Evaluate and simplify. Compute numerator \\(%d - (%d) = %d\\) and denominator \\(%d - %d = %d\\), then divide: \\(f[x_0, x_1] = \\frac{%d}{%d} = %s\\).",
		x0, y0, x1, y1, y1, y0, x1, x0,
		y1, y0, y1-y0, x1, x0, dx, y1-y0, dx, ans)

	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: exp,
		Facts: map[string]string{
			"x0":     strconv.Itoa(x0),
			"y0":     strconv.Itoa(y0),
			"x1":     strconv.Itoa(x1),
			"y1":     strconv.Itoa(y1),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// The interpolation error (multiple choice)
// ---------------------------------------------------------------------------

type interpErrorGen struct{}

func (g *interpErrorGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "The error of polynomial interpolation at a point is proportional to...",
			correct:     "the product of the distances from that point to the interpolation nodes",
			distractors: []string{"the spacing of the nodes only", "the number of nodes", "the derivative at the nodes"},
			reason:      "The node product vanishes at the nodes and grows between them and it multiplies the derivative term.",
		},
		{
			question:    "By the interpolation error formula, the error involves the (n+1)-th derivative...",
			correct:     "evaluated at some unknown point of the interval",
			distractors: []string{"at the left endpoint", "at a node", "at infinity"},
			reason:      "The remainder theorem places the unknown point strictly inside the interval.",
		},
		{
			question:    "Adding more equally spaced nodes can...",
			correct:     "make the error worse near the ends, as Runge's example shows",
			distractors: []string{"always reduce the maximum error", "never change the error", "remove the need for a derivative"},
			reason:      "Runge's phenomenon is the blow-up of the interpolant's edges when the nodes are equally spaced.",
		},
		{
			question:    "The interpolation error is zero at...",
			correct:     "the interpolation nodes",
			distractors: []string{"the midpoint", "every point", "no point"},
			reason:      "The interpolant matches the function exactly at the data.",
		},
		{
			question:    "The error formula shows that accuracy depends on...",
			correct:     "both the node spacing and the size of the derivatives",
			distractors: []string{"only the node spacing", "only the function values", "neither"},
			reason:      "The node product and the derivative bound enter as a product.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Chebyshev nodes (multiple choice)
// ---------------------------------------------------------------------------

type chebyshevGen struct{}

func (g *chebyshevGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Chebyshev nodes are...",
			correct:     "the roots of a Chebyshev polynomial, clustered toward the ends of the interval",
			distractors: []string{"equally spaced points", "the endpoints only", "random points"},
			reason:      "Their clustering tames the edge growth that equally spaced nodes provoke.",
		},
		{
			question:    "Using Chebyshev nodes helps because they...",
			correct:     "minimise the maximum absolute value of the node product in the error formula",
			distractors: []string{"make the function linear", "remove the derivative term", "use fewer nodes"},
			reason:      "The node product is exactly what the Chebyshev construction minimises.",
		},
		{
			question:    "Chebyshev nodes cluster near...",
			correct:     "the endpoints of the interval",
			distractors: []string{"the centre", "the left endpoint only", "the data values"},
			reason:      "The density rises where the equally spaced error is worst.",
		},
		{
			question:    "Runge's phenomenon on equally spaced nodes is...",
			correct:     "the error blowing up near the interval ends",
			distractors: []string{"the error at the centre", "a machine-precision effect", "a monotone convergence"},
			reason:      "It is the failure of equally spaced interpolation to converge for smooth functions with large higher derivatives.",
		},
		{
			question:    "Among all choices of nodes, the Chebyshev nodes give the node product...",
			correct:     "the smallest possible maximum absolute value",
			distractors: []string{"the largest possible maximum", "a vanishing derivative", "no zeros"},
			reason:      "This is the minimax property of the Chebyshev polynomial.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Cubic splines (multiple choice)
// ---------------------------------------------------------------------------

type splinesGen struct{}

func (g *splinesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "A cubic spline is...",
			correct:     "a piecewise cubic that matches the values and the first two derivatives at the knots",
			distractors: []string{"a single high-degree polynomial", "a piecewise linear function", "a rational function"},
			reason:      "The derivative matching makes the curve smooth across each knot.",
		},
		{
			question:    "Splines are preferred to one interpolating polynomial because...",
			correct:     "they avoid the wild oscillation of high-degree polynomials",
			distractors: []string{"they are exact", "they need no data", "they are linear"},
			reason:      "Low-degree pieces keep the interpolant close to the data between the knots.",
		},
		{
			question:    "The natural cubic spline has...",
			correct:     "zero second derivative at the two endpoints",
			distractors: []string{"zero first derivative at the endpoints", "a prescribed slope everywhere", "no boundary conditions"},
			reason:      "Two extra conditions are needed to pin down the otherwise free end pieces.",
		},
		{
			question:    "The interpolation error of a cubic spline behaves like...",
			correct:     "a constant times the fourth power of the knot spacing",
			distractors: []string{"the square of the knot spacing", "the knot spacing", "a constant"},
			reason:      "The fourth-order behaviour comes from bounding the fourth derivative on each piece.",
		},
		{
			question:    "A spline of degree one is...",
			correct:     "a piecewise linear interpolant",
			distractors: []string{"a quadratic", "a cubic", "a constant"},
			reason:      "Degree counts the highest power used on each piece.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
