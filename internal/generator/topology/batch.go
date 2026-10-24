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
		tr := triples[rand.Intn(len(triples))]
		dx, dy, dist := tr[0], tr[1], tr[2]
		x1 := rand.Intn(15) - 7
		y1 := rand.Intn(15) - 7
		x2 := x1 + dx
		y2 := y1 + dy
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

// ---------------------------------------------------------------------------
// 2. Interior, closure and boundary (numeric)
// ---------------------------------------------------------------------------

type interiorGen struct{}

func (g *interiorGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	var q, ans, reason string
	var pts []int
	kind := "boundary"
	a := rand.Intn(16) - 8
	b := a + rand.Intn(6) + 1
	switch mode {
	case 0:
		pts = []int{a, b}
		ans = "2"
		q = fmt.Sprintf("In R with the usual topology, how many boundary points does the set [%d, %d] have? (enter a number)", a, b)
		reason = fmt.Sprintf("The boundary is {%d, %d}: every neighbourhood of either endpoint meets both the set and its complement.", a, b)
	case 1:
		c := b + rand.Intn(5) + 1
		d := c + rand.Intn(5) + 1
		pts = []int{a, b, c, d}
		ans = "4"
		q = fmt.Sprintf("In R with the usual topology, how many boundary points does the set [%d, %d] union [%d, %d] have? (enter a number)", a, b, c, d)
		reason = fmt.Sprintf("The boundary is {%d, %d, %d, %d}: the four endpoints each have neighbourhoods meeting the set and its complement.", a, b, c, d)
	case 2:
		c := b + rand.Intn(6) + 1
		pts = []int{a, b, c}
		ans = "3"
		q = fmt.Sprintf("In R with the usual topology, how many boundary points does the set (%d, %d) union (%d, %d) have? (enter a number)", a, b, b, c)
		reason = fmt.Sprintf("The boundary is {%d, %d, %d}: the endpoints and the missing point %d all have neighbourhoods meeting both sides.", a, b, c, b)
	case 3:
		pts = []int{a}
		ans = "1"
		q = fmt.Sprintf("In R with the usual topology, how many boundary points does the ray (%d, infinity) have? (enter a number)", a)
		reason = fmt.Sprintf("The only boundary point is %d; every other real is either inside the ray or has a ball outside it.", a)
	default:
		kind = "interior"
		pts = []int{}
		ans = "0"
		n := 2 + rand.Intn(4)
		elems := make([]string, n)
		for i := 0; i < n; i++ {
			elems[i] = strconv.Itoa(a + i*2)
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
	cPool := []int{1, 2, 4, 5, 8, 10}
	c := cPool[rand.Intn(len(cPool))]
	a := rand.Intn(16)
	b := rand.Intn(15) + 1
	if rand.Intn(2) == 1 {
		b = -b
	}
	d := rand.Intn(10)
	var ans string
	if a%c == 0 {
		ans = strconv.Itoa(a / c)
	} else {
		ans = strconv.FormatFloat(float64(a)/float64(c), 'f', -1, 64)
	}
	var numStr string
	if a == 0 {
		numStr = strconv.Itoa(b)
	} else if b > 0 {
		numStr = fmt.Sprintf("%dn + %d", a, b)
	} else if b < 0 {
		numStr = fmt.Sprintf("%dn - %d", a, -b)
	} else {
		numStr = fmt.Sprintf("%dn", a)
	}
	var denStr string
	if d > 0 {
		denStr = fmt.Sprintf("%dn + %d", c, d)
	} else {
		denStr = fmt.Sprintf("%dn", c)
	}
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
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "A basis for a topology is a collection of open sets such that...",
			correct:     "every open set is a union of basis elements",
			distractors: []string{"every basis element is compact", "the basis is closed under complements", "the basis has finitely many sets"},
			reason:      "The basis generates the topology: each open set is assembled from basis elements, and the basis axioms make that assembly consistent.",
		},
		{
			question:    "Which collection is a basis for the usual topology on R?",
			correct:     "the open intervals (a, b)",
			distractors: []string{"the closed intervals [a, b]", "the singletons {x}", "the half-open intervals [a, b)"},
			reason:      "Open intervals generate the usual opens; closed intervals generate the discrete topology and half-open intervals generate the Sorgenfrey topology.",
		},
		{
			question:    "In a metric space, the collection of open balls is a basis for...",
			correct:     "the metric topology",
			distractors: []string{"the discrete topology", "the cofinite topology", "the indiscrete topology"},
			reason:      "Open sets in a metric space are unions of balls, which is the defining property of a basis for the metric topology.",
		},
		{
			question:    "A subbasis differs from a basis because...",
			correct:     "finite intersections of subbasis elements form a basis",
			distractors: []string{"it must be countable", "it contains the empty set", "it is closed under arbitrary unions"},
			reason:      "A subbasis is the raw generating family; taking finite intersections produces a basis, and unions of that basis produce the topology.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 7. The subspace topology
// ---------------------------------------------------------------------------

type subspaceGen struct{}

func (g *subspaceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "For a subset A of a space X, a set U is open in the subspace topology exactly when...",
			correct:     "U is the intersection of A with an open set of X",
			distractors: []string{"U is open in X", "U is closed in X", "U contains every limit point of A"},
			reason:      "The subspace opens are the traces A intersect V of the ambient opens V, by definition.",
		},
		{
			question:    "In the subspace [0, 1] of R, the set [0, 1/2) is...",
			correct:     "open in the subspace but not in R",
			distractors: []string{"open in R", "closed in the subspace", "neither open nor closed"},
			reason:      "It is [0, 1] intersect (-1, 1/2), so it is a subspace open; in R it is not open because no ball around 0 stays inside it.",
		},
		{
			question:    "The subspace topology on A is the...",
			correct:     "coarsest topology making the inclusion of A into X continuous",
			distractors: []string{"finest topology making the inclusion continuous", "quotient topology of A", "product topology on A"},
			reason:      "The inclusion is continuous exactly when each subspace open is a trace of an ambient open, which is the coarsest such choice.",
		},
		{
			question:    "A subspace of a Hausdorff space is...",
			correct:     "Hausdorff",
			distractors: []string{"not necessarily Hausdorff", "always compact", "always connected"},
			reason:      "Separating two points by disjoint ambient opens and intersecting both with the subspace gives disjoint subspace opens.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 8. Dense sets and separability
// ---------------------------------------------------------------------------

type denseGen struct{}

func (g *denseGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "A subset D of a space X is dense exactly when...",
			correct:     "every nonempty open set meets D",
			distractors: []string{"D is closed", "D is countable", "D contains every limit point"},
			reason:      "D is dense when its closure is X, which is the same as every nonempty open set meeting D.",
		},
		{
			question:    "Which subset is dense in R with the usual topology?",
			correct:     "the rationals",
			distractors: []string{"the integers", "the integers between 0 and 1", "the set {0}"},
			reason:      "Every open interval contains a rational, so the rationals meet every nonempty open set; the other sets leave gaps.",
		},
		{
			question:    "A separable space is one with...",
			correct:     "a countable dense subset",
			distractors: []string{"a countable basis", "finitely many points", "no limit points"},
			reason:      "Separability is defined by the existence of a countable dense subset; a countable basis is a different, stronger condition.",
		},
		{
			question:    "The closure of a dense set is...",
			correct:     "the whole space",
			distractors: []string{"the empty set", "the set itself", "a proper subset of the space"},
			reason:      "Dense means closure equal to the whole space; the set itself need not be closed.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
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
			Question:    fmt.Sprintf("A closed orientable surface of genus %d has first Betti number b1 = 2g. What is b1? (enter a number)", genus),
			Answer:      ans,
			Explanation: fmt.Sprintf("b1 = 2g = 2(%d) = %d for a surface of genus %d.", genus, b1, genus),
			Facts: map[string]string{
				"kind":   "surface",
				"param":  strconv.Itoa(genus),
				"answer": ans,
			},
		}
	case 2:
		k := 2 + rand.Intn(10)
		ans := strconv.Itoa(k)
		return generator.Problem{
			Question:    fmt.Sprintf("A wedge of %d circles has first Betti number equal to the number of circles. What is b1? (enter a number)", k),
			Answer:      ans,
			Explanation: fmt.Sprintf("Each circle contributes one independent loop, so a wedge of %d circles has b1 = %d.", k, k),
			Facts: map[string]string{
				"kind":   "wedge",
				"param":  strconv.Itoa(k),
				"answer": ans,
			},
		}
	case 3:
		k := 2 + rand.Intn(10)
		ans := strconv.Itoa(k)
		return generator.Problem{
			Question:    fmt.Sprintf("A space has %d connected components. What is its zeroth Betti number b0? (enter a number)", k),
			Answer:      ans,
			Explanation: fmt.Sprintf("b0 counts connected components, so %d components give b0 = %d.", k, k),
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
