package numerical

// The second numerical-analysis batch: the twenty concepts that close the
// domain with the rest of interpolation and approximation, numerical
// integration, and numerical linear algebra.
//
// The first batch measured error, found roots and started interpolating. This
// batch finishes the reconstruction of a function from data (least squares, the
// normal equations, orthogonal polynomials and the dangers of extrapolation),
// replaces the exact integral with a weighted sum of function values
// (trapezoid, midpoint, Simpson, composite rules, Romberg, Gaussian and
// adaptive quadrature, and improper integrals), and solves the linear system
// (Gaussian elimination, pivoting, LU, Jacobi, Gauss-Seidel, the power method
// and least squares by QR).
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

// ---------------------------------------------------------------------------
// Least squares (numeric)
// ---------------------------------------------------------------------------

type leastSquaresGen struct{}

func (g *leastSquaresGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var xs, ys []int
	var bVal float64
	var answer string
	var xbar, ybar, sxx, sxy float64

	collision := true
	for iter := 0; iter < 100; iter++ {
		numPoints := 3
		if tier == 2 {
			numPoints = 4
		}
		xStart := rand.Intn(3)
		step := 1
		if tier > 0 && rand.Float64() < 0.4 {
			step = 2
		}
		xs = make([]int, numPoints)
		for i := 0; i < numPoints; i++ {
			xs[i] = xStart + i*step
		}

		slopeBase := 1 + rand.Intn(2+tier*2)
		intercept := rand.Intn(4 + tier*3)
		ys = make([]int, numPoints)
		for i := 0; i < numPoints; i++ {
			noise := rand.Intn(3) - 1 // -1, 0 or 1
			ys[i] = intercept + slopeBase*xs[i] + noise
			if ys[i] < 0 {
				ys[i] = 0
			}
		}

		n := float64(len(xs))
		var sx, sy float64
		for i := range xs {
			sx += float64(xs[i])
			sy += float64(ys[i])
		}
		xbar = sx / n
		ybar = sy / n
		sxx, sxy = 0, 0
		for i := range xs {
			dx := float64(xs[i]) - xbar
			sxx += dx * dx
			sxy += dx * (float64(ys[i]) - ybar)
		}
		if sxx == 0 {
			continue
		}
		bVal = sxy / sxx
		answer = decimal6(bVal)

		// Avoid read-off: answer should not match any x or y
		collision = false
		for _, x := range xs {
			if strconv.Itoa(x) == answer {
				collision = true
				break
			}
		}
		for _, y := range ys {
			if strconv.Itoa(y) == answer {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	if collision {
		xs = []int{1, 3, 5}
		ys = []int{3, 6, 12}
		xbar, ybar, sxx, sxy = 3, 7, 8, 18
		bVal = 2.25
		answer = "2.25"
	}

	xsStrs := make([]string, len(xs))
	ysStrs := make([]string, len(ys))
	for i := range xs {
		xsStrs[i] = strconv.Itoa(xs[i])
		ysStrs[i] = strconv.Itoa(ys[i])
	}

	subgoal1 := "Subgoal 1: State governing formula. The slope b of the least-squares regression line y = a + b x is given by b = s_xy / s_xx, where s_xy = \\sum (x_i - \\bar{x})(y_i - \\bar{y}) and s_xx = \\sum (x_i - \\bar{x})^2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For the given points, the sample means are \\bar{x} = %s and \\bar{y} = %s. The centred cross-sum is s_xy = %s and the centred sum of squares is s_xx = %s.",
		decimal6(xbar), decimal6(ybar), decimal6(sxy), decimal6(sxx))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing yields slope b = %s / %s = %s.",
		decimal6(sxy), decimal6(sxx), answer)

	return generator.Problem{
		Question: "The least-squares line y = a + b x through the points with x values " + strings.Join(xsStrs, ", ") +
			" and y values " + strings.Join(ysStrs, ", ") + " has slope b equal to the centred sum of x times y divided by the centred sum of x squared. What is b? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"xs": strings.Join(xsStrs, ","), "ys": strings.Join(ysStrs, ","), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// The normal equations (numeric)
// ---------------------------------------------------------------------------

type normalEquationsGen struct{}

func (g *normalEquationsGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var xs, ys []int
	var aVal, bVal float64
	var answer string
	var xbar, ybar, sxx, sxy float64

	collision := true
	for iter := 0; iter < 100; iter++ {
		numPoints := 3
		if tier == 2 {
			numPoints = 4
		}
		xStart := rand.Intn(3)
		step := 1
		if tier > 0 && rand.Float64() < 0.4 {
			step = 2
		}
		xs = make([]int, numPoints)
		for i := 0; i < numPoints; i++ {
			xs[i] = xStart + i*step
		}

		slopeBase := 1 + rand.Intn(2+tier*2)
		intercept := 1 + rand.Intn(3+tier*3)
		ys = make([]int, numPoints)
		for i := 0; i < numPoints; i++ {
			noise := rand.Intn(3) - 1
			ys[i] = intercept + slopeBase*xs[i] + noise
			if ys[i] < 0 {
				ys[i] = 0
			}
		}

		n := float64(len(xs))
		var sx, sy float64
		for i := range xs {
			sx += float64(xs[i])
			sy += float64(ys[i])
		}
		xbar = sx / n
		ybar = sy / n
		sxx, sxy = 0, 0
		for i := range xs {
			dx := float64(xs[i]) - xbar
			sxx += dx * dx
			sxy += dx * (float64(ys[i]) - ybar)
		}
		if sxx == 0 {
			continue
		}
		bVal = sxy / sxx
		aVal = ybar - bVal*xbar
		answer = decimal6(aVal)

		// Avoid read-off: answer should not match any x or y
		collision = false
		for _, x := range xs {
			if strconv.Itoa(x) == answer {
				collision = true
				break
			}
		}
		for _, y := range ys {
			if strconv.Itoa(y) == answer {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	if collision {
		xs = []int{1, 3, 5}
		ys = []int{3, 6, 12}
		xbar, ybar, sxx, sxy = 3, 7, 8, 18
		bVal = 2.25
		aVal = 0.25
		answer = "0.25"
	}

	xsStrs := make([]string, len(xs))
	ysStrs := make([]string, len(ys))
	for i := range xs {
		xsStrs[i] = strconv.Itoa(xs[i])
		ysStrs[i] = strconv.Itoa(ys[i])
	}

	subgoal1 := "Subgoal 1: State governing formula. The normal equations for least squares determine that the regression line passes through the data centroid, giving intercept a = \\bar{y} - b \\bar{x}, where b = s_xy / s_xx."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For the sample data, sample means are \\bar{x} = %s and \\bar{y} = %s with slope b = %s.",
		decimal6(xbar), decimal6(ybar), decimal6(bVal))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the intercept gives a = %s - (%s * %s) = %s.",
		decimal6(ybar), decimal6(bVal), decimal6(xbar), answer)

	return generator.Problem{
		Question: "The normal equations for the least-squares line y = a + b x through the points with x values " + strings.Join(xsStrs, ", ") +
			" and y values " + strings.Join(ysStrs, ", ") + " give the intercept a = mean(y) − b·mean(x). What is a? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"xs": strings.Join(xsStrs, ","), "ys": strings.Join(ysStrs, ","), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// Orthogonal polynomials (multiple choice)
// ---------------------------------------------------------------------------

type orthogonalGen struct{}

func (g *orthogonalGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Least squares is easier with orthogonal basis functions because the normal equations...",
			correct:     "become diagonal, so each coefficient can be computed on its own",
			distractors: []string{"become singular", "need no data", "give the exact fit with one term"},
			reason:      "Orthogonality kills the off-diagonal inner products that couple the unknowns.",
		},
		{
			question:    "With an orthogonal basis, the coefficient of each basis function is...",
			correct:     "the inner product of the data with that function divided by the squared norm of the function",
			distractors: []string{"the sum of the residuals", "the derivative of the function", "the largest data value"},
			reason:      "The diagonal normal equation gives exactly that quotient.",
		},
		{
			question:    "Legendre polynomials are orthogonal with respect to...",
			correct:     "the unweighted integral on the interval from minus one to one",
			distractors: []string{"a discrete set of points only", "the Gaussian weight", "the unit circle"},
			reason:      "Their weight function is the constant one on the standard interval.",
		},
		{
			question:    "Chebyshev polynomials are orthogonal with respect to the weight...",
			correct:     "one over the square root of one minus x squared",
			distractors: []string{"one over one plus x squared", "a constant weight", "the exponential weight"},
			reason:      "The reciprocal square-root weight is what makes the cosine substitution an isometry.",
		},
		{
			question:    "Using the successive powers of x as a basis is a poor choice because...",
			correct:     "the powers are highly correlated, so the normal equations are ill conditioned",
			distractors: []string{"they are not functions", "they cannot be evaluated", "they have no coefficients"},
			reason:      "On a short interval the monomials look alike, which is exactly the ill conditioning orthogonal bases avoid.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Extrapolation (multiple choice)
// ---------------------------------------------------------------------------

type extrapolationGen struct{}

func (g *extrapolationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Extrapolation means estimating the function...",
			correct:     "outside the range covered by the data",
			distractors: []string{"between the data points", "by fitting a line", "by removing noise"},
			reason:      "It is interpolation taken beyond the last node, where nothing constrains the fit.",
		},
		{
			question:    "Extrapolation is risky because...",
			correct:     "the interpolant is unconstrained beyond the data, so its error can grow quickly",
			distractors: []string{"the data are always wrong", "the interpolant does not exist outside the nodes", "there are never enough points"},
			reason:      "The node product grows fast outside the interval, and the derivative term is unbounded.",
		},
		{
			question:    "The farther the extrapolation point is from the data, the...",
			correct:     "larger the error bound generally becomes",
			distractors: []string{"smaller the error becomes", "more exact the answer becomes", "less the derivatives matter"},
			reason:      "The error bound is proportional to those very distances.",
		},
		{
			question:    "Richardson extrapolation is different because it...",
			correct:     "combines a numerical sequence using its known error expansion, rather than continuing a fitted curve",
			distractors: []string{"fits a polynomial beyond the data", "ignores the error terms", "extrapolates at random"},
			reason:      "It extrapolates the limit of a sequence, not a function off the end of its data.",
		},
		{
			question:    "When asked for a value far outside the data, a careful response is to...",
			correct:     "report the estimate with a much wider uncertainty",
			distractors: []string{"report it as exact", "refuse under all circumstances", "quietly double the data"},
			reason:      "The prediction may still be useful, but its uncertainty grows with the distance.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// The trapezoid rule (numeric)
// ---------------------------------------------------------------------------

type trapezoidGen struct{}

func (g *trapezoidGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b int
	var answer string
	var est float64

	for iter := 0; iter < 10; iter++ {
		a = rand.Intn(3 + tier*2)
		width := 1 + rand.Intn(3+tier*2)
		b = a + width

		fA := a * a
		fB := b * b
		est = float64(b-a) * float64(fA+fB) / 2.0
		answer = decimal6(est)

		if answer != strconv.Itoa(a) && answer != strconv.Itoa(b) {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The single-interval trapezoid rule approximates the definite integral by T = (b - a)(f(a) + f(b)) / 2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For f(x) = x² on interval [%d, %d], the width is b - a = %d, with f(%d) = %d and f(%d) = %d.",
		a, b, b-a, a, a*a, b, b*b)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the trapezoid sum gives (%d)(%d + %d) / 2 = %s.",
		b-a, a*a, b*b, answer)

	return generator.Problem{
		Question: "The trapezoid rule estimates the integral of f over the interval from a to b by (b − a)(f(a) + f(b))/2. For f(x) = x², a = " +
			strconv.Itoa(a) + " and b = " + strconv.Itoa(b) + ", what is the estimate? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// The midpoint rule (numeric)
// ---------------------------------------------------------------------------

type midpointGen struct{}

func (g *midpointGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b int
	var m float64
	var answer string
	var est float64

	for iter := 0; iter < 10; iter++ {
		a = rand.Intn(3 + tier*2)
		width := 1 + rand.Intn(3+tier*2)
		b = a + width
		if a == 0 && b == 2 {
			b = 3
		}

		m = float64(a+b) / 2.0
		est = float64(b-a) * m * m
		answer = decimal6(est)

		if answer != strconv.Itoa(a) && answer != strconv.Itoa(b) {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The midpoint rule approximates the integral of f over [a, b] by M = (b - a) * f((a + b)/2)."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For f(x) = x² on [%d, %d], the width is b - a = %d and the midpoint is m = %s with f(m) = %s.",
		a, b, b-a, decimal6(m), decimal6(m*m))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the midpoint product gives (%d) * %s = %s.",
		b-a, decimal6(m*m), answer)

	return generator.Problem{
		Question: "The midpoint rule estimates the integral of f over the interval from a to b by (b − a)·f((a + b)/2). For f(x) = x², a = " +
			strconv.Itoa(a) + " and b = " + strconv.Itoa(b) + ", what is the estimate? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// Simpson's rule (numeric)
// ---------------------------------------------------------------------------

type simpsonGen struct{}

func (g *simpsonGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b int
	var m float64
	var answer string
	var est float64

	for iter := 0; iter < 10; iter++ {
		a = rand.Intn(3 + tier*2)
		width := 1 + rand.Intn(3+tier*2)
		b = a + width

		m = float64(a+b) / 2.0
		fA := float64(a * a)
		fB := float64(b * b)
		fM := m * m
		est = (float64(b-a) / 6.0) * (fA + 4.0*fM + fB)
		answer = decimal6(est)

		if answer != strconv.Itoa(a) && answer != strconv.Itoa(b) {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. Simpson's rule approximates the integral of f over [a, b] by S = ((b - a)/6) * (f(a) + 4 f(mid) + f(b))."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For f(x) = x² on [%d, %d], the width is %d, midpoint is %s, f(%d) = %d, f(%d) = %d and f(mid) = %s.",
		a, b, b-a, decimal6(m), a, a*a, b, b*b, decimal6(m*m))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating gives (%d/6) * (%d + 4 * %s + %d) = %s.",
		b-a, a*a, decimal6(m*m), b*b, answer)

	return generator.Problem{
		Question: "Simpson's rule estimates the integral of f over the interval from a to b by (b − a)(f(a) + 4 f(mid) + f(b))/6. For f(x) = x², a = " +
			strconv.Itoa(a) + " and b = " + strconv.Itoa(b) + ", what is the estimate? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// Composite rules (numeric)
// ---------------------------------------------------------------------------

type compositeGen struct{}

func (g *compositeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var b, n int
	var answer string
	var h float64

	for iter := 0; iter < 10; iter++ {
		n = 2 + rand.Intn(2+tier*2) // 2 to 6
		b = 1 + rand.Intn(2+tier*3) // 1 to 8

		h = float64(b) / float64(n)
		sum := 0.0
		for i := 0; i <= n; i++ {
			x := float64(i) * h
			w := 1.0
			if i == 0 || i == n {
				w = 0.5
			}
			sum += w * x * x
		}
		est := h * sum
		answer = decimal6(est)

		if answer != strconv.Itoa(b) && answer != strconv.Itoa(n) {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The composite trapezoid rule on [0, b] with n equal subintervals is T_n = h (f_0/2 + f_1 + ... + f_{n-1} + f_n/2), where h = b/n."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For f(x) = x² on [0, %d] with n = %d subintervals, the width is h = %s and endpoints carry half-weight.",
		b, n, decimal6(h))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Summing the weighted function values and multiplying by h yields T_n = %s.",
		answer)

	return generator.Problem{
		Question: "The composite trapezoid rule with n equal subintervals of width h on the interval from 0 to b is h(f_0/2 + f_1 + ... + f_{n−1} + f_n/2). For f(x) = x², b = " +
			strconv.Itoa(b) + " and n = " + strconv.Itoa(n) + ", what is the estimate? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"b": strconv.Itoa(b), "n": strconv.Itoa(n), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// Error bounds for quadrature (multiple choice)
// ---------------------------------------------------------------------------

type quadErrorGen struct{}

func (g *quadErrorGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "The trapezoid rule error on a single interval is proportional to...",
			correct:     "the cube of the width times the second derivative",
			distractors: []string{"the square of the width", "the width itself", "the fourth power of the width"},
			reason:      "Because the rule is exact for straight lines, the leading error term is the second derivative times the cube of the width.",
		},
		{
			question:    "Simpson's rule is exact for polynomials up to degree...",
			correct:     "three",
			distractors: []string{"one", "two", "four"},
			reason:      "Although it is built from a parabola, the symmetry cancels the cubic error term.",
		},
		{
			question:    "Simpson's rule error is proportional to the fourth power of the width times...",
			correct:     "the fourth derivative",
			distractors: []string{"the second derivative", "the function value", "the first derivative"},
			reason:      "The first nonzero error term needs the fourth derivative.",
		},
		{
			question:    "Halving every subinterval in the composite trapezoid rule reduces the error by about...",
			correct:     "a factor of four",
			distractors: []string{"a factor of two", "a factor of eight", "nothing"},
			reason:      "The global error is proportional to the square of the width.",
		},
		{
			question:    "Doubling the number of subintervals in the composite Simpson rule reduces the error by about...",
			correct:     "a factor of sixteen",
			distractors: []string{"a factor of four", "a factor of two", "nothing"},
			reason:      "The global error is proportional to the fourth power of the width.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Romberg integration (multiple choice)
// ---------------------------------------------------------------------------

type rombergGen struct{}

func (g *rombergGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Romberg integration combines trapezoid estimates to...",
			correct:     "cancel the leading error terms by Richardson extrapolation",
			distractors: []string{"average the estimates", "increase the step size", "fit a straight line"},
			reason:      "Each extrapolation removes the next power of the width from the error expansion.",
		},
		{
			question:    "The first column of the Romberg table holds...",
			correct:     "the composite trapezoid rule at successively halved widths",
			distractors: []string{"the midpoint rule", "the Simpson rule", "the exact integral"},
			reason:      "Everything else in the table is extrapolated from that column.",
		},
		{
			question:    "Each Richardson step raises the order of accuracy by...",
			correct:     "two",
			distractors: []string{"one", "three", "four"},
			reason:      "The trapezoid error expansion runs in even powers of the width.",
		},
		{
			question:    "Romberg integration is most effective when the integrand is...",
			correct:     "smooth, so the error expansion is regular",
			distractors: []string{"discontinuous", "noisy", "constant"},
			reason:      "A jump or a noise floor breaks the expansion that the extrapolation assumes.",
		},
		{
			question:    "The Romberg table is built from...",
			correct:     "trapezoid estimates at compatible, nested step sizes",
			distractors: []string{"Gaussian nodes", "random samples", "the derivative of the integrand"},
			reason:      "Nested widths are what let one column be combined into the next.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Gaussian quadrature (numeric)
// ---------------------------------------------------------------------------

type gaussGen struct{}

func (g *gaussGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	k := 1 + rand.Intn(4+tier*2) // 1 to 8

	est := 2.0 * math.Pow(1.0/3.0, float64(k))
	answer := decimal6(est)

	subgoal1 := "Subgoal 1: State governing formula. Two-point Gauss quadrature on [-1, 1] uses symmetric nodes at \\pm 1/\\sqrt{3} with unit weights, estimating the integral of f by f(-1/\\sqrt{3}) + f(1/\\sqrt{3})."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For f(x) = x^{2k} with k = %d (degree 2k = %d), each node contributes (\\pm 1/\\sqrt{3})^{%d} = (1/3)^%d = %s.",
		k, 2*k, 2*k, k, decimal6(math.Pow(1.0/3.0, float64(k))))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Adding both node contributions gives 2 * (1/3)^%d = %s.",
		k, answer)

	return generator.Problem{
		Question: "Two-point Gauss quadrature on the interval from minus one to one uses the nodes at plus and minus one over the square root of three, with equal weights, so it estimates the integral of f by f(node) + f(−node). For f(x) = x raised to the power 2k with k = " +
			strconv.Itoa(k) + " (power " + strconv.Itoa(2*k) + "), what is the estimate? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"k": strconv.Itoa(k), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// Adaptive quadrature (multiple choice)
// ---------------------------------------------------------------------------

type adaptiveGen struct{}

func (g *adaptiveGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Adaptive quadrature refines the mesh...",
			correct:     "where the estimated error is large",
			distractors: []string{"everywhere uniformly", "at random", "where the function is smallest"},
			reason:      "Effort is spent only where it reduces the error.",
		},
		{
			question:    "The main benefit of adaptivity is...",
			correct:     "accuracy near sharp features without wasting work on the smooth parts",
			distractors: []string{"a fixed number of samples", "simpler code", "exact answers"},
			reason:      "It concentrates subintervals where the integrand is hard.",
		},
		{
			question:    "Error estimates for adaptivity often come from...",
			correct:     "comparing two rules of different orders on the same subinterval",
			distractors: []string{"the function value alone", "a random sample", "the sign of the derivative"},
			reason:      "The difference between the two estimates is a practical bound on the error.",
		},
		{
			question:    "Adaptive methods are well suited to integrands with...",
			correct:     "localised peaks or endpoint singularities",
			distractors: []string{"constant values", "low-degree polynomial form", "linear shape"},
			reason:      "Those are exactly the cases where a uniform mesh is wasteful or inaccurate.",
		},
		{
			question:    "A recursive adaptive scheme subdivides a subinterval when...",
			correct:     "its local error estimate exceeds its share of the tolerance",
			distractors: []string{"the function is positive", "the interval is large", "the derivative vanishes"},
			reason:      "The tolerance is divided among the pieces, and only the pieces that miss it are split.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Integrating to infinity (multiple choice)
// ---------------------------------------------------------------------------

type improperGen struct{}

func (g *improperGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "To integrate over an infinite interval numerically, a standard approach is...",
			correct:     "a change of variable that maps the infinite interval to a finite one",
			distractors: []string{"to truncate and ignore the tail", "to sample at infinity", "to use the derivative"},
			reason:      "A finite interval can then be handled by an ordinary rule.",
		},
		{
			question:    "A Gauss-type rule for a half-infinite interval uses...",
			correct:     "Laguerre or Hermite weights that decay in the tail",
			distractors: []string{"equal weights", "Chebyshev weights", "no weights"},
			reason:      "The weights absorb the exponential decay of the integrand.",
		},
		{
			question:    "The danger of naive truncation of an infinite integral is...",
			correct:     "the neglected tail may still carry substantial mass",
			distractors: []string{"the integrand is undefined", "the interval is not finite", "the nodes repeat"},
			reason:      "A slowly decaying integrand can leave a large, silent error beyond the cut.",
		},
		{
			question:    "An endpoint singularity is often handled by...",
			correct:     "a substitution that removes the singularity",
			distractors: []string{"ignoring the endpoint", "halving the interval forever", "extrapolating at random"},
			reason:      "Once the integrand is bounded, an ordinary rule applies.",
		},
		{
			question:    "Improper integrals of the second kind have...",
			correct:     "an integrand that becomes unbounded inside or at the edge of the interval",
			distractors: []string{"an infinite interval only", "a bounded integrand", "no integral"},
			reason:      "The difficulty is the unbounded integrand, not the length of the interval.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// Gaussian elimination (numeric)
// ---------------------------------------------------------------------------

type eliminationGen struct{}

func (g *eliminationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, c, d, e, f int
	var answer string

	for iter := 0; iter < 25; iter++ {
		a = 1 + rand.Intn(3+tier*2)
		b = 1 + rand.Intn(3+tier*2)
		c = 1 + rand.Intn(3+tier*2)
		d = 1 + rand.Intn(3+tier*2)
		det := a*d - b*c
		if det == 0 {
			continue
		}

		x := 2 + rand.Intn(4+tier*3)
		y := 5 + rand.Intn(4+tier*3)
		if rand.Float64() < 0.3 && tier > 0 {
			x = -x
		}
		if rand.Float64() < 0.3 && tier > 0 {
			y = -y
		}
		for y == a || y == b || y == c || y == d {
			y += 2
		}

		e = a*x + b*y
		f = c*x + d*y
		if y == e || y == f {
			continue
		}

		yVal := float64(a*f-c*e) / float64(det)
		answer = decimal6(yVal)

		// Read-off avoidance: answer must not equal a, b, c, d, e or f
		collision := false
		for _, v := range []int{a, b, c, d, e, f} {
			if answer == strconv.Itoa(v) {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	det := a*d - b*c
	num := a*f - c*e

	subgoal1 := "Subgoal 1: State governing formula. Gaussian elimination eliminates x by subtracting (c/a) times the first row from the second, giving y = (a f - c e) / (a d - b c)."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For a = %d, b = %d, c = %d, d = %d, e = %d and f = %d, the numerator is a f - c e = (%d)(%d) - (%d)(%d) = %d and the pivot determinant is a d - b c = (%d)(%d) - (%d)(%d) = %d.",
		a, b, c, d, e, f, a, f, c, e, num, a, d, b, c, det)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing gives y = %d / %d = %s.",
		num, det, answer)

	return generator.Problem{
		Question: "Solve the linear system a x + b y = e and c x + d y = f by Gaussian elimination. For a = " +
			strconv.Itoa(a) + ", b = " + strconv.Itoa(b) + ", c = " + strconv.Itoa(c) + ", d = " + strconv.Itoa(d) +
			", e = " + strconv.Itoa(e) + " and f = " + strconv.Itoa(f) + ", what is y? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "c": strconv.Itoa(c), "d": strconv.Itoa(d),
			"e": strconv.Itoa(e), "f": strconv.Itoa(f), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// Partial pivoting (multiple choice)
// ---------------------------------------------------------------------------

type pivotingGen struct{}

func (g *pivotingGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Partial pivoting chooses the pivot as...",
			correct:     "the largest available entry in the current column",
			distractors: []string{"the smallest entry", "the diagonal entry always", "a random entry"},
			reason:      "A large pivot keeps the multipliers small and the elimination stable.",
		},
		{
			question:    "Partial pivoting improves...",
			correct:     "numerical stability by avoiding small pivots",
			distractors: []string{"the speed only", "the memory use", "the exactness of the arithmetic"},
			reason:      "Dividing by a tiny pivot would magnify rounding error.",
		},
		{
			question:    "Meeting a zero pivot during elimination...",
			correct:     "forces a row interchange, unless the matrix is singular at that stage",
			distractors: []string{"is impossible", "is harmless", "means the system has no solution"},
			reason:      "A zero in the pivot position needs a swap to bring a nonzero entry up.",
		},
		{
			question:    "Complete pivoting searches for the largest entry...",
			correct:     "in the whole remaining submatrix",
			distractors: []string{"in the first row only", "on the diagonal only", "in the last column only"},
			reason:      "It also allows column swaps, at more bookkeeping and with little extra practical benefit.",
		},
		{
			question:    "Partial pivoting usually keeps the growth factor...",
			correct:     "small, though matrices can be built to defeat it",
			distractors: []string{"always exactly one", "unbounded in practice", "negative"},
			reason:      "Worst-case growth is possible, but it is rare in practice.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// LU factorization (numeric)
// ---------------------------------------------------------------------------

type luGen struct{}

func (g *luGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, c, d int
	var answer string
	var m float64

	for iter := 0; iter < 15; iter++ {
		a = 2 + rand.Intn(3+tier*2) // 2 to 6
		b = 1 + rand.Intn(3)        // 1 to 3
		d = 1 + rand.Intn(3)        // 1 to 3

		mult := 7 + rand.Intn(6+tier*5) // 7 to 17
		rem := 0
		if tier > 0 && rand.Float64() < 0.5 {
			rem = 1 + rand.Intn(a-1)
		}
		c = a*mult + rem

		m = float64(c) / float64(a)
		answer = decimal6(m)

		// Read-off avoidance: answer must not match a, b, c or d
		collision := false
		for _, v := range []int{a, b, c, d} {
			if answer == strconv.Itoa(v) {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. In the LU factorization A = L U of a 2x2 matrix without pivoting, the elimination multiplier in L is l_21 = a_21 / a_11 = c / a."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. From the given matrix, the pivot entry is a_11 = a = %d and the entry below the pivot is a_21 = c = %d.",
		a, c)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing gives multiplier l_21 = %d / %d = %s.",
		c, a, answer)

	return generator.Problem{
		Question: "For the matrix whose first row is a and b and second row is c and d, the elimination multiplier in the LU factorization is c divided by a. For a = " +
			strconv.Itoa(a) + ", b = " + strconv.Itoa(b) + ", c = " + strconv.Itoa(c) + " and d = " + strconv.Itoa(d) +
			", what is the multiplier? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "c": strconv.Itoa(c), "d": strconv.Itoa(d), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// The Jacobi method (numeric)
// ---------------------------------------------------------------------------

type jacobiGen struct{}

func (g *jacobiGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, e, y0 int
	var answer string
	var nextX float64

	for iter := 0; iter < 15; iter++ {
		a = 2 + rand.Intn(3)  // 2 to 4
		b = 1 + rand.Intn(3)  // 1 to 3
		y0 = 1 + rand.Intn(3) // 1 to 3

		targetX := 7 + rand.Intn(6+tier*5) // 7 to 17
		rem := 0
		if tier > 0 && rand.Float64() < 0.5 {
			rem = 1 + rand.Intn(a-1)
		}
		e = a*targetX + b*y0 + rem

		nextX = float64(e-b*y0) / float64(a)
		answer = decimal6(nextX)

		// Read-off avoidance: answer must not match a, b, e or y0
		collision := false
		for _, v := range []int{a, b, e, y0} {
			if answer == strconv.Itoa(v) {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. One step of the Jacobi method solves row 1 for x using the iterate value from the previous step: x^{(k+1)} = (e - b y^{(k)}) / a."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For equation %d x + %d y = %d with previous iterate y = %d, the remaining term is b y = (%d)(%d) = %d.",
		a, b, e, y0, b, y0, b*y0)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating gives x = (%d - %d) / %d = %s.",
		e, b*y0, a, answer)

	return generator.Problem{
		Question: "One step of the Jacobi iteration solves each equation for its own unknown using the previous values. For the equation " +
			strconv.Itoa(a) + " x + " + strconv.Itoa(b) + " y = " + strconv.Itoa(e) + " with the previous value y = " +
			strconv.Itoa(y0) + ", what is the new x? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "e": strconv.Itoa(e),
			"y0": strconv.Itoa(y0), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// The Gauss-Seidel method (multiple choice)
// ---------------------------------------------------------------------------

type gaussSeidelGen struct{}

func (g *gaussSeidelGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Gauss-Seidel differs from Jacobi in that it...",
			correct:     "uses each updated value as soon as it is available",
			distractors: []string{"uses the previous values throughout", "needs a better guess", "does not iterate"},
			reason:      "The fresher values feed immediately into the later equations of the same sweep.",
		},
		{
			question:    "For a diagonally dominant matrix, Gauss-Seidel...",
			correct:     "converges, often about twice as fast as Jacobi",
			distractors: []string{"diverges", "converges more slowly than Jacobi always", "converges in one step"},
			reason:      "Diagonal dominance gives convergence, and the immediate reuse of values speeds it up.",
		},
		{
			question:    "Both Jacobi and Gauss-Seidel are guaranteed to converge when the matrix is...",
			correct:     "strictly diagonally dominant",
			distractors: []string{"singular", "symmetric only", "triangular only"},
			reason:      "The dominance of the diagonal controls the iteration matrix.",
		},
		{
			question:    "Gauss-Seidel can be written as a splitting in which the iteration matrix...",
			correct:     "involves the lower triangle of the matrix",
			distractors: []string{"involves only the diagonal", "involves only the upper triangle", "is the identity"},
			reason:      "Solving forward within a sweep uses the lower triangle already updated in that sweep.",
		},
		{
			question:    "A practical advantage of Gauss-Seidel is...",
			correct:     "it usually needs fewer iterations than Jacobi for the same problem",
			distractors: []string{"it needs no diagonal dominance", "it is a direct method", "it works on any matrix"},
			reason:      "The faster information flow across the mesh is the usual reason it is preferred.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// The power method (numeric)
// ---------------------------------------------------------------------------

type powerMethodGen struct{}

func (g *powerMethodGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, d int
	var answer string
	var mid, rad, lambda1 float64

	for iter := 0; iter < 10; iter++ {
		if tier == 0 {
			a = 2 + rand.Intn(4)
			d = a
			b = 1 + rand.Intn(3)
		} else if tier == 1 {
			type pyth struct {
				delta, b int
			}
			triples := []pyth{{3, 4}, {4, 3}, {6, 8}, {8, 6}}
			tr := triples[rand.Intn(len(triples))]
			m := 1 + rand.Intn(5)
			a = m + tr.delta
			d = m - tr.delta
			if d <= 0 {
				d = 1
				a = d + 2*tr.delta
			}
			b = tr.b
		} else {
			a = 1 + rand.Intn(6+tier*2)
			d = 1 + rand.Intn(6+tier*2)
			b = 1 + rand.Intn(4+tier*2)
		}

		mid = float64(a+d) / 2.0
		diff := float64(a-d) / 2.0
		rad = math.Sqrt(diff*diff + float64(b*b))
		lambda1 = mid + rad
		answer = decimal6(lambda1)

		// Read-off avoidance: answer must not match a, b or d
		collision := false
		for _, v := range []int{a, b, d} {
			if answer == strconv.Itoa(v) {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The power method converges to the dominant eigenvalue of a matrix; for a 2x2 symmetric matrix, the larger eigenvalue is \\lambda_1 = (a + d)/2 + \\sqrt{((a - d)/2)^2 + b^2}."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For a = %d, d = %d and b = %d, the diagonal midpoint is (a + d)/2 = %s and the radius term is \\sqrt{((%d - %d)/2)^2 + %d^2} = \\sqrt{%s + %d} = %s.",
		a, d, b, decimal6(mid), a, d, b, decimal6(math.Pow(float64(a-d)/2.0, 2)), b*b, decimal6(rad))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Summing the center and radius yields dominant eigenvalue \\lambda_1 = %s + %s = %s.",
		decimal6(mid), decimal6(rad), answer)

	return generator.Problem{
		Question: "The power method converges to the dominant eigenvalue of a matrix. For the symmetric matrix with diagonal entries a = " +
			strconv.Itoa(a) + " and d = " + strconv.Itoa(d) + " and off-diagonal entry b = " + strconv.Itoa(b) +
			", the dominant eigenvalue is (a + d)/2 + sqrt(((a − d)/2)² + b²). What is it? (enter a decimal)",
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "d": strconv.Itoa(d), "answer": answer,
		},
	}
}

// ---------------------------------------------------------------------------
// Least squares by QR (multiple choice)
// ---------------------------------------------------------------------------

type leastSquaresQRGen struct{}

func (g *leastSquaresQRGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Solving least squares by QR avoids...",
			correct:     "forming the normal equations and squaring the condition number",
			distractors: []string{"using any data", "the need for a matrix", "the residual"},
			reason:      "Orthogonal transformations preserve the two-norm, so the problem is not made worse conditioned.",
		},
		{
			question:    "In least squares by QR, the coefficients solve...",
			correct:     "an upper-triangular system from the first columns of R",
			distractors: []string{"the normal equations only", "a diagonal system", "no system at all"},
			reason:      "The orthogonal factor drops out of the normal equations, leaving a triangular system.",
		},
		{
			question:    "Compared with QR, the normal equations square...",
			correct:     "the condition number of the design matrix",
			distractors: []string{"the number of unknowns", "the residual norm", "the data values"},
			reason:      "That squaring is why the normal equations lose accuracy on ill-conditioned problems.",
		},
		{
			question:    "QR is preferred over the normal equations when...",
			correct:     "the design matrix is ill conditioned",
			distractors: []string{"the matrix is square and well conditioned", "there is no data", "the system is underdetermined"},
			reason:      "The more sensitive the problem, the more the squaring of the condition number hurts.",
		},
		{
			question:    "Householder reflections are used in QR because they...",
			correct:     "zero out entries stably using orthogonal transformations",
			distractors: []string{"scale the rows", "invert the matrix", "randomise the data"},
			reason:      "An orthogonal reflection does not change lengths, so it does not amplify rounding error.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}
