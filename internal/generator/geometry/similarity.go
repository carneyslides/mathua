package geometry

import (
	"fmt"
	"math/rand"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerSimilarity(reg *generator.Registry) {
	reg.Register("geo.similar.ratio", &similarRatioGen{})
	reg.Register("geo.similar.missing_side", &similarMissingSideGen{})
	reg.Register("geo.transform.dilation", &dilationGen{})
	reg.Register("geo.polygon.interior_angles", &interiorAnglesGen{})
	reg.Register("geo.polygon.exterior_angle", &exteriorAngleGen{})
	reg.Register("geo.circle.inscribed_angle", &inscribedAngleGen{})
	reg.Register("geo.circle.chord", &chordGen{})
	reg.Register("geo.triangle.special_right", &specialRightGen{})
}

// similarRatioGen asks for a side of the larger of two similar figures given the scale factor.
type similarRatioGen struct{}

func (g *similarRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	k := 2 + rand.Intn(4+int(ctx.Difficulty*6))
	a := 3 + rand.Intn(5+int(ctx.Difficulty*15))
	ans := k * a
	return generator.Problem{
		Question: fmt.Sprintf("Two triangles are similar: the ratio of a side of the larger to the corresponding side of the smaller is %d. If that side of the smaller triangle is %d, how long is the corresponding side of the larger?", k, a),
		Answer:   fmt.Sprintf("%d", ans),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. In similar figures with scale ratio k : 1, corresponding sides satisfy s_large = k * s_small. Subgoal 2: Substitute parameters. Substitute scale ratio k = %d and smaller side s_small = %d: s_large = %d * %d. Subgoal 3: Evaluate and simplify. Multiply to find s_large = %d.",
			k, a, k, a, ans),
		Facts: map[string]string{
			"side":   fmt.Sprintf("%d", ans),
			"factor": fmt.Sprintf("%d", k),
		},
	}
}

// similarMissingSideGen sets the two similar triangles' sides so the unknown is whole.
type similarMissingSideGen struct{}

func (g *similarMissingSideGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	k := 2 + rand.Intn(4+int(ctx.Difficulty*6))
	a := 3 + rand.Intn(5+int(ctx.Difficulty*15))
	b := 4 + rand.Intn(6+int(ctx.Difficulty*15))
	if b == a {
		b = a + 3
	}
	if b == k*a {
		b = k*a + 2
	}
	ans := k * a
	return generator.Problem{
		Question: fmt.Sprintf("Two similar triangles have corresponding sides in the ratio %d to 1. In the smaller triangle, one side is %d and another is %d. What is the length, in the larger triangle, of the side corresponding to the %d?", k, a, b, a),
		Answer:   fmt.Sprintf("%d", ans),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. Corresponding sides of similar triangles are proportional: s_large = k * s_small. Subgoal 2: Substitute parameters. Substitute ratio k = %d and corresponding side s_small = %d (with non-corresponding side %d): s_large = %d * %d. Subgoal 3: Evaluate and simplify. Multiply to find s_large = %d.",
			k, a, b, k, a, ans),
		Facts: map[string]string{
			"side":   fmt.Sprintf("%d", ans),
			"factor": fmt.Sprintf("%d", k),
			"given":  fmt.Sprintf("%d", a),
		},
	}
}

// dilationGen reads a dilation as a scale factor applied to every length.
type dilationGen struct{}

func (g *dilationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	k := 2 + rand.Intn(4+int(ctx.Difficulty*6))
	l := 3 + rand.Intn(5+int(ctx.Difficulty*15))
	if l == k {
		l = k + 1
	}

	if rand.Intn(2) == 0 {
		// Shrink by 1/k
		orig := k * l
		ans := l
		return generator.Problem{
			Question: fmt.Sprintf("A segment of length %d is dilated by a factor of 1/%d about a point. How long is its image?", orig, k),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula. A dilation with scale factor r transforms segment length L into image length L' = r * L. Subgoal 2: Substitute parameters. Substitute original length L = %d and scale factor r = 1/%d: L' = (1/%d) * %d = %d / %d. Subgoal 3: Evaluate and simplify. Divide to find image length L' = %d.",
				orig, k, k, orig, orig, k, ans),
			Facts: map[string]string{
				"length": fmt.Sprintf("%d", ans),
				"factor": fmt.Sprintf("1/%d", k),
			},
		}
	}

	// Enlarge by k
	ans := k * l
	return generator.Problem{
		Question: fmt.Sprintf("A segment of length %d is dilated by a factor of %d about a point. How long is its image?", l, k),
		Answer:   fmt.Sprintf("%d", ans),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. A dilation with scale factor k transforms segment length L into image length L' = k * L. Subgoal 2: Substitute parameters. Substitute original length L = %d and scale factor k = %d: L' = %d * %d. Subgoal 3: Evaluate and simplify. Multiply to find image length L' = %d.",
			l, k, k, l, ans),
		Facts: map[string]string{
			"length": fmt.Sprintf("%d", ans),
			"factor": fmt.Sprintf("%d", k),
		},
	}
}

// interiorAnglesGen asks for the interior-angle sum of an n-gon.
type interiorAnglesGen struct{}

func (g *interiorAnglesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	n := 3 + rand.Intn(6+int(ctx.Difficulty*20))
	sum := 180 * (n - 2)
	return generator.Problem{
		Question: fmt.Sprintf("What is the sum of the interior angles of a polygon with %d sides, in degrees?", n),
		Answer:   fmt.Sprintf("%d", sum),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. The sum of interior angles of an n-sided polygon is S = (n - 2) * 180 degrees. Subgoal 2: Substitute parameters. Substitute side count n = %d: S = (%d - 2) * 180 = %d * 180. Subgoal 3: Evaluate and simplify. Multiply to find S = %d degrees.",
			n, n, n-2, sum),
		Facts: map[string]string{
			"sum":   fmt.Sprintf("%d", sum),
			"sides": fmt.Sprintf("%d", n),
		},
	}
}

// exteriorAngleGen asks for one exterior angle of a regular polygon with whole-number degrees.
type exteriorAngleGen struct{}

func (g *exteriorAngleGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	divisors := [][]int{
		{3, 4, 5, 6, 8, 9, 10, 12},
		{6, 8, 9, 10, 12, 15, 18, 20},
		{12, 15, 18, 20, 24, 30, 36, 40},
	}
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	pool := divisors[tier]
	n := pool[rand.Intn(len(pool))]
	ans := 360 / n
	return generator.Problem{
		Question: fmt.Sprintf("What is the measure, in degrees, of one exterior angle of a regular polygon with %d sides?", n),
		Answer:   fmt.Sprintf("%d", ans),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. The exterior angles of any regular polygon with n sides sum to 360 degrees, so each exterior angle is theta = 360 / n degrees. Subgoal 2: Substitute parameters. Substitute number of sides n = %d: theta = 360 / %d. Subgoal 3: Evaluate and simplify. Divide to find theta = %d degrees.",
			n, n, ans),
		Facts: map[string]string{
			"angle": fmt.Sprintf("%d", ans),
			"sides": fmt.Sprintf("%d", n),
		},
	}
}

// inscribedAngleGen uses the inscribed-angle theorem with even central angles.
type inscribedAngleGen struct{}

func (g *inscribedAngleGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	half := 10 + rand.Intn(15+int(ctx.Difficulty*55))
	central := 2 * half
	ans := half
	return generator.Problem{
		Question: fmt.Sprintf("An inscribed angle in a circle subtends an arc whose central angle is %d degrees. What is the measure of the inscribed angle, in degrees?", central),
		Answer:   fmt.Sprintf("%d", ans),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. By the inscribed angle theorem, an inscribed angle measures half its intercepted central angle: theta_inscribed = (1/2) * theta_central. Subgoal 2: Substitute parameters. Substitute central angle theta_central = %d degrees: theta_inscribed = %d / 2. Subgoal 3: Evaluate and simplify. Divide by 2 to find theta_inscribed = %d degrees.",
			central, central, ans),
		Facts: map[string]string{
			"angle":   fmt.Sprintf("%d", ans),
			"central": fmt.Sprintf("%d", central),
		},
	}
}

// chordGen uses a Pythagorean distance from the centre so half the chord is whole.
type chordGen struct{}

func (g *chordGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	l1, l2, hyp := pythagoreanTriple(ctx.Difficulty)
	distance, halfChord := l1, l2
	if rand.Intn(2) == 0 {
		distance, halfChord = l2, l1
	}
	radius := hyp
	chord := 2 * halfChord
	return generator.Problem{
		Question: fmt.Sprintf("A chord of a circle lies %d units from the centre with a circle radius of %d. How long is the chord?", distance, radius),
		Answer:   fmt.Sprintf("%d", chord),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. The perpendicular from circle center to a chord bisects the chord, forming a right triangle with radius r as hypotenuse, distance d as one leg and half-chord h as the other leg: d^2 + h^2 = r^2, so h = sqrt(r^2 - d^2) and chord length is c = 2h. Subgoal 2: Substitute parameters. Substitute distance d = %d and radius r = %d: h = sqrt(%d^2 - %d^2) = sqrt(%d - %d) = sqrt(%d). Subgoal 3: Evaluate and simplify. Calculate half-chord h = %d, then double it to find chord length c = 2 * %d = %d.",
			distance, radius, radius, distance, radius*radius, distance*distance, radius*radius-distance*distance, halfChord, halfChord, chord),
		Facts: map[string]string{
			"chord":    fmt.Sprintf("%d", chord),
			"radius":   fmt.Sprintf("%d", radius),
			"distance": fmt.Sprintf("%d", distance),
		},
	}
}

// specialRightGen uses the 30-60-90 triangle with hypotenuse twice the short leg.
type specialRightGen struct{}

func (g *specialRightGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	if rand.Intn(2) == 0 {
		// Short leg given, find hypotenuse
		a := 3 + rand.Intn(10+int(ctx.Difficulty*20))
		if a == 15 || a == 30 || a == 45 || a == 60 || a == 90 {
			a += 2
		}
		c := 2 * a
		return generator.Problem{
			Question: fmt.Sprintf("A 30-60-90 triangle has its shortest leg %d units long. How long is the hypotenuse?", a),
			Answer:   fmt.Sprintf("%d", c),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula. In a 30-60-90 special right triangle, side ratios are 1 : sqrt(3) : 2, so the hypotenuse is twice the shortest leg: c = 2a. Subgoal 2: Substitute parameters. Substitute shortest leg a = %d: c = 2 * %d. Subgoal 3: Evaluate and simplify. Multiply to find hypotenuse c = %d.",
				a, a, c),
			Facts: map[string]string{
				"hypotenuse": fmt.Sprintf("%d", c),
				"short_leg":  fmt.Sprintf("%d", a),
			},
		}
	}

	// Hypotenuse given, find short leg
	half := 3 + rand.Intn(10+int(ctx.Difficulty*20))
	if half == 15 || half == 30 || half == 45 || half == 60 || half == 90 {
		half += 2
	}
	c := 2 * half
	a := half
	return generator.Problem{
		Question: fmt.Sprintf("A 30-60-90 triangle has a hypotenuse %d units long. How long is its shortest leg?", c),
		Answer:   fmt.Sprintf("%d", a),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. In a 30-60-90 special right triangle, the shortest leg opposite the 30 degree angle is half the hypotenuse: a = c / 2. Subgoal 2: Substitute parameters. Substitute hypotenuse c = %d: a = %d / 2. Subgoal 3: Evaluate and simplify. Divide to find shortest leg a = %d.",
			c, c, a),
		Facts: map[string]string{
			"hypotenuse": fmt.Sprintf("%d", c),
			"short_leg":  fmt.Sprintf("%d", a),
		},
	}
}
