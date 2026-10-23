package geometry

import (
	"fmt"
	"math/rand"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerTriangles(reg *generator.Registry) {
	reg.Register("geo.triangle.inequality", &triangleInequalityGen{})
	reg.Register("geo.similar.area_ratio", &areaRatioGen{})
	reg.Register("geo.circle.tangent_length", &tangentLengthGen{})
	reg.Register("geo.solid.pyramid_volume", &pyramidVolumeGen{})
}

// pythagoreanTriple returns (leg1, leg2, hypotenuse) dynamically scaled with difficulty.
func pythagoreanTriple(difficulty float64) (int, int, int) {
	bases := [][2]int{
		{2, 1}, // 3, 4, 5
		{3, 2}, // 5, 12, 13
		{4, 1}, // 15, 8, 17
		{4, 3}, // 7, 24, 25
		{5, 2}, // 21, 20, 29
		{5, 4}, // 9, 40, 41
		{6, 1}, // 35, 12, 37
		{6, 5}, // 11, 60, 61
		{7, 2}, // 45, 28, 53
		{7, 4}, // 33, 56, 65
		{7, 6}, // 13, 84, 85
		{8, 1}, // 63, 16, 65
		{8, 3}, // 55, 48, 73
		{8, 5}, // 39, 80, 89
	}
	tier := mathutil.DifficultyTier(difficulty)
	maxBase := 4 + tier*4
	if maxBase > len(bases) {
		maxBase = len(bases)
	}
	base := bases[rand.Intn(maxBase)]
	m, n := base[0], base[1]
	l1 := m*m - n*n
	l2 := 2 * m * n
	hyp := m*m + n*n
	scale := 1 + rand.Intn(1+int(difficulty*5))
	return l1 * scale, l2 * scale, hyp * scale
}

// triangleInequalityGen asks for the largest or smallest whole third side of a triangle given two sides.
type triangleInequalityGen struct{}

func (g *triangleInequalityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := 3 + rand.Intn(5+int(ctx.Difficulty*15))
	diff := 2 + rand.Intn(4+int(ctx.Difficulty*12))
	if diff == a-1 {
		diff++
	}
	b := a + diff
	bound := a + b
	difference := b - a

	if rand.Intn(2) == 0 {
		// Largest whole third side
		ans := bound - 1
		return generator.Problem{
			Question: fmt.Sprintf("Two sides of a triangle are %d and %d. What is the largest whole number the third side can be?", a, b),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State governing formula. By the triangle inequality, third side c must satisfy |a - b| < c < a + b. Subgoal 2: Substitute parameters. Substitute a = %d and b = %d: |%d - %d| = %d < c < %d + %d = %d. Subgoal 3: Evaluate and simplify. The third side must be strictly less than %d, so the largest whole number is %d - 1 = %d.",
				a, b, a, b, difference, a, b, bound, bound, bound, ans),
			Facts: map[string]string{
				"side":       fmt.Sprintf("%d", ans),
				"given_a":    fmt.Sprintf("%d", a),
				"given_b":    fmt.Sprintf("%d", b),
				"bound":      fmt.Sprintf("%d", bound),
				"difference": fmt.Sprintf("%d", difference),
			},
		}
	}

	// Smallest whole third side
	ans := difference + 1
	return generator.Problem{
		Question: fmt.Sprintf("Two sides of a triangle are %d and %d. What is the smallest whole number the third side can be?", a, b),
		Answer:   fmt.Sprintf("%d", ans),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. By the triangle inequality, third side c must satisfy |a - b| < c < a + b. Subgoal 2: Substitute parameters. Substitute a = %d and b = %d: |%d - %d| = %d < c < %d + %d = %d. Subgoal 3: Evaluate and simplify. The third side must be strictly greater than %d, so the smallest whole number is %d + 1 = %d.",
			a, b, a, b, difference, a, b, bound, difference, difference, ans),
		Facts: map[string]string{
			"side":       fmt.Sprintf("%d", ans),
			"given_a":    fmt.Sprintf("%d", a),
			"given_b":    fmt.Sprintf("%d", b),
			"difference": fmt.Sprintf("%d", difference),
			"bound":      fmt.Sprintf("%d", bound),
		},
	}
}

// areaRatioGen computes the area scaling ratio from linear similarity factor.
type areaRatioGen struct{}

func (g *areaRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	k := 2 + rand.Intn(6+int(ctx.Difficulty*16))
	areaRatio := k * k
	return generator.Problem{
		Question: fmt.Sprintf("Two figures are similar with a side of the larger %d times the corresponding side of the smaller. How many times larger is the larger figure's area?", k),
		Answer:   fmt.Sprintf("%d", areaRatio),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. Areas of similar figures scale quadratically with linear factor k: A_ratio = k^2. Subgoal 2: Substitute parameters. Substitute linear scale factor k = %d: A_ratio = %d^2. Subgoal 3: Evaluate and simplify. Compute %d^2 = %d.",
			k, k, k, areaRatio),
		Facts: map[string]string{
			"factor":     fmt.Sprintf("%d", k),
			"area_ratio": fmt.Sprintf("%d", areaRatio),
		},
	}
}

// tangentLengthGen computes tangent segment length to circle from external point.
type tangentLengthGen struct{}

func (g *tangentLengthGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	l1, l2, hyp := pythagoreanTriple(ctx.Difficulty)
	radius, tangent := l1, l2
	if rand.Intn(2) == 0 {
		radius, tangent = l2, l1
	}
	distance := hyp
	return generator.Problem{
		Question: fmt.Sprintf("A point lies %d units from the centre of a circle of radius %d. A tangent is drawn from the point, touching the circle. How long is the tangent segment?", distance, radius),
		Answer:   fmt.Sprintf("%d", tangent),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. A tangent segment meets the radius at a right angle, forming a right triangle where center distance d is hypotenuse, radius r is one leg and tangent segment t is the other leg: r^2 + t^2 = d^2, so t = sqrt(d^2 - r^2). Subgoal 2: Substitute parameters. Substitute radius r = %d and distance d = %d: t = sqrt(%d^2 - %d^2) = sqrt(%d - %d). Subgoal 3: Evaluate and simplify. Subtract under radical: t = sqrt(%d) = %d.",
			radius, distance, distance, radius, distance*distance, radius*radius, distance*distance-radius*radius, tangent),
		Facts: map[string]string{
			"tangent":  fmt.Sprintf("%d", tangent),
			"radius":   fmt.Sprintf("%d", radius),
			"distance": fmt.Sprintf("%d", distance),
		},
	}
}

// pyramidVolumeGen computes volume of pyramid with one-third multiplier.
type pyramidVolumeGen struct{}

func (g *pyramidVolumeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	h := 2 + rand.Intn(5+int(ctx.Difficulty*15))
	if h == 3 {
		h = 4
	}
	baseFactor := 2 + rand.Intn(5+int(ctx.Difficulty*18))
	base := 3 * baseFactor
	vol := base * h / 3
	return generator.Problem{
		Question: fmt.Sprintf("A pyramid has a base of area %d square units and a height of %d units. What is its volume, in cubic units?", base, h),
		Answer:   fmt.Sprintf("%d", vol),
		Explanation: fmt.Sprintf("Subgoal 1: State governing formula. The volume of a pyramid is one third the product of its base area and height: V = (1/3) * B * h. Subgoal 2: Substitute parameters. Substitute base area B = %d and height h = %d: V = (1/3) * %d * %d = (%d * %d) / 3. Subgoal 3: Evaluate and simplify. Multiply and divide by 3: %d / 3 = %d.",
			base, h, base, h, base, h, base*h, vol),
		Facts: map[string]string{
			"volume": fmt.Sprintf("%d", vol),
			"base":   fmt.Sprintf("%d", base),
			"height": fmt.Sprintf("%d", h),
		},
	}
}
