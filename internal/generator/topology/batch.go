package topology

// The 2026-10-18 topology batch: twelve concepts that take the domain from 30
// to the target of 42.
//
// The survey found that the domain jumps from three basics straight to named
// theorems (Urysohn, Tietze, Stone-Cech, Van Kampen, de Rham) with almost no
// computational middle, and that 29 of its 30 concepts are multiple choice.
// The batch therefore builds the metric-space toolkit the later lessons assume
// (balls, interior and boundary, limit points, convergence, completeness,
// bases, subspaces, dense sets) and adds the computable invariants the
// algebraic-topology lessons describe but never name (homeomorphism, Euler
// characteristic, genus, Betti numbers).
//
// Six of the twelve are numeric, and every numeric generator publishes the
// facts its answer follows from, so the tests recompute the answer instead of
// comparing a string with itself. The six recognition concepts use
// generator.MultipleChoice, so every distractor is a real option the router
// must reject. The new `topology.invariants` subdomain holds the four invariant
// concepts; the others join the existing `basics` and `constructions` homes.

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
)

func registerBatch(reg *generator.Registry) {
	reg.Register("topo.metric.ball", &ballGen{})
	reg.Register("topo.metric.interior", &interiorGen{})
	reg.Register("topo.metric.limit_point", &limitPointGen{})
	reg.Register("topo.metric.convergence", &convergenceGen{})
	reg.Register("topo.metric.complete", &completeGen{})
	reg.Register("topo.metric.basis", &basisGen{})
	reg.Register("topo.metric.subspace", &subspaceGen{})
	reg.Register("topo.metric.dense", &denseGen{})
	reg.Register("topo.invariant.homeomorphism", &homeomorphismGen{})
	reg.Register("topo.invariant.euler", &eulerGen{})
	reg.Register("topo.invariant.genus", &genusGen{})
	reg.Register("topo.invariant.betti", &bettiGen{})
}

// ---------------------------------------------------------------------------
// 1. Open balls (numeric)
// ---------------------------------------------------------------------------

type ballGen struct{}

func (g *ballGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	if rand.Intn(2) == 0 {
		triples := [][3]int{
			{3, 4, 5}, {5, 12, 13}, {8, 15, 17}, {7, 24, 25}, {6, 8, 10},
		}
		for attempt := 0; attempt < 50; attempt++ {
			tr := triples[rand.Intn(len(triples))]
			dx, dy, dist := tr[0], tr[1], tr[2]
			x1 := rand.Intn(15) - 7
			y1 := rand.Intn(15) - 7
			x2 := x1 + dx
			y2 := y1 + dy
			if dist == x1 || dist == y1 || dist == x2 || dist == y2 ||
				-dist == x1 || -dist == y1 || -dist == x2 || -dist == y2 {
				continue
			}
			ans := strconv.Itoa(dist)
			q := fmt.Sprintf("In the Euclidean plane, what is the distance d((%d,%d),(%d,%d))? (enter a number)", x1, y1, x2, y2)
			reason := fmt.Sprintf("The differences are %d and %d, so the distance is sqrt(%d^2 + %d^2) = %d.", dx, dy, dx, dy, dist)
			return generator.Problem{
				Question:    q,
				Answer:      ans,
				Explanation: reason,
				Facts: map[string]string{
					"kind":   "distance",
					"x1":     strconv.Itoa(x1),
					"y1":     strconv.Itoa(y1),
					"x2":     strconv.Itoa(x2),
					"y2":     strconv.Itoa(y2),
					"answer": ans,
				},
			}
		}
	}
	for attempt := 0; attempt < 50; attempt++ {
		center := rand.Intn(30) - 15
		radius := 1 + rand.Intn(15)
		side := "lower"
		var val int
		var boundName string
		if rand.Intn(2) == 1 {
			side = "upper"
			val = center + radius
			boundName = "b"
		} else {
			val = center - radius
			boundName = "a"
		}
		if val == center || val == radius || val == -center || val == -radius {
			continue
		}
		ans := strconv.Itoa(val)
		q := fmt.Sprintf("In R with the usual metric, the open ball of radius %d centred at %d is the interval (a, b). What is %s? (enter a number)", radius, center, boundName)
		reason := fmt.Sprintf("The ball is (%d - %d, %d + %d) = (%d, %d), so the %s endpoint is %d.", center, radius, center, radius, center-radius, center+radius, side, val)
		return generator.Problem{
			Question:    q,
			Answer:      ans,
			Explanation: reason,
			Facts: map[string]string{
				"kind":   "endpoint",
				"center": strconv.Itoa(center),
				"radius": strconv.Itoa(radius),
				"side":   side,
				"answer": ans,
			},
		}
	}
	return generator.Problem{
		Question:    "In R with the usual metric, the open ball of radius 3 centred at 5 is the interval (a, b). What is b? (enter a number)",
		Answer:      "8",
		Explanation: "5 + 3 = 8.",
		Facts: map[string]string{
			"kind":   "endpoint",
			"center": "5",
			"radius": "3",
			"side":   "upper",
			"answer": "8",
		},
	}
}

// ---------------------------------------------------------------------------
// 2. Interior, closure and boundary (numeric)
// ---------------------------------------------------------------------------

type interiorGen struct{}

func (g *interiorGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	for attempt := 0; attempt < 50; attempt++ {
		mode := rand.Intn(5)
		var q, ans, reason string
		var pts []int
		kind := "boundary"
		base := rand.Intn(20) - 10

		switch mode {
		case 0:
			// Disjoint closed intervals: 1 to 6 intervals => 2, 4, 6, 8, 10, 12 boundary points
			numIntervals := 1 + rand.Intn(6)
			ansVal := 2 * numIntervals
			ans = strconv.Itoa(ansVal)
			intervals := make([]string, numIntervals)
			cur := base
			pts = make([]int, 0, ansVal)
			leak := false
			for i := 0; i < numIntervals; i++ {
				a := cur
				b := a + rand.Intn(3) + 1
				cur = b + rand.Intn(3) + 2
				intervals[i] = fmt.Sprintf("[%d, %d]", a, b)
				pts = append(pts, a, b)
				if a == ansVal || b == ansVal || -a == ansVal || -b == ansVal {
					leak = true
				}
			}
			if leak {
				continue
			}
			q = fmt.Sprintf("In R with the usual topology, how many boundary points does the set %s have? (enter a number)", strings.Join(intervals, " union "))
			reason = fmt.Sprintf("The boundary consists of the %d endpoints of the pairwise disjoint closed intervals.", ansVal)
		case 1:
			// Adjacent intervals missing division points: 2, 3, 4 intervals touching => 3, 4, 5 boundary points
			m := 2 + rand.Intn(4)
			ansVal := m + 1
			ans = strconv.Itoa(ansVal)
			intervals := make([]string, m)
			cur := base
			pts = make([]int, 0, ansVal)
			leak := false
			for i := 0; i < m; i++ {
				next := cur + rand.Intn(4) + 1
				intervals[i] = fmt.Sprintf("(%d, %d)", cur, next)
				if i == 0 {
					pts = append(pts, cur)
					if cur == ansVal || -cur == ansVal {
						leak = true
					}
				}
				pts = append(pts, next)
				if next == ansVal || -next == ansVal {
					leak = true
				}
				cur = next
			}
			if leak {
				continue
			}
			q = fmt.Sprintf("In R with the usual topology, how many boundary points does the set %s have? (enter a number)", strings.Join(intervals, " union "))
			reason = fmt.Sprintf("The boundary consists of the %d outer and division boundary points.", ansVal)
		case 2:
			// Ray (a, infinity) or (-infinity, a) has 1 boundary point
			a := base
			if a == 1 || a == -1 {
				continue
			}
			pts = []int{a}
			ans = "1"
			if rand.Intn(2) == 0 {
				q = fmt.Sprintf("In R with the usual topology, how many boundary points does the ray (%d, infinity) have? (enter a number)", a)
			} else {
				q = fmt.Sprintf("In R with the usual topology, how many boundary points does the ray (-infinity, %d) have? (enter a number)", a)
			}
			reason = fmt.Sprintf("The only boundary point is %d.", a)
		case 3:
			// Closed ray [a, infinity) has 1 boundary point
			a := base
			if a == 1 || a == -1 {
				continue
			}
			pts = []int{a}
			ans = "1"
			q = fmt.Sprintf("In R with the usual topology, how many boundary points does the closed ray [%d, infinity) have? (enter a number)", a)
			reason = fmt.Sprintf("The only boundary point is %d.", a)
		default:
			// Finite set of isolated points has 0 interior points
			kind = "interior"
			pts = []int{}
			ans = "0"
			n := 2 + rand.Intn(4)
			elems := make([]string, n)
			for i := 0; i < n; i++ {
				elems[i] = strconv.Itoa(base + i*2)
			}
			q = fmt.Sprintf("In R with the usual topology, how many interior points does the set {%s} have? (enter a number)", strings.Join(elems, ", "))
			reason = "A finite set of isolated points has empty interior: no ball around any point stays inside the set."
		}

		parts := make([]string, len(pts))
		for i, p := range pts {
			parts[i] = strconv.Itoa(p)
		}
		return generator.Problem{
			Question:    q,
			Answer:      ans,
			Explanation: reason,
			Facts: map[string]string{
				"points": strings.Join(parts, ","), "kind": kind, "answer": ans,
			},
		}
	}
	return generator.Problem{
		Question:    "In R with the usual topology, how many boundary points does the set [1, 2] union [4, 5] have? (enter a number)",
		Answer:      "4",
		Explanation: "The four endpoints are boundary points.",
		Facts: map[string]string{
			"points": "1,2,4,5", "kind": "boundary", "answer": "4",
		},
	}
}

// ---------------------------------------------------------------------------
// 3. Limit points
// ---------------------------------------------------------------------------

type limitPointGen struct{}

func (g *limitPointGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "In R with the usual topology, which point is a limit point of the set {1/n : n at least 1}?",
			correct:     "0",
			distractors: []string{"1", "1/2", "2"},
			reason:      "The sequence 1/n converges to 0, and every neighbourhood of 0 contains some 1/n, so 0 is a limit point.",
		},
		{
			question:    "Is every point of the closed interval [0, 1] a limit point of [0, 1]? (yes/no)",
			correct:     "yes",
			distractors: []string{"no", "only the endpoints", "only the interior points"},
			reason:      "Every neighbourhood of any point of [0, 1] meets the interval in infinitely many other points, endpoints included.",
		},
		{
			question:    "Which subset of R has no limit points?",
			correct:     "the integers",
			distractors: []string{"the rationals", "the open interval (0, 1)", "the set {1/n : n at least 1}"},
			reason:      "Each integer has a ball of radius 1/2 containing no other integer, so no point is a limit of the integers.",
		},
		{
			question:    "A point x is a limit point of a set A exactly when...",
			correct:     "every neighbourhood of x meets A at a point other than x",
			distractors: []string{"x belongs to A", "A contains a neighbourhood of x", "x is an endpoint of A"},
			reason:      "The definition is about arbitrary closeness, not membership: the point x itself may or may not lie in A.",
		},
		{
			question:    "In the subspace {0} union {1/n : n at least 1} of R, how many limit points does the subspace have?",
			correct:     "1",
			distractors: []string{"0", "2", "infinitely many"},
			reason:      "The point 0 is the only limit point; each 1/n is isolated because a small ball around it misses the rest of the set.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 4. Convergence (numeric)
// ---------------------------------------------------------------------------

type convergenceGen struct{}

func (g *convergenceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	cPool := []int{2, 4, 5, 8, 10}
	for attempt := 0; attempt < 50; attempt++ {
		c := cPool[rand.Intn(len(cPool))]
		a := rand.Intn(16) + 1
		b := rand.Intn(15) + 1
		if rand.Intn(2) == 1 {
			b = -b
		}
		d := rand.Intn(10) + 1

		var ans string
		if a%c == 0 {
			ans = strconv.Itoa(a / c)
		} else {
			ans = strconv.FormatFloat(float64(a)/float64(c), 'f', -1, 64)
		}

		if ans == strconv.Itoa(a) || ans == strconv.Itoa(c) ||
			ans == strconv.Itoa(b) || ans == strconv.Itoa(-b) ||
			ans == strconv.Itoa(d) {
			continue
		}

		var numStr string
		if b > 0 {
			numStr = fmt.Sprintf("%dn + %d", a, b)
		} else {
			numStr = fmt.Sprintf("%dn - %d", a, -b)
		}
		denStr := fmt.Sprintf("%dn + %d", c, d)

		q := fmt.Sprintf("The sequence a_n = (%s)/(%s) converges to what limit? (enter a number)", numStr, denStr)
		reason := fmt.Sprintf("Divide numerator and denominator by n: the leading term dominates, leaving the limit %d/%d = %s.", a, c, ans)
		return generator.Problem{
			Question:    q,
			Answer:      ans,
			Explanation: reason,
			Facts: map[string]string{
				"a": strconv.Itoa(a), "b": strconv.Itoa(b),
				"c": strconv.Itoa(c), "d": strconv.Itoa(d), "answer": ans,
			},
		}
	}
	return generator.Problem{
		Question:    "The sequence a_n = (6n + 1)/(2n + 3) converges to what limit? (enter a number)",
		Answer:      "3",
		Explanation: "The leading coefficients give 6/2 = 3.",
		Facts: map[string]string{
			"a": "6", "b": "1", "c": "2", "d": "3", "answer": "3",
		},
	}
}

// ---------------------------------------------------------------------------
// 5. Completeness and Cauchy sequences
// ---------------------------------------------------------------------------

type completeGen struct{}

func (g *completeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Which metric space is complete?",
			correct:     "the real line with the usual metric",
			distractors: []string{"the rationals with the usual metric", "the open interval (0, 1) with the usual metric", "the rationals in [0, 1] with the usual metric"},
			reason:      "Every Cauchy sequence of reals converges to a real; the other spaces lose limits such as sqrt(2) or 0.",
		},
		{
			question:    "A Cauchy sequence in a complete metric space...",
			correct:     "converges to a point of the space",
			distractors: []string{"has a convergent subsequence only", "must be eventually constant", "must diverge"},
			reason:      "Completeness is exactly the statement that every Cauchy sequence has a limit inside the space.",
		},
		{
			question:    "The completion of the rationals with the usual metric is...",
			correct:     "the real line",
			distractors: []string{"the rationals again", "the complex plane", "the empty set"},
			reason:      "Filling the missing limits of Cauchy sequences of rationals produces the reals.",
		},
		{
			question:    "Every compact metric space is...",
			correct:     "complete",
			distractors: []string{"discrete", "not Hausdorff", "unbounded"},
			reason:      "A Cauchy sequence in a compact space has a convergent subsequence, and the whole Cauchy sequence then converges to the same limit.",
		},
		{
			question:    "Which sequence in (0, 1) converges to 0, showing that (0, 1) is not complete?",
			correct:     "1/n",
			distractors: []string{"1/2, 1/2, 1/2, ...", "1/2 + 1/n", "1 - 1/n"},
			reason:      "The sequence 1/n is Cauchy but its only limit is 0, which is not in (0, 1); the other options converge inside or to 1.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 6. Bases and subbases
// ---------------------------------------------------------------------------

type basisGen struct{}

func (g *basisGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(6)
	switch mode {
	case 0:
		a := rand.Intn(10) - 5
		b := a + 1 + rand.Intn(6)
		q := fmt.Sprintf("Let \\mathcal{S} = \\{(-\\infty, c), (c, \\infty) \\mid c \\in \\mathbb{R}\\} be the standard subbasis of open rays for \\mathbb{R}. What basis element is produced by the intersection $(-\\infty, %d) \\cap (%d, \\infty)$?", b, a)
		correct := fmt.Sprintf("(%d, %d)", a, b)
		dist := []string{
			fmt.Sprintf("[%d, %d]", a, b),
			fmt.Sprintf("(%d, %d)", a-1, b+1),
			fmt.Sprintf("(-\\infty, %d)", a),
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The intersection of rays (-inf, %d) and (%d, inf) is the bounded open interval (%d, %d).", b, a, a, b)
		p.Facts = map[string]string{"a": strconv.Itoa(a), "b": strconv.Itoa(b), "answer": correct}
		return p
	case 1:
		dim := 1 + rand.Intn(4)
		q := fmt.Sprintf("Which countable collection forms a basis for the Euclidean topology on \\mathbb{R}^%d?", dim)
		var correct, d1, d2, d3 string
		if dim == 1 {
			correct = "open intervals (p, q) with rational endpoints p, q in Q"
			d1 = "closed intervals [p, q] with rational endpoints p, q in Q"
			d2 = "singletons {q} for all rational points q in Q"
			d3 = "all open intervals (a, b) with arbitrary real endpoints"
		} else {
			correct = fmt.Sprintf("open balls B_r(q) with rational centers q in Q^%d and rational radii r in Q_{>0}", dim)
			d1 = fmt.Sprintf("closed balls \\bar{B}_r(q) with rational centers q in Q^%d and rational radii r in Q_{>0}", dim)
			d2 = fmt.Sprintf("singletons {q} for all rational points q in Q^%d", dim)
			d3 = fmt.Sprintf("all open balls B_r(x) with arbitrary real centers in R^%d", dim)
		}
		p := generator.MultipleChoice(q, correct, d1, d2, d3)
		p.Explanation = "Rational centers and rational radii form a countable collection that generates the Euclidean topology."
		p.Facts = map[string]string{"dim": strconv.Itoa(dim), "answer": correct}
		return p
	case 2:
		q := "A collection B of open subsets of a topological space X forms a basis for the topology if and only if..."
		correct := "every open set in X can be written as a union of elements of B"
		dist := []string{
			"every element of B is compact and closed",
			"the collection B is closed under arbitrary unions",
			"the collection B is closed under complements",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A basis generates the topology because every open set decomposes into a union of basis elements."
		p.Facts = map[string]string{"topic": "basis_definition", "answer": correct}
		return p
	case 3:
		q := "If B is a basis for a topology on X and x in B_1 \\cap B_2 for basis elements B_1, B_2, then the basis axioms guarantee..."
		correct := "there exists B_3 in B such that x in B_3 \\subseteq B_1 \\cap B_2"
		dist := []string{
			"B_1 \\cap B_2 must itself be a member of B",
			"B_1 \\cap B_2 is a singleton {x}",
			"B_1 and B_2 must be disjoint sets",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The second basis axiom requires that every point in an intersection of basis sets is contained in a smaller basis set inside that intersection."
		p.Facts = map[string]string{"topic": "basis_intersection_axiom", "answer": correct}
		return p
	case 4:
		n := 3 + rand.Intn(4)
		q := fmt.Sprintf("Let X be a set with %d elements. If a basis B contains every singleton {x} for each x in X, what topology does B generate?", n)
		correct := "the discrete topology"
		dist := []string{
			"the indiscrete topology",
			"the cofinite topology",
			"a non-Hausdorff topology",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "When singletons are basis elements, every subset is a union of singletons and therefore open, giving the discrete topology."
		p.Facts = map[string]string{"n": strconv.Itoa(n), "answer": correct}
		return p
	default:
		q := "Let (X, T_X) and (Y, T_Y) have bases B_X and B_Y. A standard basis for the product topology on X x Y is..."
		correct := "{ U x V | U in B_X and V in B_Y }"
		dist := []string{
			"{ U U V | U in B_X and V in B_Y }",
			"{ U x Y | U in B_X }",
			"all arbitrary subsets of X x Y",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The product topology basis consists of open boxes formed by products of basis elements from each factor."
		p.Facts = map[string]string{"topic": "product_basis", "answer": correct}
		return p
	}
}

// ---------------------------------------------------------------------------
// 7. The subspace topology
// ---------------------------------------------------------------------------

type subspaceGen struct{}

func (g *subspaceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		a := rand.Intn(6) - 3
		b := a + 3 + rand.Intn(5)
		c := a + 1 + rand.Intn(b-a-1)
		q := fmt.Sprintf("Consider the subspace Y = [%d, %d] of \\mathbb{R} with the Euclidean topology. Is the half-open interval [%d, %d) open in Y?", a, b, a, c)
		correct := fmt.Sprintf("Yes, because [%d, %d) = (%d, %d) \\cap [%d, %d] where (%d, %d) is open in R", a, c, a-1, c, a, b, a-1, c)
		dist := []string{
			fmt.Sprintf("No, because [%d, %d) is not open in the ambient space R", a, c),
			fmt.Sprintf("No, because Y = [%d, %d] is closed in R", a, b),
			"Yes, because every subset of Y inherits the discrete topology",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("[%d, %d) is the trace of the ambient open interval (%d, %d) on Y, so it is open in the subspace topology.", a, c, a-1, c)
		p.Facts = map[string]string{"a": strconv.Itoa(a), "b": strconv.Itoa(b), "c": strconv.Itoa(c), "answer": correct}
		return p
	case 1:
		a := rand.Intn(6) - 3
		b := a + 3 + rand.Intn(5)
		c := a + 1 + rand.Intn(b-a-1)
		q := fmt.Sprintf("Consider the subspace Y = [%d, %d] of \\mathbb{R} with the Euclidean topology. Is the half-open interval (%d, %d] open in Y?", a, b, c, b)
		correct := fmt.Sprintf("Yes, because (%d, %d] = (%d, %d) \\cap [%d, %d] where (%d, %d) is open in R", c, b, c, b+1, a, b, c, b+1)
		dist := []string{
			fmt.Sprintf("No, because (%d, %d] contains its right endpoint and cannot be open", c, b),
			fmt.Sprintf("No, because (%d, %d] is not closed in R", c, b),
			"No, because boundary points cannot belong to open subspace sets",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("(%d, %d] is the intersection of the ambient open interval (%d, %d) with Y, so it is open in the subspace Y.", c, b, c, b+1)
		p.Facts = map[string]string{"a": strconv.Itoa(a), "b": strconv.Itoa(b), "c": strconv.Itoa(c), "answer": correct}
		return p
	case 2:
		a := rand.Intn(6) - 3
		b := a + 4 + rand.Intn(4)
		c1 := a + 1
		c2 := b - 1
		q := fmt.Sprintf("Consider the subspace Y = [%d, %d] of \\mathbb{R} with the Euclidean topology. The interval (%d, %d) is...", a, b, c1, c2)
		correct := "open in both the subspace Y and the ambient space R"
		dist := []string{
			"open in the subspace Y but not in R",
			"closed in the subspace Y and open in R",
			"neither open in Y nor open in R",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("(%d, %d) is open in R and entirely contained in Y, so it is open in both.", c1, c2)
		p.Facts = map[string]string{"a": strconv.Itoa(a), "b": strconv.Itoa(b), "c1": strconv.Itoa(c1), "c2": strconv.Itoa(c2), "answer": correct}
		return p
	case 3:
		type propSpec struct {
			prop    string
			correct string
			dist    []string
			exp     string
		}
		props := []propSpec{
			{
				prop:    "Hausdorff (T_2)",
				correct: "every subspace of a Hausdorff space is Hausdorff",
				dist: []string{
					"subspaces of Hausdorff spaces are generally not Hausdorff",
					"only closed subspaces of Hausdorff spaces are Hausdorff",
					"only open subspaces of Hausdorff spaces are Hausdorff",
				},
				exp: "Hausdorff separation is hereditary: any two points can be separated by ambient opens whose traces remain disjoint.",
			},
			{
				prop:    "Metrizability",
				correct: "every subspace of a metrizable space is metrizable",
				dist: []string{
					"subspaces of metrizable spaces are rarely metrizable",
					"only compact subspaces of metric spaces are metrizable",
					"subspaces lose their metric structure",
				},
				exp: "The restriction of the ambient metric to the subset defines the subspace topology.",
			},
			{
				prop:    "Second-countability",
				correct: "every subspace of a second-countable space is second-countable",
				dist: []string{
					"subspaces can have uncountable minimal bases",
					"only countable subspaces are second-countable",
					"second-countability is not preserved by subspaces",
				},
				exp: "Intersecting a countable ambient basis with the subspace yields a countable basis for the subspace.",
			},
		}
		s := props[rand.Intn(len(props))]
		q := fmt.Sprintf("Regarding the %s property under the subspace topology, which statement is true?", s.prop)
		p := generator.MultipleChoice(q, s.correct, s.dist...)
		p.Explanation = s.exp
		p.Facts = map[string]string{"property": s.prop, "answer": s.correct}
		return p
	default:
		q := "Let A be a subset of a topological space X. The subspace topology on A is characterized as..."
		correct := "the coarsest topology on A for which the inclusion map i: A -> X is continuous"
		dist := []string{
			"the finest topology on A for which the inclusion map i: A -> X is continuous",
			"the quotient topology induced by the inclusion map",
			"the discrete topology restricted to A",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A map into X is continuous if preimages of opens are open; the coarsest topology having this property is the subspace topology."
		p.Facts = map[string]string{"topic": "subspace_universal_property", "answer": correct}
		return p
	}
}

// ---------------------------------------------------------------------------
// 8. Dense sets and separability
// ---------------------------------------------------------------------------

type denseGen struct{}

func (g *denseGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(6)
	switch mode {
	case 0:
		primes := []int{2, 3, 5, 7, 10}
		pVal := primes[rand.Intn(len(primes))]
		q := fmt.Sprintf("Let S = \\{ m / %d^k \\mid m \\in \\mathbb{Z}, k \\in \\mathbb{N} \\}. In the real line \\mathbb{R} with the Euclidean topology, the set S is...", pVal)
		correct := "dense in R, because every non-empty open interval contains an element of S"
		dist := []string{
			"nowhere dense in R, because S contains only isolated points",
			"closed in R, because S contains all its limit points",
			"discrete in R",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Numbers of the form m/%d^k can approximate any real number to arbitrary precision.", pVal)
		p.Facts = map[string]string{"base": strconv.Itoa(pVal), "answer": correct}
		return p
	case 1:
		numPts := 2 + rand.Intn(3)
		pts := make([]string, numPts)
		for i := 0; i < numPts; i++ {
			pts[i] = strconv.Itoa(i*3 - 2)
		}
		ptsStr := strings.Join(pts, ", ")
		q := fmt.Sprintf("Let E = \\{%s\\} be a finite subset of \\mathbb{R}. In \\mathbb{R} with the Euclidean topology, the set D = \\mathbb{R} \\setminus E is...", ptsStr)
		correct := "dense and open in R"
		dist := []string{
			"closed and nowhere dense in R",
			"dense and closed in R",
			"compact and dense in R",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Removing finitely many points leaves an open set whose closure is all of R, so it is both open and dense."
		p.Facts = map[string]string{"numPts": strconv.Itoa(numPts), "answer": correct}
		return p
	case 2:
		dim := 2 + rand.Intn(3)
		q := fmt.Sprintf("In \\mathbb{R}^%d with the Euclidean topology, which of the following subsets is dense?", dim)
		correct := fmt.Sprintf("the set of rational points Q^%d", dim)
		dist := []string{
			fmt.Sprintf("the set of integer points Z^%d", dim),
			fmt.Sprintf("the coordinate hyperplane { (x_1,...,x_%d) | x_1 = 0 }", dim),
			fmt.Sprintf("the closed unit ball \\bar{B}_1(0) in R^%d", dim),
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Q^%d meets every non-empty open box or ball in R^%d, so its closure is R^%d.", dim, dim, dim)
		p.Facts = map[string]string{"dim": strconv.Itoa(dim), "answer": correct}
		return p
	case 3:
		q := "A subset D of a topological space X is dense in X if and only if..."
		correct := "the closure of D is equal to X"
		dist := []string{
			"the interior of D is equal to X",
			"D contains all of its boundary points",
			"D is homeomorphic to X",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "By definition, D is dense in X when Cl(D) = X, which is equivalent to every non-empty open set meeting D."
		p.Facts = map[string]string{"topic": "dense_definition", "answer": correct}
		return p
	case 4:
		q := "Which of the following subsets is NOT dense in the real line \\mathbb{R} with the Euclidean topology?"
		correct := "the integers Z"
		dist := []string{
			"the rational numbers Q",
			"the irrational numbers R \\ Q",
			"the dyadic rationals { m/2^k | m in Z, k in N }",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The integers Z are discrete and closed in R; their closure is Z, which is not all of R."
		p.Facts = map[string]string{"topic": "not_dense", "answer": correct}
		return p
	default:
		q := "A topological space X is defined to be separable if..."
		correct := "it contains a countable dense subset"
		dist := []string{
			"every pair of distinct points can be separated by disjoint open sets",
			"every sequence in X has a convergent subsequence",
			"its topology has a finite basis",
		}
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Separability is the property of possessing a countable dense subset (such as Q in R)."
		p.Facts = map[string]string{"topic": "separable_definition", "answer": correct}
		return p
	}
}

// ---------------------------------------------------------------------------
// 9. Homeomorphisms
// ---------------------------------------------------------------------------

type homeomorphismGen struct{}

func (g *homeomorphismGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "A homeomorphism is a bijection that...",
			correct:     "is continuous with a continuous inverse",
			distractors: []string{"is continuous", "preserves distances", "is linear"},
			reason:      "Topological sameness needs both directions continuous; a continuous bijection alone need not have a continuous inverse.",
		},
		{
			question:    "Which property is a topological invariant?",
			correct:     "connectedness",
			distractors: []string{"boundedness", "the metric", "the diameter"},
			reason:      "A homeomorphism carries connected sets to connected sets, while boundedness and diameter depend on the chosen metric.",
		},
		{
			question:    "Are the open interval (0, 1) and the real line homeomorphic? (yes/no)",
			correct:     "yes",
			distractors: []string{"no", "only if both are compact", "only under the discrete metric"},
			reason:      "The map x to tan(pi(x - 1/2)) is a continuous bijection with a continuous inverse from (0, 1) to R.",
		},
		{
			question:    "A continuous bijection is a homeomorphism when...",
			correct:     "its inverse is continuous",
			distractors: []string{"the spaces are compact", "the spaces are metric", "it preserves distances"},
			reason:      "Continuity of the inverse is the missing direction; compactness or a metric can supply it under extra hypotheses, but not by themselves.",
		},
		{
			question:    "Which map is a homeomorphism from (0, 1) to R?",
			correct:     "x to tan(pi(x - 1/2))",
			distractors: []string{"x to x", "x to x^2", "x to sin(x)"},
			reason:      "The tangent map sends the open interval onto the whole line continuously with a continuous inverse; the others are not onto R.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 10. Euler characteristic (numeric)
// ---------------------------------------------------------------------------

type eulerGen struct{}

func (g *eulerGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(3)
	var q, ans, reason string
	var v, e, f int
	switch mode {
	case 0:
		n := 3 + rand.Intn(10)
		if rand.Intn(2) == 0 {
			v = n + 1
			e = 2 * n
			f = n + 1
			ans = "2"
			q = fmt.Sprintf("A %d-gonal pyramid has %d vertices, %d edges and %d faces. What is its Euler characteristic V - E + F? (enter a number)", n, v, e, f)
			reason = fmt.Sprintf("%d - %d + %d = 2, the Euler characteristic of a sphere.", v, e, f)
		} else {
			v = 2 * n
			e = 3 * n
			f = n + 2
			ans = "2"
			q = fmt.Sprintf("A %d-gonal prism has %d vertices, %d edges and %d faces. What is its Euler characteristic V - E + F? (enter a number)", n, v, e, f)
			reason = fmt.Sprintf("%d - %d + %d = 2, as for every polyhedron homeomorphic to a sphere.", v, e, f)
		}
	case 1:
		genus := 1 + rand.Intn(6)
		chi := 2 - 2*genus
		v = 6 + rand.Intn(10)
		f = 2 * (v - chi)
		e = v + f - chi
		ans = strconv.Itoa(chi)
		q = fmt.Sprintf("A triangulated closed surface of genus %d has %d vertices, %d edges and %d faces. What is its Euler characteristic V - E + F? (enter a number)", genus, v, e, f)
		reason = fmt.Sprintf("%d - %d + %d = %d, the Euler characteristic of a surface of genus %d.", v, e, f, chi, genus)
	default:
		chiChoice := []int{0, 1}[rand.Intn(2)]
		v = 6 + rand.Intn(10)
		f = 6 + rand.Intn(10)
		e = v + f - chiChoice
		ans = strconv.Itoa(chiChoice)
		surfName := "torus"
		if chiChoice == 1 {
			surfName = "real projective plane"
		}
		q = fmt.Sprintf("A 2-dimensional cell complex homeomorphic to the %s has %d vertices, %d edges and %d faces. What is its Euler characteristic V - E + F? (enter a number)", surfName, v, e, f)
		reason = fmt.Sprintf("%d - %d + %d = %d.", v, e, f, chiChoice)
	}
	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: reason,
		Facts: map[string]string{
			"v": strconv.Itoa(v), "e": strconv.Itoa(e), "f": strconv.Itoa(f),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// 11. Genus of a surface (numeric)
// ---------------------------------------------------------------------------

type genusGen struct{}

func (g *genusGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	kind := "genus"
	if rand.Intn(2) == 1 {
		kind = "chi"
	}
	genus := rand.Intn(12)
	chi := 2 - 2*genus
	if kind == "genus" {
		q := fmt.Sprintf("A closed orientable surface has Euler characteristic %d. What is its genus? (enter a number)", chi)
		ans := strconv.Itoa(genus)
		reason := fmt.Sprintf("For a closed orientable surface chi = 2 - 2g, so g = (2 - (%d))/2 = %d.", chi, genus)
		return generator.Problem{
			Question:    q,
			Answer:      ans,
			Explanation: reason,
			Facts: map[string]string{
				"kind":   "genus",
				"chi":    strconv.Itoa(chi),
				"answer": ans,
			},
		}
	}
	q := fmt.Sprintf("A closed orientable surface has genus %d. What is its Euler characteristic 2 - 2g? (enter a number)", genus)
	ans := strconv.Itoa(chi)
	reason := fmt.Sprintf("2 - 2(%d) = %d.", genus, chi)
	return generator.Problem{
		Question:    q,
		Answer:      ans,
		Explanation: reason,
		Facts: map[string]string{
			"kind":   "chi",
			"genus":  strconv.Itoa(genus),
			"answer": ans,
		},
	}
}

// ---------------------------------------------------------------------------
// 12. Betti numbers (numeric)
// ---------------------------------------------------------------------------

type bettiGen struct{}

func (g *bettiGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(6)
	switch mode {
	case 0:
		return generator.Problem{
			Question:    "The first Betti number of a circle is what? (enter a number)",
			Answer:      "1",
			Explanation: "A circle has one independent loop, so H_1 has rank 1.",
			Facts: map[string]string{
				"kind":   "circle",
				"param":  "1",
				"answer": "1",
			},
		}
	case 1:
		genus := 1 + rand.Intn(10)
		b1 := 2 * genus
		ans := strconv.Itoa(b1)
		return generator.Problem{
			Question:    fmt.Sprintf("A closed orientable surface of genus %d has what first Betti number b1? (enter a number)", genus),
			Answer:      ans,
			Explanation: fmt.Sprintf("A closed orientable surface of genus g has 2g independent 1-cycles, so b1 = 2(%d) = %d.", genus, b1),
			Facts: map[string]string{
				"kind":   "surface",
				"param":  strconv.Itoa(genus),
				"answer": ans,
			},
		}
	case 2:
		numWords := []string{
			"zero", "one", "two", "three", "four", "five",
			"six", "seven", "eight", "nine", "ten", "eleven",
			"twelve", "thirteen", "fourteen", "fifteen", "sixteen",
		}
		k := 2 + rand.Intn(10)
		ans := strconv.Itoa(k)
		return generator.Problem{
			Question:    fmt.Sprintf("A wedge sum of %s circles has what first Betti number b1? (enter a number)", numWords[k]),
			Answer:      ans,
			Explanation: fmt.Sprintf("Each circle in a wedge sum contributes one independent generator to the first homology group, giving b1 = %d.", k),
			Facts: map[string]string{
				"kind":   "wedge",
				"param":  strconv.Itoa(k),
				"answer": ans,
			},
		}
	case 3:
		numWords := []string{
			"zero", "one", "two", "three", "four", "five",
			"six", "seven", "eight", "nine", "ten", "eleven",
			"twelve", "thirteen", "fourteen", "fifteen", "sixteen",
		}
		k := 2 + rand.Intn(10)
		ans := strconv.Itoa(k)
		return generator.Problem{
			Question:    fmt.Sprintf("A topological space has %s connected components. What is its zeroth Betti number b0? (enter a number)", numWords[k]),
			Answer:      ans,
			Explanation: fmt.Sprintf("The zeroth Betti number b0 counts path components, so %d components give b0 = %d.", k, k),
			Facts: map[string]string{
				"kind":   "components",
				"param":  strconv.Itoa(k),
				"answer": ans,
			},
		}
	case 4:
		return generator.Problem{
			Question:    "A tree is a connected graph with no cycles. What is its first Betti number b1? (enter a number)",
			Answer:      "0",
			Explanation: "A tree has no independent loop, so H_1 has rank 0.",
			Facts: map[string]string{
				"kind":   "tree",
				"param":  "0",
				"answer": "0",
			},
		}
	default:
		return generator.Problem{
			Question:    "For a sphere, b0 = 1, b1 = 0 and b2 = 1. What is b0 - b1 + b2? (enter a number)",
			Answer:      "2",
			Explanation: "1 - 0 + 1 = 2, which is also the Euler characteristic of the sphere.",
			Facts: map[string]string{
				"kind":   "sphere_euler",
				"param":  "2",
				"answer": "2",
			},
		}
	}
}
