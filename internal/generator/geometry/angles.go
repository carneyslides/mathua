package geometry

// The angle facts, the midsegment and the prism.
//
// These are the school-course workhorses: the exterior angle theorem, the
// midsegment half, the base angles of an isosceles triangle, the opposite angles of a cyclic
// quadrilateral and the prism volume that a pyramid third is measured against.
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

func registerAngleFacts(reg *generator.Registry) {
	reg.Register("geo.triangle.exterior_angle", &triangleExteriorAngleGen{})
	reg.Register("geo.triangle.midsegment", &midsegmentGen{})
	reg.Register("geo.triangle.isosceles", &isoscelesGen{})
	reg.Register("geo.circle.cyclic_quad", &cyclicQuadGen{})
	reg.Register("geo.solid.prism_volume", &prismVolumeGen{})
}

// triangleExteriorAngleGen asks for an exterior angle, which is the sum of the two remote interior angles.
type triangleExteriorAngleGen struct{}

func (g *triangleExteriorAngleGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b int
	var answer string

	for iter := 0; iter < 20; iter++ {
		a = 25 + rand.Intn(25+tier*20)
		b = 25 + rand.Intn(25+tier*20)
		if a == b {
			b += 5
		}
		if a+b > 155 {
			b = 155 - a
		}
		ext := a + b
		answer = strconv.Itoa(ext)

		// Read-off avoidance: answer must not match a or b
		if ext != a && ext != b {
			break
		}
	}

	third := 180 - a - b
	subgoal1 := "Subgoal 1: State governing formula. The exterior angle theorem states that an exterior angle of a triangle is equal to the sum of the two remote interior angles: \\theta_{ext} = \\angle A + \\angle B."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For remote interior angles measuring %d degrees and %d degrees, the sum is %d + %d.",
		a, b, a, b)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the sum gives exterior angle measure \\theta_{ext} = %s degrees.",
		answer)

	return generator.Problem{
		Question:    fmt.Sprintf("In a triangle, one side is extended to form an exterior angle. The two interior angles not adjacent to it are %d and %d degrees. What is the measure of the exterior angle, in degrees?", a, b),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"interior_a": strconv.Itoa(a),
			"interior_b": strconv.Itoa(b),
			"third":      strconv.Itoa(third),
			"exterior":   answer,
			"answer":     answer,
		},
	}
}

// midsegmentGen uses the midsegment half of the third side.
type midsegmentGen struct{}

func (g *midsegmentGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var third, mid int
	var answer string

	for iter := 0; iter < 20; iter++ {
		baseK := mathutil.ScaleInt(ctx.Difficulty, 4, 15)
		third = 2 * (baseK + rand.Intn(4+tier*5))
		mid = third / 2
		answer = strconv.Itoa(mid)

		// Read-off avoidance: answer must not match third
		if mid != third {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. By the triangle midsegment theorem, a segment connecting the midpoints of two sides is parallel to the third side and has length equal to half that side: M = s / 2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. Given third side length s = %d units, the midsegment length is %d / 2.",
		third, third)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the division yields midsegment length M = %s units.",
		answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A midsegment of a triangle joins the midpoints of two of its sides. The third side is %d units long. How long is the midsegment?", third),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"third_side": strconv.Itoa(third),
			"midsegment": answer,
			"answer":     answer,
		},
	}
}

// isoscelesGen asks for a base angle of an isosceles triangle, given the vertex angle.
type isoscelesGen struct{}

func (g *isoscelesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var vertex, base int
	var answer string

	for iter := 0; iter < 20; iter++ {
		k := 12 + rand.Intn(25+tier*20)
		vertex = 2 * k
		if vertex > 140 {
			vertex = 140
		}
		// Read-off avoidance: avoid equilateral case where vertex equals base angle (60 degrees)
		if vertex == 60 {
			vertex = 64
		}
		base = (180 - vertex) / 2
		answer = strconv.Itoa(base)

		if base != vertex {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The interior angles of a triangle sum to 180 degrees and the two base angles of an isosceles triangle are congruent: 2\\theta + V = 180^\\circ, giving \\theta = (180^\\circ - V) / 2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For vertex angle V = %d degrees, the remaining sum for both base angles is 180 - %d = %d degrees.",
		vertex, vertex, 180-vertex)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing by 2 yields each base angle \\theta = %d / 2 = %s degrees.",
		180-vertex, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("An isosceles triangle has a vertex angle of %d degrees. What is the measure of each base angle, in degrees?", vertex),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"vertex_angle": strconv.Itoa(vertex),
			"base_angle":   answer,
			"answer":       answer,
		},
	}
}

// cyclicQuadGen uses the fact that opposite angles of a cyclic quadrilateral are supplementary.
type cyclicQuadGen struct{}

func (g *cyclicQuadGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, opp int
	var answer string

	for iter := 0; iter < 20; iter++ {
		a = 35 + rand.Intn(40+tier*35)
		if a > 145 {
			a = 145
		}
		// Read-off avoidance: avoid 90 degrees where opposite angle equals given angle
		if a == 90 {
			a = 95
		}
		opp = 180 - a
		answer = strconv.Itoa(opp)

		if opp != a {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. Opposite angles of a cyclic quadrilateral inscribed in a circle sum to 180 degrees: \\angle A + \\angle C = 180^\\circ, giving \\angle C = 180^\\circ - \\angle A."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For inscribed angle \\angle A = %d degrees, the opposite angle equation gives %d + \\angle C = 180.",
		a, a)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Subtracting yields opposite angle \\angle C = 180 - %d = %s degrees.",
		a, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A quadrilateral is inscribed in a circle, so its four vertices lie on the circle. One of its angles is %d degrees. What is the measure of the opposite angle, in degrees?", a),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"angle":    strconv.Itoa(a),
			"opposite": answer,
			"answer":   answer,
		},
	}
}

// prismVolumeGen computes the volume of a right prism from base area and height.
type prismVolumeGen struct{}

func (g *prismVolumeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var base, h, vol int
	var answer string

	for iter := 0; iter < 20; iter++ {
		base = mathutil.ScaleInt(ctx.Difficulty, 4, 12) + rand.Intn(5+tier*4)
		h = mathutil.ScaleInt(ctx.Difficulty, 3, 8) + rand.Intn(4+tier*3)
		vol = base * h
		answer = strconv.Itoa(vol)

		// Read-off avoidance: answer must not match base or h
		if vol != base && vol != h {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The volume of a prism equals the product of its base area B and height h: V = B \\cdot h."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. Substituting base area B = %d square units and height h = %d units gives V = (%d)(%d).",
		base, h, base, h)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Multiplying yields volume V = %s cubic units.",
		answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A prism has a base of area %d square units and a height of %d units. What is its volume, in cubic units?", base, h),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"base":   strconv.Itoa(base),
			"height": strconv.Itoa(h),
			"volume": answer,
			"answer": answer,
		},
	}
}
