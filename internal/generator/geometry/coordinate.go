package geometry

// The last geometry block: slope, distance in space, the trapezoid, similar solids volumes
// and the tangent-chord angle.
//
// These fill coordinate work beyond the distance formula, the trapezoid as the quadrilateral
// whose area averages the parallel bases, volumes of similar solids as the cube of the length ratio
// and the tangent-chord angle theorem.
//
// Every answer is an integer by construction with dynamic difficulty scaling,
// structured cognitive scaffolding and parameter transparency.
import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerFinal(reg *generator.Registry) {
	reg.Register("geo.coord.slope", &slopeGen{})
	reg.Register("geo.coord.distance_3d", &distance3DGen{})
	reg.Register("geo.quad.trapezoid_area", &trapezoidAreaGen{})
	reg.Register("geo.solid.similar_volume", &similarVolumeGen{})
	reg.Register("geo.circle.tangent_chord", &tangentChordGen{})
}

// slopeGen asks for the slope through two points, with integer offsets so the answer is whole.
type slopeGen struct{}

func (g *slopeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var x1, y1, x2, y2, run, slope int
	var answer string

	for iter := 0; iter < 20; iter++ {
		run = 2 + rand.Intn(3+tier*2)
		slope = 2 + rand.Intn(3+tier*3)
		if rand.Float64() < 0.4 && tier > 0 {
			slope = -slope
		}
		x1 = rand.Intn(9) - 4
		y1 = rand.Intn(9) - 4
		x2 = x1 + run
		y2 = y1 + slope*run
		answer = strconv.Itoa(slope)

		absSlope := slope
		if absSlope < 0 {
			absSlope = -absSlope
		}
		// Read-off avoidance: |slope| must not match |x1|, |y1|, |x2| or |y2|
		collision := false
		for _, coord := range []int{x1, y1, x2, y2} {
			absC := coord
			if absC < 0 {
				absC = -absC
			}
			if absSlope == absC {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	rise := y2 - y1
	subgoal1 := "Subgoal 1: State governing formula. The slope m of a non-vertical line passing through points (x_1, y_1) and (x_2, y_2) is given by m = (y_2 - y_1) / (x_2 - x_1)."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For coordinates (%d, %d) and (%d, %d), the rise is %d - (%d) = %d and run is %d - (%d) = %d.",
		x1, y1, x2, y2, y2, y1, rise, x2, x1, run)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing rise by run yields slope m = %d / %d = %s.",
		rise, run, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("Find the slope of the line through \\((%d,%d)\\) and \\((%d,%d)\\).", x1, y1, x2, y2),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"rise":   strconv.Itoa(rise),
			"run":    strconv.Itoa(run),
			"slope":  answer,
			"answer": answer,
		},
	}
}

// distance3DGen uses procedurally generated Pythagorean quadruples so the 3D distance is whole.
type distance3DGen struct{}

func (g *distance3DGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, c, d int
	var answer string

	// For any even a >= 2 and odd b >= 1, S = a^2 + b^2 is odd.
	// Setting c = (S - 1)/2 and d = (S + 1)/2 yields a^2 + b^2 + c^2 = d^2.
	for iter := 0; iter < 20; iter++ {
		valA := 2 * (1 + rand.Intn(3+tier*2))
		valB := 1 + 2*rand.Intn(3+tier*2)
		s := valA*valA + valB*valB
		valC := (s - 1) / 2
		valD := (s + 1) / 2

		coords := []int{valA, valB, valC}
		rand.Shuffle(len(coords), func(i, j int) {
			coords[i], coords[j] = coords[j], coords[i]
		})
		a, b, c, d = coords[0], coords[1], coords[2], valD
		answer = strconv.Itoa(d)

		// Read-off avoidance: d must not match a, b or c
		if d != a && d != b && d != c {
			break
		}
	}

	sumSq := a*a + b*b + c*c
	subgoal1 := "Subgoal 1: State governing formula. The 3D Euclidean distance between the origin (0,0,0) and point (a, b, c) is given by d = \\sqrt{a^2 + b^2 + c^2}."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For coordinates a = %d, b = %d and c = %d, the sum of squares is (%d)^2 + (%d)^2 + (%d)^2 = %d + %d + %d = %d.",
		a, b, c, a, b, c, a*a, b*b, c*c, sumSq)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Taking the square root yields distance d = \\sqrt{%d} = %s units.",
		sumSq, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("Find the distance between the points \\((0,0,0)\\) and \\((%d,%d,%d)\\).", a, b, c),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"distance": answer,
			"answer":   answer,
		},
	}
}

// trapezoidAreaGen uses the average of the two parallel sides.
type trapezoidAreaGen struct{}

func (g *trapezoidAreaGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var b1, b2, h, area int
	var answer string

	for iter := 0; iter < 20; iter++ {
		b1 = 2 * (2 + rand.Intn(4+tier*3))
		b2 = 2 * (3 + rand.Intn(4+tier*3))
		if b1 == b2 {
			b2 += 2
		}
		h = 2 + rand.Intn(4+tier*3)
		area = ((b1 + b2) / 2) * h
		answer = strconv.Itoa(area)

		// Read-off avoidance: answer must not match b1, b2 or h
		if area != b1 && area != b2 && area != h {
			break
		}
	}

	avgBase := (b1 + b2) / 2
	subgoal1 := "Subgoal 1: State governing formula. The area of a trapezoid with parallel bases b_1 and b_2 and height h is given by A = ((b_1 + b_2) / 2) \\cdot h."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For parallel bases b_1 = %d and b_2 = %d and height h = %d, the average base is (%d + %d) / 2 = %d.",
		b1, b2, h, b1, b2, avgBase)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Multiplying by height gives area A = (%d)(%d) = %s square units.",
		avgBase, h, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A trapezoid has parallel sides of %d and %d and a height of %d. What is its area?", b1, b2, h),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"base_a": strconv.Itoa(b1),
			"base_b": strconv.Itoa(b2),
			"height": strconv.Itoa(h),
			"area":   answer,
			"answer": answer,
		},
	}
}

// similarVolumeGen applies the cube of the length ratio to volumes.
type similarVolumeGen struct{}

func (g *similarVolumeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var k, volRatio int
	var answer string

	for iter := 0; iter < 20; iter++ {
		k = 2 + rand.Intn(10+tier*5)
		volRatio = k * k * k
		answer = strconv.Itoa(volRatio)

		// Read-off avoidance: answer must not match k
		if volRatio != k {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. When geometric solids are similar with linear scale factor k, their volumes scale by the cube of the linear factor: V_2 / V_1 = k^3."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. Given linear scale factor k = %d, the volume scaling ratio is (%d)^3.",
		k, k)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the cube gives volume ratio (%d)^3 = %s.",
		k, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("Two solids are similar and every length in the larger is %d times the corresponding length in the smaller. How many times larger is the larger solid's volume?", k),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"factor":       strconv.Itoa(k),
			"area_ratio":   strconv.Itoa(k * k),
			"volume_ratio": answer,
			"answer":       answer,
		},
	}
}

// tangentChordGen uses the angle between a tangent and a chord, which is half the intercepted arc.
type tangentChordGen struct{}

func (g *tangentChordGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var arc, ang int
	var answer string

	for iter := 0; iter < 20; iter++ {
		arc = 2 * (15 + rand.Intn(25+tier*25))
		ang = arc / 2
		answer = strconv.Itoa(ang)

		// Read-off avoidance: answer must not match arc
		if ang != arc {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The measure of an angle formed by a tangent and a chord at the point of contact equals half the measure of the intercepted arc: \\theta = (1/2) \\cdot \\text{arc}."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For an intercepted arc measuring %d degrees, the angle is %d / 2.",
		arc, arc)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing yields angle measure \\theta = %d / 2 = %s degrees.",
		arc, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A tangent touches a circle at a point and a chord is drawn from that point. The arc intercepted by the angle between the tangent and the chord has measure %d degrees. What is that angle, in degrees?", arc),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"arc":    strconv.Itoa(arc),
			"angle":  answer,
			"answer": answer,
		},
	}
}
