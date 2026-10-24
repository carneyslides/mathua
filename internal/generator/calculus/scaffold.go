package calculus

import (
	"github.com/chuma-beep/mathua/internal/generator"
)

// Ensure generators implement generator.ScaffoldProblem.
var (
	_ generator.ScaffoldProblem = (*integralSubstitutionGen)(nil)
	_ generator.ScaffoldProblem = (*integralPartsGen)(nil)
	_ generator.ScaffoldProblem = (*integralPartialFractionsGen)(nil)
	_ generator.ScaffoldProblem = (*derivOptimizationGen)(nil)
	_ generator.ScaffoldProblem = (*derivRelatedRatesGen)(nil)
	_ generator.ScaffoldProblem = (*lhopitalGen)(nil)
	_ generator.ScaffoldProblem = (*lagrangeGen)(nil)
)

func (g *integralSubstitutionGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For \\(\\int 3x^2 e^{x^3} \\, dx\\), choose substitution \\(u = x^3\\). What is the exponent \\(p\\) in \\(du = 3x^p \\, dx\\)?",
			Answer:        "2",
			Explanation:   "Differentiating u = x^3 gives du/dx = 3x^2, so the exponent is 2.",
			PrereqConcept: "calc.deriv.power_rule",
		},
		{
			Ordinal:       2,
			Prompt:        "The integral transforms to \\(\\int e^u \\, du = e^u + C\\). Evaluate \\(e^u\\) at \\(u = 0\\).",
			Answer:        "1",
			Explanation:   "e^0 = 1.",
			PrereqConcept: "calc.integral.power_rule",
		},
		{
			Ordinal:       3,
			Prompt:        "Substitute back to obtain \\(e^{x^k} + C\\). What is the exponent \\(k\\)?",
			Answer:        "3",
			Explanation:   "Substituting u = x^3 gives e^(x^3) + C, so k = 3.",
			PrereqConcept: "calc.deriv.chain_rule",
		},
	}
}

func (g *integralPartsGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "In integration by parts for \\(\\int x e^{2x} \\, dx\\), choose \\(u = x\\). What is \\(du/dx\\)?",
			Answer:        "1",
			Explanation:   "Differentiating x gives 1.",
			PrereqConcept: "calc.deriv.product_rule",
		},
		{
			Ordinal:       2,
			Prompt:        "For \\(dv = e^{2x} dx\\), integrate to find \\(v = \\frac{1}{m} e^{2x}\\). What is \\(m\\)?",
			Answer:        "2",
			Explanation:   "Integrating e^(2x) gives (1/2) e^(2x), so m = 2.",
			PrereqConcept: "calc.integral.power_rule",
		},
		{
			Ordinal:       3,
			Prompt:        "Evaluate the boundary product term \\(u v = \\frac{x}{2} e^{2x}\\) at \\(x = 2\\) divided by \\(e^4\\).",
			Answer:        "1",
			Explanation:   "At x = 2, (2/2) e^(2*2) = 1 e^4; dividing by e^4 gives 1.",
			PrereqConcept: "calc.integral.substitution",
		},
	}
}

func (g *integralPartialFractionsGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "Factor denominator \\(x^2 - 1 = (x - r)(x + r)\\). What is the positive root \\(r\\)?",
			Answer:        "1",
			Explanation:   "x^2 - 1 = (x - 1)(x + 1), so r = 1.",
			PrereqConcept: "alg.factor.trinomial",
		},
		{
			Ordinal:       2,
			Prompt:        "In \\(\\frac{1}{x^2 - 1} = \\frac{A}{x - 1} + \\frac{B}{x + 1}\\), clear fractions: \\(1 = A(x+1) + B(x-1)\\). Setting \\(x = 1\\) yields \\(2A = ?\\)",
			Answer:        "1",
			Explanation:   "Setting x = 1 gives 1 = A(1 + 1) + B(0) = 2A, so 2A = 1.",
			PrereqConcept: "alg.eq.multi_step",
		},
		{
			Ordinal:       3,
			Prompt:        "Setting \\(x = -1\\) yields \\(-2B = ?\\)",
			Answer:        "1",
			Explanation:   "Setting x = -1 gives 1 = A(0) + B(-1 - 1) = -2B, so -2B = 1.",
			PrereqConcept: "calc.integral.parts",
		},
	}
}

func (g *derivOptimizationGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "A rectangle with perimeter 20 has width \\(w\\) and height \\(10 - w\\). Area is \\(A(w) = w(10 - w) = C w - w^2\\). What is \\(C\\)?",
			Answer:        "10",
			Explanation:   "Expanding w(10 - w) gives 10w - w^2, so C = 10.",
			PrereqConcept: "calc.deriv.power_rule",
		},
		{
			Ordinal:       2,
			Prompt:        "Compute derivative \\(A'(w) = 10 - 2w = 0\\) to find the critical width \\(w\\).",
			Answer:        "5",
			Explanation:   "Solving 10 - 2w = 0 gives 2w = 10, so w = 5.",
			PrereqConcept: "alg.eq.multi_step",
		},
		{
			Ordinal:       3,
			Prompt:        "Evaluate the maximum area \\(A(5) = 5 \\times 5\\).",
			Answer:        "25",
			Explanation:   "5 * 5 = 25.",
			PrereqConcept: "calc.deriv.optimization",
		},
	}
}

func (g *derivRelatedRatesGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For area \\(A = \\pi r^2\\), differentiate with respect to \\(r\\): \\(\\frac{dA}{dr} = k \\pi r\\). What is \\(k\\)?",
			Answer:        "2",
			Explanation:   "The power rule gives d/dr(pi r^2) = 2 pi r, so k = 2.",
			PrereqConcept: "calc.deriv.power_rule",
		},
		{
			Ordinal:       2,
			Prompt:        "Using the chain rule \\(\\frac{dA}{dt} = \\frac{dA}{dr} \\frac{dr}{dt} = 2\\pi r \\frac{dr}{dt}\\), compute the scalar multiplier \\(2 r \\frac{dr}{dt}\\) when \\(r = 5\\) and \\(\\frac{dr}{dt} = 3\\).",
			Answer:        "30",
			Explanation:   "2 * 5 * 3 = 30.",
			PrereqConcept: "calc.deriv.chain_rule",
		},
		{
			Ordinal:       3,
			Prompt:        "What is the rate of area growth \\(\\frac{dA}{dt}\\) in units of \\(\\pi\\) cm\\(^2\\)/s?",
			Answer:        "30",
			Explanation:   "The rate is 30 pi cm^2/s, giving scalar 30.",
			PrereqConcept: "calc.deriv.related_rates",
		},
	}
}

func (g *lhopitalGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For \\(\\lim_{x \\to 0} \\frac{\\sin(x)}{x}\\), what is the value of the numerator \\(\\sin(0)\\)?",
			Answer:        "0",
			Explanation:   "sin(0) = 0, verifying the 0/0 indeterminate form.",
			PrereqConcept: "calc.limit.concept",
		},
		{
			Ordinal:       2,
			Prompt:        "Differentiate the denominator \\(g(x) = x\\): what is \\(g'(x)\\)?",
			Answer:        "1",
			Explanation:   "The derivative of x is 1.",
			PrereqConcept: "calc.deriv.power_rule",
		},
		{
			Ordinal:       3,
			Prompt:        "Apply L'Hôpital's rule \\(\\lim_{x \\to 0} \\frac{\\cos(x)}{1}\\). What is \\(\\cos(0)\\)?",
			Answer:        "1",
			Explanation:   "cos(0) = 1, so the limit is 1.",
			PrereqConcept: "calc.limit.lhopital",
		},
	}
}

func (g *lagrangeGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "Maximize \\(f(x, y) = xy\\) subject to \\(x + y = 10\\). Lagrange condition \\(\\nabla f = \\lambda \\nabla g\\) gives \\(y = \\lambda\\) and \\(x = \\lambda\\), so \\(x = y\\). What is the ratio \\(x/y\\)?",
			Answer:        "1",
			Explanation:   "Since x = y, the ratio is 1.",
			PrereqConcept: "calc.deriv.partial",
		},
		{
			Ordinal:       2,
			Prompt:        "Substitute \\(x = y\\) into constraint \\(x + y = 10\\) to find optimal coordinate \\(x\\).",
			Answer:        "5",
			Explanation:   "2x = 10 implies x = 5.",
			PrereqConcept: "alg.eq.multi_step",
		},
		{
			Ordinal:       3,
			Prompt:        "Evaluate maximum product \\(f(5, 5) = 5 \\times 5\\).",
			Answer:        "25",
			Explanation:   "5 * 5 = 25.",
			PrereqConcept: "calc.deriv.optimization",
		},
	}
}
