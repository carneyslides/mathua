package trigonometry

// The six right-triangle ratios turned into computed questions.
//
// Each of these concepts used to be a fixed question with a fixed answer. The
// ratio itself is the definition, so the honest computed question is not "what
// is the ratio" but "given the ratio, what is the side": a triangle similar to
// one of the Pythagorean triples fixes the ratio and a scale factor fixes the
// side, so the answer is always an exact integer rather than a rounded decimal.
// Every generator publishes the ratio, the side it starts from and the answer,
// which is what the solution schema interpolates.

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
)

type ratioTriple struct {
	a, b, hyp int
}

var ratioTriples = []ratioTriple{
	{3, 4, 5}, {5, 12, 13}, {8, 15, 17}, {7, 24, 25},
	{9, 40, 41}, {20, 21, 29}, {12, 35, 37}, {28, 45, 53},
	{11, 60, 61}, {13, 84, 85}, {16, 63, 65}, {33, 56, 65},
	{36, 77, 85}, {39, 80, 89},
}

// pickRightTriangle returns the leg opposite the angle, the leg adjacent to
// it, the hypotenuse of the primitive triple and a dynamic scale factor.
// The triangle actually presented is the triple multiplied by the scale.
func pickRightTriangle(difficulty float64) (opposite, adjacent, hypotenuse, scale int) {
	maxTriples := 3 + int(difficulty*float64(len(ratioTriples)-3))
	if maxTriples > len(ratioTriples) {
		maxTriples = len(ratioTriples)
	}
	t := ratioTriples[rand.Intn(maxTriples)]
	if rand.Intn(2) == 0 {
		opposite, adjacent = t.a, t.b
	} else {
		opposite, adjacent = t.b, t.a
	}
	hypotenuse = t.hyp
	maxScale := 4 + int(difficulty*12)
	scale = 2 + rand.Intn(maxScale)
	return opposite, adjacent, hypotenuse, scale
}

func ratioProblem(question, answer, explanation string, facts map[string]string) generator.Problem {
	return generator.Problem{
		Question:    question,
		Answer:      answer,
		Explanation: explanation,
		Facts:       facts,
	}
}

// ----- sine: opposite over hypotenuse -----

type sineRatioGen struct{}

func (g *sineRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	opposite, _, hypotenuse, scale := pickRightTriangle(ctx.Difficulty)
	given := hypotenuse * scale
	answer := opposite * scale
	for answer == opposite || answer == hypotenuse || answer == given {
		scale++
		given = hypotenuse * scale
		answer = opposite * scale
	}
	ratio := fmt.Sprintf("%d/%d", opposite, hypotenuse)
	return ratioProblem(
		fmt.Sprintf("A right triangle has \\(\\sin\\theta = \\frac{%d}{%d}\\). If its hypotenuse is %d, what is the length of the side opposite \\(\\theta\\)?",
			opposite, hypotenuse, given),
		strconv.Itoa(answer),
		fmt.Sprintf("The sine of an angle is defined as opposite leg over hypotenuse: \\(\\sin\\theta = \\frac{\\text{opposite}}{\\text{hypotenuse}}\\). Given \\(\\sin\\theta = \\frac{%d}{%d}\\) with hypotenuse %d, solve for the opposite side: \\(\\text{opposite} = \\frac{%d}{%d} \\times %d\\). Evaluating the expression yields %d.",
			opposite, hypotenuse, given, opposite, hypotenuse, given, answer),
		map[string]string{
			"ratio":  ratio,
			"given":  strconv.Itoa(given),
			"answer": strconv.Itoa(answer),
		},
	)
}

// ----- cosine: adjacent over hypotenuse -----

type cosineRatioGen struct{}

func (g *cosineRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	_, adjacent, hypotenuse, scale := pickRightTriangle(ctx.Difficulty)
	given := hypotenuse * scale
	answer := adjacent * scale
	for answer == adjacent || answer == hypotenuse || answer == given {
		scale++
		given = hypotenuse * scale
		answer = adjacent * scale
	}
	ratio := fmt.Sprintf("%d/%d", adjacent, hypotenuse)
	return ratioProblem(
		fmt.Sprintf("A right triangle has \\(\\cos\\theta = \\frac{%d}{%d}\\). If its hypotenuse is %d, what is the length of the side adjacent to \\(\\theta\\)?",
			adjacent, hypotenuse, given),
		strconv.Itoa(answer),
		fmt.Sprintf("The cosine of an angle is defined as adjacent leg over hypotenuse: \\(\\cos\\theta = \\frac{\\text{adjacent}}{\\text{hypotenuse}}\\). Given \\(\\cos\\theta = \\frac{%d}{%d}\\) with hypotenuse %d, solve for the adjacent side: \\(\\text{adjacent} = \\frac{%d}{%d} \\times %d\\). Evaluating the expression yields %d.",
			adjacent, hypotenuse, given, adjacent, hypotenuse, given, answer),
		map[string]string{
			"ratio":  ratio,
			"given":  strconv.Itoa(given),
			"answer": strconv.Itoa(answer),
		},
	)
}

// ----- tangent: opposite over adjacent -----

type tangentRatioGen struct{}

func (g *tangentRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	opposite, adjacent, _, scale := pickRightTriangle(ctx.Difficulty)
	given := adjacent * scale
	answer := opposite * scale
	for answer == opposite || answer == adjacent || answer == given {
		scale++
		given = adjacent * scale
		answer = opposite * scale
	}
	ratio := fmt.Sprintf("%d/%d", opposite, adjacent)
	return ratioProblem(
		fmt.Sprintf("A right triangle has \\(\\tan\\theta = \\frac{%d}{%d}\\). If the side adjacent to \\(\\theta\\) is %d, what is the length of the opposite side?",
			opposite, adjacent, given),
		strconv.Itoa(answer),
		fmt.Sprintf("The tangent of an angle is defined as opposite leg over adjacent leg: \\(\\tan\\theta = \\frac{\\text{opposite}}{\\text{adjacent}}\\). Given \\(\\tan\\theta = \\frac{%d}{%d}\\) with adjacent leg %d, solve for the opposite side: \\(\\text{opposite} = \\frac{%d}{%d} \\times %d\\). Evaluating the expression yields %d.",
			opposite, adjacent, given, opposite, adjacent, given, answer),
		map[string]string{
			"ratio":  ratio,
			"given":  strconv.Itoa(given),
			"answer": strconv.Itoa(answer),
		},
	)
}

// ----- secant: hypotenuse over adjacent -----

type secantRatioGen struct{}

func (g *secantRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	_, adjacent, hypotenuse, scale := pickRightTriangle(ctx.Difficulty)
	given := adjacent * scale
	answer := hypotenuse * scale
	for answer == hypotenuse || answer == adjacent || answer == given {
		scale++
		given = adjacent * scale
		answer = hypotenuse * scale
	}
	ratio := fmt.Sprintf("%d/%d", hypotenuse, adjacent)
	return ratioProblem(
		fmt.Sprintf("A right triangle has \\(\\sec\\theta = \\frac{%d}{%d}\\). If the side adjacent to \\(\\theta\\) is %d, what is the hypotenuse?",
			hypotenuse, adjacent, given),
		strconv.Itoa(answer),
		fmt.Sprintf("The secant of an angle is the reciprocal of cosine, defined as hypotenuse over adjacent leg: \\(\\sec\\theta = \\frac{\\text{hypotenuse}}{\\text{adjacent}}\\). Given \\(\\sec\\theta = \\frac{%d}{%d}\\) with adjacent leg %d, solve for the hypotenuse: \\(\\text{hypotenuse} = \\frac{%d}{%d} \\times %d\\). Evaluating the expression yields %d.",
			hypotenuse, adjacent, given, hypotenuse, adjacent, given, answer),
		map[string]string{
			"ratio":  ratio,
			"given":  strconv.Itoa(given),
			"answer": strconv.Itoa(answer),
		},
	)
}

// ----- cosecant: hypotenuse over opposite -----

type cosecantRatioGen struct{}

func (g *cosecantRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	opposite, _, hypotenuse, scale := pickRightTriangle(ctx.Difficulty)
	given := opposite * scale
	answer := hypotenuse * scale
	for answer == hypotenuse || answer == opposite || answer == given {
		scale++
		given = opposite * scale
		answer = hypotenuse * scale
	}
	ratio := fmt.Sprintf("%d/%d", hypotenuse, opposite)
	return ratioProblem(
		fmt.Sprintf("A right triangle has \\(\\csc\\theta = \\frac{%d}{%d}\\). If the side opposite \\(\\theta\\) is %d, what is the hypotenuse?",
			hypotenuse, opposite, given),
		strconv.Itoa(answer),
		fmt.Sprintf("The cosecant of an angle is the reciprocal of sine, defined as hypotenuse over opposite leg: \\(\\csc\\theta = \\frac{\\text{hypotenuse}}{\\text{opposite}}\\). Given \\(\\csc\\theta = \\frac{%d}{%d}\\) with opposite leg %d, solve for the hypotenuse: \\(\\text{hypotenuse} = \\frac{%d}{%d} \\times %d\\). Evaluating the expression yields %d.",
			hypotenuse, opposite, given, hypotenuse, opposite, given, answer),
		map[string]string{
			"ratio":  ratio,
			"given":  strconv.Itoa(given),
			"answer": strconv.Itoa(answer),
		},
	)
}

// ----- cotangent: adjacent over opposite -----

type cotangentRatioGen struct{}

func (g *cotangentRatioGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	opposite, adjacent, _, scale := pickRightTriangle(ctx.Difficulty)
	given := opposite * scale
	answer := adjacent * scale
	for answer == adjacent || answer == opposite || answer == given {
		scale++
		given = opposite * scale
		answer = adjacent * scale
	}
	ratio := fmt.Sprintf("%d/%d", adjacent, opposite)
	return ratioProblem(
		fmt.Sprintf("A right triangle has \\(\\cot\\theta = \\frac{%d}{%d}\\). If the side opposite \\(\\theta\\) is %d, what is the length of the adjacent side?",
			adjacent, opposite, given),
		strconv.Itoa(answer),
		fmt.Sprintf("The cotangent of an angle is the reciprocal of tangent, defined as adjacent leg over opposite leg: \\(\\cot\\theta = \\frac{\\text{adjacent}}{\\text{opposite}}\\). Given \\(\\cot\\theta = \\frac{%d}{%d}\\) with opposite leg %d, solve for the adjacent leg: \\(\\text{adjacent} = \\frac{%d}{%d} \\times %d\\). Evaluating the expression yields %d.",
			adjacent, opposite, given, adjacent, opposite, given, answer),
		map[string]string{
			"ratio":  ratio,
			"given":  strconv.Itoa(given),
			"answer": strconv.Itoa(answer),
		},
	)
}

// registerFuncRatios registers the six ratio concepts that used to be fixed
// questions. The old registrations are removed from Register, because a concept
// id may be registered only once.
func registerFuncRatios(reg *generator.Registry) {
	reg.Register("trig.func.sine", &sineRatioGen{})
	reg.Register("trig.func.cosine", &cosineRatioGen{})
	reg.Register("trig.func.tangent", &tangentRatioGen{})
	reg.Register("trig.func.secant", &secantRatioGen{})
	reg.Register("trig.func.cosecant", &cosecantRatioGen{})
	reg.Register("trig.func.cotangent", &cotangentRatioGen{})
}
