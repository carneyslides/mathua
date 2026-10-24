package trigonometry

// The raise that takes the domain from forty-eight concepts to the sixty-eight
// its target names. The existing surface taught the definitions, the special
// angles, the sum and difference formulas, the laws of sines and cosines and a
// first pass at equations and graphs. This batch fills the edges: periodicity
// and the two applications of angular measure, the product-to-sum and
// sum-to-product forms, the triple angle and the power-reduction identities,
// the variants of the Pythagorean identity, general solutions and linear
// equations, inequalities beyond the first quadrant, the tangent graph and the
// transformed, inverse and reciprocal graphs, the hyperbolic identities, their
// inverses and derivatives, and polar coordinates and vector projection.
//
// Nine concepts are numeric and eleven are multiple choice. Every numeric
// answer is an integer or an exact reduced fraction by construction: the period
// divides 360, the angular speed divides 360 times a whole number of turns, a
// bearing is a whole number of degrees, the triple angle and the power reduction
// are exact rationals from a Pythagorean triple, the hyperbolic inverse is the
// third side of a triple, the polar abscissa is a whole radius times a rational
// cosine, and the scalar projection is a dot product divided by a whole length.

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
)

func registerRaise(reg *generator.Registry) {
	// Angle measure and its applications.
	reg.Register("trig.basics.periodicity", &periodicityGen{})
	reg.Register("trig.basics.angular_velocity", &angularVelocityGen{})
	reg.Register("trig.basics.bearings", &bearingsGen{})
	// Identities.
	reg.Register("trig.ident.product_to_sum", &productToSumGen{})
	reg.Register("trig.ident.sum_to_product", &sumToProductGen{})
	reg.Register("trig.ident.triple_angle", &tripleAngleGen{})
	reg.Register("trig.ident.power_reduction", &powerReductionGen{})
	reg.Register("trig.ident.pythagorean_variants", &pythagoreanVariantsGen{})
	// Equations and inequalities.
	reg.Register("trig.eq.general_solution", &generalSolutionGen{})
	reg.Register("trig.eq.linear", &linearEqGen{})
	reg.Register("trig.ineq.advanced", &advancedIneqGen{})
	// Graphs.
	reg.Register("trig.graph.tan", &tanGraphGen{})
	reg.Register("trig.graph.transformations", &graphTransformGen{})
	reg.Register("trig.graph.inverse", &inverseGraphGen{})
	reg.Register("trig.graph.reciprocal", &reciprocalGraphGen{})
	// Hyperbolic functions.
	reg.Register("trig.hyperbolic.identities", &hyperbolicIdentGen{})
	reg.Register("trig.hyperbolic.inverse", &hyperbolicInverseGen{})
	reg.Register("trig.hyperbolic.derivative", &hyperbolicDerivGen{})
	// Advanced applications.
	reg.Register("trig.adv.polar", &polarGen{})
	reg.Register("trig.adv.vector_projection", &projectionGen{})
}

// trigFrac renders a reduced fraction, and a whole number when the denominator
// is one.
func trigFrac(num, den int) string {
	if den < 0 {
		num, den = -num, -den
	}
	d := gcd(num, den)
	num, den = num/d, den/d
	if den == 1 {
		return strconv.Itoa(num)
	}
	return strconv.Itoa(num) + "/" + strconv.Itoa(den)
}

type trigMCClaim struct {
	q           string
	correct     string
	distractors []string
	reason      string
}

func trigPickMC(claims []trigMCClaim) generator.Problem {
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.q, c.correct, c.distractors...)
	p.Explanation = c.reason
	if p.Facts == nil {
		p.Facts = map[string]string{}
	}
	p.Facts["answer"] = c.correct
	return p
}

// ---------------------------------------------------------------------------
// Periodicity (numeric)
// ---------------------------------------------------------------------------

type periodicityGen struct{}

func (g *periodicityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	divisors := []int{2, 3, 4, 5, 6, 8, 9, 10, 12, 15, 18}
	b := divisors[rand.Intn(len(divisors))]
	period := 360 / b
	return generator.Problem{
		Question: fmt.Sprintf("The function \\(f(x) = \\sin(%d x)\\) is periodic. What is its period, in degrees?", b),
		Answer:   strconv.Itoa(period),
		Explanation: fmt.Sprintf("A sine wave repeats every \\(360^{\\circ}\\) before the argument is scaled. Scaling the argument by %d makes the wave complete %d cycles in that span, so one cycle takes \\(360^{\\circ}/%d = %d^{\\circ}\\). Multiplying the input compresses the graph horizontally and shortens the period, which is why the period is a division rather than a multiplication.",
			b, b, b, period),
		Facts: map[string]string{"b": strconv.Itoa(b), "answer": strconv.Itoa(period)},
	}
}

// ---------------------------------------------------------------------------
// Angular velocity (numeric)
// ---------------------------------------------------------------------------

type angularVelocityGen struct{}

func (g *angularVelocityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	times := []int{2, 3, 4, 5, 6, 8, 10, 12}
	t := times[rand.Intn(len(times))]
	revs := 1 + rand.Intn(6)
	for (360*revs)%t != 0 {
		revs = 1 + rand.Intn(6)
	}
	rate := 360 * revs / t
	return generator.Problem{
		Question: fmt.Sprintf("A wheel turns %d complete revolutions in %d seconds. What is its angular speed, in degrees per second?", revs, t),
		Answer:   strconv.Itoa(rate),
		Explanation: fmt.Sprintf("One complete revolution sweeps \\(360^{\\circ}\\), so %d revolutions sweep \\(%d \\times 360^{\\circ} = %d^{\\circ}\\). Dividing that angle by the %d seconds gives \\(%d^{\\circ}\\) per second. Angular speed measures the angle swept per unit time and is independent of the radius, while linear speed along the rim is the angular speed times the radius.",
			revs, revs, 360*revs, t, rate),
		Facts: map[string]string{"revs": strconv.Itoa(revs), "seconds": strconv.Itoa(t), "answer": strconv.Itoa(rate)},
	}
}

// ---------------------------------------------------------------------------
// Bearings (numeric)
// ---------------------------------------------------------------------------

type bearingsGen struct{}

func (g *bearingsGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	theta := 10 * (1 + rand.Intn(7)) // 10 to 70 degrees
	quadrant := rand.Intn(4)
	var description string
	var bearing int
	switch quadrant {
	case 0:
		description = fmt.Sprintf("north %d degrees east", theta)
		bearing = theta
	case 1:
		description = fmt.Sprintf("south %d degrees east", theta)
		bearing = 180 - theta
	case 2:
		description = fmt.Sprintf("south %d degrees west", theta)
		bearing = 180 + theta
	default:
		description = fmt.Sprintf("north %d degrees west", theta)
		bearing = 360 - theta
	}
	return generator.Problem{
		Question: fmt.Sprintf("A ship travels on the heading %s. What is the bearing of that course, in degrees measured clockwise from north?", description),
		Answer:   strconv.Itoa(bearing),
		Explanation: fmt.Sprintf("A bearing is measured clockwise from north, so north is 0 degrees, east is 90, south is 180 and west is 270. The course %s lies %d degrees from the nearer north-south line on that side, which places it at %d degrees. Writing the bearing as a single clockwise angle removes the ambiguity that a compass phrase leaves open.",
			description, theta, bearing),
		Facts: map[string]string{"theta": strconv.Itoa(theta), "quadrant": strconv.Itoa(quadrant), "answer": strconv.Itoa(bearing)},
	}
}

// ---------------------------------------------------------------------------
// Triple angle (numeric)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Power reduction (numeric)
// ---------------------------------------------------------------------------

type powerReductionGen struct{}

func (g *powerReductionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct{ cosNum, cosDen int }
	claims := []claim{
		{7, 25}, {-7, 25}, {1, 4}, {-1, 4}, {3, 5}, {-3, 5}, {1, 2}, {0, 1},
	}
	c := claims[rand.Intn(len(claims))]
	// sin^2 t = (1 - cos 2t)/2 = (den - num)/(2 den)
	answer := trigFrac(c.cosDen-c.cosNum, 2*c.cosDen)
	return generator.Problem{
		Question: fmt.Sprintf("If \\(\\cos(2\\theta) = %s\\), what is \\(\\sin^2\\theta\\)? Give an exact fraction.", trigFrac(c.cosNum, c.cosDen)),
		Answer:   answer,
		Explanation: fmt.Sprintf("The power-reduction identity is \\(\\sin^2\\theta = (1 - \\cos(2\\theta))/2\\). Substituting the given value gives \\((1 - %s)/2 = %s\\). The identity is the double-angle formula solved for the squared sine, and it is what turns an even power of a sine into a first power of a cosine for integration.",
			trigFrac(c.cosNum, c.cosDen), answer),
		Facts: map[string]string{"cos_num": strconv.Itoa(c.cosNum), "cos_den": strconv.Itoa(c.cosDen), "answer": answer},
	}
}

// ---------------------------------------------------------------------------
// Linear equations (numeric)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// The tangent graph (multiple choice)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Graph transformations (multiple choice)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Inverse graphs (multiple choice)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Reciprocal graphs (multiple choice)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Hyperbolic inverses (numeric)
// ---------------------------------------------------------------------------

type hyperbolicInverseGen struct{}

func (g *hyperbolicInverseGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type triple struct{ opp, adj, hyp int }
	triples := []triple{{3, 4, 5}, {5, 12, 13}, {8, 15, 17}, {7, 24, 25}, {20, 21, 29}, {9, 40, 41}}
	tr := triples[rand.Intn(len(triples))]
	answer := trigFrac(tr.hyp, tr.adj)
	return generator.Problem{
		Question: fmt.Sprintf("If \\(\\sinh x = %d/%d\\) and \\(x\\) is positive, what is \\(\\cosh x\\)? Give an exact fraction.", tr.opp, tr.adj),
		Answer:   answer,
		Explanation: fmt.Sprintf("The hyperbolic identity gives \\(\\cosh^2 x = 1 + \\sinh^2 x = 1 + %d/%d = %d/%d\\), so \\(\\cosh x = %s\\). The positive root is taken because the hyperbolic cosine is at least one for every real input, and the triple %d, %d, %d is the arithmetic that makes the root exact.",
			tr.opp*tr.opp, tr.adj*tr.adj, tr.hyp*tr.hyp, tr.adj*tr.adj, answer, tr.opp, tr.adj, tr.hyp),
		Facts: map[string]string{"opp": strconv.Itoa(tr.opp), "adj": strconv.Itoa(tr.adj), "hyp": strconv.Itoa(tr.hyp), "answer": answer},
	}
}

// ---------------------------------------------------------------------------
// Polar coordinates (numeric)
// ---------------------------------------------------------------------------

type polarGen struct{}

func (g *polarGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		theta  int
		cosNum int
		cosDen int
	}
	claims := []claim{
		{0, 1, 1},
		{60, 1, 2},
		{90, 0, 1},
		{120, -1, 2},
		{180, -1, 1},
	}
	c := claims[rand.Intn(len(claims))]
	r := 2 * (1 + rand.Intn(9))
	for r%c.cosDen != 0 {
		r = 2 * (1 + rand.Intn(9))
	}
	x := r * c.cosNum / c.cosDen
	return generator.Problem{
		Question: fmt.Sprintf("A point has polar coordinates \\(r = %d\\) and \\(\\theta = %d^{\\circ}\\). What is its x-coordinate in rectangular coordinates?", r, c.theta),
		Answer:   strconv.Itoa(x),
		Explanation: fmt.Sprintf("The conversion from polar to rectangular is \\(x = r\\cos\\theta\\) and \\(y = r\\sin\\theta\\). Here \\(x = %d \\cdot \\cos(%d^{\\circ}) = %d \\cdot %s = %d\\). The radius is a distance from the origin, so it is never negative, while the angle carries the direction and therefore the sign of each coordinate.",
			r, c.theta, r, trigFrac(c.cosNum, c.cosDen), x),
		Facts: map[string]string{"r": strconv.Itoa(r), "theta": strconv.Itoa(c.theta), "cos_num": strconv.Itoa(c.cosNum), "cos_den": strconv.Itoa(c.cosDen), "answer": strconv.Itoa(x)},
	}
}

// ---------------------------------------------------------------------------
// Vector projection (numeric)
// ---------------------------------------------------------------------------

type projectionGen struct{}

func (g *projectionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct{ u1, u2 int }
	claims := []claim{
		{2, 1}, {1, 3}, {5, 0}, {0, 5}, {3, 4}, {4, 3}, {6, 2}, {-1, 2},
	}
	c := claims[rand.Intn(len(claims))]
	// v = (3, 4) has length 5, so the scalar projection is (3 u1 + 4 u2)/5.
	num := 3*c.u1 + 4*c.u2
	for num%5 != 0 {
		c = claims[rand.Intn(len(claims))]
		num = 3*c.u1 + 4*c.u2
	}
	answer := num / 5
	return generator.Problem{
		Question: fmt.Sprintf("What is the scalar projection of the vector \\(\\mathbf{u} = (%d, %d)\\) onto the vector \\(\\mathbf{v} = (3, 4)\\)?", c.u1, c.u2),
		Answer:   strconv.Itoa(answer),
		Explanation: fmt.Sprintf("The scalar projection is the dot product divided by the length of \\(\\mathbf{v}\\): \\((%d)(3) + (%d)(4) = %d\\), and the length of \\(\\mathbf{v}\\) is \\(\\sqrt{9 + 16} = 5\\), so the projection is \\(%d/5 = %d\\). The sign of the result says whether \\(\\mathbf{u}\\) points along \\(\\mathbf{v}\\) or against it.",
			c.u1, c.u2, num, num, answer),
		Facts: map[string]string{"u1": strconv.Itoa(c.u1), "u2": strconv.Itoa(c.u2), "answer": strconv.Itoa(answer)},
	}
}
