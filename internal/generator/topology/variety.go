package topology

import (
	"fmt"
	"math/rand"

	"github.com/chuma-beep/mathua/internal/generator"
)

// distinctDistractors ensures option candidates do not collide with the correct
// answer or with each other, providing fallbacks when collisions occur.
func distinctDistractors(correct string, candidates ...string) []string {
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
		"not enough information is provided",
		"only on compact subspaces",
		"none of the other choices holds",
		"holds only for finite sets",
		"diverges everywhere",
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
// 1. Hausdorff Spaces (topo.hausdorff)
// ---------------------------------------------------------------------------

type hausdorffGen struct{}

func (g *hausdorffGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(6) {
	case 0:
		dim := 2 + rand.Intn(7)
		q := fmt.Sprintf("Which of the following topological spaces fails to be Hausdorff ($T_2$)?")
		correct := fmt.Sprintf("the cofinite topology on an infinite set $\\mathbb{R}^{%d}$", dim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("Euclidean space $\\mathbb{R}^{%d}$ with standard metric topology", dim),
			fmt.Sprintf("the discrete topology on $\\mathbb{R}^{%d}$", dim),
			fmt.Sprintf("the product space $\\mathbb{R}^{%d} \\times [0, 1]$ with product topology", dim),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "In the cofinite topology on any infinite set, every pair of nonempty open sets intersects because their complements are finite, preventing distinct points from having disjoint open neighborhoods."
		return p
	case 1:
		a := 1 + rand.Intn(8)
		b := a + 2 + rand.Intn(8)
		diff := b - a
		maxR := float64(diff) / 2.0
		q := fmt.Sprintf("In the real line $\\mathbb{R}$ with Euclidean metric, consider distinct points $x = %d$ and $y = %d$. What is the supremum of radius $r > 0$ such that the open balls $B(x, r)$ and $B(y, r)$ are disjoint?", a, b)
		correct := fmt.Sprintf("%.1f", maxR)
		dist := distinctDistractors(correct,
			fmt.Sprintf("%d", diff),
			fmt.Sprintf("%.1f", float64(diff)/4.0),
			fmt.Sprintf("%d", diff*2),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Open balls around distinct points are disjoint if and only if the sum of their radii does not exceed the distance d(x, y), so r <= d(x, y)/2."
		return p
	case 2:
		n := 2 + rand.Intn(6)
		q := fmt.Sprintf("In any Hausdorff topological space $X$, every finite subset of cardinality %d is guaranteed to be...", n)
		correct := "closed"
		dist := distinctDistractors(correct, "open", "dense", "clopen")
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Every Hausdorff space is T_1, which means every singleton set {x} is closed. A finite union of closed singletons is closed."
		return p
	case 3:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("A topological space $X = \\mathbb{R}^{%d}$ is Hausdorff if and only if the diagonal $\\Delta = \\{(x, x) \\in X \\times X\\}$ is...", dim)
		correct := fmt.Sprintf("closed in $\\mathbb{R}^{%d} \\times \\mathbb{R}^{%d}$", dim, dim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("open in $\\mathbb{R}^{%d} \\times \\mathbb{R}^{%d}$", dim, dim),
			fmt.Sprintf("dense in $\\mathbb{R}^{%d} \\times \\mathbb{R}^{%d}$", dim, dim),
			fmt.Sprintf("compact in $\\mathbb{R}^{%d} \\times \\mathbb{R}^{%d}$", dim, dim),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A space is Hausdorff if and only if its diagonal is closed in the Cartesian product space."
		return p
	case 4:
		k := 2 + rand.Intn(5)
		q := fmt.Sprintf("Let $f: X \\to Y$ be a continuous bijection where $X$ is a compact space and $Y$ is a Hausdorff space of dimension %d. Then $f$ is guaranteed to be...", k)
		correct := "a homeomorphism"
		dist := distinctDistractors(correct,
			"an open map that fails to be closed",
			"a non-invertible quotient map",
			"not necessarily continuous in the inverse direction",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A continuous map from a compact space to a Hausdorff space maps closed sets to compact sets, which are closed in Y. A closed continuous bijection is a homeomorphism."
		return p
	default:
		c := 1 + rand.Intn(9)
		q := fmt.Sprintf("In the line with two origins $\\tilde{\\mathbb{R}} = (\\mathbb{R} \\setminus \\{0\\}) \\cup \\{0_a, 0_b\\}$, what prevents the points $0_a$ and $0_b$ from having disjoint open neighborhoods?")
		correct := "every open neighborhood of $0_a$ and $0_b$ contains an interval $(-\\varepsilon, \\varepsilon) \\setminus \\{0\\}$"
		dist := distinctDistractors(correct,
			fmt.Sprintf("the distance between $0_a$ and $0_b$ is exactly %d", c),
			"the space is compact",
			"the space fails to be first-countable",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Any open neighborhood of 0_a intersects any open neighborhood of 0_b in a punctured symmetric interval, so 0_a and 0_b cannot be separated by disjoint open sets."
		return p
	}
}

// ---------------------------------------------------------------------------
// 2. Quotient Spaces (topo.quotient)
// ---------------------------------------------------------------------------

type quotientGen struct{}

func (g *quotientGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(6) {
	case 0:
		lenVal := 2 + rand.Intn(9)
		q := fmt.Sprintf("What topological space is homeomorphic to the quotient space of $[0, %d]$ formed by identifying the endpoints $0 \\sim %d$?", lenVal, lenVal)
		correct := "$S^1$ (the circle)"
		dist := distinctDistractors(correct,
			"$\\mathbb{R}$",
			"$D^2$ (the closed 2-disc)",
			"a single isolated point",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Identifying the two endpoints of a closed interval glues them together into a circle S^1."
		return p
	case 1:
		dim := 2 + rand.Intn(5)
		q := fmt.Sprintf("What topological space is homeomorphic to the quotient space $D^{%d} / \\partial D^{%d}$, where the boundary sphere of the closed unit %d-disc is collapsed to a point?", dim, dim, dim)
		correct := fmt.Sprintf("$S^{%d}$", dim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("$S^{%d}$", dim-1),
			fmt.Sprintf("$D^{%d}$", dim),
			fmt.Sprintf("$\\mathbb{R}^{%d}$", dim),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Collapsing the boundary sphere S^%d of an %d-dimensional disc D^%d to a single point produces the sphere S^%d.", dim-1, dim, dim, dim)
		return p
	case 2:
		m := 2 + rand.Intn(9)
		q := fmt.Sprintf("The quotient group and space $\\mathbb{R} / (%d\\mathbb{Z})$ with the quotient topology is homeomorphic to...", m)
		correct := "$S^1$ (the circle)"
		dist := distinctDistractors(correct,
			"$\\mathbb{R}$",
			fmt.Sprintf("$[0, %d]$", m),
			"$\\mathbb{Z}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The exponential wrapping map x -> exp(2 pi i x / m) is a continuous open surjection from R to S^1 whose fibers are the cosets of mZ."
		return p
	case 3:
		k := 2 + rand.Intn(5)
		q := fmt.Sprintf("Let $X = \\bigsqcup_{j=1}^{%d} S_j^1$ be the disjoint topological sum of %d circles. If one point from each circle is identified into a single basepoint, the quotient space is...", k, k)
		correct := fmt.Sprintf("the wedge sum $\\bigvee_{j=1}^{%d} S^1$", k)
		dist := distinctDistractors(correct,
			fmt.Sprintf("the product torus $T^{%d}$", k),
			fmt.Sprintf("the sphere $S^{%d}$", k),
			"a disjoint discrete space",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Identifying one point on each of k circles produces the bouquet or wedge sum of k circles."
		return p
	case 4:
		w := 1 + rand.Intn(5)
		h := 1 + rand.Intn(5)
		q := fmt.Sprintf("Consider the rectangle $[0, %d] \\times [0, %d]$. Identifying opposite edges by $(x, 0) \\sim (x, %d)$ and $(0, y) \\sim (%d, y)$ yields which topological surface?", w, h, h, w)
		correct := "the torus $T^2$"
		dist := distinctDistractors(correct,
			"the Klein bottle",
			"the real projective plane $\\mathbb{R}P^2$",
			"the 2-sphere $S^2$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Gluing both pairs of opposite edges in the parallel orientation creates a 2-dimensional torus T^2."
		return p
	default:
		w := 1 + rand.Intn(5)
		h := 1 + rand.Intn(5)
		q := fmt.Sprintf("Consider the rectangle $[0, %d] \\times [0, %d]$. Identifying opposite vertical edges with a reversal $(0, y) \\sim (%d, %d-y)$ while leaving horizontal edges untouched yields...", w, h, w, h)
		correct := "the Mobius strip"
		dist := distinctDistractors(correct,
			"a cylinder",
			"the torus $T^2$",
			"the Klein bottle",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A single pair of edges glued with a twist produces the non-orientable surface with boundary known as the Mobius strip."
		return p
	}
}

// ---------------------------------------------------------------------------
// 3. Product Spaces (topo.product)
// ---------------------------------------------------------------------------

type productGen struct{}

func (g *productGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(6) {
	case 0:
		coord1 := 1 + rand.Intn(4)
		coord2 := coord1 + 1 + rand.Intn(4)
		a := 1 + rand.Intn(4)
		b := a + 1 + rand.Intn(4)
		q := fmt.Sprintf("In the product space $\\mathbb{R}^\\mathbb{N}$ with product topology, the cylinder set $U = \\{x : x_%d \\in (0, %d) \\text{ and } x_%d \\in (%d, %d)\\}$ is...", coord1, a, coord2, a, b)
		correct := "open in both the product and box topologies"
		dist := distinctDistractors(correct,
			"open in the box topology but not the product topology",
			"closed but not open in the product topology",
			"compact in the product topology",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Because only two coordinates are restricted, U is a basic open cylinder in the product topology and is also open in the box topology."
		return p
	case 1:
		k := 2 + rand.Intn(5)
		q := fmt.Sprintf("Consider the product $X = \\prod_{n=1}^\\infty [-%d, %d]$. By Tychonoff's theorem, $X$ is...", k, k)
		correct := "compact in the product topology"
		dist := distinctDistractors(correct,
			"compact in the box topology",
			"not compact in any topology",
			"discrete",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Each interval [-k, k] is compact, so by Tychonoff's theorem their arbitrary Cartesian product is compact in the product topology."
		return p
	case 2:
		target := 1 + rand.Intn(8)
		q := fmt.Sprintf("For the infinite product space $\\prod_{n=1}^\\infty X_n$, the coordinate projection map $\\pi_%d: \\prod_{n=1}^\\infty X_n \\to X_%d$ is guaranteed to be...", target, target)
		correct := "continuous and open"
		dist := distinctDistractors(correct,
			"continuous and closed",
			"a homeomorphism",
			"a compact embedding",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Coordinate projections from a product space are continuous and map open basic cylinder sets to open sets, making them open maps."
		return p
	case 3:
		n := 2 + rand.Intn(6)
		q := fmt.Sprintf("Let $(M_1, d_1), \\dots, (M_%d, d_%d)$ be %d metric spaces. What metric on the finite product space $\\prod_{j=1}^{%d} M_j$ generates the product topology?", n, n, n, n)
		correct := "$d(x, y) = \\max_{1 \\le j \\le n} d_j(x_j, y_j)$"
		dist := distinctDistractors(correct,
			"$d(x, y) = \\prod_{j=1}^n d_j(x_j, y_j)$",
			"$d(x, y) = \\min_{1 \\le j \\le n} d_j(x_j, y_j)$",
			"$d(x, y) = 1$ for all $x \\ne y$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The supremum or maximum metric generates the product topology on a finite product of metric spaces."
		return p
	case 4:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("In the product space $(\\mathbb{R}^{%d})^\\mathbb{N}$, a basic open set in the product topology restricts...", dim)
		correct := "only finitely many coordinates to proper open subsets"
		dist := distinctDistractors(correct,
			"infinitely many coordinates to proper open subsets",
			"all coordinates to proper open subsets",
			"exactly one single coordinate",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "By definition of the product topology, basic open sets restrict only a finite subset of coordinates, with all remaining coordinates unconstrained."
		return p
	default:
		d := 2 + rand.Intn(5)
		q := fmt.Sprintf("On the Cartesian product $\\mathbb{R}^{%d}$, how do the box topology and the product topology compare?", d)
		correct := "they are identical"
		dist := distinctDistractors(correct,
			"the box topology is strictly finer",
			"the product topology is strictly finer",
			"neither topology is comparable",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "For finite Cartesian products, the box topology and the product topology coincide."
		return p
	}
}

// ---------------------------------------------------------------------------
// 4. Urysohn Lemma (topo.urysohn)
// ---------------------------------------------------------------------------

type urysohnGen struct{}

func (g *urysohnGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		a := 1 + rand.Intn(4)
		b := a + 1 + rand.Intn(3)
		c := b + 1 + rand.Intn(3)
		d := c + 1 + rand.Intn(3)
		q := fmt.Sprintf("Let $A = [%d, %d]$ and $B = [%d, %d]$ be disjoint closed subsets of $\\mathbb{R}$. By Urysohn's Lemma, there exists a continuous function $f: \\mathbb{R} \\to [0, 1]$ satisfying...", a, b, c, d)
		correct := fmt.Sprintf("$f([%d, %d]) = \\{0\\}$ and $f([%d, %d]) = \\{1\\}$", a, b, c, d)
		dist := distinctDistractors(correct,
			fmt.Sprintf("$f([%d, %d]) = \\{1\\}$ and $f([%d, %d]) = \\{1\\}$", a, b, c, d),
			fmt.Sprintf("$f([%d, %d] \\cup [%d, %d]) = \\{0\\}$", a, b, c, d),
			"$f(\\mathbb{R}) = \\{0, 1\\}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Urysohn's Lemma constructs a continuous function into [0, 1] taking constant value 0 on A and constant value 1 on B."
		return p
	case 1:
		v0 := rand.Intn(3)
		v1 := v0 + 1 + rand.Intn(4)
		q := fmt.Sprintf("If $A$ and $B$ are disjoint closed subsets in a normal space $X$, Urysohn's Lemma guarantees a continuous function $f: X \\to [%d, %d]$ such that $f(A) = \\{%d\\}$ and $f(B) = \\{%d\\}$. Which separation axiom is required on $X$?", v0, v1, v0, v1)
		correct := "normal ($T_4$)"
		dist := distinctDistractors(correct,
			"Hausdorff ($T_2$)",
			"regular ($T_3$)",
			"indiscrete",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Normality is the necessary and sufficient condition for Urysohn's separation lemma to hold."
		return p
	case 2:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("Does Urysohn's Lemma hold in every Euclidean space $\\mathbb{R}^{%d}$ with standard metric?", dim)
		correct := fmt.Sprintf("yes, because $\\mathbb{R}^{%d}$ is a metric space and all metric spaces are normal", dim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("no, $\\mathbb{R}^{%d}$ fails normality for $n \\ge 2$", dim),
			"only for bounded subsets",
			"only for convex subsets",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Every metric space is normal, so Urysohn's Lemma applies to all Euclidean spaces."
		return p
	case 3:
		denom := []int{2, 4, 8, 16}[rand.Intn(4)]
		q := fmt.Sprintf("The inductive construction in Urysohn's Lemma defines nested open sets $U_r$ satisfying $\\overline{U_r} \\subseteq U_s$ for $r < s$. What dense subset of $[0, 1]$ indexed with denominator %d and powers of 2 is used?", denom)
		correct := "the dyadic rationals $k / 2^n$"
		dist := distinctDistractors(correct,
			"the irrational numbers",
			"the integer lattice",
			"the Cantor ternary set",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The proof uses the dense set of dyadic rationals in [0, 1] to construct the level sets of the continuous function."
		return p
	default:
		xVal := 1 + rand.Intn(4)
		distB := 2 + rand.Intn(4)
		total := xVal + distB
		val := float64(xVal) / float64(total)
		q := fmt.Sprintf("In the metric space $(\\mathbb{R}, d)$, the explicit Urysohn function separating $A = \\{0\\}$ from $B = \\{%d\\}$ is $f(x) = \\frac{d(x, A)}{d(x, A) + d(x, B)}$. What is $f(%d)$?", total, xVal)
		correct := fmt.Sprintf("%.2f", val)
		dist := distinctDistractors(correct,
			fmt.Sprintf("%.2f", 1.0-val),
			fmt.Sprintf("%.2f", float64(xVal)/float64(distB)),
			"0.00",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("d(%d, 0) = %d and d(%d, %d) = %d, so f(%d) = %d / (%d + %d) = %.2f.", xVal, xVal, xVal, total, distB, xVal, xVal, xVal, distB, val)
		return p
	}
}

// ---------------------------------------------------------------------------
// 5. Metrization Theorems (topo.metrization)
// ---------------------------------------------------------------------------

type metrizationGen struct{}

func (g *metrizationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		dim := 1 + rand.Intn(6)
		q := fmt.Sprintf("Euclidean space $\\mathbb{R}^{%d}$ is metrizable. Which combination of properties from the Urysohn Metrization Theorem guarantees this?", dim)
		correct := "regular, Hausdorff and second-countable"
		dist := distinctDistractors(correct,
			"compact, connected and separable",
			"locally compact and complete",
			"normal and indiscrete",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The Urysohn Metrization Theorem states that every second-countable regular Hausdorff space is metrizable."
		return p
	case 1:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("What is the cardinality of the standard countable topological basis of open balls with rational centers and rational radii in $\\mathbb{R}^{%d}$?", dim)
		correct := "countably infinite ($\\aleph_0$)"
		dist := distinctDistractors(correct,
			"uncountable ($\\mathfrak{c}$)",
			fmt.Sprintf("finite of order %d", dim),
			"empty",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The centers Q^%d and positive rational radii Q^+ are both countable sets, and their Cartesian product is countable.", dim)
		return p
	case 2:
		q := "Why does the Sorgenfrey line $\\mathbb{R}_l$ (the real line with the lower-limit topology) fail to be metrizable?"
		correct := "it is separable but fails to be second-countable"
		dist := distinctDistractors(correct,
			"it fails to be Hausdorff",
			"it fails to be regular",
			"it is compact",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "In any metrizable space, separability implies second-countability. The Sorgenfrey line is separable because Q is dense, but its basis cannot be countable, so it is not metrizable."
		return p
	case 3:
		dim := 1 + rand.Intn(4)
		q := fmt.Sprintf("A compact Hausdorff space $X = (S^1)^{%d}$ is metrizable if and only if $X$ is...", dim)
		correct := "second-countable"
		dist := distinctDistractors(correct,
			"first-countable",
			"totally disconnected",
			"locally compact",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "For any compact Hausdorff space, metrizability is logically equivalent to possessing a countable topological basis."
		return p
	default:
		q := "The Nagata-Smirnov Metrization Theorem characterizes metrizability using which type of basis?"
		correct := "a countably locally finite basis"
		dist := distinctDistractors(correct,
			"a strictly finite basis",
			"a linearly ordered basis",
			"a pairwise disjoint basis",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Nagata-Smirnov states that a space is metrizable if and only if it is regular, Hausdorff and has a countably locally finite basis."
		return p
	}
}

// ---------------------------------------------------------------------------
// 6. Path-Connectedness (topo.connected.path)
// ---------------------------------------------------------------------------

type pathConnectedGen struct{}

func (g *pathConnectedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		dim := 1 + rand.Intn(5)
		if dim == 1 {
			q := "How many path components does the punctured real line $\\mathbb{R} \\setminus \\{0\\}$ have?"
			correct := "2"
			dist := distinctDistractors(correct, "1", "3", "infinitely many")
			p := generator.MultipleChoice(q, correct, dist...)
			p.Explanation = "Removing 0 disconnects R into two open rays (-inf, 0) and (0, inf), which are its two path components."
			return p
		}
		q := fmt.Sprintf("How many path components does the punctured Euclidean space $\\mathbb{R}^{%d} \\setminus \\{0\\}$ have?", dim)
		correct := "1 (it is path-connected)"
		dist := distinctDistractors(correct,
			"2",
			fmt.Sprintf("%d", dim),
			"infinitely many",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("For dimension n >= 2, any two points in R^%d \\ {0} can be joined by a path avoiding the origin.", dim)
		return p
	case 1:
		k := 2 + rand.Intn(5)
		q := fmt.Sprintf("How many path components does the subspace $S = \\bigcup_{j=1}^{%d} (3j, 3j+1) \\subset \\mathbb{R}$ have?", k)
		correct := fmt.Sprintf("%d", k)
		dist := distinctDistractors(correct,
			"1",
			fmt.Sprintf("%d", k-1),
			fmt.Sprintf("%d", k+1),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Each open interval is path-connected, and the gaps between distinct intervals prevent any continuous path between them."
		return p
	case 2:
		pts := 2 + rand.Intn(5)
		q := fmt.Sprintf("If %d distinct points are removed from the real line $\\mathbb{R}$, how many path components does the resulting punctured space have?", pts)
		correct := fmt.Sprintf("%d", pts+1)
		dist := distinctDistractors(correct,
			fmt.Sprintf("%d", pts),
			fmt.Sprintf("%d", pts-1),
			"1",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Removing k points from the real line divides the line into k + 1 disjoint open intervals."
		return p
	case 3:
		q := "The topologist's sine curve $T = \\{(x, \\sin(1/x)) : x \\in (0, 1]\\} \\cup \\{(0, y) : y \\in [-1, 1]\\}$ is a classic example of a space that is..."
		correct := "connected but not path-connected"
		dist := distinctDistractors(correct,
			"path-connected but not connected",
			"neither connected nor path-connected",
			"locally path-connected",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The curve is connected as the closure of a connected graph, but no continuous path can join a point on the graph to the limit segment at x=0."
		return p
	default:
		pts := 1 + rand.Intn(4)
		q := fmt.Sprintf("Let $S^2$ be the unit 2-sphere. If %d distinct points are removed, the resulting subspace $S^2 \\setminus \\{p_1, \\dots, p_%d\\}$ is...", pts, pts)
		correct := "path-connected"
		dist := distinctDistractors(correct,
			fmt.Sprintf("disconnected into %d components", pts+1),
			"compact",
			"not locally path-connected",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Removing a finite number of points from the 2-sphere leaves the surface path-connected."
		return p
	}
}

// ---------------------------------------------------------------------------
// 7. Local Compactness (topo.compact.local)
// ---------------------------------------------------------------------------

type localCompactGen struct{}

func (g *localCompactGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("In Euclidean space $\\mathbb{R}^{%d}$, what standard compact neighborhood exists around any given point $x_0$?", dim)
		correct := fmt.Sprintf("a closed %d-dimensional ball $\\overline{B}(x_0, r)$ of positive radius $r$", dim)
		dist := distinctDistractors(correct,
			"the open ball $B(x_0, r)$",
			"the singleton set $\\{x_0\\}$ only",
			fmt.Sprintf("the entire unbounded space $\\mathbb{R}^{%d}$", dim),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("In R^%d, closed and bounded sets are compact by Heine-Borel, so closed balls provide compact neighborhoods.", dim)
		return p
	case 1:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("Is the punctured Euclidean space $\\mathbb{R}^{%d} \\setminus \\{0\\}$ locally compact?", dim)
		correct := "yes, because open subspaces of locally compact Hausdorff spaces are locally compact"
		dist := distinctDistractors(correct,
			"no, because the missing origin prevents local compactness",
			"only if the dimension is 1",
			"no, because it is unbounded",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Any open subspace of a locally compact Hausdorff space inherits local compactness."
		return p
	case 2:
		a := 1 + rand.Intn(5)
		b := a + 1 + rand.Intn(5)
		q := fmt.Sprintf("Consider the rational subspace $S = [%d, %d] \\cap \\mathbb{Q}$. Is $S$ locally compact in the subspace topology?", a, b)
		correct := "no, because no point in $S$ has a neighborhood with compact closure in $S$"
		dist := distinctDistractors(correct,
			"yes, because it is a bounded subset",
			"yes, because it is countable",
			"no, because it is unbounded",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Every neighborhood in Q contains missing irrational limit points, so closures are never compact in the subspace topology."
		return p
	case 3:
		q := "By Riesz's Lemma, a normed vector space $V$ over $\\mathbb{R}$ is locally compact if and only if..."
		correct := "$V$ is finite-dimensional"
		dist := distinctDistractors(correct,
			"$V$ is a Banach space",
			"$V$ is separable",
			"$V$ possesses an inner product",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "In infinite-dimensional normed vector spaces, the closed unit ball is never compact, so local compactness characterizes finite-dimensional spaces."
		return p
	default:
		n := 2 + rand.Intn(6)
		q := fmt.Sprintf("Consider a discrete space $X$ with %d elements. Is $X$ locally compact?", n)
		correct := "yes, every singleton $\\{x\\}$ is both open and compact"
		dist := distinctDistractors(correct,
			"no, discrete spaces fail compactness axioms",
			"only if $X$ is infinite",
			"no, discrete spaces are not Hausdorff",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "In any discrete space, every singleton {x} is an open neighborhood that is finite and therefore compact."
		return p
	}
}

// ---------------------------------------------------------------------------
// 8. Tietze Extension Theorem (topo.separation.tietze)
// ---------------------------------------------------------------------------

type tietzeGen struct{}

func (g *tietzeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		a := 1 + rand.Intn(5)
		b := a + 1 + rand.Intn(5)
		c := 1 + rand.Intn(4)
		d := c + 1 + rand.Intn(4)
		q := fmt.Sprintf("Let $A = \\{%d, %d\\} \\subset \\mathbb{R}$. If $f(%d) = %d$ and $f(%d) = %d$, does the Tietze Extension Theorem guarantee a continuous extension $F: \\mathbb{R} \\to [%d, %d]$?", a, b, a, c, b, d, c, d)
		correct := fmt.Sprintf("yes, because $A$ is a closed subset of the normal space $\\mathbb{R}$ and $f(A) \\subseteq [%d, %d]$", c, d)
		dist := distinctDistractors(correct,
			"no, extensions only exist on compact intervals",
			"only if the extension is linear",
			"no, finite subsets fail the hypotheses of Tietze",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Finite subsets in R are closed. By Tietze's theorem, any continuous map from a closed subset into a bounded interval extends continuously to the whole space."
		return p
	case 1:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("In the Euclidean space $\\mathbb{R}^{%d}$, if $A$ is a closed subset and $f: A \\to \\mathbb{R}$ is an unbounded continuous function, does $f$ extend continuously to $\\mathbb{R}^{%d}$?", dim, dim)
		correct := fmt.Sprintf("yes, Tietze's theorem guarantees an unbounded continuous extension $F: \\mathbb{R}^{%d} \\to \\mathbb{R}$", dim)
		dist := distinctDistractors(correct,
			"no, the target must be a compact interval",
			"only if $A$ is bounded",
			"only if $A$ is convex",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The full version of Tietze's theorem applies to all real-valued continuous functions without requiring boundedness."
		return p
	case 2:
		a := 1 + rand.Intn(4)
		q := fmt.Sprintf("Consider $X = [%d, %d]$ and the non-closed subset $A = (%d, %d]$. Which continuous function on $A$ fails to extend continuously to $X$, demonstrating that $A$ must be closed in Tietze's theorem?", a, a+2, a, a+2)
		correct := fmt.Sprintf("$f(x) = \\sin(1 / (x - %d))$", a)
		dist := distinctDistractors(correct,
			fmt.Sprintf("$f(x) = (x - %d)^2$", a),
			fmt.Sprintf("$f(x) = 2x + %d$", a),
			"$f(x) = e^x$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The function sin(1/(x-a)) oscillates infinitely as x approaches a, having no limit, so it cannot extend continuously to the boundary point a."
		return p
	case 3:
		q := "A topological space $X$ is normal if and only if every continuous map from a closed subset $A \\subseteq X$ into which target space extends continuously to $X$?"
		correct := "$[-1, 1]$"
		dist := distinctDistractors(correct,
			"$S^1$",
			"$\\{0, 1\\}$ with discrete topology",
			"$\\mathbb{Z}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The ability to extend continuous functions from closed sets into the unit interval is logically equivalent to normality."
		return p
	default:
		q := "How does Urysohn's Lemma relate to the Tietze Extension Theorem?"
		correct := "Urysohn's Lemma is the special case of Tietze applied to two disjoint closed sets mapped to 0 and 1"
		dist := distinctDistractors(correct,
			"Tietze applies only to metric spaces while Urysohn applies to all spaces",
			"they are independent theorems with no logical relation",
			"Urysohn's lemma implies Tietze only for compact spaces",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Setting A = E0 u E1 with f(E0)=0 and f(E1)=1 embeds Urysohn's lemma directly into Tietze's theorem."
		return p
	}
}

// ---------------------------------------------------------------------------
// 9. Homotopy and Fundamental Group (topo.homotopy)
// ---------------------------------------------------------------------------

type homotopyGen struct{}

func (g *homotopyGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(6) {
	case 0:
		n := 1 + rand.Intn(5)
		if n == 1 {
			q := "What is the fundamental group $\\pi_1(S^1)$ of the 1-dimensional circle?"
			correct := "$\\mathbb{Z}$"
			dist := distinctDistractors(correct, "0", "$\\mathbb{Z} \\times \\mathbb{Z}$", "$\\mathbb{Z}_2$")
			p := generator.MultipleChoice(q, correct, dist...)
			p.Explanation = "Loops around the circle are classified by their integer winding number, so pi_1(S^1) = Z."
			return p
		}
		q := fmt.Sprintf("What is the fundamental group $\\pi_1(S^{%d})$ of the %d-dimensional sphere?", n, n)
		correct := "0 (the trivial group)"
		dist := distinctDistractors(correct,
			"$\\mathbb{Z}$",
			fmt.Sprintf("$\\mathbb{Z}_{%d}$", n),
			"$\\mathbb{Z} \\times \\mathbb{Z}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("For n >= 2, the sphere S^%d is simply connected, so its fundamental group is trivial.", n)
		return p
	case 1:
		k := 2 + rand.Intn(4)
		q := fmt.Sprintf("What is the fundamental group $\\pi_1(T^{%d})$ of the %d-dimensional torus $T^{%d} = (S^1)^{%d}$?", k, k, k, k)
		correct := fmt.Sprintf("$\\mathbb{Z}^{%d}$", k)
		dist := distinctDistractors(correct,
			"$\\mathbb{Z}$",
			"0",
			fmt.Sprintf("$\\mathbb{Z}_{%d}$", k),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The fundamental group of a Cartesian product of path-connected spaces is the product of their fundamental groups, yielding Z^%d.", k)
		return p
	case 2:
		k := 1 + rand.Intn(4)
		if k == 1 {
			q := "What is the fundamental group $\\pi_1(\\mathbb{R}^2 \\setminus \\{p\\})$ of the plane punctured at 1 point?"
			correct := "$\\mathbb{Z}$"
			dist := distinctDistractors(correct, "0", "$\\mathbb{Z} \\times \\mathbb{Z}$", "$\\mathbb{R}$")
			p := generator.MultipleChoice(q, correct, dist...)
			p.Explanation = "The punctured plane deformation retracts onto the circle S^1, whose fundamental group is Z."
			return p
		}
		q := fmt.Sprintf("What is the fundamental group of the plane with %d distinct points removed, $\\mathbb{R}^2 \\setminus \\{p_1, \\dots, p_%d\\}$?", k, k)
		correct := fmt.Sprintf("the free group on %d generators $F_%d$", k, k)
		dist := distinctDistractors(correct,
			fmt.Sprintf("$\\mathbb{Z}^{%d}$", k),
			"0",
			"$\\mathbb{Z}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The plane minus %d points deformation retracts onto a wedge sum of %d circles, whose fundamental group is the free group F_%d.", k, k, k)
		return p
	case 3:
		dim := 2 + rand.Intn(5)
		q := fmt.Sprintf("What is the fundamental group $\\pi_1(\\mathbb{R}^{%d} \\setminus \\{0\\})$ of Euclidean space punctured at the origin?", dim)
		if dim == 2 {
			correct := "$\\mathbb{Z}$"
			dist := distinctDistractors(correct, "0", "$\\mathbb{Z}_2$", "$\\mathbb{R}$")
			p := generator.MultipleChoice(q, correct, dist...)
			p.Explanation = "R^2 \\ {0} deformation retracts onto S^1, which has fundamental group Z."
			return p
		}
		correct := "0 (the trivial group)"
		dist := distinctDistractors(correct,
			"$\\mathbb{Z}$",
			fmt.Sprintf("$\\mathbb{Z}_{%d}$", dim),
			"$\\mathbb{R}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("For n >= 3, R^%d \\ {0} deformation retracts onto the sphere S^%d, which is simply connected.", dim, dim-1)
		return p
	case 4:
		q := "What is the fundamental group $\\pi_1(\\mathbb{R}P^2)$ of the real projective plane?"
		correct := "$\\mathbb{Z}_2$"
		dist := distinctDistractors(correct, "$\\mathbb{Z}$", "0", "$\\mathbb{Z} \\times \\mathbb{Z}$")
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The 2-sphere S^2 is a simply connected 2-sheeted universal cover of RP^2, so pi_1(RP^2) = Z_2."
		return p
	default:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("Euclidean space $\\mathbb{R}^{%d}$ is contractible. What is its fundamental group $\\pi_1(\\mathbb{R}^{%d})$?", dim, dim)
		correct := "0 (the trivial group)"
		dist := distinctDistractors(correct,
			"$\\mathbb{Z}$",
			fmt.Sprintf("$\\mathbb{Z}^{%d}$", dim),
			"$\\mathbb{R}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Any contractible space is homotopy equivalent to a single point, so its fundamental group is trivial."
		return p
	}
}

// ---------------------------------------------------------------------------
// 10. Topological Manifolds (topo.manifold)
// ---------------------------------------------------------------------------

type manifoldGen struct{}

func (g *manifoldGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		p := 1 + rand.Intn(4)
		qDim := 1 + rand.Intn(4)
		totalDim := p + qDim
		q := fmt.Sprintf("What is the topological dimension of the product manifold $S^{%d} \\times S^{%d}$?", p, qDim)
		correct := fmt.Sprintf("%d", totalDim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("%d", p*qDim),
			fmt.Sprintf("%d", totalDim+1),
			fmt.Sprintf("%d", totalDim-1),
		)
		prob := generator.MultipleChoice(q, correct, dist...)
		prob.Explanation = fmt.Sprintf("The dimension of a product of manifolds of dimensions %d and %d is %d + %d = %d.", p, qDim, p, qDim, totalDim)
		return prob
	case 1:
		g := rand.Intn(5)
		chi := 2 - 2*g
		q := fmt.Sprintf("What is the Euler characteristic $\\chi$ of a closed orientable surface of genus %d?", g)
		correct := fmt.Sprintf("%d", chi)
		dist := distinctDistractors(correct,
			fmt.Sprintf("%d", 2-g),
			fmt.Sprintf("%d", 2*g),
			fmt.Sprintf("%d", 1-2*g),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The Euler characteristic of a closed orientable surface of genus g is chi = 2 - 2g. For genus %d, chi = %d.", g, chi)
		return p
	case 2:
		n := 2 + rand.Intn(4)
		dimGL := n * n
		q := fmt.Sprintf("What is the dimension of the general linear Lie group $GL_{%d}(\\mathbb{R})$ as a topological manifold?", n)
		correct := fmt.Sprintf("%d", dimGL)
		dist := distinctDistractors(correct,
			fmt.Sprintf("%d", n),
			fmt.Sprintf("%d", dimGL-1),
			fmt.Sprintf("%d", 2*n),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("GL_%d(R) is an open subset of the vector space of %dx%d matrices M_%d(R), so its manifold dimension is %d^2 = %d.", n, n, n, n, n, dimGL)
		return p
	case 3:
		dim := 2 + rand.Intn(5)
		q := fmt.Sprintf("If $M$ is a compact %d-manifold with boundary $\\partial M$, what is the manifold dimension of its boundary $\\partial M$?", dim)
		correct := fmt.Sprintf("%d", dim-1)
		dist := distinctDistractors(correct,
			fmt.Sprintf("%d", dim),
			fmt.Sprintf("%d", dim-2),
			"0",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The boundary of an n-manifold with boundary is an (n-1)-manifold without boundary. For n = %d, the boundary has dimension %d.", dim, dim-1)
		return p
	default:
		q := "A topological $n$-manifold is defined as a topological space that is locally Euclidean of dimension $n$, together with which two topological properties?"
		correct := "Hausdorff and second-countable"
		dist := distinctDistractors(correct,
			"compact and metrizable",
			"connected and regular",
			"normal and first-countable",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A manifold must be Hausdorff to ensure limits and neighborhoods are well-behaved and second-countable to admit partitions of unity."
		return p
	}
}

// ---------------------------------------------------------------------------
// 11. Local Path-Connectedness (topo.connected.local_path)
// ---------------------------------------------------------------------------

type localPathGen struct{}

func (g *localPathGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		dim := 1 + rand.Intn(4)
		q := fmt.Sprintf("In the Euclidean space $\\mathbb{R}^{%d}$, open balls are path-connected. Does this imply that $\\mathbb{R}^{%d}$ is locally path-connected?", dim, dim)
		correct := "yes, the open balls form a topological basis of path-connected neighborhoods"
		dist := distinctDistractors(correct,
			"no, local path-connectedness fails in Euclidean space",
			"only if the space is compactified",
			"only for dimension 1",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A space is locally path-connected if it has a basis of path-connected open sets, which open balls in R^n provide."
		return p
	case 1:
		k := 2 + rand.Intn(5)
		dim := 2 + rand.Intn(3)
		q := fmt.Sprintf("Let $X = \\bigsqcup_{j=1}^{%d} B_j$ be the disjoint union of %d open balls in $\\mathbb{R}^{%d}$. In this locally path-connected space, how do connected components and path components compare?", k, k, dim)
		correct := fmt.Sprintf("there are exactly %d components, and connected components coincide with path components", k)
		dist := distinctDistractors(correct,
			"there are infinitely many path components but only 1 connected component",
			"connected components and path components never coincide",
			"the space is disconnected into single points",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "In any locally path-connected space, path components are open and coincide with connected components."
		return p
	case 2:
		q := "The comb space $C = ([0, 1] \\times \\{0\\}) \\cup (\\{0\\} \\times [0, 1]) \\cup \\bigcup_{n=1}^\\infty (\\{1/n\\} \\times [0, 1])$ is a classic counterexample of a space that is..."
		correct := "path-connected but not locally path-connected"
		dist := distinctDistractors(correct,
			"locally path-connected but not path-connected",
			"neither connected nor path-connected",
			"a 1-dimensional topological manifold",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The comb space is path-connected via its base, but points on the vertical spine at x=0 have arbitrarily small neighborhoods that fail to be path-connected."
		return p
	case 3:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("If a topological space $X$ is both connected and locally path-connected in dimension %d, then $X$ is guaranteed to be...", dim)
		correct := "path-connected"
		dist := distinctDistractors(correct,
			"compact",
			"simply connected",
			"contractible",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "In a locally path-connected space, path components are clopen. If the space is connected, the only non-empty clopen set is the entire space, so it is path-connected."
		return p
	default:
		q := "In the subspace topology of the real line, why is the space of rational numbers $\\mathbb{Q}$ NOT locally path-connected?"
		correct := "every nonempty open set in $\\mathbb{Q}$ is totally disconnected"
		dist := distinctDistractors(correct,
			"$\\mathbb{Q}$ is countable",
			"$\\mathbb{Q}$ is not Hausdorff",
			"singletons in $\\mathbb{Q}$ are not closed",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Every open interval in Q has missing irrational numbers, so every open set is totally disconnected and contains no paths of positive length."
		return p
	}
}

// ---------------------------------------------------------------------------
// 12. One-Point Compactification (topo.compact.one_point)
// ---------------------------------------------------------------------------

type onePointGen struct{}

func (g *onePointGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(6) {
	case 0:
		dim := 1 + rand.Intn(4)
		q := fmt.Sprintf("The one-point (Alexandroff) compactification of Euclidean space $\\mathbb{R}^{%d}$ is homeomorphic to which topological space?", dim)
		correct := fmt.Sprintf("the sphere $S^{%d}$", dim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("the closed disc $D^{%d}$", dim),
			fmt.Sprintf("the torus $T^{%d}$", dim),
			fmt.Sprintf("$\\mathbb{R}^{%d}$", dim),
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("Inverse stereographic projection maps R^%d homeomorphically onto S^%d minus the north pole, so the one-point compactification of R^%d is S^%d.", dim, dim, dim, dim)
		return p
	case 1:
		a := 1 + rand.Intn(5)
		b := a + 1 + rand.Intn(5)
		q := fmt.Sprintf("The one-point compactification of the open interval $(%d, %d)$ is homeomorphic to...", a, b)
		correct := "$S^1$ (the circle)"
		dist := distinctDistractors(correct,
			fmt.Sprintf("[%d, %d]", a, b),
			"a single isolated point",
			"$\\mathbb{R}$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Any open interval is homeomorphic to the real line R, whose one-point compactification is the circle S^1."
		return p
	case 2:
		dim := 1 + rand.Intn(4)
		r := 1 + rand.Intn(5)
		q := fmt.Sprintf("What is the one-point compactification of the open ball $B_{%d}(0, %d) \\subset \\mathbb{R}^{%d}$?", dim, r, dim)
		correct := fmt.Sprintf("the sphere $S^{%d}$", dim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("the closed ball $\\overline{B}_{%d}(0, %d)$", dim, r),
			fmt.Sprintf("the torus $T^{%d}$", dim),
			"a single point",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("The open ball is homeomorphic to R^%d, so adding one point at infinity yields S^%d.", dim, dim)
		return p
	case 3:
		q := "The one-point compactification $X^* = X \\cup \\{\\infty\\}$ of a topological space $X$ is compact and Hausdorff if and only if $X$ is..."
		correct := "locally compact and Hausdorff"
		dist := distinctDistractors(correct,
			"compact",
			"second-countable and connected",
			"metrizable and bounded",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Alexandroff's theorem establishes that X* is Hausdorff if and only if X is locally compact Hausdorff."
		return p
	case 4:
		q := "In the one-point compactification $X^* = X \\cup \\{\\infty\\}$, an open neighborhood of the point at infinity $\\infty$ has the form..."
		correct := "$(X \\setminus K) \\cup \\{\\infty\\}$ where $K \\subseteq X$ is compact and closed"
		dist := distinctDistractors(correct,
			"$K \\cup \\{\\infty\\}$ where $K$ is open in $X$",
			"$\\{\\infty\\}$ as an isolated open point",
			"$U \\cap X$ for any open subset $U$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The topology of the one-point compactification defines open neighborhoods of infinity as complements of compact closed subsets of X."
		return p
	default:
		n := 2 + rand.Intn(6)
		q := fmt.Sprintf("If a compact Hausdorff space $X$ with %d points is compactified with one point $X^* = X \\cup \\{\\infty\\}$, what role does $\\infty$ play in $X^*$?", n)
		correct := "$\\infty$ is an isolated point, so $X^* = X \\sqcup \\{\\infty\\}$ is a discrete union"
		dist := distinctDistractors(correct,
			"$\\infty$ is a limit point of all points in $X$",
			"$X^*$ is no longer compact",
			"$X^*$ is no longer Hausdorff",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "When X is already compact, the complement X \\ X = empty set is compact, which means {infinity} is open, making infinity an isolated point."
		return p
	}
}

// ---------------------------------------------------------------------------
// 13. Complete Regularity (topo.separation.completely_regular)
// ---------------------------------------------------------------------------

type completelyRegularGen struct{}

func (g *completelyRegularGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(6) {
	case 0:
		a := 1 + rand.Intn(4)
		b := a + 1 + rand.Intn(3)
		c := b + 1 + rand.Intn(3)
		q := fmt.Sprintf("Consider $\\mathbb{R}$ with standard topology, closed set $A = [%d, %d]$ and point $x_0 = %d \\notin A$. Which function $f: \\mathbb{R} \\to [0, 1]$ satisfies complete regularity ($f(x_0)=1$ and $f|_A=0$)?", b, c, a)
		correct := fmt.Sprintf("$f(x) = \\max(0, 1 - \\frac{|x - %d|}{%d})$", a, b-a)
		dist := distinctDistractors(correct,
			fmt.Sprintf("$f(x) = \\frac{|x - %d|}{%d}$", a, b-a),
			"$f(x) = 1$ for all $x$",
			"$f(x) = 0$ for all $x$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = fmt.Sprintf("At x = %d, f(%d) = 1. For x in [%d, %d], |x - %d| >= %d, so f(x) = 0.", a, a, b, c, a, b-a)
		return p
	case 1:
		dim := 1 + rand.Intn(5)
		q := fmt.Sprintf("Is the Euclidean space $\\mathbb{R}^{%d}$ completely regular ($T_{3.5}$)?", dim)
		correct := fmt.Sprintf("yes, because $\\mathbb{R}^{%d}$ is a metric space and all metric spaces are completely regular", dim)
		dist := distinctDistractors(correct,
			fmt.Sprintf("no, $\\mathbb{R}^{%d}$ fails complete regularity for $n \\ge 2$", dim),
			"only if restricted to bounded domains",
			"no, complete regularity requires compact spaces",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Every metric space is normal (T_4), and every normal space is completely regular (T_3.5)."
		return p
	case 2:
		q := "What is the correct logical chain of separation axioms?"
		correct := "normal ($T_4$) $\\implies$ completely regular ($T_{3.5}$) $\\implies$ regular ($T_3$) $\\implies$ Hausdorff ($T_2$)"
		dist := distinctDistractors(correct,
			"regular $\\implies$ completely regular $\\implies$ normal $\\implies$ Hausdorff",
			"Hausdorff $\\implies$ regular $\\implies$ completely regular $\\implies$ normal",
			"completely regular $\\implies$ normal $\\implies$ regular $\\implies$ Hausdorff",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Normality implies complete regularity by Urysohn's Lemma, complete regularity implies regularity, and regularity with T_1 implies Hausdorff."
		return p
	case 3:
		k := 2 + rand.Intn(5)
		q := fmt.Sprintf("Consider a product of %d completely regular topological spaces $\\prod_{j=1}^{%d} X_j$. Is the product space guaranteed to be completely regular?", k, k)
		correct := "yes, complete regularity is preserved under arbitrary Cartesian products"
		dist := distinctDistractors(correct,
			"no, products of completely regular spaces fail regularity",
			"only for finite products of compact spaces",
			"only if each factor is metrizable",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "Complete regularity is productive and hereditary, meaning both arbitrary products and arbitrary subspaces remain completely regular."
		return p
	case 4:
		q := "The Niemytzki plane (the upper half-plane with tangent-disc topology at the boundary) is a classic example of a space that is..."
		correct := "completely regular and Hausdorff, but NOT normal"
		dist := distinctDistractors(correct,
			"normal, but NOT completely regular",
			"regular, but NOT Hausdorff",
			"compact and metrizable",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "The Niemytzki plane is completely regular (T_3.5) and Hausdorff, but fails normality because the real axis is a discrete closed subspace whose subsets cannot be separated by disjoint open sets."
		return p
	default:
		q := "By the Tychonoff Embedding Theorem, a topological space $X$ is Tychonoff (completely regular and $T_1$) if and only if $X$ is homeomorphic to a subspace of..."
		correct := "a cube $[0, 1]^J$ for some indexing set $J$"
		dist := distinctDistractors(correct,
			"the Euclidean space $\\mathbb{R}^n$",
			"a discrete space",
			"the Hilbert space $\\ell^2$",
		)
		p := generator.MultipleChoice(q, correct, dist...)
		p.Explanation = "A space is Tychonoff if and only if it embeds as a subspace of a product of unit intervals [0, 1]^J."
		return p
	}
}
