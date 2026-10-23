package calculus

// The vector-calculus theorems and the coordinate changes that make the multiple integrals of
// multivariable.go computable.
//
// These are the concepts that connect an integral over a region to an integral over its boundary
// (Green's theorem), and the two differential operators that decide whether a field is a gradient
// (curl and divergence). Every answer is an exact integer by construction, and every generator
// offers at least three distinct question texts at difficulty 0.3, because the practice feed asks
// for three questions at a time and the bank test gates on the smallest of the difficulties it
// sweeps rather than on the average.
import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"unicode"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerVector(reg *generator.Registry) {
	reg.Register("calc.vec.curl", &curlGen{})
	reg.Register("calc.vec.divergence", &divergenceGen{})
	reg.Register("calc.vec.conservative_test", &conservativeTestGen{})
	reg.Register("calc.vec.greens_theorem", &greensTheoremGen{})
	reg.Register("calc.multi.transform_jacobian", &transformJacobianGen{})
	reg.Register("calc.multi.double_integral_region", &doubleIntegralRegionGen{})
	reg.Register("calc.multi.saddle_test", &saddleTestGen{})
	reg.Register("calc.multi.lagrange", &lagrangeGen{})
}

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

// small returns a random int in [lo, hi], and never panics when hi < lo.
func small(lo, hi int) int {
	if hi < lo {
		hi = lo
	}
	return lo + rand.Intn(hi-lo+1)
}

// curlGen asks for the scalar curl of a linear field, which is the constant dQ/dx - dP/dy.
type curlGen struct{}

func (g *curlGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, curlVal int
	var q, ex string
	for {
		maxBound := 6 + tier*4
		a = small(1, maxBound)
		b = small(1, maxBound)
		if tier > 0 && rand.Intn(2) == 0 {
			a = -a
		}
		if tier > 0 && rand.Intn(2) == 0 {
			b = -b
		}
		curlVal = b - a
		ansStr := strconv.Itoa(curlVal)
		q = fmt.Sprintf("For the field \\(\\mathbf{F} = (%dy,\\ %dx)\\), find the scalar curl \\(\\frac{\\partial Q}{\\partial x} - \\frac{\\partial P}{\\partial y}\\).", a, b)
		if curlVal == a || curlVal == -a || curlVal == b || curlVal == -b || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("The scalar curl of a planar vector field \\(\\mathbf{F} = (P, Q)\\) is defined as \\(\\operatorname{curl} \\mathbf{F} = \\frac{\\partial Q}{\\partial x} - \\frac{\\partial P}{\\partial y}\\). Compute partial derivatives for component functions \\(P(x,y) = %dy\\) and \\(Q(x,y) = %dx\\), yielding \\(\\frac{\\partial P}{\\partial y} = %d\\) and \\(\\frac{\\partial Q}{\\partial x} = %d\\). Substitute into the curl definition to evaluate \\(\\operatorname{curl} \\mathbf{F} = %d - (%d) = %d\\).",
			a, b, a, b, b, a, curlVal)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(curlVal),
		Explanation: ex,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"b":      strconv.Itoa(b),
			"curl":   strconv.Itoa(curlVal),
			"answer": strconv.Itoa(curlVal),
		},
	}
}

// divergenceGen asks for the divergence of a linear field, which is a + b.
type divergenceGen struct{}

func (g *divergenceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, divVal int
	var q, ex string
	for {
		maxBound := 3 + tier*3
		a = small(1, maxBound)
		b = small(1, maxBound)
		if tier > 0 && rand.Intn(2) == 0 {
			a = -a
		}
		if tier > 0 && rand.Intn(2) == 0 {
			b = -b
		}
		divVal = a + b
		ansStr := strconv.Itoa(divVal)
		q = fmt.Sprintf("For the field \\(\\mathbf{F} = (%dx,\\ %dy)\\), find the divergence \\(\\frac{\\partial P}{\\partial x} + \\frac{\\partial Q}{\\partial y}\\).", a, b)
		if divVal == a || divVal == -a || divVal == b || divVal == -b || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("The divergence of a planar vector field \\(\\mathbf{F} = (P, Q)\\) measures net outward flux density: \\(\\operatorname{div} \\mathbf{F} = \\nabla \\cdot \\mathbf{F} = \\frac{\\partial P}{\\partial x} + \\frac{\\partial Q}{\\partial y}\\). Evaluate the partial derivatives of components \\(P(x,y) = %dx\\) and \\(Q(x,y) = %dy\\), giving \\(\\frac{\\partial P}{\\partial x} = %d\\) and \\(\\frac{\\partial Q}{\\partial y} = %d\\). Sum the diagonal derivatives to obtain \\(\\operatorname{div} \\mathbf{F} = %d + (%d) = %d\\).",
			a, b, a, b, a, b, divVal)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(divVal),
		Explanation: ex,
		Facts: map[string]string{
			"a":          strconv.Itoa(a),
			"b":          strconv.Itoa(b),
			"divergence": strconv.Itoa(divVal),
			"answer":     strconv.Itoa(divVal),
		},
	}
}

// conservativeTestGen asks for the cross-partial difference, which is zero exactly for a gradient.
type conservativeTestGen struct{}

func (g *conservativeTestGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var b, c, diffVal int
	var q, ex string
	for {
		maxBound := 6 + tier*4
		b = small(1, maxBound)
		c = small(1, maxBound)
		if tier > 0 && rand.Intn(2) == 0 {
			b = -b
		}
		if tier > 0 && rand.Intn(2) == 0 {
			c = -c
		}
		diffVal = c - b
		ansStr := strconv.Itoa(diffVal)
		q = fmt.Sprintf("For \\(\\mathbf{F} = (2x %+dy,\\ %dx + 5y)\\), compute \\(\\frac{\\partial Q}{\\partial x} - \\frac{\\partial P}{\\partial y}\\). A value of zero means the field is conservative.", b, c)
		if diffVal == b || diffVal == -b || diffVal == c || diffVal == -c || diffVal == 2 || diffVal == 5 || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("A continuously differentiable field \\(\\mathbf{F} = (P, Q)\\) on a simply connected domain is conservative if and only if \\(\\frac{\\partial Q}{\\partial x} - \\frac{\\partial P}{\\partial y} = 0\\). Differentiate components \\(P = 2x %+dy\\) with respect to y and \\(Q = %dx + 5y\\) with respect to x: \\(\\frac{\\partial P}{\\partial y} = %d\\) and \\(\\frac{\\partial Q}{\\partial x} = %d\\). Calculate the cross-partial difference \\(\\frac{\\partial Q}{\\partial x} - \\frac{\\partial P}{\\partial y} = %d - (%d) = %d\\).",
			b, c, b, c, c, b, diffVal)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(diffVal),
		Explanation: ex,
		Facts: map[string]string{
			"b":          strconv.Itoa(b),
			"c":          strconv.Itoa(c),
			"difference": strconv.Itoa(diffVal),
			"answer":     strconv.Itoa(diffVal),
		},
	}
}

// greensTheoremGen uses Green's theorem to turn a boundary integral into an area, which for a
// rectangle is the product of the sides.
type greensTheoremGen struct{}

func (g *greensTheoremGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, areaVal int
	var q, ex string
	for {
		maxBound := 3 + tier*3
		a = small(2, maxBound)
		b = small(2, maxBound+1)
		areaVal = a * b
		ansStr := strconv.Itoa(areaVal)
		q = fmt.Sprintf("Use Green's theorem to evaluate \\(\\oint_C x\\,dy\\) counterclockwise around the rectangle with corners \\((0,0)\\), \\((%d,0)\\), \\((%d,%d)\\) and \\((0,%d)\\).", a, a, b, b)
		if areaVal == a || areaVal == b || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("By Green's theorem in the plane, \\(\\oint_C P\\,dx + Q\\,dy = \\iint_R \\left(\\frac{\\partial Q}{\\partial x} - \\frac{\\partial P}{\\partial y}\\right)\\,dA\\). With \\(P = 0\\) and \\(Q = x\\), the integrand is \\(\\frac{\\partial Q}{\\partial x} - \\frac{\\partial P}{\\partial y} = 1 - 0 = 1\\), reducing the boundary circulation to the enclosed area \\(\\iint_R 1\\,dA\\). Evaluate the area of the rectangle with width %d and height %d to get \\(%d \\times %d = %d\\).",
			a, b, a, b, areaVal)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(areaVal),
		Explanation: ex,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"b":      strconv.Itoa(b),
			"area":   strconv.Itoa(areaVal),
			"answer": strconv.Itoa(areaVal),
		},
	}
}

// transformJacobianGen asks for the determinant of a linear change of variables, which is ad - bc.
type transformJacobianGen struct{}

func (g *transformJacobianGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, c, d, det int
	var q, ex string
	for {
		maxBound := 5 + tier*3
		a = small(1, maxBound)
		b = small(1, maxBound)
		c = small(1, maxBound)
		d = small(1, maxBound)
		det = a*d - b*c
		if det == 0 {
			continue
		}
		ansStr := strconv.Itoa(det)
		q = fmt.Sprintf("For the change of variables \\(x = %du + %dv,\\ y = %du + %dv\\), find \\(\\frac{\\partial(x,y)}{\\partial(u,v)}\\).", a, b, c, d)
		if det == a || det == -a || det == b || det == -b || det == c || det == -c || det == d || det == -d || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("The Jacobian determinant for transformation \\((u, v) \\mapsto (x, y)\\) is given by \\(\\frac{\\partial(x,y)}{\\partial(u,v)} = \\det \\begin{pmatrix} x_u & x_v \\\\ y_u & y_v \\end{pmatrix} = x_u y_v - x_v y_u\\). Compute partial derivatives from \\(x = %du + %dv\\) and \\(y = %du + %dv\\), giving matrix entries \\(x_u = %d\\), \\(x_v = %d\\), \\(y_u = %d\\) and \\(y_v = %d\\). Evaluate the determinant \\((%d)(%d) - (%d)(%d) = %d - %d = %d\\).",
			a, b, c, d, a, b, c, d, a, d, b, c, a*d, b*c, det)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(det),
		Explanation: ex,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"b":      strconv.Itoa(b),
			"c":      strconv.Itoa(c),
			"d":      strconv.Itoa(d),
			"det":    strconv.Itoa(det),
			"answer": strconv.Itoa(det),
		},
	}
}

// doubleIntegralRegionGen integrates 1 over a triangle whose area is a^2/2, with a even so the
// answer is whole.
type doubleIntegralRegionGen struct{}

func (g *doubleIntegralRegionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, areaVal int
	var q, ex string
	for {
		maxBound := 7 + tier*4
		k := small(2, maxBound)
		a = 2 * k
		areaVal = a * a / 2
		ansStr := strconv.Itoa(areaVal)
		q = fmt.Sprintf("Evaluate \\(\\iint_R 1\\,dA\\) where R is the triangle \\(0 \\le y \\le x \\le %d\\).", a)
		if areaVal == a || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("Set up the iterated integral over the triangular region \\(R\\): \\(\\iint_R 1\\,dA = \\int_0^{%d} \\int_0^x 1\\,dy\\,dx\\). Evaluate the inner integral \\(\\int_0^x 1\\,dy = [y]_0^x = x\\), reducing the problem to \\(\\int_0^{%d} x\\,dx\\). Integrate with respect to x: \\(\\left[\\frac{x^2}{2}\\right]_0^{%d} = \\frac{(%d)^2}{2} = \\frac{%d}{2} = %d\\).",
			a, a, a, a, a*a, areaVal)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(areaVal),
		Explanation: ex,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"area":   strconv.Itoa(areaVal),
			"answer": strconv.Itoa(areaVal),
		},
	}
}

// saddleTestGen asks for the second-derivative discriminant of f = a x^2 + b y^2 at its critical
// point, which is 4ab.
type saddleTestGen struct{}

func (g *saddleTestGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, discVal int
	var q, ex string
	for {
		maxBound := 5 + tier*3
		a = small(1, maxBound)
		b = small(1, maxBound)
		if tier > 0 && rand.Intn(2) == 0 {
			b = -b
		}
		discVal = 4 * a * b
		ansStr := strconv.Itoa(discVal)
		q = fmt.Sprintf("For \\(f(x,y) = %dx^2 %+dy^2\\), evaluate the discriminant \\(D = f_{xx}f_{yy} - (f_{xy})^2\\) at the origin.", a, b)
		if discVal == a || discVal == -a || discVal == b || discVal == -b || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("The discriminant in the Second Derivative Test is given by \\(D = f_{xx} f_{yy} - (f_{xy})^2\\). Calculate second-order partial derivatives for \\(f(x,y) = %dx^2 %+dy^2\\): \\(f_{xx} = %d\\), \\(f_{yy} = %d\\) and \\(f_{xy} = 0\\). Evaluate the discriminant product at the origin: \\(D = (%d)(%d) - (0)^2 = %d\\).",
			a, b, 2*a, 2*b, 2*a, 2*b, discVal)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(discVal),
		Explanation: ex,
		Facts: map[string]string{
			"a":      strconv.Itoa(a),
			"b":      strconv.Itoa(b),
			"D":      strconv.Itoa(discVal),
			"answer": strconv.Itoa(discVal),
		},
	}
}

// lagrangeGen maximises a product under a fixed sum, which is (s/2)^2 with s even.
type lagrangeGen struct{}

func (g *lagrangeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var s, halfS, maxVal int
	var q, ex string
	for {
		maxBound := 8 + tier*5
		k := small(3, maxBound)
		s = 2 * k
		halfS = s / 2
		maxVal = halfS * halfS
		ansStr := strconv.Itoa(maxVal)
		q = fmt.Sprintf("Use Lagrange multipliers to find the largest value of \\(xy\\) subject to \\(x + y = %d\\) with \\(x, y > 0\\).", s)
		if maxVal == s || containsWord(q, ansStr) {
			continue
		}
		ex = fmt.Sprintf("Formulate the Lagrange condition \\(\\nabla f = \\lambda \\nabla g\\) for objective \\(f(x,y) = xy\\) and constraint \\(g(x,y) = x + y = %d\\), giving system \\(y = \\lambda\\) and \\(x = \\lambda\\). Equating components yields \\(x = y\\); substitute into the constraint to get \\(2x = %d\\), so \\(x = y = %d\\). Evaluate the maximum product: \\(f(%d, %d) = %d \\times %d = %d\\).",
			s, s, halfS, halfS, halfS, halfS, halfS, maxVal)
		break
	}
	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(maxVal),
		Explanation: ex,
		Facts: map[string]string{
			"s":       strconv.Itoa(s),
			"maximum": strconv.Itoa(maxVal),
			"answer":  strconv.Itoa(maxVal),
		},
	}
}
