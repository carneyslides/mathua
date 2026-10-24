package odes

import (
	"github.com/chuma-beep/mathua/internal/generator"
)

// Ensure generators implement generator.ScaffoldProblem.
var (
	_ generator.ScaffoldProblem = (*separableGen)(nil)
	_ generator.ScaffoldProblem = (*linearFirstGen)(nil)
	_ generator.ScaffoldProblem = (*exactGen)(nil)
	_ generator.ScaffoldProblem = (*homogeneousGen)(nil)
	_ generator.ScaffoldProblem = (*nonhomogeneousGen)(nil)
	_ generator.ScaffoldProblem = (*variationParamsGen)(nil)
)

func (g *separableGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For \\(y \\frac{dy}{dx} = 2x\\), separate variables and integrate the \\(x\\)-side: evaluate \\(\\int_0^3 2x \\, dx\\).",
			Answer:        "9",
			Explanation:   "Integrating 2x gives x^2; evaluating from 0 to 3 yields 3^2 - 0 = 9.",
			PrereqConcept: "calc.integral.power_rule",
		},
		{
			Ordinal:       2,
			Prompt:        "Given initial condition \\(y(0) = 4\\), evaluate the initial integration constant \\([y(0)]^2\\).",
			Answer:        "16",
			Explanation:   "Squaring the initial value gives 4^2 = 16.",
			PrereqConcept: "ode.basics.concept",
		},
		{
			Ordinal:       3,
			Prompt:        "Combine to find \\([y(3)]^2 = 9 + 16\\).",
			Answer:        "25",
			Explanation:   "Adding the separated integral to the constant yields 9 + 16 = 25.",
			PrereqConcept: "calc.integral.substitution",
		},
	}
}

func (g *linearFirstGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For \\(y' + 2y = 4\\), find the exponent coefficient \\(k\\) in integrating factor \\(\\mu(x) = e^{kx}\\).",
			Answer:        "2",
			Explanation:   "The integrating factor is exp(integral 2 dx) = exp(2x), giving exponent coefficient 2.",
			PrereqConcept: "calc.deriv.exp_log",
		},
		{
			Ordinal:       2,
			Prompt:        "Multiply by \\(\\mu\\) and integrate \\(\\int 4 e^{2x} \\, dx = A e^{2x}\\). What is \\(A\\)?",
			Answer:        "2",
			Explanation:   "Integrating yields (4/2) e^(2x) = 2 e^(2x), so A = 2.",
			PrereqConcept: "calc.integral.power_rule",
		},
		{
			Ordinal:       3,
			Prompt:        "For general solution \\(y(x) = 2 + C e^{-2x}\\) with \\(y(0) = 5\\), solve for \\(C\\).",
			Answer:        "3",
			Explanation:   "Setting x = 0 gives 2 + C = 5, hence C = 3.",
			PrereqConcept: "ode.first_order.separable",
		},
	}
}

func (g *exactGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "For \\((3x^2 y + 2x) dx + (x^3 + 4y) dy = 0\\), compute \\(\\frac{\\partial M}{\\partial y}\\) at \\((1, 1)\\).",
			Answer:        "3",
			Explanation:   "Differentiating M = 3x^2 y + 2x with respect to y treats x as constant, giving 3x^2 = 3(1)^2 = 3.",
			PrereqConcept: "calc.deriv.implicit",
		},
		{
			Ordinal:       2,
			Prompt:        "Integrate \\(M(x, y)\\) with respect to \\(x\\) to obtain potential function \\(\\Psi(x, y) = x^3 y + x^2 + 2y^2\\). Evaluate \\(\\Psi(1, 1)\\).",
			Answer:        "4",
			Explanation:   "Substituting (1, 1) gives (1)^3(1) + (1)^2 + 2(1)^2 = 1 + 1 + 2 = 4.",
			PrereqConcept: "calc.integral.indefinite",
		},
		{
			Ordinal:       3,
			Prompt:        "Given initial condition \\(y(1) = 1\\), what is the first integral constant \\(C = \\Psi(1, 1)\\)?",
			Answer:        "4",
			Explanation:   "The first integral is Psi(x, y) = C, so C = Psi(1, 1) = 4.",
			PrereqConcept: "ode.first_order.linear",
		},
	}
}

func (g *homogeneousGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "Compute the discriminant \\(\\Delta = b^2 - 4ac\\) for characteristic equation of \\(y'' - 5y' + 6y = 0\\).",
			Answer:        "1",
			Explanation:   "Here a = 1, b = -5, c = 6. The discriminant is (-5)^2 - 4(1)(6) = 25 - 24 = 1.",
			PrereqConcept: "alg.quad.formula",
		},
		{
			Ordinal:       2,
			Prompt:        "Find the smaller root of \\(r^2 - 5r + 6 = 0\\).",
			Answer:        "2",
			Explanation:   "Factoring gives (r - 2)(r - 3) = 0, so the roots are 2 and 3; the smaller root is 2.",
			PrereqConcept: "alg.quad.formula",
		},
		{
			Ordinal:       3,
			Prompt:        "Find the larger root of \\(r^2 - 5r + 6 = 0\\).",
			Answer:        "3",
			Explanation:   "The larger root is 3.",
			PrereqConcept: "ode.first_order.linear",
		},
	}
}

func (g *nonhomogeneousGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "At what angular frequency \\(\\omega_0\\) does pure resonance occur for \\(y'' + 16y = \\cos(\\omega t)\\)?",
			Answer:        "4",
			Explanation:   "Resonance occurs when omega = omega_0 = sqrt(16) = 4.",
			PrereqConcept: "ode.second_order.homogeneous",
		},
		{
			Ordinal:       2,
			Prompt:        "Find constant particular solution \\(y_p\\) for \\(y'' + 3y' + 2y = 8\\).",
			Answer:        "4",
			Explanation:   "For constant y_p, derivatives vanish: 2*y_p = 8, so y_p = 4.",
			PrereqConcept: "alg.eq.vars_both_sides",
		},
		{
			Ordinal:       3,
			Prompt:        "For \\(y(x) = C_1 \\cos(4x) + y_p\\) with \\(y_p = 4\\) and initial condition \\(y(0) = 9\\), find \\(C_1\\).",
			Answer:        "5",
			Explanation:   "At x = 0, y(0) = C_1 + 4 = 9, so C_1 = 9 - 4 = 5.",
			PrereqConcept: "ode.second_order.homogeneous",
		},
	}
}

func (g *variationParamsGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "Compute Wronskian \\(W(e^{2x}, e^{5x})\\) evaluated at \\(x = 0\\).",
			Answer:        "3",
			Explanation:   "W = y1 y2' - y1' y2 = (1)(5) - (2)(1) = 5 - 2 = 3.",
			PrereqConcept: "ode.wronskian",
		},
		{
			Ordinal:       2,
			Prompt:        "In variation of parameters with \\(W = 3\\) and forcing function \\(g(x) = 6\\), compute parameter derivative magnitude \\(|u_1'(0)| = |-y_2(0) g(0) / W|\\).",
			Answer:        "2",
			Explanation:   "Evaluating gives |-(1)(6) / 3| = |-2| = 2.",
			PrereqConcept: "ode.second_order.homogeneous",
		},
		{
			Ordinal:       3,
			Prompt:        "If \\(u_1(x) = -2x\\) and \\(y_1(x) = 1\\), evaluate particular solution component \\(u_1(1) y_1(1)\\) magnitude.",
			Answer:        "2",
			Explanation:   "|-2(1)(1)| = 2.",
			PrereqConcept: "ode.second_order.nonhomogeneous",
		},
	}
}
