package topology

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func Register(reg *generator.Registry) {
	reg.Register("topo.basics.metric", &metricGen{})
	reg.Register("topo.basics.open_closed", &openClosedGen{})
	reg.Register("topo.basics.continuous", &continuousGen{})
	reg.Register("topo.compact", &compactGen{})
	reg.Register("topo.connected", &connectedGen{})
	reg.Register("topo.hausdorff", &hausdorffGen{})
	reg.Register("topo.quotient", &quotientGen{})
	reg.Register("topo.product", &productGen{})
	reg.Register("topo.urysohn", &urysohnGen{})
	reg.Register("topo.metrization", &metrizationGen{})
	reg.Register("topo.connected.path", &pathConnectedGen{})
	reg.Register("topo.compact.local", &localCompactGen{})
	reg.Register("topo.separation.tietze", &tietzeGen{})
	reg.Register("topo.homotopy", &homotopyGen{})
	reg.Register("topo.manifold", &manifoldGen{})
	reg.Register("topo.connected.local_path", &localPathGen{})
	reg.Register("topo.compact.one_point", &onePointGen{})
	reg.Register("topo.separation.completely_regular", &completelyRegularGen{})
	reg.Register("topo.homotopy.fundamental", &fundamentalGen{})
	reg.Register("topo.manifold.orientability", &orientabilityGen{})
	reg.Register("topo.embedding", &embeddingGen{})
	reg.Register("topo.knot", &knotGen{})
	reg.Register("topo.cohomology", &cohomologyGen{})
	reg.Register("topo.compactification.stone_cech", &stoneCechGen{})
	reg.Register("topo.dimension", &dimensionGen{})
	reg.Register("topo.covering", &coveringGen{})
	reg.Register("topo.fundamental.van_kampen", &vanKampenGen{})
	reg.Register("topo.fiber_bundle", &fiberBundleGen{})
	reg.Register("topo.homology.cellular", &cellularHomologyGen{})
	reg.Register("topo.cohomology.de_rham", &deRhamGen{})
	registerBatch(reg)
}

// ----- 1. metric -----

type metricGen struct{}

func (g *metricGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	for attempt := 0; attempt < 100; attempt++ {
		mode := rand.Intn(3)
		switch mode {
		case 0:
			// Taxicab L1 distance in R^2
			x1 := rand.Intn(7) + 1
			y1 := rand.Intn(7) + 1
			dx := rand.Intn(6) + 2
			dy := rand.Intn(6) + 2
			x2 := x1 + dx
			y2 := y1 + dy
			dist := dx + dy
			ansStr := strconv.Itoa(dist)
			q := fmt.Sprintf("Under the taxicab metric \\(d_1(P, Q) = |x_1 - x_2| + |y_1 - y_2|\\) on \\(\\mathbb{R}^2\\), compute the distance between the points \\((%d, %d)\\) and \\((%d, %d)\\).", x1, y1, x2, y2)
			if strings.Contains(q, ansStr) {
				continue
			}
			ex := fmt.Sprintf("The coordinate differences are \\(|%d - %d| = %d\\) and \\(|%d - %d| = %d\\). Summing them yields \\(d_1 = %d + %d = %d\\).",
				x1, x2, dx, y1, y2, dy, dx, dy, dist)
			return generator.Problem{
				Question:    q,
				Answer:      ansStr,
				Explanation: ex,
				Facts: map[string]string{
					"x1":     strconv.Itoa(x1),
					"y1":     strconv.Itoa(y1),
					"x2":     strconv.Itoa(x2),
					"y2":     strconv.Itoa(y2),
					"answer": ansStr,
				},
			}
		case 1:
			// Discrete metric ball cardinality for radius > 1
			n := rand.Intn(15) + 6 // 6 to 20
			ansStr := strconv.Itoa(n)
			q := fmt.Sprintf("Let \\(X\\) be a finite set with %d distinct elements equipped with the discrete metric. How many points are contained in the open ball \\(B_2(x_0) = \\{y \\in X : d(x_0, y) < 2\\}\\)?", n)
			if strings.Contains(q, ansStr) {
				continue
			}
			ex := fmt.Sprintf("In the discrete metric, any two distinct elements have distance 1, and every element has distance 0 from itself. Because radius 2 exceeds 1, every point in \\(X\\) satisfies \\(d(x_0, y) < 2\\). Thus the ball equals the entire set \\(X\\), containing %d elements.", n)
			return generator.Problem{
				Question:    q,
				Answer:      ansStr,
				Explanation: ex,
				Facts: map[string]string{
					"n":      strconv.Itoa(n),
					"answer": ansStr,
				},
			}
		default:
			// Triangle inequality maximum bound
			dXY := rand.Intn(10) + 4
			dYZ := rand.Intn(8) + 3
			for dXY == dYZ {
				dYZ = rand.Intn(8) + 3
			}
			maxBound := dXY + dYZ
			ansStr := strconv.Itoa(maxBound)
			q := fmt.Sprintf("In a metric space \\((X, d)\\), three points satisfy \\(d(x, y) = %d\\) and \\(d(y, z) = %d\\). What is the maximum possible value of the distance \\(d(x, z)\\)?", dXY, dYZ)
			if strings.Contains(q, ansStr) {
				continue
			}
			ex := fmt.Sprintf("By the triangle inequality, \\(d(x, z) \\le d(x, y) + d(y, z)\\). Substituting gives \\(d(x, z) \\le %d + %d = %d\\).", dXY, dYZ, maxBound)
			return generator.Problem{
				Question:    q,
				Answer:      ansStr,
				Explanation: ex,
				Facts: map[string]string{
					"d_xy":   strconv.Itoa(dXY),
					"d_yz":   strconv.Itoa(dYZ),
					"answer": ansStr,
				},
			}
		}
	}
	return generator.Problem{
		Question:    "Under the taxicab metric on \\(\\mathbb{R}^2\\), compute the distance between \\((1, 1)\\) and \\((4, 5)\\).",
		Answer:      "7",
		Explanation: "The sum of absolute differences is 3 + 4 = 7.",
	}
}

// ----- 2. open_closed -----

type openClosedGen struct{}

func (g *openClosedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Is (0,1) open, closed, both, or neither? (type one word)", "open", "Every point of (0,1) has an ε-neighborhood inside it, but 0 and 1 are limit points outside, so open and not closed."},
		{"Is [0,1] open, closed, both, or neither? (type one word)", "closed", "[0,1] contains all its limit points, so closed; endpoints have no ε-neighborhood inside, so not open."},
		{"Is R open, closed, both, or neither? (type one word)", "both", "R is open by definition and closed because its complement (empty set) is open."},
		{"Is (0,1] open, closed, both, or neither? (type one word)", "neither", "(0,1] misses limit point 0 so not closed, and 1 has no ε-neighborhood inside so not open."},
		{"Is the interval/set (0,1) open in R? (yes/no)", "yes", "(0,1) is open in R because every point has an ε-neighborhood contained in the interval."},
		{"Is the interval/set (0,1) closed in R? (yes/no)", "no", "(0,1) is not closed because 0 and 1 are limit points not in the set."},
	}
	hard := []entry{
		{"Is the interval/set {0} closed in R? (yes/no)", "yes", "A singleton {0} is closed in R (it contains its only limit point)."},
		{"Is the interval/set {1/n : n∈N} ∪ {0} closed in R? (yes/no)", "yes", "This set is closed because it contains its limit point 0."},
		{"Is the interval/set (0,∞) open in R? (yes/no)", "yes", "(0,∞) is open in R because every positive x has an ε-neighborhood in (0,∞)."},
		{"Is the interval/set [0,∞) closed in R? (yes/no)", "yes", "[0,∞) is closed in R because it contains all its limit points."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 3. continuous -----

type continuousGen struct{}

func (g *continuousGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"At which point is f(x)=1/x discontinuous? (type like x=0)", "x=0", "The limit as x→0 does not exist (approaches ±∞), so 1/x breaks at x=0."},
		{"Which function is discontinuous at x=0: x^2 or 1/x? (type it)", "1/x", "x^2 is a polynomial hence continuous everywhere; 1/x blows up at x=0."},
		{"Is \\(f(x)=x^{2}\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "yes", "f(x)=x^2 is a polynomial, and all polynomials are continuous on R."},
		{"Is the floor function \\(\\lfloor x \\rfloor\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "no", "The floor function is discontinuous at every integer because the left and right limits differ."},
		{"Is \\(f(x)=|x|\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "yes", "f(x)=|x| is continuous on R (although not differentiable at x=0, continuity holds)."},
		{"Is \\(f(x)=\\tan(x)\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "no", "tan(x) is discontinuous at x=π/2 + nπ where the function approaches ±∞."},
	}
	hard := []entry{
		{"Is \\(f(x)=\\sin(x)\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "yes", "sin(x) is continuous on R — a trigonometric function with no breaks."},
		{"Is \\(f(x)=e^{x}\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "yes", "f(x)=e^x is continuous on R as an exponential function."},
		{"Is \\(f(x)=\\sqrt{x}\\) continuous on \\([0,\\infty)\\)? (yes/no)", "yes", "f(x)=√x is continuous on its domain [0,∞)."},
		{"Is \\(f(x)=1/(x^{2}-1)\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "no", "f(x)=1/(x^2-1) is discontinuous at x=±1 where the denominator is zero."},
		{"Is \\(f(x)=x^{3}-3x+1\\) continuous on \\(\\mathbb{R}\\)? (yes/no)", "yes", "f(x)=x^3-3x+1 is a polynomial, and all polynomials are continuous on R."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 4. compact -----

type compactGen struct{}

func (g *compactGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Which Heine-Borel condition fails for (0,1): closed or bounded? (type one word)", "closed", "(0,1) is bounded but misses limit points 0 and 1, so it fails closedness."},
		{"Which Heine-Borel condition fails for R: closed or bounded? (type one word)", "bounded", "R is closed but unbounded, so it fails boundedness and is not compact."},
		{"Is the set [0,1] compact in \\(\\mathbb{R}\\)? (yes/no)", "yes", "Heine-Borel: [0,1] is closed and bounded in R, hence compact."},
		{"Is the set (0,1) compact in \\(\\mathbb{R}\\)? (yes/no)", "no", "(0,1) is not closed (0 and 1 are limit points outside), so not compact."},
		{"Is the set R compact in \\(\\mathbb{R}\\)? (yes/no)", "no", "R is not bounded, so not compact (Heine-Borel)."},
	}
	hard := []entry{
		{"Is the set \\(\\{0\\} \\cup \\{1/n : n\\in\\mathbb{N}\\}\\) compact in \\(\\mathbb{R}\\)? (yes/no)", "yes", "This set is closed and bounded, and every open cover has a finite subcover."},
		{"Is the set \\((0,1]\\) compact in \\(\\mathbb{R}\\)? (yes/no)", "no", "(0,1] is not closed (0 missing), so not compact."},
		{"Is the set \\([0,1] \\cup [2,3]\\) compact in \\(\\mathbb{R}\\)? (yes/no)", "yes", "Finite union of compact sets is compact; both intervals are compact."},
		{"Is the set \\(\\{x : x\\geq 0\\}\\) compact in \\(\\mathbb{R}\\)? (yes/no)", "no", "Unbounded above, so not compact."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 5. connected -----

type connectedGen struct{}

func (g *connectedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"How many connected components does (0,1) union (2,3) have?", "2", "The two disjoint open intervals separate, giving one component each."},
		{"How many connected components does R have?", "1", "R is connected (intervals are connected), so a single component."},
		{"Is the set (0,1) connected in \\(\\mathbb{R}\\)? (yes/no)", "yes", "(0,1) is an interval, hence connected in R."},
		{"Is the set Q connected in \\(\\mathbb{R}\\)? (yes/no)", "no", "Q is totally disconnected: between any two rationals an irrational separates them."},
		{"Is the set \\(\\{0,1\\}\\) connected in \\(\\mathbb{R}\\)? (yes/no)", "no", "Two isolated points can be separated by disjoint opens, so not connected."},
	}
	hard := []entry{
		{"Is the set \\(\\mathbb{R}\\) connected in \\(\\mathbb{R}\\)? (yes/no)", "yes", "R is connected (intervals are connected)."},
		{"Is the set \\([0,1]\\) connected in \\(\\mathbb{R}\\)? (yes/no)", "yes", "Closed intervals are connected."},
		{"Is the set \\((0,1]\\) connected in \\(\\mathbb{R}\\)? (yes/no)", "yes", "(0,1] is an interval, hence connected."},
		{"Is the set \\((0,1) \\cup (2,3)\\) connected in \\(\\mathbb{R}\\)? (yes/no)", "no", "Two disjoint open intervals can be separated by open sets, so not connected."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 6. hausdorff -----

type hausdorffGen struct{}

func (g *hausdorffGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Which topology on {a,b} is not Hausdorff: discrete or indiscrete? (type one word)", "indiscrete", "In the indiscrete topology only {} and the whole set are open, so distinct points cannot be separated."},
		{"Which topology on R is not Hausdorff: usual, discrete, or cofinite? (type one word)", "cofinite", "Any two nonempty cofinite opens intersect (complements are finite), so separation fails."},
		{"Is R with the usual metric Hausdorff? (yes/no)", "yes", "Every metric space is Hausdorff: distinct points have disjoint epsilon-balls."},
		{"Is the cofinite topology on R Hausdorff? (yes/no)", "no", "Any two nonempty opens intersect (their complements are finite), so separation fails."},
	}
	hard := []entry{
		{"Is the discrete topology on any set Hausdorff? (yes/no)", "yes", "Discrete spaces are metric (discrete metric), hence Hausdorff."},
		{"Is R^2 with the Euclidean metric Hausdorff? (yes/no)", "yes", "Metric spaces are Hausdorff."},
		{"Is the Sierpinski space Hausdorff? (yes/no)", "no", "Points cannot be separated by disjoint opens in the Sierpinski topology."},
		{"Is the indiscrete topology on {a,b} Hausdorff? (yes/no)", "no", "In indiscrete topology only {} and the whole set are open, so distinct points cannot be separated."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 7. quotient -----

type quotientGen struct{}

func (g *quotientGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"What space is [0,1] with 0~1 homeomorphic to? (type like S^1)", "S^1", "Glueing the endpoints of an interval closes the loop into a circle."},
		{"Is the quotient map \\(q: X \\to X/\\sim\\) always continuous by definition? (yes/no)", "yes", "The quotient topology is the finest topology making q continuous."},
		{"Is the quotient of a compact space by any equivalence relation always Hausdorff? (yes/no)", "no", "The quotient of compact need not be Hausdorff; e.g. collapsing a non-closed set."},
		{"Does the quotient topology make the projection an open map? (yes/no)", "no", "Quotient maps need not be open; they are continuous and closed under compact-Hausdorff hypotheses."},
	}
	hard := []entry{
		{"What space is the disc D^2 with its boundary collapsed to a point homeomorphic to? (type like S^2)", "S^2", "Collapsing the boundary circle of a 2-disc closes it into a 2-sphere."},
		{"Is the quotient of a compact Hausdorff space by a closed equivalence relation Hausdorff? (yes/no)", "yes", "Closed saturated sets keep the quotient Hausdorff in the compact Hausdorff setting."},
		{"If \\(X=[0,1]\\) and \\(0\\sim1\\), is \\(X/\\sim\\) homeomorphic to \\(S^{1}\\)? (yes/no)", "yes", "Glueing the endpoints of an interval yields a circle."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 8. product -----

type productGen struct{}

func (g *productGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Which topology is finer on infinite products: box or product? (type one word)", "box", "The box topology constrains every factor, giving strictly more opens than the product topology."},
		{"Are projections \\(\\pi_i\\) from a product continuous in the product topology? (yes/no)", "yes", "The product topology is the coarsest making all projections continuous."},
		{"Is the product of two compact spaces compact? (yes/no)", "yes", "Finite products of compact spaces are compact (finite Tychonoff)."},
		{"Does a basic product open constrain infinitely many coordinates? (yes/no)", "no", "A basic open constrains only finitely many factors; the rest are whole spaces."},
	}
	hard := []entry{
		{"Is an arbitrary product of compact spaces compact (Tychonoff)? (yes/no)", "yes", "Tychonoff: arbitrary products of compact spaces are compact."},
		{"Is the box topology the same as the product topology on infinite products? (yes/no)", "no", "Box has more opens; they agree only for finite products."},
		{"Is \\(\\mathbb{R}^{\\mathbb{N}}\\) with product topology metrizable? (yes/no)", "yes", "Countable product of metrizable spaces is metrizable."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 9. urysohn -----

type urysohnGen struct{}

func (g *urysohnGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"What interval does the Urysohn function map into? (type like [0,1])", "[0,1]", "Urysohn builds f: X→[0,1] with f=0 on A and f=1 on B."},
		{"What value does the Urysohn function take on the closed set A?", "0", "The lemma constructs f with f(A)=0 and f(B)=1 for disjoint closed A,B."},
		{"Does Urysohn's lemma separate disjoint closed sets in a normal space? (yes/no)", "yes", "In normal spaces, disjoint closed sets can be separated by a continuous function."},
		{"Does Urysohn's lemma apply without normality? (yes/no)", "no", "Normality is the key hypothesis; without it the separating function may not exist."},
	}
	hard := []entry{
		{"Is every metric space normal (hence Urysohn's lemma applies)? (yes/no)", "yes", "Metric spaces are normal, so the lemma holds."},
		{"Does Urysohn's lemma require the space to be Hausdorff? (yes/no)", "no", "It requires normality, which implies Hausdorff only in some formulations; normality is the key."},
		{"Can Urysohn's function be chosen with \\(f(A)=0\\) and \\(f(B)=1\\) for disjoint closed \\(A,B\\)? (yes/no)", "yes", "The lemma constructs f with exactly those values."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 10. metrization -----

type metrizationGen struct{}

func (g *metrizationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Which line is not metrizable: real or Sorgenfrey? (type one word)", "Sorgenfrey", "The Sorgenfrey line is regular and Hausdorff but fails the countability hypotheses for metrizability."},
		{"How many hypotheses does Urysohn metrization combine: 2 or 3? (type a number)", "3", "Regular + Hausdorff + second-countable: three hypotheses imply metrizable."},
		{"Is every compact Hausdorff second-countable space metrizable? (yes/no)", "yes", "Compact Hausdorff + second-countable → metrizable by Urysohn."},
		{"Is the Sorgenfrey line metrizable? (yes/no)", "no", "The Sorgenfrey line is regular and Hausdorff but not second-countable in the metrizable sense; it is not metrizable."},
	}
	hard := []entry{
		{"Does Urysohn metrization state that a second-countable regular Hausdorff space is metrizable? (yes/no)", "yes", "Regular + second-countable + Hausdorff implies metrizable."},
		{"Does metrizability imply the space is Hausdorff? (yes/no)", "yes", "Every metric space is Hausdorff, so metrizable spaces are Hausdorff."},
		{"Are metrizable spaces normal? (yes/no)", "yes", "Metric spaces are normal, so metrizable spaces inherit normality."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 11. path-connected -----

type pathConnectedGen struct{}

func (g *pathConnectedGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"How many path components does (0,1) union (2,3) have?", "2", "No path can jump between the disjoint intervals, so each is its own component."},
		{"Does path-connected imply connected? (yes/no)", "yes", "Path-connected spaces are always connected."},
		{"Is [0,1] path-connected? (yes/no)", "yes", "[0,1] is an interval, hence path-connected via linear paths."},
		{"Is Q path-connected? (yes/no)", "no", "Q is totally disconnected, hence not path-connected."},
	}
	hard := []entry{
		{"Is the topologist's sine curve connected but not path-connected? (yes/no)", "yes", "Classic example: connected but no path between the curve and the limit segment."},
		{"Does path-connectedness imply local path-connectedness? (yes/no)", "no", "Comb space is path-connected but not locally path-connected."},
		{"Is R^2 minus a point path-connected? (yes/no)", "yes", "Punctured plane remains path-connected (go around the hole)."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 12. local compact -----

type localCompactGen struct{}

func (g *localCompactGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Which space is not locally compact: R or Q? (type one word)", "Q", "Q has no compact neighbourhood of any point in the subspace topology."},
		{"Is R locally compact? (yes/no)", "yes", "Every point has a compact neighbourhood [x-ε,x+ε]."},
		{"Is Q locally compact? (yes/no)", "no", "Q has no compact neighbourhood of any point in the subspace topology."},
		{"Is [0,1] locally compact? (yes/no)", "yes", "Compact spaces are locally compact."},
	}
	hard := []entry{
		{"Is R^n locally compact? (yes/no)", "yes", "Euclidean spaces are locally compact."},
		{"Is an infinite-dimensional Hilbert space locally compact? (yes/no)", "no", "Unit ball is not compact in infinite dimensions, so not locally compact."},
		{"Does one-point compactification apply to locally compact Hausdorff spaces? (yes/no)", "yes", "One-point compactification gives a compact Hausdorff space."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 13. Tietze -----

type tietzeGen struct{}

func (g *tietzeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"From which subsets does Tietze extend functions: open or closed? (type one word)", "closed", "Tietze extends continuous functions from closed subsets of normal spaces."},
		{"Does Tietze extension extend continuous f: A→[0,1] from closed A⊂X normal to all X? (yes/no)", "yes", "Tietze: closed subset of normal space, bounded continuous function extends."},
		{"Does Tietze fail if A is not closed? (yes/no)", "yes", "Closedness is essential; e.g. Q ⊂ R not closed cannot extend all functions."},
		{"Does Tietze extend from non-closed sets in general? (yes/no)", "no", "Closedness is essential; extension from arbitrary subsets fails in general."},
	}
	hard := []entry{
		{"Can any continuous f: A→R from closed A in normal X be extended to X? (yes/no)", "yes", "Full Tietze for real-valued (possibly unbounded) functions on normal spaces."},
		{"Does Tietze imply Urysohn's lemma? (yes/no)", "yes", "Urysohn is special case of Tietze with two closed sets."},
		{"Is Tietze equivalent to normality? (yes/no)", "yes", "Normal iff every bounded continuous function from a closed subset extends."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 14. homotopy -----

type homotopyGen struct{}

func (g *homotopyGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"What is pi_1(S^1)? (type like Z)", "Z", "Loops around the circle are classified by winding number, an integer."},
		{"What is pi_1(S^2)? (type 0 for trivial)", "0", "The 2-sphere is simply connected: every loop contracts."},
		{"Is simply connected equivalent to pi_1 trivial? (yes/no)", "yes", "Every loop contracts to a point iff fundamental group is trivial."},
		{"Does homotopy define an equivalence relation? (yes/no)", "yes", "Reflexive, symmetric, transitive via glueing homotopies."},
	}
	hard := []entry{
		{"Is π_1(S^2) trivial? (yes/no)", "yes", "2-sphere is simply connected."},
		{"Is π_1(T^2) ≅ Z×Z? (yes/no)", "yes", "Torus fundamental group is product of two circle factors."},
		{"Does covering space p: R→S^1 give universal cover of circle? (yes/no)", "yes", "R is simply connected cover of S^1."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 15. manifold -----

type manifoldGen struct{}

func (g *manifoldGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"What dimension is S^2 as a manifold?", "2", "Each point of the sphere has an R^2 neighbourhood via stereographic charts."},
		{"What dimension is S^1 as a manifold?", "1", "Each point of the circle has an interval (R) neighbourhood."},
		{"Is S^1 a 1-manifold? (yes/no)", "yes", "Circle locally looks like an interval (R)."},
		{"Is a manifold required to be Hausdorff and second-countable? (yes/no)", "yes", "Standard definition includes Hausdorff and second-countable plus locally Euclidean."},
	}
	hard := []entry{
		{"Is torus T^2 = S^1×S^1 a compact 2-manifold? (yes/no)", "yes", "Product of compact manifolds is compact manifold."},
		{"Is every manifold metrizable? (yes/no)", "yes", "Urysohn plus manifold hypotheses imply metrizable."},
		{"Is R with discrete topology a 0-manifold? (yes/no)", "yes", "Discrete spaces are 0-dimensional manifolds."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 16. locally path-connected -----

type localPathGen struct{}

func (g *localPathGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Which space is path-connected but not locally path-connected: interval or comb? (type one word)", "comb", "The comb space is path-connected but has no path-connected neighbourhood at the baseline."},
		{"Is R locally path-connected? (yes/no)", "yes", "Intervals are path-connected neighbourhoods."},
		{"Is Q locally path-connected? (yes/no)", "no", "No neighbourhood in Q contains a path-connected open set."},
		{"Does locally path-connected + connected imply path-connected? (yes/no)", "yes", "Classic theorem for locally path-connected spaces."},
	}
	hard := []entry{
		{"Is comb space locally path-connected at the base? (yes/no)", "no", "The baseline has no path-connected neighbourhood."},
		{"Is topologist's sine curve locally connected? (yes/no)", "no", "No small connected neighbourhood of points on the limit segment."},
		{"Does manifold ⇒ locally path-connected? (yes/no)", "yes", "Manifolds are locally Euclidean, hence locally path-connected."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 17. one-point compactification -----

type onePointGen struct{}

func (g *onePointGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"What is the one-point compactification of R? (type like S^1)", "S^1", "R ∪ {∞} ≅ S^1 via stereographic projection."},
		{"What is the one-point compactification of R^2? (type like S^2)", "S^2", "R^2 ∪ {∞} ≅ S^2 via stereographic projection."},
		{"Is locally compact Hausdorff required for one-point compactification? (yes/no)", "yes", "Alexandroff compactification needs locally compact Hausdorff."},
		{"Is one-point compactification of R homeomorphic to S^1? (yes/no)", "yes", "R ∪ {∞} ≅ S^1."},
	}
	hard := []entry{
		{"Is one-point compactification of Q compact? (yes/no)", "no", "Q is not locally compact, so Alexandroff does not give Hausdorff compact."},
		{"Does one-point compactification add exactly one point? (yes/no)", "yes", "By definition."},
		{"Is compact Hausdorff space its own one-point compactification? (yes/no)", "no", "Already compact, no point added."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 18. completely regular -----

type completelyRegularGen struct{}

func (g *completelyRegularGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"What interval do the separating functions map into? (type like [0,1])", "[0,1]", "Complete regularity separates x from closed A by f: X→[0,1] with f(x)=1, f|_A=0."},
		{"Is every metric space completely regular? (yes/no)", "yes", "Metric ⇒ normal ⇒ completely regular."},
		{"Is every Hausdorff space completely regular? (yes/no)", "no", "Hausdorff alone does not give separating functions; complete regularity is strictly stronger."},
		{"Does completely regular mean points and closed sets separated by functions? (yes/no)", "yes", "For each closed A and x∉A, ∃f:X→[0,1] with f(x)=1, f|_A=0."},
	}
	hard := []entry{
		{"Is Tychonoff = completely regular + T1? (yes/no)", "yes", "Tychonoff spaces are T1 and completely regular."},
		{"Does product of completely regular spaces remain completely regular? (yes/no)", "yes", "Product preserves complete regularity."},
		{"Is Niemytzki plane completely regular? (yes/no)", "yes", "It is Tychonoff."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

// ----- 19. fundamental group calculations -----

type fundamentalGen struct{}

func (g *fundamentalGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(6)

	switch mode {
	case 0:
		// n-torus T^n = (S^1)^n, pi_1(T^n) = Z^n (rank n).
		n := rand.Intn(5) + 2
		q := fmt.Sprintf("What is the rank (number of independent \\(\\mathbb{Z}\\) factors) of the fundamental group \\(\\pi_1(T^{%d})\\) of the %d-dimensional torus \\(T^{%d} = (S^1)^{%d}\\)?",
			n, n, n, n)
		exp := fmt.Sprintf("The %d-torus is the product of %d circles, so \\(\\pi_1(T^{%d}) \\cong \\pi_1(S^1)^{%d} \\cong \\mathbb{Z}^{%d}\\). The rank of \\(\\mathbb{Z}^{%d}\\) is %d.",
			n, n, n, n, n, n, n)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(n),
			Explanation: exp,
			Facts: map[string]string{
				"kind":   "torus",
				"n":      strconv.Itoa(n),
				"answer": strconv.Itoa(n),
			},
		}

	case 1:
		// Wedge of k circles, pi_1 = F_k (free group of rank k).
		k := rand.Intn(6) + 2
		q := fmt.Sprintf("By the Seifert-van Kampen theorem, the fundamental group of a wedge sum of %d circles \\(\\bigvee_{i=1}^{%d} S^1\\) is the free group \\(F_{%d}\\). What is the rank (number of free generators) of \\(\\pi_1\\left(\\bigvee_{i=1}^{%d} S^1\\right)\\)?",
			k, k, k, k)
		exp := fmt.Sprintf("A wedge of %d circles at a common basepoint has fundamental group isomorphic to the free product of %d copies of \\(\\pi_1(S^1) \\cong \\mathbb{Z}\\), which is the free group on %d generators. The rank is %d.",
			k, k, k, k)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(k),
			Explanation: exp,
			Facts: map[string]string{
				"kind":   "wedge_circles",
				"k":      strconv.Itoa(k),
				"answer": strconv.Itoa(k),
			},
		}

	case 2:
		// Punctured plane R^2 \ {p_1, ..., p_k}, deformation retracts to wedge of k circles.
		k := rand.Intn(6) + 2
		q := fmt.Sprintf("The Euclidean plane minus %d distinct points, \\(\\mathbb{R}^2 \\setminus \\{p_1, \\dots, p_{%d}\\}\\), deformation retracts onto a wedge sum of circles. What is the rank of its fundamental group?",
			k, k)
		exp := fmt.Sprintf("Removing %d points from \\(\\mathbb{R}^2\\) leaves a space homotopy equivalent to a bouquet of %d circles. Its fundamental group is the free group on %d generators, so the rank is %d.",
			k, k, k, k)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(k),
			Explanation: exp,
			Facts: map[string]string{
				"kind":   "punctured_plane",
				"points": strconv.Itoa(k),
				"answer": strconv.Itoa(k),
			},
		}

	case 3:
		// Lens space L(p, q), pi_1 = Z_p (order p).
		primes := []int{2, 3, 5, 7, 11, 13}
		p := primes[rand.Intn(len(primes))]
		qVal := 1
		for qCandidate := 2; qCandidate < p; qCandidate++ {
			a, b := p, qCandidate
			for b != 0 {
				a, b = b, a%b
			}
			if a == 1 {
				qVal = qCandidate
				break
			}
		}
		q := fmt.Sprintf("The 3-dimensional lens space \\(L(%d, %d)\\) has fundamental group isomorphic to the cyclic group \\(\\mathbb{Z}/%d\\mathbb{Z}\\). What is the order of \\(\\pi_1(L(%d, %d))\\)?",
			p, qVal, p, p, qVal)
		exp := fmt.Sprintf("By definition, the lens space \\(L(%d, %d)\\) is the quotient of \\(S^3\\) by a free action of \\(\\mathbb{Z}/%d\\mathbb{Z}\\). Since \\(S^3\\) is simply connected, \\(\\pi_1(L(%d, %d)) \\cong \\mathbb{Z}/%d\\mathbb{Z}\\), which has order %d.",
			p, qVal, p, p, qVal, p, p)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(p),
			Explanation: exp,
			Facts: map[string]string{
				"kind":   "lens_space",
				"p":      strconv.Itoa(p),
				"q":      strconv.Itoa(qVal),
				"answer": strconv.Itoa(p),
			},
		}

	case 4:
		// Surface of genus g, rank of abelianized fundamental group is 2g.
		gVal := rand.Intn(5) + 1
		rank := 2 * gVal
		q := fmt.Sprintf("For a closed connected orientable surface \\(\\Sigma_{%d}\\) of genus %d, what is the rank of the abelianized fundamental group \\(\\pi_1(\\Sigma_{%d})^{ab} \\cong H_1(\\Sigma_{%d}, \\mathbb{Z})\\)?",
			gVal, gVal, gVal, gVal)
		exp := fmt.Sprintf("The fundamental group of a surface of genus %d has %d generator pairs \\(a_i, b_i\\) subject to a single relation \\(\\prod [a_i, b_i] = 1\\). In the abelianization, all commutators become trivial, leaving \\(\\mathbb{Z}^{2g} = \\mathbb{Z}^{%d}\\) of rank %d.",
			gVal, gVal, rank, rank)
		return generator.Problem{
			Question:    q,
			Answer:      strconv.Itoa(rank),
			Explanation: exp,
			Facts: map[string]string{
				"kind":   "surface_genus",
				"genus":  strconv.Itoa(gVal),
				"answer": strconv.Itoa(rank),
			},
		}

	default:
		// Sphere S^n for n >= 2, simply connected, order of pi_1 is 1.
		n := rand.Intn(5) + 2
		q := fmt.Sprintf("For the %d-dimensional sphere \\(S^{%d}\\), what is the order of the fundamental group \\(\\pi_1(S^{%d})\\)?",
			n, n, n)
		exp := fmt.Sprintf("Every sphere \\(S^n\\) with \\(n \\ge 2\\) is simply connected: every loop can be homotoped away from a point and contracted to the basepoint. The fundamental group is trivial (order 1).")
		return generator.Problem{
			Question:    q,
			Answer:      "1",
			Explanation: exp,
			Facts: map[string]string{
				"kind":   "sphere",
				"n":      strconv.Itoa(n),
				"answer": "1",
			},
		}
	}
}

// ----- 20. orientability -----

type orientabilityGen struct{}

func (g *orientabilityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct {
		question string
		answer   string
		reason   string
	}
	easy := []entry{
		{"Which band is non-orientable: cylinder or Mobius? (type one word)", "Mobius", "The Mobius band has one side: a loop around it flips orientation."},
		{"Is S^2 orientable? (yes/no)", "yes", "Sphere is orientable (outward normal field)."},
		{"Is Mobius band orientable? (yes/no)", "no", "Möbius has one side, non-orientable."},
		{"Is torus T^2 orientable? (yes/no)", "yes", "Torus is orientable."},
	}
	hard := []entry{
		{"Is Klein bottle non-orientable? (yes/no)", "yes", "Contains Möbius band."},
		{"Does orientability mean consistent choice of local orientation? (yes/no)", "yes", "Atlas with positive Jacobian transitions."},
		{"Is R^n orientable? (yes/no)", "yes", "Euclidean space orientable."},
	}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.question, Answer: e.answer, Explanation: e.reason}
}

type embeddingGen struct{}

func (g *embeddingGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct{ q, a, e string }
	easy := []entry{
		{"What is the Whitney target for a closed n-manifold? (type like R^2n)", "R^2n", "Whitney embeds closed n-manifolds in R^2n."},
		{"Is an embedding an injective immersion homeomorphic onto its image? (yes/no)", "yes", "Embedding definition."},
		{"Is the Alexander horned sphere tame? (yes/no)", "no", "It is the classic wild embedding of S^2 in R^3."},
		{"Does Whitney embedding embed n-manifolds in R^2n? (yes/no)", "yes", "Whitney theorem."},
	}
	hard := []entry{{"Does Nash embed Riemannian manifold isometrically in R^n? (yes/no)", "yes", "Nash embedding."}, {"Is Alexander horned sphere embedding wild? (yes/no)", "yes", "Wild embedding."}, {"Does Schoenflies generalize Jordan curve? (yes/no)", "yes", "Sphere."}}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.q, Answer: e.a, Explanation: e.e}
}

type knotGen struct{}

func (g *knotGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct{ q, a, e string }
	easy := []entry{
		{"What is the crossing number of the trefoil?", "3", "Three crossings: the simplest nontrivial knot."},
		{"Is the unknot trivial? (yes/no)", "yes", "Unknotted circle bounds a disc."},
		{"Is the figure-eight knot the same as the trefoil? (yes/no)", "no", "Figure-eight is 4_1 with four crossings; trefoil is 3_1."},
		{"Does the trefoil have crossing number 3? (yes/no)", "yes", "Simplest nontrivial knot."},
	}
	hard := []entry{{"Is Jones polynomial knot invariant? (yes/no)", "yes", "Jones."}, {"Does Gordon-Luecke show knot determined by complement? (yes/no)", "yes", "Knot complement."}, {"Is figure-eight knot 4_1? (yes/no)", "yes", "Four crossings."}}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.q, Answer: e.a, Explanation: e.e}
}

type cohomologyGen struct{}

func (g *cohomologyGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct{ q, a, e string }
	easy := []entry{
		{"What is H^0 of a connected space? (type like Z)", "Z", "H^0 counts components: one copy of Z per component."},
		{"Is cohomology dual to homology? (yes/no)", "yes", "Universal coefficients pair them."},
		{"Does de Rham cohomology use differential forms? (yes/no)", "yes", "Smooth manifolds: closed forms modulo exact."},
		{"Is H^0 the group counting connected components? (yes/no)", "yes", "H^0 records components."},
	}
	hard := []entry{{"Does cohomology ring have cup product? (yes/no)", "yes", "Ring structure."}, {"Is singular cohomology homotopy invariant? (yes/no)", "yes", "Homotopy."}, {"Does Poincaré duality hold for compact orientable manifold? (yes/no)", "yes", "Duality."}}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.q, Answer: e.a, Explanation: e.e}
}

type stoneCechGen struct{}

func (g *stoneCechGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*4)
	type entry struct{ q, a, e string }
	easy := []entry{
		{"Which compactification is largest: one-point or Stone-Cech? (type with hyphen)", "Stone-Cech", "Stone-Cech is the universal (largest) compactification of Tychonoff spaces."},
		{"Does beta N contain a remainder N*? (yes/no)", "yes", "The growth beta N minus N is the remainder."},
		{"Is beta R metrizable? (yes/no)", "no", "Stone-Cech remainders are huge and non-metrizable."},
		{"Is Tychonoff exactly embeddable in compact Hausdorff? (yes/no)", "yes", "Characterization of Tychonoff spaces."},
	}
	hard := []entry{{"Is βR not metrizable? (yes/no)", "yes", "Huge."}, {"Does Stone-Čech of discrete is ultrafilters? (yes/no)", "yes", "βD."}, {"Is completely regular iff subspace of compact Hausdorff? (yes/no)", "yes", "Tychonoff."}}
	pool := easy
	if scale > 3 {
		pool = append(easy, hard...)
	}
	e := pool[rand.Intn(len(pool))]
	return generator.Problem{Question: e.q, Answer: e.a, Explanation: e.e}
}

// dimensionGen asks for the Lebesgue covering dimension of a space whose
// dimension is a definite integer, and publishes the space, the dimension and
// the rule that fixes it as facts so the authored schema in
// data/lessons/solutions can explain any instance.
//
// It replaced a table of yes/no recognition rows. Covering dimension is a
// number, so the learner should produce the number rather than pick "yes" from
// a list, which is the recognition-without-production shape the pool ratchet
// exists to catch.
type dimensionGen struct{}

func (g *dimensionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	// The ambient dimension offered grows with difficulty: 1 at the easy end,
	// up to 5 once the learner is answering well.
	n := rand.Intn(int(1+ctx.Difficulty*4)) + 1

	type family struct {
		space string // LaTeX body, without the surrounding delimiters
		dim   int
		rule  string
	}
	families := []family{
		{fmt.Sprintf(`\mathbb{R}^{%d}`, n), n,
			fmt.Sprintf(`Euclidean n-space has covering dimension n, and here n is %d`, n)},
		{fmt.Sprintf(`S^{%d}`, n), n,
			fmt.Sprintf(`an n-sphere is an n-dimensional manifold, and here n is %d`, n)},
		{fmt.Sprintf(`T^{%d}`, n), n,
			fmt.Sprintf(`an n-torus has n independent circular directions, and here n is %d`, n)},
		{`\text{the Cantor set}`, 0,
			`the Cantor set is totally disconnected, and a totally disconnected compact metric space has covering dimension 0`},
	}
	f := families[rand.Intn(len(families))]

	return generator.Problem{
		Question:    fmt.Sprintf(`What is the Lebesgue covering dimension of \(%s\)? (enter a number)`, f.space),
		Answer:      strconv.Itoa(f.dim),
		Explanation: fmt.Sprintf(`The covering dimension of \(%s\) is %d: %s.`, f.space, f.dim, f.rule),
		Facts: map[string]string{
			"space": f.space,
			"dim":   strconv.Itoa(f.dim),
			"rule":  f.rule,
		},
	}
}

type coveringGen struct{}

func (g *coveringGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)
	switch mode {
	case 0:
		// Composed covering map degree: p(z) = z^a, q(z) = z^b => (p o q)(z) = z^(a*b)
		a := rand.Intn(3) + 2 // 2 to 4
		b := rand.Intn(3) + 2 // 2 to 4
		deg := a * b
		ansStr := strconv.Itoa(deg)
		q := fmt.Sprintf("Let \\(p: S^1 \\to S^1\\) be the %d-sheeted covering \\(z \\mapsto z^{%d}\\), and let \\(q: S^1 \\to S^1\\) be the %d-sheeted covering \\(z \\mapsto z^{%d}\\). How many sheets does the composite covering map \\(p \\circ q\\) have?", a, a, b, b)
		ex := fmt.Sprintf("The composite covering map is given by \\((p \\circ q)(z) = p(z^{%d}) = (z^{%d})^{%d} = z^{%d}\\), which has degree %d and therefore %d sheets.", b, b, a, deg, deg, deg)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"a":      strconv.Itoa(a),
				"b":      strconv.Itoa(b),
				"answer": ansStr,
			},
		}
	case 1:
		// Genus of covering surface: chi(tilde{X}) = d * chi(X) = d * (2 - 2g) = 2 - 2*tilde{g}
		// If g=2, chi(X) = -2. For d=3: chi(tilde{X}) = -6 => 2 - 2*tilde{g} = -6 => tilde{g} = 4.
		// For d=2: chi = -4 => 2 - 2*tilde{g} = -4 => tilde{g} = 3.
		// For d=4: chi = -8 => 2 - 2*tilde{g} = -8 => tilde{g} = 5.
		type covSurface struct {
			g, d, tildeG int
		}
		cases := []covSurface{
			{2, 2, 3},
			{2, 3, 4},
			{2, 4, 5},
			{3, 2, 5},
		}
		c := cases[rand.Intn(len(cases))]
		ansStr := strconv.Itoa(c.tildeG)
		q := fmt.Sprintf("Let \\(X\\) be a closed orientable surface of genus %d. If \\(\\tilde{X} \\to X\\) is a connected %d-fold covering space, what is the genus of \\(\\tilde{X}\\)?", c.g, c.d)
		chiX := 2 - 2*c.g
		chiCover := c.d * chiX
		ex := fmt.Sprintf("The Euler characteristic of \\(X\\) is \\(\\chi(X) = 2 - 2(%d) = %d\\). The covering space has Euler characteristic \\(\\chi(\\tilde{X}) = %d \\times (%d) = %d\\). Solving \\(2 - 2\\tilde{g} = %d\\) yields genus \\(\\tilde{g} = %d\\).", c.g, chiX, c.d, chiX, chiCover, chiCover, c.tildeG)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"g":      strconv.Itoa(c.g),
				"d":      strconv.Itoa(c.d),
				"answer": ansStr,
			},
		}
	case 2:
		// Euler characteristic of covering space chi(tilde{X}) = d * chi(X)
		// We use strictly negative chi(X) <= -1 with d >= 2, ensuring chi(tilde{X}) < chi(X)
		type covChiCase struct {
			d, chiX, ans int
		}
		cases := []covChiCase{
			{2, -2, -4},
			{3, -2, -6},
			{4, -2, -8},
			{2, -3, -6},
			{3, -3, -9},
			{2, -1, -2},
			{3, -1, -3},
		}
		c := cases[rand.Intn(len(cases))]
		ansStr := strconv.Itoa(c.ans)
		q := fmt.Sprintf("If a compact surface \\(X\\) has Euler characteristic \\(\\chi(X) = %d\\) and \\(\\tilde{X} \\to X\\) is a %d-sheeted covering space, what is \\(\\chi(\\tilde{X})\\)?", c.chiX, c.d)
		ex := fmt.Sprintf("The Euler characteristic of an \\(n\\)-sheeted covering satisfies \\(\\chi(\\tilde{X}) = n \\cdot \\chi(X) = %d \\times (%d) = %d\\).", c.d, c.chiX, c.ans)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"d":      strconv.Itoa(c.d),
				"chiX":   strconv.Itoa(c.chiX),
				"answer": ansStr,
			},
		}
	default:
		// Universal cover fundamental group
		correct := `\{0\}`
		dist1 := `\(\mathbb{Z}\)`
		dist2 := `\(\mathbb{Z} \times \mathbb{Z}\)`
		dist3 := `\(\mathbb{Z}/2\mathbb{Z}\)`
		choices := []string{correct, dist1, dist2, dist3}
		rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
		q := fmt.Sprintf("By definition, what is the fundamental group \\(\\pi_1(\\tilde{X})\\) of the universal covering space of a connected manifold \\(X\\)?\n\nA) %s\n\nB) %s\n\nC) %s\n\nD) %s",
			choices[0], choices[1], choices[2], choices[3])
		ex := "The universal covering space is simply connected by definition, which means its fundamental group is trivial: \\(\\pi_1(\\tilde{X}) = \\{0\\}\\)."
		return generator.Problem{
			Question:    q,
			Answer:      correct,
			Explanation: ex,
			Facts: map[string]string{
				"answer": correct,
			},
		}
	}
}

type vanKampenGen struct{}

func (g *vanKampenGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)
	switch mode {
	case 0:
		// Standard symmetric presentation of genus g surface fundamental group has 2g generators
		gVal := rand.Intn(3) + 2 // 2 to 4
		ans := 2 * gVal
		ansStr := strconv.Itoa(ans)
		q := fmt.Sprintf("By the Seifert–van Kampen theorem, the fundamental group of a closed orientable surface of genus %d has standard presentation \\(\\langle a_1, b_1, \\dots, a_g, b_g \\mid \\prod_{i=1}^g [a_i, b_i] = 1\\rangle\\). How many generators appear in this presentation?", gVal)
		ex := fmt.Sprintf("Each of the %d handles contributes a pair of loops \\(a_i\\) and \\(b_i\\), giving exactly \\(2g = 2(%d) = %d\\) generators.", gVal, gVal, ans)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"genus":  strconv.Itoa(gVal),
				"answer": ansStr,
			},
		}
	case 1:
		// Fundamental group of torus T^2
		correct := `\(\mathbb{Z} \times \mathbb{Z}\)`
		dist1 := `\(\mathbb{Z} * \mathbb{Z}\)`
		dist2 := `\(\mathbb{Z}\)`
		dist3 := `\(\{0\}\)`
		choices := []string{correct, dist1, dist2, dist3}
		rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
		q := fmt.Sprintf("By the Seifert–van Kampen theorem applied to the standard planar square cell with boundary \\(a b a^{-1} b^{-1}\\), what is the fundamental group \\(\\pi_1(T^2)\\) of the 2-torus?\n\nA) %s\n\nB) %s\n\nC) %s\n\nD) %s",
			choices[0], choices[1], choices[2], choices[3])
		ex := "The fundamental group of the 1-skeleton is \\(F_2 = \\langle a, b \\rangle\\), and attaching the 2-cell imposes the single relation \\(a b a^{-1} b^{-1} = 1\\), giving the abelian presentation \\(\\langle a, b \\mid [a, b] = 1\\rangle \\cong \\mathbb{Z} \\times \\mathbb{Z}\\)."
		return generator.Problem{
			Question:    q,
			Answer:      correct,
			Explanation: ex,
			Facts: map[string]string{
				"answer": correct,
			},
		}
	case 2:
		// Klein bottle abelianization
		correct := `\(\mathbb{Z} \times \mathbb{Z}/2\mathbb{Z}\)`
		dist1 := `\(\mathbb{Z} \times \mathbb{Z}\)`
		dist2 := `\(\{0\}\)`
		dist3 := `\(\mathbb{Z} * \mathbb{Z}\)`
		choices := []string{correct, dist1, dist2, dist3}
		rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
		q := fmt.Sprintf("By the Seifert–van Kampen theorem, the Klein bottle has fundamental group presentation \\(\\langle a, b \\mid a b a^{-1} b = 1\\rangle\\). What is the first homology group \\(H_1(K; \\mathbb{Z})\\) obtained by abelianizing this group?\n\nA) %s\n\nB) %s\n\nC) %s\n\nD) %s",
			choices[0], choices[1], choices[2], choices[3])
		ex := "Abelianizing the relation gives \\(a + b - a + b = 0\\), which simplifies to \\(2b = 0\\). Since \\(a\\) is unconstrained and \\(b\\) has order 2, the abelianization is \\(\\mathbb{Z} \\times \\mathbb{Z}/2\\mathbb{Z}\\)."
		return generator.Problem{
			Question:    q,
			Answer:      correct,
			Explanation: ex,
			Facts: map[string]string{
				"answer": correct,
			},
		}
	default:
		// Rank of H_1 of wedge of k 2-tori: each torus has H_1 = Z^2, so wedge has rank 2k
		k := rand.Intn(4) + 2 // 2 to 5
		ans := 2 * k
		ansStr := strconv.Itoa(ans)
		q := fmt.Sprintf("What is the rank of the first homology group \\(H_1\\left(\\bigvee_{i=1}^{%d} T^2; \\mathbb{Z}\\right)\\) of a wedge sum of %d 2-dimensional tori?", k, k)
		ex := fmt.Sprintf("For a wedge sum, \\(H_1\\left(\\bigvee_{i=1}^{%d} T^2\\right) \\cong \\bigoplus_{i=1}^{%d} H_1(T^2) \\cong \\bigoplus_{i=1}^{%d} \\mathbb{Z}^2 \\cong \\mathbb{Z}^{%d}\\), which has rank %d.", k, k, k, ans, ans)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"k":      strconv.Itoa(k),
				"answer": ansStr,
			},
		}
	}
}

type fiberBundleGen struct{}

func (g *fiberBundleGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)
	switch mode {
	case 0:
		// Hopf fibrations fiber dimension
		type hopfCase struct {
			total, base, fiber int
		}
		cases := []hopfCase{
			{3, 2, 1},
			{7, 4, 3},
			{15, 8, 7},
		}
		c := cases[rand.Intn(len(cases))]
		ansStr := strconv.Itoa(c.fiber)
		q := fmt.Sprintf("In the Hopf fibration \\(S^{%d} \\to S^{%d}\\), what is the topological dimension of the fiber sphere?", c.total, c.base)
		ex := fmt.Sprintf("The Hopf fibration has total space \\(S^{%d}\\) and base \\(S^{%d}\\). The fiber is \\(S^{%d}\\), which has dimension %d.", c.total, c.base, c.fiber, c.fiber)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"total":  strconv.Itoa(c.total),
				"base":   strconv.Itoa(c.base),
				"fiber":  ansStr,
				"answer": ansStr,
			},
		}
	case 1:
		// Total space dimension of tangent bundle TM for n-dimensional manifold: dim(TM) = 2n
		n := rand.Intn(4) + 3 // 3 to 6
		ans := 2 * n
		ansStr := strconv.Itoa(ans)
		q := fmt.Sprintf("What is the manifold dimension of the total space of the tangent bundle \\(T M\\) of an %d-dimensional smooth manifold \\(M\\)?", n)
		ex := fmt.Sprintf("The tangent bundle \\(T M\\) has base dimension \\(\\dim(M) = %d\\) and fiber dimension \\(\\dim(T_p M) = %d\\), so the total space has dimension \\(%d + %d = %d\\).", n, n, n, n, ans)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"n":      strconv.Itoa(n),
				"answer": ansStr,
			},
		}
	case 2:
		// Euler characteristic product for surface bundles
		// chi(B) = 2 - 2*gB, chi(F) = 2 - 2*gF
		type bundleCase struct {
			gB, gF, chiE int
		}
		cases := []bundleCase{
			{2, 2, 4},
			{2, 3, 8},
			{3, 2, 8},
			{3, 3, 16},
		}
		c := cases[rand.Intn(len(cases))]
		ansStr := strconv.Itoa(c.chiE)
		q := fmt.Sprintf("For a fiber bundle \\(F \\hookrightarrow E \\to B\\) where base \\(B\\) is a closed surface of genus %d and fiber \\(F\\) is a closed surface of genus %d, what is the Euler characteristic \\(\\chi(E)\\)?", c.gB, c.gF)
		chiB := 2 - 2*c.gB
		chiF := 2 - 2*c.gF
		ex := fmt.Sprintf("The Euler characteristics are \\(\\chi(B) = 2 - 2(%d) = %d\\) and \\(\\chi(F) = 2 - 2(%d) = %d\\). By multiplicativity, \\(\\chi(E) = \\chi(B) \\cdot \\chi(F) = (%d) \\times (%d) = %d\\).", c.gB, chiB, c.gF, chiF, chiB, chiF, c.chiE)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"gB":     strconv.Itoa(c.gB),
				"gF":     strconv.Itoa(c.gF),
				"answer": ansStr,
			},
		}
	default:
		// Vector bundle fiber type
		correct := `\(\mathbb{R}^n\)`
		dist1 := `\(S^n\)`
		dist2 := `\(\mathbb{R}^{n+1}\)`
		dist3 := `\(S^{n-1}\)`
		choices := []string{correct, dist1, dist2, dist3}
		rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
		q := fmt.Sprintf("What is the typical fiber of the tangent bundle \\(T M\\) of an \\(n\\)-dimensional smooth manifold \\(M\\)?\n\nA) %s\n\nB) %s\n\nC) %s\n\nD) %s",
			choices[0], choices[1], choices[2], choices[3])
		ex := "The tangent space \\(T_p M\\) at any point \\(p \\in M\\) is an \\(n\\)-dimensional real vector space, isomorphic to \\(\\mathbb{R}^n\\)."
		return generator.Problem{
			Question:    q,
			Answer:      correct,
			Explanation: ex,
			Facts: map[string]string{
				"answer": correct,
			},
		}
	}
}

type cellularHomologyGen struct{}

func (g *cellularHomologyGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	for attempt := 0; attempt < 100; attempt++ {
		mode := rand.Intn(4)
		switch mode {
		case 0:
			// Closed orientable surface of genus g: b_1 = 2g
			gVal := rand.Intn(6) + 2 // 2 to 7
			b1 := 2 * gVal
			ansStr := strconv.Itoa(b1)
			q := fmt.Sprintf("What is the first Betti number \\(b_1 = \\operatorname{rank} H_1(\\Sigma_{%d}; \\mathbb{Z})\\) of a closed orientable surface of genus %d?", gVal, gVal)
			if strings.Contains(q, ansStr) {
				continue
			}
			ex := fmt.Sprintf("A closed orientable surface of genus \\(g\\) has CW decomposition with 1 0-cell, \\(2g\\) 1-cells and 1 2-cell. The 2-cell boundary word is a product of commutators \\(\\prod [a_i, b_i]\\), which abelianizes to 0 in \\(H_1\\). Thus \\(H_1(\\Sigma_g) \\cong \\mathbb{Z}^{2g}\\), and \\(b_1 = 2 \\times %d = %d\\).", gVal, b1)
			return generator.Problem{
				Question:    q,
				Answer:      ansStr,
				Explanation: ex,
				Facts: map[string]string{
					"genus":  strconv.Itoa(gVal),
					"answer": ansStr,
				},
			}
		case 1:
			// Punctured surface Sigma_{g, p}: b_1 = 2g + p - 1
			gVal := rand.Intn(4) + 1 // 1 to 4
			pVal := rand.Intn(3) + 1 // 1 to 3
			b1 := 2*gVal + pVal - 1
			ansStr := strconv.Itoa(b1)
			q := fmt.Sprintf("Consider an orientable surface of genus %d with %d puncture(s). What is its first Betti number \\(b_1 = \\operatorname{rank} H_1(S; \\mathbb{Z})\\)?", gVal, pVal)
			if strings.Contains(q, ansStr) {
				continue
			}
			ex := fmt.Sprintf("An orientable surface of genus \\(g\\) with \\(p \\ge 1\\) punctures deformation retracts onto a 1-dimensional CW complex (a wedge of circles). Its Euler characteristic is \\(\\chi = 2 - 2g - p\\). Since \\(H_2(S) = 0\\) and \\(H_0(S) \\cong \\mathbb{Z}\\), we have \\(1 - b_1 = 2 - 2g - p\\), giving \\(b_1 = 2g + p - 1 = 2(%d) + %d - 1 = %d\\).", gVal, pVal, b1)
			return generator.Problem{
				Question:    q,
				Answer:      ansStr,
				Explanation: ex,
				Facts: map[string]string{
					"genus":     strconv.Itoa(gVal),
					"punctures": strconv.Itoa(pVal),
					"answer":    ansStr,
				},
			}
		case 2:
			// Connected sum of k real projective planes: b_1 = k - 1
			kVal := rand.Intn(5) + 2 // 2 to 6
			b1 := kVal - 1
			ansStr := strconv.Itoa(b1)
			q := fmt.Sprintf("What is the free rank (first Betti number \\(b_1\\)) of the integral homology group \\(H_1(N_{%d}; \\mathbb{Z})\\) for the connected sum of %d real projective planes?", kVal, kVal)
			if strings.Contains(q, ansStr) {
				continue
			}
			ex := fmt.Sprintf("The first homology of the non-orientable surface \\(N_k = k \\mathbb{RP}^2\\) is \\(H_1(N_k; \\mathbb{Z}) \\cong \\mathbb{Z}^{k-1} \\oplus \\mathbb{Z}_2\\). The torsion-free rank is the Betti number \\(b_1 = k - 1 = %d - 1 = %d\\).", kVal, b1)
			return generator.Problem{
				Question:    q,
				Answer:      ansStr,
				Explanation: ex,
				Facts: map[string]string{
					"k":      strconv.Itoa(kVal),
					"answer": ansStr,
				},
			}
		default:
			// Explicit cellular boundary matrix reduction via Smith Normal Form over Z
			c0 := 2
			c1 := 4
			c2 := 2
			d1 := mathutil.NewIntMatrix(c0, c1)
			d2 := mathutil.NewIntMatrix(c1, c2)
			// d1: C1 -> C0 has rank 1 (maps e1, e2, e3, e4 to v2 - v1)
			d1.Data[0][0] = -1; d1.Data[1][0] = 1
			d1.Data[0][1] = -1; d1.Data[1][1] = 1
			// d2: C2 -> C1 maps 2-cells into ker(d1)
			d2.Data[0][0] = 1; d2.Data[1][0] = -1
			d2.Data[2][1] = 1; d2.Data[3][1] = -1
			_, b1, _ := mathutil.CellularBettiNumbers(c0, c1, c2, d1, d2)
			ansStr := strconv.Itoa(b1)
			q := fmt.Sprintf("A CW complex has %d 0-cells, %d 1-cells and %d 2-cells. Smith normal form reduction on the boundary matrices yields \\(\\operatorname{rank}(d_1) = 1\\) and \\(\\operatorname{rank}(d_2) = 2\\). What is the first Betti number \\(b_1 = \\operatorname{rank} H_1(X; \\mathbb{Z})\\)?", c0, c1, c2)
			if strings.Contains(q, ansStr) {
				continue
			}
			ex := fmt.Sprintf("Smith normal form over \\(\\mathbb{Z}\\) computes \\(\\dim \\ker(d_1) = c_1 - \\operatorname{rank}(d_1) = %d - 1 = 3\\). Subtracting \\(\\operatorname{rank}(d_2) = 2\\) yields the first Betti number \\(b_1 = 3 - 2 = %d\\).",
				c1, b1)
			return generator.Problem{
				Question:    q,
				Answer:      ansStr,
				Explanation: ex,
				Facts: map[string]string{
					"c0":     strconv.Itoa(c0),
					"c1":     strconv.Itoa(c1),
					"c2":     strconv.Itoa(c2),
					"rank1":  "1",
					"rank2":  "2",
					"answer": ansStr,
				},
			}
		}
	}
	return generator.Problem{
		Question:    "What is the first Betti number \\(b_1\\) of a closed orientable surface of genus 3?",
		Answer:      "6",
		Explanation: "The first Betti number of a closed surface of genus g is 2g = 6.",
	}
}

type deRhamGen struct{}

func (g *deRhamGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(4)
	switch mode {
	case 0:
		// Poincare lemma: H^k_dR(R^n) = 0 for k >= 1
		k := rand.Intn(3) + 2 // 2 to 4
		n := k + rand.Intn(3) + 1
		ansStr := "0"
		q := fmt.Sprintf("By the Poincaré lemma, what is the dimension of the de Rham cohomology group \\(H^{%d}_{\\mathrm{dR}}(\\mathbb{R}^{%d})\\)?", k, n)
		ex := fmt.Sprintf("Since \\(\\mathbb{R}^{%d}\\) is contractible, the Poincaré lemma implies that every closed %d-form on \\(\\mathbb{R}^{%d}\\) is exact for %d \\ge 1, so \\(H^{%d}_{\\mathrm{dR}}(\\mathbb{R}^{%d}) = 0\\).", n, k, n, k, k, n)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"k":      strconv.Itoa(k),
				"n":      strconv.Itoa(n),
				"answer": ansStr,
			},
		}
	case 1:
		// Second Betti number b_2 of 4-manifold with b_1 = 0
		// chi = 1 - 0 + b_2 - 0 + 1 = b_2 + 2 => b_2 = chi - 2
		type b2Case struct {
			chi, b2 int
		}
		cases := []b2Case{
			{7, 5},
			{8, 6},
			{10, 8},
		}
		c := cases[rand.Intn(len(cases))]
		ansStr := strconv.Itoa(c.b2)
		q := fmt.Sprintf("Let \\(M\\) be a connected, compact, orientable 4-dimensional smooth manifold with Euler characteristic \\(\\chi(M) = %d\\) and \\(\\dim H^1_{\\mathrm{dR}}(M) = 0\\). What is the second Betti number \\(b_2 = \\dim H^2_{\\mathrm{dR}}(M)\\)?", c.chi)
		ex := fmt.Sprintf("By Poincaré duality, \\(b_0 = b_4 = 1\\) and \\(b_3 = b_1 = 0\\). The Euler characteristic is \\(\\chi(M) = b_0 - b_1 + b_2 - b_3 + b_4 = 1 - 0 + b_2 - 0 + 1 = b_2 + 2\\). Therefore \\(b_2 = %d - 2 = %d\\).", c.chi, c.b2)
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"chi":    strconv.Itoa(c.chi),
				"answer": ansStr,
			},
		}
	case 2:
		// Sum of Betti numbers of 2-torus: b0=1, b1=2, b2=1 => sum = 4
		ansStr := "4"
		q := "What is the total sum of all de Rham cohomology dimensions \\(\\sum_{k=0}^2 \\dim H^k_{\\mathrm{dR}}(T^2)\\) for the 2-dimensional torus \\(T^2\\)?"
		ex := "The Betti numbers of the 2-torus are \\(b_0 = 1\\), \\(b_1 = 2\\) and \\(b_2 = 1\\). Their sum is \\(1 + 2 + 1 = 4\\)."
		return generator.Problem{
			Question:    q,
			Answer:      ansStr,
			Explanation: ex,
			Facts: map[string]string{
				"answer": ansStr,
			},
		}
	default:
		// Top de Rham cohomology of sphere S^n
		type sphereCase struct {
			dim int
			ans string
		}
		cases := []sphereCase{
			{2, "1"},
			{3, "1"},
			{4, "1"},
		}
		c := cases[rand.Intn(len(cases))]
		q := fmt.Sprintf("What is the dimension of the top de Rham cohomology group \\(H^{%d}_{\\mathrm{dR}}(S^{%d})\\) of the smooth %d-sphere?", c.dim, c.dim, c.dim)
		ex := fmt.Sprintf("The %d-sphere is a connected, compact, orientable %d-dimensional manifold, so by Poincaré duality \\(H^{%d}_{\\mathrm{dR}}(S^{%d}) \\cong \\mathbb{R}\\), which has dimension 1.", c.dim, c.dim, c.dim, c.dim)
		return generator.Problem{
			Question:    q,
			Answer:      c.ans,
			Explanation: ex,
			Facts: map[string]string{
				"dim":    strconv.Itoa(c.dim),
				"answer": c.ans,
			},
		}
	}
}
