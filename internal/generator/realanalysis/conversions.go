package realanalysis

// Conversions of real analysis multiple-choice concepts to algorithmic,
// computed numeric generators.
//
// Every generator publishes the concrete parameters its answer follows from,
// so the fact verification test recomputes the answer independently.

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
)

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// Metric spaces (ra.metric.spaces)
// ---------------------------------------------------------------------------

type metricSpacesGen struct{}

func (g *metricSpacesGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)

	switch mode {
	case 0:
		// Euclidean metric d_2 on R^2 with exact integer distances via Pythagorean triples.
		type triple struct{ a, b, c int }
		triples := []triple{
			{3, 4, 5},
			{5, 12, 13},
			{8, 15, 17},
			{7, 24, 25},
			{9, 40, 41},
			{12, 35, 37},
			{20, 21, 29},
		}
		tr := triples[rand.Intn(len(triples))]
		k := rand.Intn(3) + 1
		dx := tr.a * k
		dy := tr.b * k
		if rand.Intn(2) == 1 {
			dx, dy = dy, dx
		}
		px := rand.Intn(21) - 10
		py := rand.Intn(21) - 10
		sx := 1
		if rand.Intn(2) == 1 {
			sx = -1
		}
		sy := 1
		if rand.Intn(2) == 1 {
			sy = -1
		}
		qx := px + sx*dx
		qy := py + sy*dy
		d := tr.c * k

		q := fmt.Sprintf("In the metric space (\\mathbb{R}^2, d_2) equipped with the standard Euclidean metric d_2(p, q) = \\sqrt{(p_1 - q_1)^2 + (p_2 - q_2)^2}, compute the distance between the points p = (%d, %d) and q = (%d, %d).",
			px, py, qx, qy)
		exp := fmt.Sprintf("The coordinate differences are %d - (%d) = %d and %d - (%d) = %d. Their squares are %d and %d, summing to %d. Taking the square root gives d_2(p, q) = \\sqrt{%d} = %d.",
			px, qx, px-qx, py, qy, py-qy, (px-qx)*(px-qx), (py-qy)*(py-qy), d*d, d*d, d)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(d),
			Explanation: exp,
			Facts: map[string]string{
				"metric": "d_2",
				"px":     strconv.Itoa(px),
				"py":     strconv.Itoa(py),
				"qx":     strconv.Itoa(qx),
				"qy":     strconv.Itoa(qy),
				"answer": strconv.Itoa(d),
			},
		}

	case 1:
		// Taxicab / Manhattan metric d_1 on R^3.
		px, py, pz := rand.Intn(21)-10, rand.Intn(21)-10, rand.Intn(21)-10
		qx, qy, qz := rand.Intn(21)-10, rand.Intn(21)-10, rand.Intn(21)-10
		if px == qx && py == qy && pz == qz {
			qx += 3
		}
		dx := intAbs(px - qx)
		dy := intAbs(py - qy)
		dz := intAbs(pz - qz)
		d1 := dx + dy + dz

		q := fmt.Sprintf("In the metric space (\\mathbb{R}^3, d_1) equipped with the taxicab (Manhattan) metric d_1(p, q) = |p_1 - q_1| + |p_2 - q_2| + |p_3 - q_3|, compute the distance between the points p = (%d, %d, %d) and q = (%d, %d, %d).",
			px, py, pz, qx, qy, qz)
		exp := fmt.Sprintf("The absolute coordinate differences are |%d - (%d)| = %d, |%d - (%d)| = %d, and |%d - (%d)| = %d. Summing these distances gives d_1(p, q) = %d + %d + %d = %d.",
			px, qx, dx, py, qy, dy, pz, qz, dz, dx, dy, dz, d1)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(d1),
			Explanation: exp,
			Facts: map[string]string{
				"metric": "d_1",
				"px":     strconv.Itoa(px),
				"py":     strconv.Itoa(py),
				"pz":     strconv.Itoa(pz),
				"qx":     strconv.Itoa(qx),
				"qy":     strconv.Itoa(qy),
				"qz":     strconv.Itoa(qz),
				"answer": strconv.Itoa(d1),
			},
		}

	case 2:
		// Chebyshev / supremum metric d_infty on R^3.
		px, py, pz := rand.Intn(31)-15, rand.Intn(31)-15, rand.Intn(31)-15
		qx, qy, qz := rand.Intn(31)-15, rand.Intn(31)-15, rand.Intn(31)-15
		if px == qx && py == qy && pz == qz {
			qx += 4
		}
		dx := intAbs(px - qx)
		dy := intAbs(py - qy)
		dz := intAbs(pz - qz)
		dInf := intMax(dx, intMax(dy, dz))

		q := fmt.Sprintf("In the metric space (\\mathbb{R}^3, d_\\infty) equipped with the Chebyshev (maximum) metric d_\\infty(p, q) = \\max(|p_1 - q_1|, |p_2 - q_2|, |p_3 - q_3|), compute the distance between the points p = (%d, %d, %d) and q = (%d, %d, %d).",
			px, py, pz, qx, qy, qz)
		exp := fmt.Sprintf("The absolute coordinate differences are |%d - (%d)| = %d, |%d - (%d)| = %d, and |%d - (%d)| = %d. Taking the maximum gives d_\\infty(p, q) = \\max(%d, %d, %d) = %d.",
			px, qx, dx, py, qy, dy, pz, qz, dz, dx, dy, dz, dInf)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(dInf),
			Explanation: exp,
			Facts: map[string]string{
				"metric": "d_infty",
				"px":     strconv.Itoa(px),
				"py":     strconv.Itoa(py),
				"pz":     strconv.Itoa(pz),
				"qx":     strconv.Itoa(qx),
				"qy":     strconv.Itoa(qy),
				"qz":     strconv.Itoa(qz),
				"answer": strconv.Itoa(dInf),
			},
		}

	case 3:
		// Diameter of a finite subset S = {p1, p2, p3} under d_1.
		p1x, p1y := rand.Intn(17)-8, rand.Intn(17)-8
		p2x, p2y := rand.Intn(17)-8, rand.Intn(17)-8
		p3x, p3y := rand.Intn(17)-8, rand.Intn(17)-8
		d12 := intAbs(p1x-p2x) + intAbs(p1y-p2y)
		d13 := intAbs(p1x-p3x) + intAbs(p1y-p3y)
		d23 := intAbs(p2x-p3x) + intAbs(p2y-p3y)
		diam := intMax(d12, intMax(d13, d23))
		if diam == 0 {
			p2x += 2
			d12 = intAbs(p1x-p2x) + intAbs(p1y-p2y)
			d23 = intAbs(p2x-p3x) + intAbs(p2y-p3y)
			diam = intMax(d12, intMax(d13, d23))
		}

		q := fmt.Sprintf("Let S = {p_1, p_2, p_3} \\subset \\mathbb{R}^2 be the three-point subset with p_1 = (%d, %d), p_2 = (%d, %d), and p_3 = (%d, %d). What is the diameter \\operatorname{diam}(S) = \\max_{u, v \\in S} d_1(u, v) under the taxicab metric d_1(u, v) = |u_1 - v_1| + |u_2 - v_2|?",
			p1x, p1y, p2x, p2y, p3x, p3y)
		exp := fmt.Sprintf("The pairwise taxicab distances are d_1(p_1, p_2) = %d, d_1(p_1, p_3) = %d, and d_1(p_2, p_3) = %d. The diameter is the maximum pairwise distance: \\operatorname{diam}(S) = \\max(%d, %d, %d) = %d.",
			d12, d13, d23, d12, d13, d23, diam)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(diam),
			Explanation: exp,
			Facts: map[string]string{
				"metric": "diameter_d1",
				"p1x":    strconv.Itoa(p1x),
				"p1y":    strconv.Itoa(p1y),
				"p2x":    strconv.Itoa(p2x),
				"p2y":    strconv.Itoa(p2y),
				"p3x":    strconv.Itoa(p3x),
				"p3y":    strconv.Itoa(p3y),
				"answer": strconv.Itoa(diam),
			},
		}

	default:
		// Open balls in a discrete metric space.
		n := rand.Intn(20) + 5
		r := 1
		if rand.Intn(2) == 1 {
			r = 2
		}

		if r == 1 {
			q := fmt.Sprintf("Let (X, d) be a discrete metric space with |X| = %d points, where d(x, y) = 1 for x \\neq y and d(x, x) = 0. For any point x \\in X, what is the cardinality |B_1(x)| of the open ball B_1(x) = {y \\in X : d(x, y) < 1}?", n)
			exp := "In the discrete metric, any point distinct from x has distance d(y, x) = 1. Because the open ball of radius 1 requires d(y, x) < 1, only x itself qualifies (since d(x, x) = 0 < 1). Therefore B_1(x) = {x}, which has cardinality 1."
			return generator.Problem{
				Question:    q,
				Answer:      "1",
				Explanation: exp,
				Facts: map[string]string{
					"metric":   "discrete_ball",
					"set_size": strconv.Itoa(n),
					"radius":   "1",
					"answer":   "1",
				},
			}
		}

		q := fmt.Sprintf("Let (X, d) be a discrete metric space with |X| = %d points, where d(x, y) = 1 for x \\neq y and d(x, x) = 0. For any point x \\in X, what is the cardinality |B_2(x)| of the open ball B_2(x) = {y \\in X : d(x, y) < 2}?", n)
		exp := fmt.Sprintf("In the discrete metric, every distance satisfies d(y, x) \\le 1 < 2. Thus every point in X belongs to the open ball B_2(x) = X, which has cardinality |X| = %d.", n)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(n),
			Explanation: exp,
			Facts: map[string]string{
				"metric":   "discrete_ball",
				"set_size": strconv.Itoa(n),
				"radius":   "2",
				"answer":   strconv.Itoa(n),
			},
		}
	}
}

// registerConversions registers the concepts converted from multiple choice.
func registerConversions(reg *generator.Registry) {
	reg.Register("ra.metric.spaces", &metricSpacesGen{})
}
