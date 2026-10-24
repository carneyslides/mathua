package trigonometry

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
)

// distinctIdentDistractors returns three options that differ from the answer and from
// each other, falling back to diverse expressions when collisions occur.
func distinctIdentDistractors(correct string, candidates ...string) []string {
	seen := map[string]bool{correct: true}
	out := make([]string, 0, 3)
	for _, cand := range candidates {
		if cand != "" && !seen[cand] {
			seen[cand] = true
			out = append(out, cand)
			if len(out) == 3 {
				break
			}
		}
	}
	fallbacks := []string{
		"0",
		"1",
		"-1",
		"undefined",
		"no real solution",
	}
	for _, fb := range fallbacks {
		if len(out) == 3 {
			break
		}
		if !seen[fb] {
			seen[fb] = true
			out = append(out, fb)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// 1. Product to sum (trig.ident.product_to_sum)
// ---------------------------------------------------------------------------

type productToSumGen struct{}

func (g *productToSumGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	a := 2 + rand.Intn(7)
	b := 1 + rand.Intn(a-1)
	sum := a + b
	diff := a - b

	switch rand.Intn(4) {
	case 0:
		// sin(ax) cos(bx) = 1/2 [sin((a+b)x) + sin((a-b)x)]
		q := fmt.Sprintf("Which expression rewrites the product $\\sin(%d x) \\cos(%d x)$ as a sum or difference?", a, b)
		correct := fmt.Sprintf("(\\sin(%d x) + \\sin(%d x)) / 2", sum, diff)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("(\\cos(%d x) - \\cos(%d x)) / 2", diff, sum),
			fmt.Sprintf("(\\sin(%d x) - \\sin(%d x)) / 2", sum, diff),
			fmt.Sprintf("(\\cos(%d x) + \\cos(%d x)) / 2", diff, sum),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The product-to-sum identity sin A cos B = (sin(A + B) + sin(A - B)) / 2 gives (sin(%d x) + sin(%d x)) / 2.", sum, diff)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 1:
		// cos(ax) cos(bx) = 1/2 [cos((a-b)x) + cos((a+b)x)]
		q := fmt.Sprintf("Which expression rewrites the product $\\cos(%d x) \\cos(%d x)$ as a sum or difference?", a, b)
		correct := fmt.Sprintf("(\\cos(%d x) + \\cos(%d x)) / 2", diff, sum)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("(\\cos(%d x) - \\cos(%d x)) / 2", diff, sum),
			fmt.Sprintf("(\\sin(%d x) + \\sin(%d x)) / 2", sum, diff),
			fmt.Sprintf("2 \\cos(%d x) \\cos(%d x)", sum, diff),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The product-to-sum identity cos A cos B = (cos(A - B) + cos(A + B)) / 2 gives (cos(%d x) + cos(%d x)) / 2.", diff, sum)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 2:
		// sin(ax) sin(bx) = 1/2 [cos((a-b)x) - cos((a+b)x)]
		q := fmt.Sprintf("Which expression rewrites the product $\\sin(%d x) \\sin(%d x)$ as a sum or difference?", a, b)
		correct := fmt.Sprintf("(\\cos(%d x) - \\cos(%d x)) / 2", diff, sum)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("(\\cos(%d x) + \\cos(%d x)) / 2", diff, sum),
			fmt.Sprintf("(\\sin(%d x) - \\sin(%d x)) / 2", sum, diff),
			fmt.Sprintf("2 \\sin(%d x) \\sin(%d x)", sum, diff),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The product-to-sum identity sin A sin B = (cos(A - B) - cos(A + B)) / 2 gives (cos(%d x) - cos(%d x)) / 2.", diff, sum)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	default:
		// Scaled: 2 sin(ax) cos(bx) = sin((a+b)x) + sin((a-b)x)
		q := fmt.Sprintf("Rewrite $2 \\sin(%d x) \\cos(%d x)$ in expanded sum form.", a, b)
		correct := fmt.Sprintf("\\sin(%d x) + \\sin(%d x)", sum, diff)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("\\cos(%d x) - \\cos(%d x)", diff, sum),
			fmt.Sprintf("\\sin(%d x) - \\sin(%d x)", sum, diff),
			fmt.Sprintf("\\cos(%d x) + \\cos(%d x)", diff, sum),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Multiplying the identity by 2 yields 2 sin A cos B = sin(A + B) + sin(A - B) = sin(%d x) + sin(%d x).", sum, diff)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	}
}

// ---------------------------------------------------------------------------
// 2. Sum to product (trig.ident.sum_to_product)
// ---------------------------------------------------------------------------

type sumToProductGen struct{}

func (g *sumToProductGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	// Pick a and b of the same parity so (a+b)/2 and (a-b)/2 are integers
	parity := rand.Intn(2) // 0 for even, 1 for odd
	m := 2 + rand.Intn(5)
	k := 1 + rand.Intn(m-1)
	a := m + k
	b := m - k
	if parity == 1 && a%2 == 0 {
		a++
		b++
		m = (a + b) / 2
		k = (a - b) / 2
	}

	switch rand.Intn(4) {
	case 0:
		// sin A + sin B = 2 sin((A+B)/2) cos((A-B)/2)
		q := fmt.Sprintf("Which expression rewrites $\\sin(%d x) + \\sin(%d x)$ as a product of trigonometric functions?", a, b)
		correct := fmt.Sprintf("2 \\sin(%d x) \\cos(%d x)", m, k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("2 \\cos(%d x) \\cos(%d x)", m, k),
			fmt.Sprintf("2 \\sin(%d x) \\sin(%d x)", m, k),
			fmt.Sprintf("-2 \\sin(%d x) \\cos(%d x)", m, k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The sum-to-product identity sin A + sin B = 2 sin((A + B)/2) cos((A - B)/2) gives 2 sin(%d x) cos(%d x).", m, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 1:
		// cos A + cos B = 2 cos((A+B)/2) cos((A-B)/2)
		q := fmt.Sprintf("Which expression rewrites $\\cos(%d x) + \\cos(%d x)$ as a product of trigonometric functions?", a, b)
		correct := fmt.Sprintf("2 \\cos(%d x) \\cos(%d x)", m, k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("2 \\sin(%d x) \\sin(%d x)", m, k),
			fmt.Sprintf("-2 \\sin(%d x) \\sin(%d x)", m, k),
			fmt.Sprintf("2 \\cos(%d x) \\sin(%d x)", m, k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The sum-to-product identity cos A + cos B = 2 cos((A + B)/2) cos((A - B)/2) gives 2 cos(%d x) cos(%d x).", m, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 2:
		// cos A - cos B = -2 sin((A+B)/2) sin((A-B)/2)
		q := fmt.Sprintf("Which expression rewrites $\\cos(%d x) - \\cos(%d x)$ as a product of trigonometric functions?", a, b)
		correct := fmt.Sprintf("-2 \\sin(%d x) \\sin(%d x)", m, k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("2 \\sin(%d x) \\sin(%d x)", m, k),
			fmt.Sprintf("-2 \\cos(%d x) \\cos(%d x)", m, k),
			fmt.Sprintf("2 \\cos(%d x) \\sin(%d x)", m, k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The sum-to-product identity cos A - cos B = -2 sin((A + B)/2) sin((A - B)/2) gives -2 sin(%d x) sin(%d x).", m, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	default:
		// sin A - sin B = 2 cos((A+B)/2) sin((A-B)/2)
		q := fmt.Sprintf("Which expression rewrites $\\sin(%d x) - \\sin(%d x)$ as a product of trigonometric functions?", a, b)
		correct := fmt.Sprintf("2 \\cos(%d x) \\sin(%d x)", m, k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("2 \\sin(%d x) \\cos(%d x)", m, k),
			fmt.Sprintf("-2 \\cos(%d x) \\sin(%d x)", m, k),
			fmt.Sprintf("2 \\cos(%d x) \\cos(%d x)", m, k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The sum-to-product identity sin A - sin B = 2 cos((A + B)/2) sin((A - B)/2) gives 2 cos(%d x) sin(%d x).", m, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	}
}

// ---------------------------------------------------------------------------
// 3. Pythagorean variants (trig.ident.pythagorean_variants)
// ---------------------------------------------------------------------------

type pythagoreanVariantsGen struct{}

func (g *pythagoreanVariantsGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	k := 2 + rand.Intn(7)
	switch rand.Intn(5) {
	case 0:
		// k sec^2(t) - k tan^2(t) = k
		q := fmt.Sprintf("What is the simplified constant value of $%d \\sec^2(t) - %d \\tan^2(t)$?", k, k)
		correct := fmt.Sprintf("%d", k)
		dist := distinctIdentDistractors(correct, "0", "1", fmt.Sprintf("-%d", k))
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Factoring gives %d(sec^2 t - tan^2 t). Since 1 + tan^2 t = sec^2 t, sec^2 t - tan^2 t = 1, so the expression equals %d.", k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 1:
		// k csc^2(t) - k cot^2(t) = k
		q := fmt.Sprintf("What is the simplified constant value of $%d \\csc^2(t) - %d \\cot^2(t)$?", k, k)
		correct := fmt.Sprintf("%d", k)
		dist := distinctIdentDistractors(correct, "0", "1", fmt.Sprintf("-%d", k))
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Factoring gives %d(csc^2 t - cot^2 t). Since 1 + cot^2 t = csc^2 t, csc^2 t - cot^2 t = 1, so the expression equals %d.", k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 2:
		// 1 + tan^2(k x) = sec^2(k x)
		q := fmt.Sprintf("Express $1 + \\tan^2(%d x)$ in terms of a single squared trigonometric function.", k)
		correct := fmt.Sprintf("\\sec^2(%d x)", k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("\\csc^2(%d x)", k),
			fmt.Sprintf("\\cot^2(%d x)", k),
			fmt.Sprintf("\\cos^2(%d x)", k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Dividing sin^2(%d x) + cos^2(%d x) = 1 by cos^2(%d x) yields tan^2(%d x) + 1 = sec^2(%d x).", k, k, k, k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 3:
		// 1 + cot^2(k x) = csc^2(k x)
		q := fmt.Sprintf("Express $1 + \\cot^2(%d x)$ in terms of a single squared trigonometric function.", k)
		correct := fmt.Sprintf("\\csc^2(%d x)", k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("\\sec^2(%d x)", k),
			fmt.Sprintf("\\tan^2(%d x)", k),
			fmt.Sprintf("\\sin^2(%d x)", k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Dividing sin^2(%d x) + cos^2(%d x) = 1 by sin^2(%d x) yields 1 + cot^2(%d x) = csc^2(%d x).", k, k, k, k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	default:
		// sec^2(kx) - 1 = tan^2(kx)
		q := fmt.Sprintf("Simplify $\\sec^2(%d x) - 1$.", k)
		correct := fmt.Sprintf("\\tan^2(%d x)", k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("\\cot^2(%d x)", k),
			fmt.Sprintf("\\cos^2(%d x)", k),
			"1",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("From the Pythagorean identity 1 + tan^2(%d x) = sec^2(%d x), subtracting 1 yields tan^2(%d x).", k, k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	}
}

// ---------------------------------------------------------------------------
// 4. General solutions (trig.eq.general_solution)
// ---------------------------------------------------------------------------

type generalSolutionGen struct{}

func (g *generalSolutionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type spec struct {
		fn      string
		val     string
		correct string
		dist1   string
		dist2   string
		dist3   string
		reason  string
	}

	specs := []spec{
		{
			fn:      "\\sin t",
			val:     "1/2",
			correct: "t = 30 + 360k or t = 150 + 360k",
			dist1:   "t = 30 + 180k",
			dist2:   "t = 150 + 360k only",
			dist3:   "t = 30 + 360k only",
			reason:  "Sine takes value 1/2 at 30 and its reflection 150 in the first rotation, each repeating every full 360 turn.",
		},
		{
			fn:      "\\sin t",
			val:     "\\sqrt{2}/2",
			correct: "t = 45 + 360k or t = 135 + 360k",
			dist1:   "t = 45 + 180k",
			dist2:   "t = 45 + 360k only",
			dist3:   "t = 135 + 360k only",
			reason:  "Sine equals sqrt(2)/2 at 45 and 135 degrees, repeating every 360k.",
		},
		{
			fn:      "\\sin t",
			val:     "\\sqrt{3}/2",
			correct: "t = 60 + 360k or t = 120 + 360k",
			dist1:   "t = 60 + 180k",
			dist2:   "t = 60 + 360k only",
			dist3:   "t = 120 + 360k only",
			reason:  "Sine equals sqrt(3)/2 at 60 and 120 degrees, repeating every 360k.",
		},
		{
			fn:      "\\cos t",
			val:     "1/2",
			correct: "t = 60 + 360k or t = 300 + 360k",
			dist1:   "t = 60 + 180k",
			dist2:   "t = 120 + 360k or t = 240 + 360k",
			dist3:   "t = 60 + 360k only",
			reason:  "Cosine is symmetric about the horizontal axis, taking 1/2 at 60 and 300 (or -60) degrees, repeating every 360k.",
		},
		{
			fn:      "\\cos t",
			val:     "\\sqrt{2}/2",
			correct: "t = 45 + 360k or t = 315 + 360k",
			dist1:   "t = 45 + 180k",
			dist2:   "t = 135 + 360k or t = 225 + 360k",
			dist3:   "t = 45 + 360k only",
			reason:  "Cosine equals sqrt(2)/2 at 45 and 315 degrees, repeating every 360k.",
		},
		{
			fn:      "\\cos t",
			val:     "\\sqrt{3}/2",
			correct: "t = 30 + 360k or t = 330 + 360k",
			dist1:   "t = 30 + 180k",
			dist2:   "t = 150 + 360k or t = 210 + 360k",
			dist3:   "t = 30 + 360k only",
			reason:  "Cosine equals sqrt(3)/2 at 30 and 330 degrees, repeating every 360k.",
		},
		{
			fn:      "\\tan t",
			val:     "1",
			correct: "t = 45 + 180k",
			dist1:   "t = 45 + 360k",
			dist2:   "t = 45 + 360k or t = 225 + 360k",
			dist3:   "t = 135 + 180k",
			reason:  "Tangent has fundamental period 180 degrees, so t = 45 + 180k generates all solutions.",
		},
		{
			fn:      "\\tan t",
			val:     "\\sqrt{3}",
			correct: "t = 60 + 180k",
			dist1:   "t = 60 + 360k",
			dist2:   "t = 30 + 180k",
			dist3:   "t = 120 + 180k",
			reason:  "Tangent has fundamental period 180 degrees, taking value sqrt(3) at 60 + 180k.",
		},
		{
			fn:      "\\tan t",
			val:     "0",
			correct: "t = 180k",
			dist1:   "t = 360k",
			dist2:   "t = 90 + 180k",
			dist3:   "t = 90 + 360k",
			reason:  "Tangent is zero at integer multiples of 180 degrees.",
		},
		{
			fn:      "\\cos t",
			val:     "0",
			correct: "t = 90 + 180k",
			dist1:   "t = 180k",
			dist2:   "t = 90 + 360k",
			dist3:   "t = 270 + 360k",
			reason:  "Cosine vanishes at 90 and 270 degrees, which can be written concisely as 90 + 180k.",
		},
	}

	s := specs[rand.Intn(len(specs))]
	q := fmt.Sprintf("The general solution of $%s = %s$, in degrees (where $k \\in \\mathbb{Z}$), is...", s.fn, s.val)
	dist := distinctIdentDistractors(s.correct, s.dist1, s.dist2, s.dist3)
	p := generator.MultipleChoice(q, s.correct, dist...)
	p.Explanation = s.reason
	if p.Facts == nil {
		p.Facts = map[string]string{}
	}
	p.Facts["answer"] = s.correct
	return p
}

// ---------------------------------------------------------------------------
// 5. Advanced inequalities (trig.ineq.advanced)
// ---------------------------------------------------------------------------

type advancedIneqGen struct{}

func (g *advancedIneqGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type ineqSpec struct {
		ineq    string
		domain  string
		correct string
		dist1   string
		dist2   string
		dist3   string
		reason  string
	}

	specs := []ineqSpec{
		{
			ineq:    "\\sin t > 1/2",
			domain:  "0 \\le t < 360",
			correct: "30 < t < 150",
			dist1:   "0 < t < 30 or 150 < t < 360",
			dist2:   "30 < t < 330",
			dist3:   "150 < t < 360",
			reason:  "Sine exceeds 1/2 from 30 to 150 degrees in the upper half of the circle.",
		},
		{
			ineq:    "\\sin t > \\sqrt{2}/2",
			domain:  "0 \\le t < 360",
			correct: "45 < t < 135",
			dist1:   "0 < t < 45 or 135 < t < 360",
			dist2:   "45 < t < 315",
			dist3:   "135 < t < 360",
			reason:  "Sine exceeds sqrt(2)/2 in the first rotation between 45 and 135 degrees.",
		},
		{
			ineq:    "\\sin t > \\sqrt{3}/2",
			domain:  "0 \\le t < 360",
			correct: "60 < t < 120",
			dist1:   "0 < t < 60 or 120 < t < 360",
			dist2:   "60 < t < 300",
			dist3:   "120 < t < 360",
			reason:  "Sine exceeds sqrt(3)/2 between 60 and 120 degrees.",
		},
		{
			ineq:    "\\cos t < 0",
			domain:  "0 \\le t < 360",
			correct: "90 < t < 270",
			dist1:   "0 < t < 90 or 270 < t < 360",
			dist2:   "270 < t < 360",
			dist3:   "0 < t < 180",
			reason:  "Cosine is negative in the second and third quadrants, running from 90 to 270 degrees.",
		},
		{
			ineq:    "\\cos t > 1/2",
			domain:  "0 \\le t < 360",
			correct: "0 <= t < 60 or 300 < t < 360",
			dist1:   "60 < t < 300",
			dist2:   "0 < t < 60",
			dist3:   "300 < t < 360",
			reason:  "Cosine exceeds 1/2 in quadrant I (0 to 60) and quadrant IV (300 to 360).",
		},
		{
			ineq:    "\\cos t > \\sqrt{2}/2",
			domain:  "0 \\le t < 360",
			correct: "0 <= t < 45 or 315 < t < 360",
			dist1:   "45 < t < 315",
			dist2:   "0 < t < 45",
			dist3:   "315 < t < 360",
			reason:  "Cosine exceeds sqrt(2)/2 in quadrant I (0 to 45) and quadrant IV (315 to 360).",
		},
		{
			ineq:    "\\tan t \\ge 1",
			domain:  "0 \\le t < 180",
			correct: "45 <= t < 90",
			dist1:   "0 <= t <= 45",
			dist2:   "45 <= t <= 90",
			dist3:   "90 < t < 180",
			reason:  "Tangent reaches 1 at 45 and grows without bound as t approaches the vertical asymptote at 90, which is excluded.",
		},
		{
			ineq:    "\\tan t \\ge \\sqrt{3}",
			domain:  "0 \\le t < 180",
			correct: "60 <= t < 90",
			dist1:   "0 <= t <= 60",
			dist2:   "60 <= t <= 90",
			dist3:   "90 < t < 180",
			reason:  "Tangent reaches sqrt(3) at 60 and remains above it until the vertical asymptote at 90.",
		},
	}

	s := specs[rand.Intn(len(specs))]
	q := fmt.Sprintf("The solution of $%s$ in the interval $%s$ (in degrees) is...", s.ineq, s.domain)
	dist := distinctIdentDistractors(s.correct, s.dist1, s.dist2, s.dist3)
	p := generator.MultipleChoice(q, s.correct, dist...)
	p.Explanation = s.reason
	if p.Facts == nil {
		p.Facts = map[string]string{}
	}
	p.Facts["answer"] = s.correct
	return p
}

// ---------------------------------------------------------------------------
// 6. Hyperbolic identities (trig.hyperbolic.identities)
// ---------------------------------------------------------------------------

type hyperbolicIdentGen struct{}

func (g *hyperbolicIdentGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	k := 2 + rand.Intn(7)

	switch rand.Intn(5) {
	case 0:
		// k cosh^2(x) - k sinh^2(x) = k
		q := fmt.Sprintf("What is the simplified constant value of $%d \\cosh^2(x) - %d \\sinh^2(x)$ for all $x$?", k, k)
		correct := fmt.Sprintf("%d", k)
		dist := distinctIdentDistractors(correct, "0", "1", fmt.Sprintf("-%d", k))
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Factoring gives %d(cosh^2 x - sinh^2 x). By the fundamental hyperbolic identity cosh^2 x - sinh^2 x = 1, this equals %d.", k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 1:
		// 1 - tanh^2(x) = sech^2(x)
		q := fmt.Sprintf("What is the simplified form of $%d - %d \\tanh^2(x)$?", k, k)
		correct := fmt.Sprintf("%d \\operatorname{sech}^2(x)", k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("%d \\operatorname{csch}^2(x)", k),
			fmt.Sprintf("%d \\coth^2(x)", k),
			fmt.Sprintf("%d + %d \\tanh^2(x)", k, k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Dividing cosh^2 x - sinh^2 x = 1 by cosh^2 x gives 1 - tanh^2 x = sech^2 x, so multiplying by %d yields %d sech^2 x.", k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 2:
		// Double argument: cosh(2x)
		q := "Which identity expresses $\\cosh(2x)$ in terms of $\\cosh(x)$ and $\\sinh(x)$?"
		correct := "\\cosh(2x) = \\cosh^2(x) + \\sinh^2(x)"
		dist := distinctIdentDistractors(correct,
			"\\cosh(2x) = \\cosh^2(x) - \\sinh^2(x)",
			"\\cosh(2x) = 2 \\sinh(x) \\cosh(x)",
			"\\cosh(2x) = 1 - 2 \\cosh^2(x)",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The hyperbolic double-argument formula carries a plus sign, cosh(2x) = cosh^2 x + sinh^2 x, reflecting the geometry of the hyperbola."
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 3:
		// sinh(2x) = 2 sinh x cosh x
		q := "Which identity expresses $\\sinh(2x)$ in terms of $\\sinh(x)$ and $\\cosh(x)$?"
		correct := "\\sinh(2x) = 2 \\sinh(x) \\cosh(x)"
		dist := distinctIdentDistractors(correct,
			"\\sinh(2x) = \\cosh^2(x) + \\sinh^2(x)",
			"\\sinh(2x) = \\cosh^2(x) - \\sinh^2(x)",
			"\\sinh(2x) = \\sinh(x) \\cosh(x)",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The hyperbolic sine of a double angle is twice the product of the hyperbolic sine and cosine: sinh(2x) = 2 sinh x cosh x."
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	default:
		// (cosh x + sinh x)^n = cosh(nx) + sinh(nx)
		n := 2 + rand.Intn(4)
		q := fmt.Sprintf("Using $\\cosh(x) + \\sinh(x) = e^x$, what is $(\\cosh(x) + \\sinh(x))^{%d}$ in terms of hyperbolic functions of $%d x$?", n, n)
		correct := fmt.Sprintf("\\cosh(%d x) + \\sinh(%d x)", n, n)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("\\cosh^{%d}(x) + \\sinh^{%d}(x)", n, n),
			fmt.Sprintf("\\cosh(%d x) - \\sinh(%d x)", n, n),
			fmt.Sprintf("%d \\cosh(x) + %d \\sinh(x)", n, n),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("(cosh x + sinh x)^%d = (e^x)^%d = e^(%d x) = cosh(%d x) + sinh(%d x).", n, n, n, n, n)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	}
}

// ---------------------------------------------------------------------------
// 7. Hyperbolic derivatives (trig.hyperbolic.derivative)
// ---------------------------------------------------------------------------

type hyperbolicDerivGen struct{}

func (g *hyperbolicDerivGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	k := 2 + rand.Intn(6)

	switch rand.Intn(5) {
	case 0:
		// d/dx [sinh(kx)] = k cosh(kx)
		q := fmt.Sprintf("What is the derivative $\\frac{d}{dx}[\\sinh(%d x)]$?", k)
		correct := fmt.Sprintf("%d \\cosh(%d x)", k, k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("-%d \\cosh(%d x)", k, k),
			fmt.Sprintf("%d \\sinh(%d x)", k, k),
			fmt.Sprintf("\\cosh(%d x)", k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("By the chain rule, d/dx[sinh(%d x)] = %d cosh(%d x) with a positive sign.", k, k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 1:
		// d/dx [cosh(kx)] = k sinh(kx)
		q := fmt.Sprintf("What is the derivative $\\frac{d}{dx}[\\cosh(%d x)]$?", k)
		correct := fmt.Sprintf("%d \\sinh(%d x)", k, k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("-%d \\sinh(%d x)", k, k),
			fmt.Sprintf("%d \\cosh(%d x)", k, k),
			fmt.Sprintf("\\sinh(%d x)", k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("By the chain rule, d/dx[cosh(%d x)] = %d sinh(%d x) with a positive sign, unlike circular cosine which takes a minus.", k, k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 2:
		// d/dx [tanh(kx)] = k sech^2(kx)
		q := fmt.Sprintf("What is the derivative $\\frac{d}{dx}[\\tanh(%d x)]$?", k)
		correct := fmt.Sprintf("%d \\operatorname{sech}^2(%d x)", k, k)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("-%d \\operatorname{sech}^2(%d x)", k, k),
			fmt.Sprintf("%d \\operatorname{csch}^2(%d x)", k, k),
			fmt.Sprintf("\\operatorname{sech}^2(%d x)", k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Differentiating tanh(%d x) gives %d sech^2(%d x).", k, k, k)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	case 3:
		// Chain rule on power: d/dx [sinh^n(x)]
		n := 2 + rand.Intn(4)
		q := fmt.Sprintf("What is the derivative $\\frac{d}{dx}[\\sinh^{%d}(x)]$?", n)
		correct := fmt.Sprintf("%d \\sinh^{%d}(x) \\cosh(x)", n, n-1)
		dist := distinctIdentDistractors(correct,
			fmt.Sprintf("-%d \\sinh^{%d}(x) \\cosh(x)", n, n-1),
			fmt.Sprintf("%d \\cosh^{%d}(x)", n, n),
			fmt.Sprintf("\\sinh^{%d}(x)", n-1),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Applying the power rule and chain rule gives %d sinh^(%d)(x) cosh(x).", n, n-1)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	default:
		// Higher order derivative of sinh or cosh
		derivOrder := 2 + rand.Intn(6)
		isEven := derivOrder%2 == 0
		q := fmt.Sprintf("What is the %d-th derivative $\\frac{d^{%d}}{dx^{%d}}[\\sinh(x)]$?", derivOrder, derivOrder, derivOrder)
		var correct string
		var dist []string
		if isEven {
			correct = "\\sinh(x)"
			dist = distinctIdentDistractors(correct, "\\cosh(x)", "-\\sinh(x)", "-\\cosh(x)")
		} else {
			correct = "\\cosh(x)"
			dist = distinctIdentDistractors(correct, "\\sinh(x)", "-\\cosh(x)", "-\\sinh(x)")
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Since d/dx[sinh x] = cosh x and d/dx[cosh x] = sinh x without sign changes, the derivatives alternate between sinh and cosh with period 2. Because %d is %s, the result is %s.",
			derivOrder, strconv.FormatBool(isEven), correct)
		if p.Facts == nil {
			p.Facts = map[string]string{}
		}
		p.Facts["answer"] = correct
		return p
	}
}
