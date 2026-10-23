package linalg

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

var matrixRe = regexp.MustCompile(`\[\[([^\]]+)\],\[([^\]]+)\]\]`)

func Register(reg *generator.Registry) {
	reg.Register("linalg.vector.concept", &vectorConceptGen{})
	reg.Register("linalg.vector.add", &vectorAddGen{})
	reg.Register("linalg.vector.dot", &vectorDotGen{})
	reg.Register("linalg.matrix.concept", &matrixConceptGen{})
	reg.Register("linalg.matrix.add", &matrixAddGen{})
	reg.Register("linalg.matrix.mult", &matrixMultGen{})
	reg.Register("linalg.matrix.identity", &matrixIdentityGen{})
	reg.Register("linalg.det.2x2", &det2x2Gen{})
	reg.Register("linalg.det.3x3", &det3x3Gen{})
	reg.Register("linalg.systems.matrix", &systemsMatrixGen{})
	reg.Register("linalg.sys.cramer", &cramerGen{})
	reg.Register("linalg.eigen.concept", &eigenConceptGen{})
	reg.Register("linalg.eigen.compute", &eigenComputeGen{})
	reg.Register("linalg.lintrans.concept", &transformationsGen{})
	reg.Register("linalg.vec.span", &spanGen{})
	reg.Register("linalg.vec.basis", &basisGen{})

	reg.Register("linalg.eigen.diagonalization", &diagonalizationGen{})
	reg.Register("linalg.matrix.rank", &rankGen{})
	reg.Register("linalg.vector.cosine_similarity", &cosineSimilarityGen{})
	reg.Register("linalg.vector.parametric", &parametricGen{})
	reg.Register("linalg.inner_product", &innerProductGen{})
	reg.Register("linalg.decomp.lu", &luGen{})
	reg.Register("linalg.decomp.qr", &qrGen{})
	reg.Register("linalg.decomp.svd", &svdGen{})
	reg.Register("linalg.decomp.jordan", &jordanGen{})
	reg.Register("linalg.subspace", &subspaceGen{})
	registerFoundations(reg)
	registerBatch(reg)
}

type vectorConceptGen struct{}

func (g *vectorConceptGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	triples := [][3]int{
		{3, 4, 5}, {4, 3, 5},
		{6, 8, 10}, {8, 6, 10},
		{5, 12, 13}, {12, 5, 13},
		{8, 15, 17}, {15, 8, 17},
		{9, 12, 15}, {12, 9, 15},
		{7, 24, 25}, {24, 7, 25},
	}
	if ctx.Difficulty > 0.6 {
		triples = append(triples,
			[][3]int{{20, 21, 29}, {9, 40, 41}, {11, 60, 61}}...)
	}
	t := triples[rand.Intn(len(triples))]
	x, y, mag := t[0], t[1], t[2]
	if rand.Intn(2) == 0 {
		x = -x
	}
	if rand.Intn(2) == 0 {
		y = -y
	}
	return generator.Problem{
		Question:    fmt.Sprintf("What is the magnitude of vector \\((%d,%d)\\)?", x, y),
		Answer:      fmt.Sprintf("%d", mag),
		Explanation: fmt.Sprintf("|(%d,%d)| = sqrt(%d²+%d²) = sqrt(%d+%d) = sqrt(%d) = %d", x, y, x, y, x*x, y*y, x*x+y*y, mag),
	}
}

type vectorAddGen struct{}

func (g *vectorAddGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	a, b := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	c, d := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	if rand.Intn(2) == 0 {
		sumX, sumY := a+c, b+d
		return generator.Problem{
			Question:    fmt.Sprintf("Add vectors: \\((%d,%d) + (%d,%d) =\\) ?", a, b, c, d),
			Answer:      fmt.Sprintf("(%d,%d)", sumX, sumY),
			Explanation: fmt.Sprintf("(%d,%d) + (%d,%d) = (%d,%d)", a, b, c, d, sumX, sumY),
		}
	}
	scalar := rand.Intn(max(1, scale)) + 2
	if rand.Intn(2) == 0 {
		scalar = -scalar
	}
	return generator.Problem{
		Question:    fmt.Sprintf("Compute scalar multiplication: \\(%d \\cdot (%d,%d) =\\) ?", scalar, a, b),
		Answer:      fmt.Sprintf("(%d,%d)", scalar*a, scalar*b),
		Explanation: fmt.Sprintf("%d * (%d,%d) = (%d,%d)", scalar, a, b, scalar*a, scalar*b),
	}
}

type vectorDotGen struct{}

func (g *vectorDotGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	a, b := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	c, d := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	dot := a*c + b*d
	return generator.Problem{
		Question:    fmt.Sprintf("Compute the dot product: \\((%d,%d) \\cdot (%d,%d) =\\) ?", a, b, c, d),
		Answer:      fmt.Sprintf("%d", dot),
		Explanation: fmt.Sprintf("(%d,%d) · (%d,%d) = %d*%d + %d*%d = %d+%d = %d", a, b, c, d, a, c, b, d, a*c, b*d, dot),
	}
}

type matrixConceptGen struct{}

func (g *matrixConceptGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	var rows, cols int
	if rand.Intn(2) == 0 {
		rows, cols = 2, 2
	} else {
		rows, cols = 3, 3
	}
	m := make([][]int, rows)
	for i := range m {
		m[i] = make([]int, cols)
		for j := range m[i] {
			m[i][j] = rand.Intn(max(1, scale*3)) + 1
		}
	}
	r := rand.Intn(rows) + 1
	c := rand.Intn(cols) + 1
	answer := m[r-1][c-1]
	return generator.Problem{
		Question:    fmt.Sprintf("In the matrix \\(%s\\), what is the \\((%d,%d)\\) entry?", formatMatrix(m), r, c),
		Answer:      fmt.Sprintf("%d", answer),
		Explanation: fmt.Sprintf("The entry at row %d, column %d is %d.", r, c, answer),
	}
}

type matrixAddGen struct{}

// A matrix answer is written with brackets and commas. The concept is numeric,
// so the field offered digits only: the syntax existed in the accepted answer
// and nowhere in the interface.
func (g *matrixAddGen) AnswerFormat() generator.AnswerFormat {
	return generator.AnswerFormat{Hint: "Matrix with rows in brackets, e.g. [[1,2],[3,4]]", InputMode: "text"}
}

func (g *matrixAddGen) Grade(expected, userAnswer string) grader.Result {
	return gradeMatrix(expected, userAnswer)
}

func (g *matrixAddGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	a := make([][]int, 2)
	b := make([][]int, 2)
	for i := range a {
		a[i] = make([]int, 2)
		b[i] = make([]int, 2)
		for j := range a[i] {
			a[i][j] = rand.Intn(max(1, scale*2)) - scale
			b[i][j] = rand.Intn(max(1, scale*2)) - scale
		}
	}
	if rand.Intn(2) == 0 {
		sum := make([][]int, 2)
		for i := range sum {
			sum[i] = make([]int, 2)
			for j := range sum[i] {
				sum[i][j] = a[i][j] + b[i][j]
			}
		}
		return generator.Problem{
			Question:    fmt.Sprintf("Add matrices: \\(%s + %s =\\) ?", formatMatrix(a), formatMatrix(b)),
			Answer:      formatMatrix(sum),
			Explanation: fmt.Sprintf("%s + %s = %s (add corresponding entries)", formatMatrix(a), formatMatrix(b), formatMatrix(sum)),
		}
	}
	scalar := rand.Intn(max(1, scale)) + 2
	if rand.Intn(2) == 0 {
		scalar = -scalar
	}
	prod := make([][]int, 2)
	for i := range prod {
		prod[i] = make([]int, 2)
		for j := range prod[i] {
			prod[i][j] = scalar * a[i][j]
		}
	}
	return generator.Problem{
		Question:    fmt.Sprintf("Compute scalar multiplication: \\(%d \\cdot %s =\\) ?", scalar, formatMatrix(a)),
		Answer:      formatMatrix(prod),
		Explanation: fmt.Sprintf("%d * %s = %s", scalar, formatMatrix(a), formatMatrix(prod)),
	}
}

type matrixMultGen struct{}

func (g *matrixMultGen) AnswerFormat() generator.AnswerFormat {
	return generator.AnswerFormat{Hint: "Matrix with rows in brackets, e.g. [[1,2],[3,4]]", InputMode: "text"}
}

func (g *matrixMultGen) Grade(expected, userAnswer string) grader.Result {
	return gradeMatrix(expected, userAnswer)
}

func (g *matrixMultGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	a := make([][]int, 2)
	b := make([][]int, 2)
	for i := range a {
		a[i] = make([]int, 2)
		b[i] = make([]int, 2)
		for j := range a[i] {
			a[i][j] = rand.Intn(max(1, scale*2)) - scale
			b[i][j] = rand.Intn(max(1, scale*2)) - scale
		}
	}
	c := make([][]int, 2)
	for i := range c {
		c[i] = make([]int, 2)
	}
	c[0][0] = a[0][0]*b[0][0] + a[0][1]*b[1][0]
	c[0][1] = a[0][0]*b[0][1] + a[0][1]*b[1][1]
	c[1][0] = a[1][0]*b[0][0] + a[1][1]*b[1][0]
	c[1][1] = a[1][0]*b[0][1] + a[1][1]*b[1][1]
	return generator.Problem{
		Question:    fmt.Sprintf("Multiply matrices: \\(%s \\times %s =\\) ?", formatMatrix(a), formatMatrix(b)),
		Answer:      formatMatrix(c),
		Explanation: fmt.Sprintf("%s × %s = %s", formatMatrix(a), formatMatrix(b), formatMatrix(c)),
	}
}

type matrixIdentityGen struct{}

func (g *matrixIdentityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Identity vector multiplication I_2 v = v
		v1 := mathutil.ScaleInt(ctx.Difficulty, 1, 10)
		v2 := mathutil.ScaleInt(ctx.Difficulty, 1, 10)
		if rand.Intn(2) == 0 {
			v1 = -v1
		}
		if rand.Intn(2) == 0 {
			v2 = -v2
		}
		comp := 1
		ans := v1
		if rand.Intn(2) == 0 {
			comp = 2
			ans = v2
		}
		return generator.Problem{
			Question: fmt.Sprintf("Let \\(I_2\\) be the \\(2 \\times 2\\) identity matrix and \\(v = \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\). Find component %d of the product vector \\(I_2 v\\).", v1, v2, comp),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State the governing property of the identity matrix: \\(I_n v = v\\) for any vector \\(v\\).\n"+
				"Subgoal 2: Substitute vector components: \\(I_2 \\begin{pmatrix} %d \\\\ %d \\end{pmatrix} = \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\).\n"+
				"Subgoal 3: Extract component %d, giving %d.", v1, v2, v1, v2, comp, ans),
		}
	case 1:
		// Determinant of scaled identity det(c * I_n) = c^n
		n := 2
		if ctx.Difficulty > 0.5 {
			n = 3
		}
		c := mathutil.ScaleInt(ctx.Difficulty, 2, 5)
		ans := 1
		for i := 0; i < n; i++ {
			ans *= c
		}
		return generator.Problem{
			Question: fmt.Sprintf("Find the determinant of the scaled \\(%d \\times %d\\) identity matrix \\(%d I_%d\\).", n, n, c, n),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State the determinant scaling identity for identity matrices: \\(\\det(c I_n) = c^n\\).\n"+
				"Subgoal 2: Substitute dimension \\(n = %d\\) and scalar \\(c = %d\\): \\((%d)^{%d}\\).\n"+
				"Subgoal 3: Evaluate the exponent to obtain %d.", n, c, c, n, ans),
		}
	case 2:
		// Trace of identity matrix tr(I_n) = n
		n := mathutil.ScaleInt(ctx.Difficulty, 2, 8)
		return generator.Problem{
			Question: fmt.Sprintf("Compute the trace of the \\(%d \\times %d\\) identity matrix \\(I_%d\\).", n, n, n),
			Answer:   fmt.Sprintf("%d", n),
			Explanation: fmt.Sprintf("Subgoal 1: State the trace definition as the sum of all diagonal elements: \\(\\operatorname{tr}(A) = \\sum_{i=1}^n a_{ii}\\).\n"+
				"Subgoal 2: Substitute the diagonal elements of \\(I_%d\\), each equal to 1: \\(\\sum_{i=1}^{%d} 1\\).\n"+
				"Subgoal 3: Evaluate the sum to find \\(\\operatorname{tr}(I_%d) = %d\\).", n, n, n, n),
		}
	case 3:
		// Unimodular matrix inverse entry
		rng := rand.New(rand.NewSource(rand.Int63()))
		mat := mathutil.RandomUnimodular2x2(rng, 1+int(ctx.Difficulty*2))
		if mat.Det() == -1 {
			mat.A = -mat.A
			mat.B = -mat.B
		}
		inv, _ := mat.Inverse()
		pos := rand.Intn(4)
		row, col, ans := 1, 1, inv.A
		switch pos {
		case 1:
			row, col, ans = 1, 2, inv.B
		case 2:
			row, col, ans = 2, 1, inv.C
		case 3:
			row, col, ans = 2, 2, inv.D
		}
		return generator.Problem{
			Question: fmt.Sprintf("For the invertible matrix \\(A = %s\\) with \\(\\det(A) = 1\\), what is the \\((%d,%d)\\) entry of its inverse \\(A^{-1}\\)?", mat.LaTeX(), row, col),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State the \\(2 \\times 2\\) matrix inverse formula: \\(A^{-1} = \\frac{1}{\\det(A)} \\begin{pmatrix} d & -b \\\\ -c & a \\end{pmatrix}\\).\n"+
				"Subgoal 2: Substitute matrix parameters with \\(\\det(A) = 1\\) to construct \\(A^{-1} = %s\\).\n"+
				"Subgoal 3: Read off entry \\((%d,%d)\\), which equals %d.", inv.LaTeX(), row, col, ans),
		}
	default:
		// Determinant of inverse matrix reciprocal
		d := mathutil.ScaleInt(ctx.Difficulty, 2, 7)
		if rand.Intn(2) == 0 {
			d = -d
		}
		return generator.Problem{
			Question: fmt.Sprintf("If an invertible matrix \\(A\\) has \\(\\det(A) = %d\\), compute the value of \\(%d \\cdot \\det(A^{-1})\\).", d, d),
			Answer:   "1",
			Explanation: fmt.Sprintf("Subgoal 1: State the determinant reciprocal rule for invertible matrices: \\(\\det(A^{-1}) = \\frac{1}{\\det(A)}\\).\n"+
				"Subgoal 2: Substitute \\(\\det(A) = %d\\) into the identity: \\(\\det(A^{-1}) = \\frac{1}{%d}\\).\n"+
				"Subgoal 3: Multiply by %d: \\(%d \\cdot \\frac{1}{%d} = 1\\).", d, d, d, d, d),
		}
	}
}

type det2x2Gen struct{}

func (g *det2x2Gen) Generate(ctx generator.GeneratorContext) generator.Problem {
	span := mathutil.ScaleInt(ctx.Difficulty, 3, 12)
	a := rand.Intn(2*span+1) - span
	b := rand.Intn(2*span+1) - span
	c := rand.Intn(2*span+1) - span
	d := rand.Intn(2*span+1) - span
	det := a*d - b*c
	return generator.Problem{
		Question: fmt.Sprintf("Find the determinant of the matrix \\(%s\\).", formatMatrix2([][]int{{a, b}, {c, d}})),
		Answer:   fmt.Sprintf("%d", det),
		Explanation: fmt.Sprintf("Subgoal 1: State the governing formula for a \\(2 \\times 2\\) determinant: \\(\\det(A) = ad - bc\\).\n"+
			"Subgoal 2: Substitute matrix parameters \\(a = %d, b = %d, c = %d, d = %d\\): \\((%d)(%d) - (%d)(%d)\\).\n"+
			"Subgoal 3: Evaluate products and subtract: %d - (%d) = %d.", a, b, c, d, a, d, b, c, a*d, b*c, det),
	}
}

type det3x3Gen struct{}

func (g *det3x3Gen) Generate(ctx generator.GeneratorContext) generator.Problem {
	span := mathutil.ScaleInt(ctx.Difficulty, 2, 6)
	m := make([][]int, 3)
	for i := range m {
		m[i] = make([]int, 3)
		for j := range m[i] {
			m[i][j] = rand.Intn(2*span+1) - span
		}
	}
	a, b, c := m[0][0], m[0][1], m[0][2]
	d, e, f := m[1][0], m[1][1], m[1][2]
	gVal, h, i := m[2][0], m[2][1], m[2][2]
	m11 := e*i - f*h
	m12 := d*i - f*gVal
	m13 := d*h - e*gVal
	det := a*m11 - b*m12 + c*m13
	return generator.Problem{
		Question: fmt.Sprintf("Find the determinant of the matrix \\(%s\\).", formatMatrix(m)),
		Answer:   fmt.Sprintf("%d", det),
		Explanation: fmt.Sprintf("Subgoal 1: State Laplace cofactor expansion along the first row: \\(\\det(A) = a(ei - fh) - b(di - fg) + c(dh - eg)\\).\n"+
			"Subgoal 2: Substitute row entries and evaluate \\(2 \\times 2\\) minors: \\(M_{11} = %d, M_{12} = %d, M_{13} = %d\\).\n"+
			"Subgoal 3: Combine with alternating signs: \\(%d(%d) - (%d)(%d) + (%d)(%d) = %d\\).", m11, m12, m13, a, m11, b, m12, c, m13, det),
	}
}

type systemsMatrixGen struct{}

func (g *systemsMatrixGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	x := rand.Intn(max(1, scale*2)) - scale
	y := rand.Intn(max(1, scale*2)) - scale
	a, b := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	c, d := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	if a == 0 && c == 0 {
		a = 1
		c = 2
	}
	if b == 0 && d == 0 {
		b = 1
		d = 3
	}
	e1 := a*x + b*y
	e2 := c*x + d*y
	return generator.Problem{
		Question:    fmt.Sprintf("Solve the linear system using matrices: \\(%dx + %dy = %d\\), \\(%dx + %dy = %d\\)", a, b, e1, c, d, e2),
		Answer:      fmt.Sprintf("(%d,%d)", x, y),
		Explanation: fmt.Sprintf("The solution is x=%d, y=%d.", x, y),
	}
}

type cramerGen struct{}

func (g *cramerGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	rng := rand.New(rand.NewSource(rand.Int63()))
	mat := mathutil.RandomUnimodular2x2(rng, 1+int(ctx.Difficulty*3))
	a, b := mat.A, mat.B
	c, d := mat.C, mat.D
	detA := a*d - b*c

	x := mathutil.ScaleInt(ctx.Difficulty, 1, 8)
	y := mathutil.ScaleInt(ctx.Difficulty, 1, 8)
	if rand.Intn(2) == 0 {
		x = -x
	}
	if rand.Intn(2) == 0 {
		y = -y
	}
	e1 := a*x + b*y
	e2 := c*x + d*y

	askX := rand.Intn(2) == 0
	if askX {
		detX := e1*d - b*e2
		valX := detX / detA
		return generator.Problem{
			Question: fmt.Sprintf("Use Cramer's rule to find \\(x\\) for the linear system: \\(%dx + %dy = %d\\), \\(%dx + %dy = %d\\).", a, b, e1, c, d, e2),
			Answer:   fmt.Sprintf("%d", valX),
			Explanation: fmt.Sprintf("Subgoal 1: State Cramer's rule for coordinate \\(x\\): \\(x = \\frac{\\det(A_x)}{\\det(A)}\\).\n"+
				"Subgoal 2: Compute coefficient determinant \\(\\det(A) = (%d)(%d) - (%d)(%d) = %d\\) and modified determinant \\(\\det(A_x) = (%d)(%d) - (%d)(%d) = %d\\).\n"+
				"Subgoal 3: Divide determinants: \\(x = \\frac{%d}{%d} = %d\\).", a, d, b, c, detA, e1, d, b, e2, detX, detX, detA, valX),
		}
	}
	detY := a*e2 - e1*c
	valY := detY / detA
	return generator.Problem{
		Question: fmt.Sprintf("Use Cramer's rule to find \\(y\\) for the linear system: \\(%dx + %dy = %d\\), \\(%dx + %dy = %d\\).", a, b, e1, c, d, e2),
		Answer:   fmt.Sprintf("%d", valY),
		Explanation: fmt.Sprintf("Subgoal 1: State Cramer's rule for coordinate \\(y\\): \\(y = \\frac{\\det(A_y)}{\\det(A)}\\).\n"+
			"Subgoal 2: Compute coefficient determinant \\(\\det(A) = (%d)(%d) - (%d)(%d) = %d\\) and modified determinant \\(\\det(A_y) = (%d)(%d) - (%d)(%d) = %d\\).\n"+
			"Subgoal 3: Divide determinants: \\(y = \\frac{%d}{%d} = %d\\).", a, d, b, c, detA, a, e2, e1, c, detY, detY, detA, valY),
	}
}

type eigenConceptGen struct{}

func (g *eigenConceptGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Av = lambda v
		lambda := smallInt(1, 5+int(ctx.Difficulty*3))
		k := smallInt(1, 3)
		b := smallInt(-2, 3)
		a := lambda - b*k
		d := smallInt(-2, 3)
		c := k * (lambda - d)
		return generator.Problem{
			Question:    fmt.Sprintf("The vector \\(v = \\begin{pmatrix} 1 \\\\ %d \\end{pmatrix}\\) is an eigenvector of matrix \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\). Find the corresponding eigenvalue \\(\\lambda\\).", k, a, b, c, d),
			Answer:      fmt.Sprintf("%d", lambda),
			Explanation: fmt.Sprintf("An eigenvector satisfies \\(Av = \\lambda v\\). Computing the first component of \\(Av\\) gives \\(%d(1) + (%d)(%d) = %d\\). Therefore \\(\\lambda = %d\\).", a, b, k, lambda, lambda),
		}
	case 1:
		// tr(A) = lambda1 + lambda2, with A = P diag(lambda1, lambda2) P^-1 for
		// P = [[1,1],[1,2]]. Building A this way makes the stated lambda1 an
		// actual eigenvalue; arbitrary entries would make the premise false.
		l1 := smallInt(-4, 4)
		l2 := smallInt(-4, 4)
		a := 2*l1 - l2
		b := l2 - l1
		c := 2*l1 - 2*l2
		d := 2*l2 - l1
		tr := a + d
		return generator.Problem{
			Question:    fmt.Sprintf("The matrix \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\) has one eigenvalue \\(\\lambda_1 = %d\\). Using the trace formula \\(\\operatorname{tr}(A) = \\lambda_1 + \\lambda_2\\), compute the second eigenvalue \\(\\lambda_2\\).", a, b, c, d, l1),
			Answer:      fmt.Sprintf("%d", l2),
			Explanation: fmt.Sprintf("The trace of a square matrix is the sum of its eigenvalues: \\(\\operatorname{tr}(A) = %d + (%d) = %d\\). Thus \\(\\lambda_2 = \\operatorname{tr}(A) - \\lambda_1 = %d - (%d) = %d\\).", a, d, tr, tr, l1, l2),
		}
	case 2:
		// det(A) = lambda1 * lambda2
		a := smallInt(-4, 4)
		d := smallInt(-4, 4)
		b := smallInt(-3, 3)
		c := smallInt(-3, 3)
		det := a*d - b*c
		return generator.Problem{
			Question:    fmt.Sprintf("For the matrix \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\), what is the product of its eigenvalues \\(\\lambda_1 \\lambda_2\\)?", a, b, c, d),
			Answer:      fmt.Sprintf("%d", det),
			Explanation: fmt.Sprintf("The product of all eigenvalues of a matrix equals its determinant: \\(\\lambda_1 \\lambda_2 = \\det(A) = (%d)(%d) - (%d)(%d) = %d\\).", a, d, b, c, det),
		}
	case 3:
		// c * lambda
		lambda := smallInt(-5, 5)
		for lambda == 0 {
			lambda = smallInt(-5, 5)
		}
		c := smallInt(2, 5)
		ans := c * lambda
		return generator.Problem{
			Question:    fmt.Sprintf("If matrix \\(A\\) has an eigenvalue \\(\\lambda = %d\\), what is the corresponding eigenvalue of the scaled matrix \\(%d A\\)?", lambda, c),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("For any scalar \\(c\\), the eigenvalues of \\(cA\\) are \\(c\\lambda\\). Here \\(%d \\cdot (%d) = %d\\).", c, lambda, ans),
		}
	default:
		// lambda + c
		lambda := smallInt(-5, 5)
		c := smallInt(-5, 5)
		for c == 0 {
			c = smallInt(-5, 5)
		}
		ans := lambda + c
		return generator.Problem{
			Question:    fmt.Sprintf("If matrix \\(A\\) has an eigenvalue \\(\\lambda = %d\\), what is the corresponding eigenvalue of the shifted matrix \\(A + %d I\\)?", lambda, c),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("If \\(Av = \\lambda v\\), then \\((A + %d I)v = Av + %d v = (%d + %d)v = %d v\\). The eigenvalue is %d.", c, c, lambda, c, ans, ans),
		}
	}
}

type eigenComputeGen struct{}

func (g *eigenComputeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	a := rand.Intn(max(1, scale*2)) - scale
	c := rand.Intn(max(1, scale*2)) - scale
	d := rand.Intn(max(1, scale*2)) - scale
	if a == 0 {
		a = 2
	}
	if d == 0 {
		d = 1
	}
	return generator.Problem{
		Question:    fmt.Sprintf("Find the eigenvalues of the triangular matrix \\(%s\\).", formatMatrix2([][]int{{a, c}, {0, d}})),
		Answer:      fmt.Sprintf("%d,%d", a, d),
		Explanation: fmt.Sprintf("For a triangular matrix, eigenvalues are the diagonal entries: λ₁=%d, λ₂=%d.", a, d),
	}
}

type transformationsGen struct{}

func (g *transformationsGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	a, b := rand.Intn(max(1, scale*2))+1, rand.Intn(max(1, scale*2))+1
	x, y := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	return generator.Problem{
		Question:    fmt.Sprintf("Under the linear transformation \\(T(x,y) = (%dx,%dy)\\), where does the point \\((%d,%d)\\) map to?", a, b, x, y),
		Answer:      fmt.Sprintf("(%d,%d)", a*x, b*y),
		Explanation: fmt.Sprintf("T(%d,%d) = (%d*%d, %d*%d) = (%d,%d)", x, y, a, x, b, y, a*x, b*y),
	}
}

type spanGen struct{}

func (g *spanGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Dimension of span of two vectors in R^2
		dependent := rand.Intn(2) == 0
		if dependent {
			v1 := smallInt(1, 4)
			v2 := smallInt(1, 4)
			m := smallInt(2, 4)
			w1 := m * v1
			w2 := m * v2
			return generator.Problem{
				Question:    fmt.Sprintf("What is the dimension of the subspace \\(\\operatorname{span}\\{(%d,%d), (%d,%d)\\}\\) in \\(\\mathbb{R}^2\\)?", v1, v2, w1, w2),
				Answer:      "1",
				Explanation: fmt.Sprintf("The second vector is %d times the first vector, so they are linearly dependent. The dimension of their span is 1.", m),
			}
		}
		v1 := smallInt(1, 4)
		v2 := 0
		w1 := 0
		w2 := smallInt(1, 4)
		return generator.Problem{
			Question:    fmt.Sprintf("What is the dimension of the subspace \\(\\operatorname{span}\\{(%d,%d), (%d,%d)\\}\\) in \\(\\mathbb{R}^2\\)?", v1, v2, w1, w2),
			Answer:      "2",
			Explanation: "The determinant of the two vectors is nonzero, so they are linearly independent and span a 2-dimensional subspace.",
		}
	case 1:
		// Linear combination scalar c
		c := smallInt(2, 6+int(ctx.Difficulty*3))
		x := smallInt(1, 5)
		y := smallInt(1, 5)
		cx := c * x
		cy := c * y
		return generator.Problem{
			Question:    fmt.Sprintf("Find the scalar \\(c\\) such that \\((%d,%d) = c(%d,%d)\\), showing that \\((%d,%d)\\) lies in \\(\\operatorname{span}\\{(%d,%d)\\}\\).", cx, cy, x, y, cx, cy, x, y),
			Answer:      fmt.Sprintf("%d", c),
			Explanation: fmt.Sprintf("Equating coordinates gives \\(%d = c(%d)\\), which yields \\(c = %d\\).", cx, x, c),
		}
	case 2:
		// Parameter k making vectors linearly dependent
		a := smallInt(1, 4)
		b := smallInt(1, 4)
		m := smallInt(2, 4)
		cVal := m * a
		k := m * b
		return generator.Problem{
			Question:    fmt.Sprintf("For what integer value of \\(k\\) do the vectors \\(v = (%d,%d)\\) and \\(w = (%d,k)\\) fail to span \\(\\mathbb{R}^2\\)?", a, b, cVal),
			Answer:      fmt.Sprintf("%d", k),
			Explanation: fmt.Sprintf("The vectors fail to span \\(\\mathbb{R}^2\\) when their determinant is zero: \\((%d)(k) - (%d)(%d) = 0\\), which gives \\(k = %d\\).", a, b, cVal, k),
		}
	case 3:
		// Dimension of 3 vectors with 0 in z coordinate
		x := smallInt(1, 5)
		y := smallInt(1, 5)
		return generator.Problem{
			Question:    fmt.Sprintf("Consider the vectors \\(v_1 = (1,0,0)\\), \\(v_2 = (0,1,0)\\) and \\(v_3 = (%d,%d,0)\\). What is the dimension of \\(\\operatorname{span}\\{v_1, v_2, v_3\\}\\)?", x, y),
			Answer:      "2",
			Explanation: "All three vectors lie in the xy-plane because their z-components are zero. The first two vectors form a basis for that plane, so the dimension is 2.",
		}
	default:
		// Span of k standard basis vectors in R^n
		k := smallInt(2, 4+int(ctx.Difficulty*2))
		n := k + smallInt(1, 3)
		return generator.Problem{
			Question:    fmt.Sprintf("Let \\(S\\) be a subset of %d distinct standard unit basis vectors in \\(\\mathbb{R}^{%d}\\). What is the dimension of \\(\\operatorname{span}(S)\\)?", k, n),
			Answer:      fmt.Sprintf("%d", k),
			Explanation: fmt.Sprintf("Distinct standard basis vectors are mutually orthogonal and linearly independent. Therefore %d such vectors span a subspace of dimension %d.", k, k),
		}
	}
}

type basisGen struct{}

func (g *basisGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Polynomial space P_n
		n := smallInt(1, 5+int(ctx.Difficulty*3))
		ans := n + 1
		return generator.Problem{
			Question:    fmt.Sprintf("What is the dimension of the vector space \\(\\mathcal{P}_{%d}\\) of real polynomials of degree at most %d?", n, n),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("A standard basis for \\(\\mathcal{P}_{%d}\\) is \\(\\{1, x, x^2, \\dots, x^{%d}\\}\\), which has %d elements. Thus \\(\\dim(\\mathcal{P}_{%d}) = %d\\).", n, n, ans, n, ans),
		}
	case 1:
		// Matrix space M_{m x n}
		m := smallInt(2, 4+int(ctx.Difficulty*2))
		n := smallInt(2, 4+int(ctx.Difficulty*2))
		ans := m * n
		return generator.Problem{
			Question:    fmt.Sprintf("What is the dimension of the vector space \\(\\mathcal{M}_{%d \\times %d}(\\mathbb{R})\\) of all %d-by-%d real matrices?", m, n, m, n),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("The standard basis consists of single-entry matrix units \\(E_{ij}\\). With %d rows and %d columns, the dimension is %d.", m, n, ans),
		}
	case 2:
		// Diagonal matrices D_n
		n := smallInt(2, 6+int(ctx.Difficulty*4))
		return generator.Problem{
			Question:    fmt.Sprintf("What is the dimension of the subspace of all %d-by-%d diagonal matrices in \\(\\mathcal{M}_{%d \\times %d}(\\mathbb{R})\\)?", n, n, n, n),
			Answer:      fmt.Sprintf("%d", n),
			Explanation: fmt.Sprintf("A diagonal matrix is completely determined by its %d diagonal entries, so its basis has %d matrices.", n, n),
		}
	case 3:
		// Coordinates with respect to an upper triangular basis
		a := smallInt(1, 3)
		c1 := smallInt(1, 5+int(ctx.Difficulty*3))
		c2 := smallInt(1, 4+int(ctx.Difficulty*2))
		w1 := c1 + a*c2
		w2 := c2
		return generator.Problem{
			Question:    fmt.Sprintf("Given the ordered basis \\(\\mathcal{B} = \\{(1,0), (%d,1)\\}\\), find the first coordinate \\(c_1\\) of the vector \\((%d,%d)\\) with respect to \\(\\mathcal{B}\\).", a, w1, w2),
			Answer:      fmt.Sprintf("%d", c1),
			Explanation: fmt.Sprintf("We solve \\(c_1(1,0) + c_2(%d,1) = (%d,%d)\\). The second component gives \\(c_2 = %d\\). The first component gives \\(c_1 + %d(%d) = %d\\), so \\(c_1 = %d - %d = %d\\).", a, w1, w2, w2, a, w2, w1, w1, a*w2, c1),
		}
	default:
		// Symmetric matrices Sym_n
		n := smallInt(2, 5)
		ans := n * (n + 1) / 2
		return generator.Problem{
			Question:    fmt.Sprintf("What is the dimension of the subspace of all %d-by-%d real symmetric matrices in \\(\\mathcal{M}_{%d \\times %d}(\\mathbb{R})\\)?", n, n, n, n),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("A symmetric matrix is determined by its diagonal and upper triangular entries. The number of independent entries is \\(\\frac{%d(%d + 1)}{2} = %d\\).", n, n, ans),
		}
	}
}

func formatMatrix(m [][]int) string {
	if len(m) == 2 && len(m[0]) == 2 {
		return formatMatrix2(m)
	}
	if len(m) == 3 && len(m[0]) == 3 {
		return formatMatrix3(m)
	}
	return fmt.Sprintf("%v", m)
}

func formatMatrix2(m [][]int) string {
	return fmt.Sprintf("[[%d,%d],[%d,%d]]", m[0][0], m[0][1], m[1][0], m[1][1])
}

func formatMatrix3(m [][]int) string {
	return fmt.Sprintf("[[%d,%d,%d],[%d,%d,%d],[%d,%d,%d]]",
		m[0][0], m[0][1], m[0][2],
		m[1][0], m[1][1], m[1][2],
		m[2][0], m[2][1], m[2][2])
}

func parseMatrix(s string) ([][]int, bool) {
	matches := matrixRe.FindStringSubmatch(s)
	if len(matches) != 3 {
		return nil, false
	}
	rows := make([][]int, 2)
	for i, part := range matches[1:] {
		parts := strings.Split(part, ",")
		if len(parts) != 2 {
			return nil, false
		}
		row := make([]int, 2)
		for j, p := range parts {
			val, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				return nil, false
			}
			row[j] = val
		}
		rows[i] = row
	}
	return rows, true
}

func gradeMatrix(expected, userAnswer string) grader.Result {
	eMat, ok := parseMatrix(expected)
	if !ok {
		return grader.Result{Correct: false, Score: 0, Feedback: "Invalid matrix format"}
	}
	aMat, ok := parseMatrix(userAnswer)
	if !ok {
		return grader.Result{Correct: false, Score: 0, Feedback: "Invalid matrix format"}
	}
	if len(eMat) != len(aMat) || len(eMat[0]) != len(aMat[0]) {
		return grader.Result{Correct: false, Score: 0, Feedback: "Matrix dimension mismatch"}
	}
	for i := range eMat {
		for j := range eMat[i] {
			if eMat[i][j] != aMat[i][j] {
				return grader.Result{Correct: false, Score: 0, Feedback: fmt.Sprintf("Mismatch at (%d,%d)", i+1, j+1)}
			}
		}
	}
	return grader.Result{Correct: true, Score: 1}
}

// Diagonalization: algebraic properties of diagonalizable matrices PDP⁻¹
type diagonalizationGen struct{}

func (g *diagonalizationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	rng := rand.New(rand.NewSource(rand.Int63()))
	l1 := mathutil.ScaleInt(ctx.Difficulty, 1, 5)
	l2 := l1 + mathutil.ScaleInt(ctx.Difficulty, 1, 4)
	if rand.Intn(2) == 0 {
		l1 = -l1
	}
	mat := mathutil.IntegerEigenvalues2x2(rng, l1, l2, 1+int(ctx.Difficulty*2))

	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Trace of diagonal matrix D
		tr := l1 + l2
		return generator.Problem{
			Question: fmt.Sprintf("The matrix \\(A = %s\\) is diagonalizable as \\(A = PDP^{-1}\\) with \\(D\\) diagonal. What is the trace of \\(D\\)?", mat.LaTeX()),
			Answer:   fmt.Sprintf("%d", tr),
			Explanation: fmt.Sprintf("Subgoal 1: State the similarity invariance of trace: \\(\\operatorname{tr}(D) = \\operatorname{tr}(A) = \\lambda_1 + \\lambda_2\\).\n"+
				"Subgoal 2: Compute the trace of \\(A\\) by summing its diagonal entries: \\((%d) + (%d)\\).\n"+
				"Subgoal 3: Evaluate the sum to find \\(\\operatorname{tr}(D) = %d\\).", mat.A, mat.D, tr),
		}
	case 1:
		// Determinant of diagonal matrix D
		det := l1 * l2
		return generator.Problem{
			Question: fmt.Sprintf("The matrix \\(A = %s\\) is diagonalizable as \\(A = PDP^{-1}\\) with \\(D\\) diagonal. What is the determinant of \\(D\\)?", mat.LaTeX()),
			Answer:   fmt.Sprintf("%d", det),
			Explanation: fmt.Sprintf("Subgoal 1: State the similarity invariance of determinant: \\(\\det(D) = \\det(A) = \\lambda_1 \\lambda_2\\).\n"+
				"Subgoal 2: Compute the determinant of \\(A\\): \\((%d)(%d) - (%d)(%d)\\).\n"+
				"Subgoal 3: Evaluate products to find \\(\\det(D) = %d\\).", mat.A, mat.D, mat.B, mat.C, det),
		}
	case 2:
		// Power trace tr(A^2)
		ans := l1*l1 + l2*l2
		return generator.Problem{
			Question: fmt.Sprintf("Let \\(A = PDP^{-1}\\) be diagonalizable with diagonal matrix \\(D = \\begin{pmatrix} %d & 0 \\\\ 0 & %d \\end{pmatrix}\\). Compute \\(\\operatorname{tr}(A^2)\\).", l1, l2),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State the power trace formula: \\(\\operatorname{tr}(A^2) = \\operatorname{tr}(D^2) = \\lambda_1^2 + \\lambda_2^2\\).\n"+
				"Subgoal 2: Substitute the eigenvalues \\(\\lambda_1 = %d\\) and \\(\\lambda_2 = %d\\): \\((%d)^2 + (%d)^2\\).\n"+
				"Subgoal 3: Evaluate squares and sum: %d + %d = %d.", l1, l2, l1, l2, l1*l1, l2*l2, ans),
		}
	case 3:
		// Larger eigenvalue of diagonalizable matrix
		larger := l1
		if l2 > larger {
			larger = l2
		}
		return generator.Problem{
			Question: fmt.Sprintf("The matrix \\(A = %s\\) has distinct integer eigenvalues. Find its largest eigenvalue.", mat.LaTeX()),
			Answer:   fmt.Sprintf("%d", larger),
			Explanation: fmt.Sprintf("Subgoal 1: State the characteristic equation \\(\\det(A - \\lambda I) = \\lambda^2 - \\operatorname{tr}(A)\\lambda + \\det(A) = 0\\).\n"+
				"Subgoal 2: Substitute \\(\\operatorname{tr}(A) = %d\\) and \\(\\det(A) = %d\\) to solve for roots \\(\\lambda_1 = %d\\) and \\(\\lambda_2 = %d\\).\n"+
				"Subgoal 3: Select the maximum root: %d.", mat.Trace(), mat.Det(), l1, l2, larger),
		}
	default:
		// Shifted determinant det(A - c I)
		c := mathutil.ScaleInt(ctx.Difficulty, 1, 4)
		ans := (l1 - c) * (l2 - c)
		return generator.Problem{
			Question: fmt.Sprintf("A diagonalizable matrix \\(A\\) has eigenvalues \\(\\lambda_1 = %d\\) and \\(\\lambda_2 = %d\\). Compute \\(\\det(A - %d I)\\).", l1, l2, c),
			Answer:   fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Subgoal 1: State the spectral shift property: eigenvalues of \\(A - cI\\) are \\(\\lambda_1 - c\\) and \\(\\lambda_2 - c\\), so \\(\\det(A - cI) = (\\lambda_1 - c)(\\lambda_2 - c)\\).\n"+
				"Subgoal 2: Substitute eigenvalues and shift scalar: \\((%d - %d)(%d - %d)\\).\n"+
				"Subgoal 3: Multiply the shifted factors to obtain %d.", l1, c, l2, c, ans),
		}
	}
}

// Rank of a matrix
type rankGen struct{}

func (g *rankGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	rng := rand.New(rand.NewSource(rand.Int63()))
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Full rank 2x2 matrix via RandomUnimodular2x2
		mat := mathutil.RandomUnimodular2x2(rng, 1+int(ctx.Difficulty*2))
		return generator.Problem{
			Question: fmt.Sprintf("What is the rank of matrix \\(%s\\)?", mat.LaTeX()),
			Answer:   "2",
			Explanation: fmt.Sprintf("Subgoal 1: State the full-rank criterion: an \\(n \\times n\\) matrix has rank \\(n\\) if and only if \\(\\det(A) \\ne 0\\).\n"+
				"Subgoal 2: Compute \\(\\det(A) = (%d)(%d) - (%d)(%d) = %d \\ne 0\\).\n"+
				"Subgoal 3: Conclude that both rows are linearly independent, so \\(\\operatorname{rank}(A) = 2\\).", mat.A, mat.D, mat.B, mat.C, mat.Det()),
		}
	case 1:
		// Rank 1 outer product 2x2 matrix
		u1 := mathutil.ScaleInt(ctx.Difficulty, 1, 5)
		u2 := mathutil.ScaleInt(ctx.Difficulty, 1, 5)
		k := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		if rand.Intn(2) == 0 {
			k = -k
		}
		m := [][]int{{u1, u2}, {k * u1, k * u2}}
		return generator.Problem{
			Question: fmt.Sprintf("What is the rank of matrix \\(%s\\)?", formatMatrix2(m)),
			Answer:   "1",
			Explanation: fmt.Sprintf("Subgoal 1: State the rank definition as the count of linearly independent rows.\n"+
				"Subgoal 2: Observe that row 2 is a scalar multiple of row 1 (ratio %d).\n"+
				"Subgoal 3: Conclude there is only 1 linearly independent row, so \\(\\operatorname{rank}(A) = 1\\).", k),
		}
	case 2:
		// Upper triangular 3x3 matrix with controlled pivots
		pivots := []int{mathutil.ScaleInt(ctx.Difficulty, 1, 4), mathutil.ScaleInt(ctx.Difficulty, 1, 4), mathutil.ScaleInt(ctx.Difficulty, 1, 4)}
		targetRank := 1 + rand.Intn(3)
		for i := targetRank; i < 3; i++ {
			pivots[i] = 0
		}
		m := [][]int{
			{pivots[0], smallInt(0, 3), smallInt(0, 3)},
			{0, pivots[1], smallInt(0, 3)},
			{0, 0, pivots[2]},
		}
		return generator.Problem{
			Question: fmt.Sprintf("What is the rank of the upper triangular matrix \\(%s\\)?", formatMatrix(m)),
			Answer:   fmt.Sprintf("%d", targetRank),
			Explanation: fmt.Sprintf("Subgoal 1: State the echelon rank property: for an upper triangular matrix, rank equals the count of nonzero diagonal pivots.\n"+
				"Subgoal 2: Inspect diagonal entries: %d, %d and %d.\n"+
				"Subgoal 3: Count the %d nonzero pivot(s) to determine \\(\\operatorname{rank}(A) = %d\\).", pivots[0], pivots[1], pivots[2], targetRank, targetRank),
		}
	case 3:
		// Rank-Nullity theorem application
		n := mathutil.ScaleInt(ctx.Difficulty, 3, 6)
		nullity := 1 + rand.Intn(n-1)
		rank := n - nullity
		return generator.Problem{
			Question: fmt.Sprintf("Let \\(T: \\mathbb{R}^{%d} \\to \\mathbb{R}^{%d}\\) be a linear transformation with nullity \\(\\dim(\\ker(T)) = %d\\). What is the rank \\(\\dim(\\operatorname{im}(T))\\)?", n, n, nullity),
			Answer:   fmt.Sprintf("%d", rank),
			Explanation: fmt.Sprintf("Subgoal 1: State the Rank-Nullity Theorem: \\(\\operatorname{rank}(T) + \\operatorname{nullity}(T) = n\\), where \\(n\\) is the domain dimension.\n"+
				"Subgoal 2: Substitute domain dimension \\(n = %d\\) and nullity = %d: \\(\\operatorname{rank}(T) + %d = %d\\).\n"+
				"Subgoal 3: Solve for rank: \\(%d - %d = %d\\).", n, nullity, nullity, n, n, nullity, rank),
		}
	default:
		// Collinear 3x3 rows
		a, b, cVal := smallInt(1, 3), smallInt(1, 3), smallInt(1, 3)
		m := [][]int{
			{a, b, cVal},
			{2 * a, 2 * b, 2 * cVal},
			{3 * a, 3 * b, 3 * cVal},
		}
		return generator.Problem{
			Question: fmt.Sprintf("What is the rank of matrix \\(%s\\)?", formatMatrix(m)),
			Answer:   "1",
			Explanation: fmt.Sprintf("Subgoal 1: Recall that the rank is the dimension of the subspace spanned by the row vectors.\n"+
				"Subgoal 2: Observe that row 2 = 2 \\times row 1 and row 3 = 3 \\times row 1, meaning all rows are scalar multiples of row 1.\n"+
				"Subgoal 3: Conclude the row space is 1-dimensional, so \\(\\operatorname{rank}(A) = 1\\)."),
		}
	}
}

// Cosine similarity between two vectors
type cosineSimilarityGen struct{}

func (g *cosineSimilarityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Parallel vectors with positive scalar k > 0
		a := mathutil.ScaleInt(ctx.Difficulty, 1, 6)
		b := mathutil.ScaleInt(ctx.Difficulty, 1, 6)
		k := mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		uStr := fmt.Sprintf("(%d,%d)", a, b)
		vStr := fmt.Sprintf("(%d,%d)", k*a, k*b)
		return generator.Problem{
			Question: fmt.Sprintf("Compute the cosine similarity between \\(u=%s\\) and \\(v=%s\\).", uStr, vStr),
			Answer:   "1",
			Explanation: fmt.Sprintf("Subgoal 1: State the cosine similarity formula: \\(\\cos(\\theta) = \\frac{u \\cdot v}{\\|u\\| \\|v\\|}\\).\n"+
				"Subgoal 2: Note \\(v = %d u\\) with positive scalar \\(k = %d\\), so \\(u\\) and \\(v\\) point in the identical direction (\\(\\theta = 0\\)).\n"+
				"Subgoal 3: Evaluate \\(\\cos(0) = 1\\).", k, k),
		}
	case 1:
		// Anti-parallel vectors with negative scalar k < 0
		a := mathutil.ScaleInt(ctx.Difficulty, 1, 6)
		b := mathutil.ScaleInt(ctx.Difficulty, 1, 6)
		k := -mathutil.ScaleInt(ctx.Difficulty, 2, 4)
		uStr := fmt.Sprintf("(%d,%d)", a, b)
		vStr := fmt.Sprintf("(%d,%d)", k*a, k*b)
		return generator.Problem{
			Question: fmt.Sprintf("Compute the cosine similarity between \\(u=%s\\) and \\(v=%s\\).", uStr, vStr),
			Answer:   "-1",
			Explanation: fmt.Sprintf("Subgoal 1: State the cosine similarity formula: \\(\\cos(\\theta) = \\frac{u \\cdot v}{\\|u\\| \\|v\\|}\\).\n"+
				"Subgoal 2: Note \\(v = %d u\\) with negative scalar \\(k = %d\\), so \\(u\\) and \\(v\\) point in opposite directions (\\(\\theta = \\pi\\)).\n"+
				"Subgoal 3: Evaluate \\(\\cos(\\pi) = -1\\).", k, k),
		}
	case 2:
		// Orthogonal 2D vectors (a, b) and (-b, a)
		a := mathutil.ScaleInt(ctx.Difficulty, 1, 8)
		b := mathutil.ScaleInt(ctx.Difficulty, 1, 8)
		uStr := fmt.Sprintf("(%d,%d)", a, b)
		vStr := fmt.Sprintf("(%d,%d)", -b, a)
		return generator.Problem{
			Question: fmt.Sprintf("Compute the cosine similarity between orthogonal vectors \\(u=%s\\) and \\(v=%s\\).", uStr, vStr),
			Answer:   "0",
			Explanation: fmt.Sprintf("Subgoal 1: State the cosine similarity formula: \\(\\cos(\\theta) = \\frac{u \\cdot v}{\\|u\\| \\|v\\|}\\).\n"+
				"Subgoal 2: Compute the dot product: \\(u \\cdot v = (%d)(%d) + (%d)(%d) = %d + (%d) = 0\\).\n"+
				"Subgoal 3: Divide numerator zero by the nonzero norms to obtain \\(\\cos(\\theta) = 0\\).", a, -b, b, a, -a*b, a*b),
		}
	case 3:
		// Orthogonal 3D vectors
		a := mathutil.ScaleInt(ctx.Difficulty, 1, 6)
		b := mathutil.ScaleInt(ctx.Difficulty, 1, 6)
		cVal := mathutil.ScaleInt(ctx.Difficulty, 1, 6)
		uStr := fmt.Sprintf("(%d,0,0)", a)
		vStr := fmt.Sprintf("(0,%d,%d)", b, cVal)
		return generator.Problem{
			Question: fmt.Sprintf("Compute the cosine similarity between vectors \\(u=%s\\) and \\(v=%s\\).", uStr, vStr),
			Answer:   "0",
			Explanation: fmt.Sprintf("Subgoal 1: State the cosine similarity formula: \\(\\cos(\\theta) = \\frac{u \\cdot v}{\\|u\\| \\|v\\|}\\).\n"+
				"Subgoal 2: Compute dot product: \\((%d)(0) + (0)(%d) + (0)(%d) = 0\\).\n"+
				"Subgoal 3: Evaluate \\(\\frac{0}{\\|u\\| \\|v\\|} = 0\\).", a, b, cVal),
		}
	default:
		// Identical standard unit vectors
		return generator.Problem{
			Question: "Compute the cosine similarity between unit vectors \\(u=(1,0)\\) and \\(v=(1,0)\\).",
			Answer:   "1",
			Explanation: "Subgoal 1: State the cosine similarity formula for identical unit vectors: \\(\\cos(\\theta) = \\frac{u \\cdot v}{\\|u\\| \\|v\\|}\\).\n" +
				"Subgoal 2: Compute \\(u \\cdot v = (1)(1) + (0)(0) = 1\\) and \\(\\|u\\| = 1, \\|v\\| = 1\\).\n" +
				"Subgoal 3: Evaluate \\(\\frac{1}{1 \\times 1} = 1\\).",
		}
	}
}

// Parametric form of a line
type parametricGen struct{}

func (g *parametricGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		x0, y0, dx, dy int
	}
	directions := []entry{
		{1, 2, 3, 4},
		{0, 0, 1, 2},
		{2, 3, -1, 2},
		{-1, 1, 2, -3},
		{0, 1, 1, 0},
		{1, 0, 0, 1},
		{3, 0, 1, 1},
		{-2, 4, 3, -1},
	}
	e := directions[rand.Intn(len(directions))]
	return generator.Problem{
		Question:    fmt.Sprintf("Find the parametric equations for the line through \\((%d,%d)\\) with direction vector \\((%d,%d)\\).", e.x0, e.y0, e.dx, e.dy),
		Answer:      fmt.Sprintf("x=%d+%dt,y=%d+%dt", e.x0, e.dx, e.y0, e.dy),
		Explanation: fmt.Sprintf("The parametric form is (x,y) = (%d,%d) + t(%d,%d), so x=%d+%dt, y=%d+%dt.", e.x0, e.y0, e.dx, e.dy, e.x0, e.dx, e.y0, e.dy),
	}
}

type innerProductGen struct{}

func (g *innerProductGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Weighted inner product
		w1 := smallInt(1, 4)
		w2 := smallInt(1, 4)
		u1 := smallInt(1, 4)
		u2 := smallInt(1, 4)
		v1 := smallInt(1, 4)
		v2 := smallInt(1, 4)
		ans := w1*u1*v1 + w2*u2*v2
		return generator.Problem{
			Question:    fmt.Sprintf("Compute the weighted inner product \\(\\langle u, v \\rangle = %d u_1 v_1 + %d u_2 v_2\\) for \\(u = (%d,%d)\\) and \\(v = (%d,%d)\\).", w1, w2, u1, u2, v1, v2),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Substituting components gives \\(\\langle u, v \\rangle = %d(%d)(%d) + %d(%d)(%d) = %d\\).", w1, u1, v1, w2, u2, v2, ans),
		}
	case 1:
		// Orthogonality parameter k
		a := smallInt(1, 4)
		b := smallInt(1, 4)
		m := smallInt(1, 3)
		c := m * b
		k := -m * a
		return generator.Problem{
			Question:    fmt.Sprintf("For what integer value of \\(k\\) are the vectors \\(u = (%d,%d)\\) and \\(v = (%d,k)\\) orthogonal under the standard dot product?", a, b, c),
			Answer:      fmt.Sprintf("%d", k),
			Explanation: fmt.Sprintf("Vectors are orthogonal when their inner product is zero: \\((%d)(%d) + (%d)(k) = 0\\), which gives \\(k = %d\\).", a, c, b, k),
		}
	case 2:
		// Matrix-induced inner product
		d1 := smallInt(1, 3)
		d2 := smallInt(1, 3)
		u1 := smallInt(1, 3)
		u2 := smallInt(1, 3)
		v1 := smallInt(1, 3)
		v2 := smallInt(1, 3)
		ans := d1*u1*v1 + d2*u2*v2
		return generator.Problem{
			Question:    fmt.Sprintf("For the diagonal matrix \\(A = \\begin{pmatrix} %d & 0 \\\\ 0 & %d \\end{pmatrix}\\), compute the induced inner product \\(\\langle u, v \\rangle_A = u^T A v\\) for \\(u = (%d,%d)\\) and \\(v = (%d,%d)\\).", d1, d2, u1, u2, v1, v2),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Evaluating \\(u^T A v = %d(%d)(%d) + %d(%d)(%d)\\) gives %d.", d1, u1, v1, d2, u2, v2, ans),
		}
	case 3:
		// Polynomial inner product \int_0^1 (2m x) * b dx = m b
		m := smallInt(1, 4)
		a := 2 * m
		b := smallInt(1, 5)
		ans := m * b
		return generator.Problem{
			Question:    fmt.Sprintf("On the vector space of polynomials with inner product \\(\\langle f, g \\rangle = \\int_0^1 f(x) g(x) \\, dx\\), compute \\(\\langle %d x, %d \\rangle\\).", a, b),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("The integral is \\(\\int_0^1 %d x \\cdot %d \\, dx = %d \\left[ \\frac{x^2}{2} \\right]_0^1 = %d \\left(\\frac{1}{2}\\right) = %d\\).", a, b, a*b, a*b, ans),
		}
	default:
		// Orthogonal projection coefficient
		v1 := smallInt(1, 3)
		v2 := smallInt(1, 3)
		vNormSq := v1*v1 + v2*v2
		m := smallInt(1, 4+int(ctx.Difficulty*2))
		u1 := m * v1
		u2 := m * v2
		dot := u1*v1 + u2*v2
		return generator.Problem{
			Question:    fmt.Sprintf("Find the scalar projection coefficient \\(c = \\frac{\\langle u, v \\rangle}{\\langle v, v \\rangle}\\) of \\(u = (%d,%d)\\) onto \\(v = (%d,%d)\\) under the standard dot product.", u1, u2, v1, v2),
			Answer:      fmt.Sprintf("%d", m),
			Explanation: fmt.Sprintf("We have \\(\\langle u, v \\rangle = %d(%d) + %d(%d) = %d\\) and \\(\\langle v, v \\rangle = %d^2 + %d^2 = %d\\). The projection coefficient is %d / %d = %d.", u1, v1, u2, v2, dot, v1, v2, vNormSq, dot, vNormSq, m),
		}
	}
}

type luGen struct{}

func (g *luGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// det(A) = det(U)
		u11 := smallInt(1, 4)
		u22 := smallInt(1, 4)
		u33 := smallInt(1, 4)
		ans := u11 * u22 * u33
		return generator.Problem{
			Question:    fmt.Sprintf("An \\(n \\times n\\) matrix \\(A\\) has LU factorization \\(A = LU\\) where \\(L\\) is unit lower triangular and \\(U\\) has diagonal entries %d, %d and %d. Compute \\(\\det(A)\\).", u11, u22, u33),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Since \\(L\\) is unit lower triangular, \\(\\det(L) = 1\\). The determinant of \\(U\\) is the product of its diagonal entries: \\(\\det(U) = (%d)(%d)(%d) = %d\\). Therefore \\(\\det(A) = \\det(L)\\det(U) = %d\\).", u11, u22, u33, ans, ans),
		}
	case 1:
		// Forward solve Ly = b
		l21 := smallInt(1, 4)
		b1 := smallInt(1, 5)
		y2 := smallInt(1, 5)
		b2 := l21*b1 + y2
		return generator.Problem{
			Question:    fmt.Sprintf("In the first stage of solving \\(Ax = b\\) via \\(A = LU\\), solve \\(Ly = b\\) for \\(y_2\\) where \\(L = \\begin{pmatrix} 1 & 0 \\\\ %d & 1 \\end{pmatrix}\\) and \\(b = \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\).", l21, b1, b2),
			Answer:      fmt.Sprintf("%d", y2),
			Explanation: fmt.Sprintf("Forward substitution gives \\(y_1 = %d\\). The second equation is \\(%d(y_1) + y_2 = %d\\), so \\(y_2 = %d - %d(%d) = %d\\).", b1, l21, b2, b2, l21, b1, y2),
		}
	case 2:
		// Back substitution Ux = y
		u11 := smallInt(1, 3)
		u12 := smallInt(1, 3)
		u22 := smallInt(1, 3)
		x2 := smallInt(1, 4)
		y2 := u22 * x2
		x1 := smallInt(1, 4)
		y1 := u11*x1 + u12*x2
		_ = y1
		return generator.Problem{
			Question:    fmt.Sprintf("In the back substitution stage of solving \\(Ax = b\\), solve \\(Ux = y\\) for \\(x_2\\) where \\(U = \\begin{pmatrix} %d & %d \\\\ 0 & %d \\end{pmatrix}\\) and \\(y = \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\).", u11, u12, u22, y1, y2),
			Answer:      fmt.Sprintf("%d", x2),
			Explanation: fmt.Sprintf("From the second row, \\(%d x_2 = %d\\), which yields \\(x_2 = %d / %d = %d\\).", u22, y2, y2, u22, x2),
		}
	case 3:
		// Multiplier l_{21} in Gaussian elimination
		a11 := smallInt(1, 4)
		m := smallInt(2, 4)
		a21 := m * a11
		a12 := smallInt(1, 4)
		a22 := smallInt(1, 5)
		return generator.Problem{
			Question:    fmt.Sprintf("In computing the LU factorization \\(A = LU\\) of \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\) without pivoting, what is the entry \\(l_{21}\\) of the unit lower triangular factor \\(L\\)?", a11, a12, a21, a22),
			Answer:      fmt.Sprintf("%d", m),
			Explanation: fmt.Sprintf("The multiplier used to eliminate entry \\(a_{21}\\) using row 1 is \\(l_{21} = \\frac{a_{21}}{a_{11}} = \\frac{%d}{%d} = %d\\).", a21, a11, m),
		}
	default:
		// Diagonal entry u_{22} in U
		l21 := smallInt(1, 3)
		u11 := smallInt(1, 4)
		u12 := smallInt(1, 4)
		u22 := smallInt(1, 5)
		a11 := u11
		a12 := u12
		a21 := l21 * u11
		a22 := l21*u12 + u22
		_ = a11
		_ = a21
		return generator.Problem{
			Question:    fmt.Sprintf("In computing the LU factorization \\(A = LU\\) of \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\) where \\(L = \\begin{pmatrix} 1 & 0 \\\\ %d & 1 \\end{pmatrix}\\), what is the diagonal entry \\(u_{22}\\)?", a11, a12, a21, a22, l21),
			Answer:      fmt.Sprintf("%d", u22),
			Explanation: fmt.Sprintf("Matrix multiplication gives \\(a_{22} = l_{21} u_{12} + u_{22} = %d(%d) + u_{22} = %d\\). Thus \\(u_{22} = %d - %d = %d\\).", l21, u12, a22, a22, l21*u12, u22),
		}
	}
}

type qrGen struct{}

func (g *qrGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	triples := [][3]int{
		{3, 4, 5},
		{4, 3, 5},
		{6, 8, 10},
		{8, 6, 10},
		{5, 12, 13},
		{12, 5, 13},
	}
	trip := triples[rand.Intn(len(triples))]
	a, b, r := trip[0], trip[1], trip[2]

	u := 1 + rand.Intn(3)
	v := 1 + rand.Intn(3)
	if rand.Intn(2) == 0 {
		u = -u
	}
	c := u*a - v*b
	d := u*b + v*a

	switch rand.Intn(5) {
	case 0:
		return generator.Problem{
			Question: fmt.Sprintf("The first column of real matrix \\(A\\) is \\(a_1 = \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\). In the QR decomposition \\(A = QR\\), find the diagonal entry \\(r_{11} = \\|a_1\\|\\).", a, b),
			Answer:   fmt.Sprintf("%d", r),
			Explanation: fmt.Sprintf("The entry \\(r_{11}\\) is the Euclidean norm of the first column: \\(r_{11} = \\sqrt{%d^2 + %d^2} = \\sqrt{%d} = %d\\).",
				a, b, a*a+b*b, r),
		}
	case 1:
		r12 := u * r
		return generator.Problem{
			Question: fmt.Sprintf("Consider matrix \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\) with columns \\(a_1\\) and \\(a_2\\). In the QR factorization \\(A = QR\\), compute the upper entry \\(r_{12} = q_1^T a_2\\).", a, c, b, d),
			Answer:   fmt.Sprintf("%d", r12),
			Explanation: fmt.Sprintf("The first normalized column is \\(q_1 = \\frac{1}{%d} \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\). Taking the inner product gives \\(r_{12} = q_1^T a_2 = \\frac{1}{%d}(%d \\cdot (%d) + %d \\cdot (%d)) = \\frac{%d}{%d} = %d\\).",
				r, a, b, r, a, c, b, d, a*c+b*d, r, r12),
		}
	case 2:
		r22 := v * r
		return generator.Problem{
			Question: fmt.Sprintf("Consider matrix \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\). In the QR decomposition \\(A = QR\\) with positive diagonal entries, find \\(r_{22}\\).", a, c, b, d),
			Answer:   fmt.Sprintf("%d", r22),
			Explanation: fmt.Sprintf("Orthogonalizing \\(a_2\\) gives \\(v_2 = a_2 - r_{12} q_1 = \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\). Its Euclidean length gives \\(r_{22} = \\|v_2\\| = %d\\).",
				-v*b, v*a, r22),
		}
	case 3:
		d1 := 1 + rand.Intn(5)
		d2 := 1 + rand.Intn(5)
		d3 := 1 + rand.Intn(4)
		detR := d1 * d2 * d3
		return generator.Problem{
			Question: fmt.Sprintf("A \\(3 \\times 3\\) matrix \\(A\\) has QR decomposition \\(A = QR\\), where \\(Q\\) is an orthogonal matrix with \\(\\det(Q) = 1\\) and \\(R\\) has diagonal entries %d, %d and %d. Find \\(\\det(A)\\).", d1, d2, d3),
			Answer:   fmt.Sprintf("%d", detR),
			Explanation: fmt.Sprintf("Because \\(A = QR\\), the determinant satisfies \\(\\det(A) = \\det(Q) \\det(R)\\). For triangular matrix \\(R\\), the determinant is the product of diagonal entries: \\(\\det(R) = %d \\times %d \\times %d = %d\\). With \\(\\det(Q) = 1\\), \\(\\det(A) = %d\\).",
				d1, d2, d3, detR, detR),
		}
	default:
		r11 := 1 + rand.Intn(4)
		r12 := rand.Intn(4) + 1
		r22 := 1 + rand.Intn(4)
		x2 := 1 + rand.Intn(5)
		x1 := 1 + rand.Intn(5)
		c2 := r22 * x2
		c1 := r11*x1 + r12*x2
		return generator.Problem{
			Question: fmt.Sprintf("To solve the least-squares problem \\(Ax = b\\) via QR decomposition \\(Rx = Q^T b\\), let \\(R = \\begin{pmatrix} %d & %d \\\\ 0 & %d \\end{pmatrix}\\) and \\(Q^T b = \\begin{pmatrix} %d \\\\ %d \\end{pmatrix}\\). Find \\(x_1\\).", r11, r12, r22, c1, c2),
			Answer:   fmt.Sprintf("%d", x1),
			Explanation: fmt.Sprintf("From the second equation, \\(%d x_2 = %d\\) gives \\(x_2 = %d\\). Back substitution into the first row yields \\(%d x_1 + %d(%d) = %d\\), so \\(x_1 = %d\\).",
				r22, c2, x2, r11, r12, x2, c1, x1),
		}
	}
}

type svdGen struct{}

func (g *svdGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		s1 := 2 + rand.Intn(8)
		s2 := 1 + rand.Intn(s1 - 1)
		eig1 := s1 * s1
		eig2 := s2 * s2
		return generator.Problem{
			Question: fmt.Sprintf("The symmetric matrix \\(A^T A\\) has eigenvalues %d, %d and 0. Compute the largest singular value \\(\\sigma_1\\) of \\(A\\).", eig1, eig2),
			Answer:   fmt.Sprintf("%d", s1),
			Explanation: fmt.Sprintf("The singular values of \\(A\\) are the non-negative square roots of the eigenvalues of \\(A^T A\\): \\(\\sigma_1 = \\sqrt{%d} = %d\\).",
				eig1, s1),
		}
	case 1:
		s1 := 4 + rand.Intn(7)
		s2 := 2 + rand.Intn(s1 - 2)
		s3 := 1 + rand.Intn(s2 - 1)
		return generator.Problem{
			Question: fmt.Sprintf("A matrix \\(A\\) has singular values %d, %d and %d. Find the spectral norm \\(\\|A\\|_2\\) of the matrix.", s1, s2, s3),
			Answer:   fmt.Sprintf("%d", s1),
			Explanation: fmt.Sprintf("The matrix 2-norm (spectral norm) equals the largest singular value: \\(\\|A\\|_2 = \\sigma_1 = %d\\).", s1),
		}
	case 2:
		s1 := 2 + rand.Intn(5)
		s2 := 1 + rand.Intn(4)
		s3 := 1 + rand.Intn(3)
		frobSq := s1*s1 + s2*s2 + s3*s3
		return generator.Problem{
			Question: fmt.Sprintf("A matrix \\(A\\) has singular values %d, %d and %d. Compute the squared Frobenius norm \\(\\|A\\|_F^2\\).", s1, s2, s3),
			Answer:   fmt.Sprintf("%d", frobSq),
			Explanation: fmt.Sprintf("The squared Frobenius norm is the sum of the squares of the singular values: \\(\\|A\\|_F^2 = %d^2 + %d^2 + %d^2 = %d + %d + %d = %d\\).",
				s1, s2, s3, s1*s1, s2*s2, s3*s3, frobSq),
		}
	case 3:
		minSig := 1 + rand.Intn(3)
		mult := 2 + rand.Intn(5)
		maxSig := minSig * mult
		midSig := minSig + 1
		if midSig >= maxSig {
			midSig = minSig
		}
		return generator.Problem{
			Question: fmt.Sprintf("An invertible real matrix \\(A\\) has singular values %d, %d and %d. Find the 2-norm condition number \\(\\kappa_2(A) = \\frac{\\sigma_{\\max}}{\\sigma_{\\min}}\\).", maxSig, midSig, minSig),
			Answer:   fmt.Sprintf("%d", mult),
			Explanation: fmt.Sprintf("The 2-norm condition number is the ratio of the largest to the smallest singular value: \\(\\kappa_2(A) = \\frac{\\sigma_{\\max}}{\\sigma_{\\min}} = \\frac{%d}{%d} = %d\\).",
				maxSig, minSig, mult),
		}
	default:
		numNonZero := 2 + rand.Intn(2)
		numZero := 1 + rand.Intn(2)
		vals := make([]string, 0, numNonZero+numZero)
		for i := 0; i < numNonZero; i++ {
			vals = append(vals, fmt.Sprintf("%d", 2+rand.Intn(7)))
		}
		for i := 0; i < numZero; i++ {
			vals = append(vals, "0")
		}
		var listStr string
		if len(vals) == 3 {
			listStr = fmt.Sprintf("%s, %s and %s", vals[0], vals[1], vals[2])
		} else {
			listStr = fmt.Sprintf("%s, %s, %s and %s", vals[0], vals[1], vals[2], vals[3])
		}
		return generator.Problem{
			Question: fmt.Sprintf("A real matrix \\(A\\) has singular values %s. Find the rank of \\(A\\).", listStr),
			Answer:   fmt.Sprintf("%d", numNonZero),
			Explanation: fmt.Sprintf("The rank of a matrix equals the number of nonzero singular values in its singular value decomposition. With %d nonzero singular values, \\(\\operatorname{rank}(A) = %d\\).",
				numNonZero, numNonZero),
		}
	}
}

type jordanGen struct{}

func (g *jordanGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	switch rand.Intn(5) {
	case 0:
		alg := 3 + rand.Intn(3)
		geom := 1 + rand.Intn(alg - 1)
		eig := 1 + rand.Intn(5)
		return generator.Problem{
			Question: fmt.Sprintf("An eigenvalue \\(\\lambda = %d\\) of matrix \\(A\\) has algebraic multiplicity %d and geometric multiplicity \\(\\dim \\ker(A - %d I) = %d\\). How many Jordan blocks corresponding to \\(\\lambda\\) are in the Jordan canonical form of \\(A\\)?", eig, alg, eig, geom),
			Answer:   fmt.Sprintf("%d", geom),
			Explanation: fmt.Sprintf("The number of Jordan blocks corresponding to an eigenvalue equals its geometric multiplicity: \\(\\dim \\ker(A - \\lambda I) = %d\\).", geom),
		}
	case 1:
		l1 := 1 + rand.Intn(4)
		l2 := 1 + rand.Intn(4)
		l3 := 1 + rand.Intn(4)
		tr := 2*l1 + l2 + l3
		return generator.Problem{
			Question: fmt.Sprintf("A \\(4 \\times 4\\) matrix has Jordan blocks \\(J_2(%d)\\), \\(J_1(%d)\\) and \\(J_1(%d)\\). Compute the trace of the matrix.", l1, l2, l3),
			Answer:   fmt.Sprintf("%d", tr),
			Explanation: fmt.Sprintf("The trace of a matrix equals the sum of the diagonal entries in its Jordan normal form: \\(\\operatorname{tr}(A) = 2(%d) + %d + %d = %d + %d + %d = %d\\).",
				l1, l2, l3, 2*l1, l2, l3, tr),
		}
	case 2:
		l1 := 2 + rand.Intn(3)
		l2 := 2 + rand.Intn(3)
		detA := l1 * l1 * l2
		return generator.Problem{
			Question: fmt.Sprintf("A \\(3 \\times 3\\) matrix has Jordan canonical form consisting of a \\(2 \\times 2\\) block \\(J_2(%d)\\) and a \\(1 \\times 1\\) block \\(J_1(%d)\\). Find \\(\\det(A)\\).", l1, l2),
			Answer:   fmt.Sprintf("%d", detA),
			Explanation: fmt.Sprintf("Because Jordan normal form is triangular, the determinant is the product of all diagonal entries: \\(\\det(A) = %d^2 \\times %d = %d \\times %d = %d\\).",
				l1, l2, l1*l1, l2, detA),
		}
	case 3:
		alg := 3 + rand.Intn(3)
		geom := 1 + rand.Intn(alg - 1)
		defect := alg - geom
		eig := 2 + rand.Intn(4)
		return generator.Problem{
			Question: fmt.Sprintf("An eigenspace for eigenvalue \\(\\lambda = %d\\) has algebraic multiplicity %d and geometric multiplicity %d. How many generalized eigenvectors that are not ordinary eigenvectors are needed to complete a Jordan chain basis for this eigenspace?", eig, alg, geom),
			Answer:   fmt.Sprintf("%d", defect),
			Explanation: fmt.Sprintf("The number of additional generalized eigenvectors is the defect between algebraic multiplicity and geometric multiplicity: %d - %d = %d.",
				alg, geom, defect),
		}
	default:
		e1 := 2 + rand.Intn(3)
		e2 := 1 + rand.Intn(2)
		l1 := 1 + rand.Intn(4)
		l2 := l1 + 1 + rand.Intn(3)
		return generator.Problem{
			Question: fmt.Sprintf("A matrix has minimal polynomial \\(m(\\lambda) = (\\lambda - %d)^{%d}(\\lambda - %d)^{%d}\\). What is the size of the largest Jordan block corresponding to eigenvalue %d?", l1, e1, l2, e2, l1),
			Answer:   fmt.Sprintf("%d", e1),
			Explanation: fmt.Sprintf("The size of the largest Jordan block for an eigenvalue equals its multiplicity as a root of the minimal polynomial. For eigenvalue %d, the exponent in \\(m(\\lambda)\\) is %d, so the largest block has size %d.",
				l1, e1, e1),
		}
	}
}

// ----- subspace -----

type subspaceGen struct{}

func (g *subspaceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	mode := rand.Intn(5)
	switch mode {
	case 0:
		// Grassmann's dimension formula: dim(U + W) = dim(U) + dim(W) - dim(U cap W)
		d12 := smallInt(1, 3)
		d1 := d12 + smallInt(1, 3+int(ctx.Difficulty*2))
		d2 := d12 + smallInt(1, 3+int(ctx.Difficulty*2))
		ans := d1 + d2 - d12
		return generator.Problem{
			Question:    fmt.Sprintf("Two subspaces \\(U\\) and \\(W\\) of a vector space satisfy \\(\\dim(U) = %d\\), \\(\\dim(W) = %d\\) and \\(\\dim(U \\cap W) = %d\\). What is \\(\\dim(U + W)\\)?", d1, d2, d12),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("By Grassmann's dimension formula, \\(\\dim(U + W) = \\dim(U) + \\dim(W) - \\dim(U \\cap W) = %d + %d - %d = %d\\).", d1, d2, d12, ans),
		}
	case 1:
		// Direct sum: dim(U \oplus W) = dim(U) + dim(W)
		d1 := smallInt(1, 4+int(ctx.Difficulty*2))
		d2 := smallInt(1, 4+int(ctx.Difficulty*2))
		ans := d1 + d2
		return generator.Problem{
			Question:    fmt.Sprintf("If \\(V = U \\oplus W\\) is the direct sum of subspaces \\(U\\) and \\(W\\) with \\(\\dim(U) = %d\\) and \\(\\dim(W) = %d\\), what is \\(\\dim(V)\\)?", d1, d2),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("For a direct sum \\(U \\cap W = \\{0\\}\\), so \\(\\dim(U \\oplus W) = \\dim(U) + \\dim(W) = %d + %d = %d\\).", d1, d2, ans),
		}
	case 2:
		// Orthogonal complement dimension: dim(W^\perp) = n - dim(W)
		n := smallInt(3, 6+int(ctx.Difficulty*3))
		dimW := smallInt(1, n-1)
		ans := n - dimW
		return generator.Problem{
			Question:    fmt.Sprintf("In \\(\\mathbb{R}^{%d}\\), let \\(W\\) be a subspace of dimension %d. What is the dimension of the orthogonal complement \\(W^\\perp\\)?", n, dimW),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("In an \\(n\\)-dimensional space, \\(\\dim(W) + \\dim(W^\\perp) = n\\). Here \\(\\dim(W^\\perp) = %d - %d = %d\\).", n, dimW, ans),
		}
	case 3:
		// Hyperplane defined by a single linear constraint
		n := smallInt(3, 7+int(ctx.Difficulty*3))
		ans := n - 1
		return generator.Problem{
			Question:    fmt.Sprintf("What is the dimension of the subspace in \\(\\mathbb{R}^{%d}\\) defined by the single equation \\(x_1 + x_2 + \\dots + x_{%d} = 0\\)?", n, n),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("A single nonzero homogeneous linear equation in \\(\\mathbb{R}^{%d}\\) imposes 1 constraint, so the dimension is %d - 1 = %d.", n, n, ans),
		}
	default:
		// k independent linear equations
		n := smallInt(4, 7+int(ctx.Difficulty*2))
		k := smallInt(2, n-2)
		ans := n - k
		return generator.Problem{
			Question:    fmt.Sprintf("What is the dimension of the subspace in \\(\\mathbb{R}^{%d}\\) defined by %d linearly independent homogeneous linear equations?", n, k),
			Answer:      fmt.Sprintf("%d", ans),
			Explanation: fmt.Sprintf("Each independent constraint reduces the dimension of the solution space by 1. Therefore the dimension is %d - %d = %d.", n, k, ans),
		}
	}
}
