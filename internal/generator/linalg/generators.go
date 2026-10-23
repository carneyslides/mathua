package linalg

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
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

func (g *matrixIdentityGen) AnswerFormat() generator.AnswerFormat {
	return generator.AnswerFormat{Hint: "Matrix with rows in brackets, e.g. [[1,2],[3,4]]", InputMode: "text"}
}

func (g *matrixIdentityGen) Grade(expected, userAnswer string) grader.Result {
	return gradeMatrix(expected, userAnswer)
}

func (g *matrixIdentityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	candidates := [][4]int{
		{2, 1, 1, 1}, // det=1
		{3, 2, 2, 1}, // det=-1
		{3, 4, 2, 3}, // det=1
		{4, 3, 3, 2}, // det=-1
		{5, 3, 3, 2}, // det=1
		{5, 7, 2, 3}, // det=1
		{4, 5, 3, 4}, // det=1
		{3, 1, 4, 1}, // det=-1
		{2, 5, 1, 3}, // det=1
		{1, 2, 3, 5}, // det=-1
		{2, 3, 5, 7}, // det=-1
		{7, 4, 5, 3}, // det=1
		{6, 5, 5, 4}, // det=-1
		{3, 5, 2, 3}, // det=-1
		{4, 7, 1, 2}, // det=1
		{7, 5, 4, 3}, // det=1
	}
	c := candidates[rand.Intn(len(candidates))]
	a, b, cVal, d := c[0], c[1], c[2], c[3]
	det := a*d - b*cVal
	if det < 0 {
		det = -det
	}
	inv := make([][]int, 2)
	for i := range inv {
		inv[i] = make([]int, 2)
	}
	inv[0][0] = d / det
	inv[0][1] = -b / det
	inv[1][0] = -cVal / det
	inv[1][1] = a / det
	return generator.Problem{
		Question:    fmt.Sprintf("Find the inverse of the matrix \\(%s\\).", formatMatrix2([][]int{{a, b}, {cVal, d}})),
		Answer:      formatMatrix(inv),
		Explanation: fmt.Sprintf("det = %d*%d - %d*%d = %d, inverse = (1/%d)×[[%d,%d],[%d,%d]] = %s", a, d, b, cVal, a*d-b*cVal, det, d, -b, -cVal, a, formatMatrix(inv)),
	}
}

type det2x2Gen struct{}

func (g *det2x2Gen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	a, b := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	c, d := rand.Intn(max(1, scale*2))-scale, rand.Intn(max(1, scale*2))-scale
	det := a*d - b*c
	return generator.Problem{
		Question:    fmt.Sprintf("Find the determinant of \\(%s\\).", formatMatrix2([][]int{{a, b}, {c, d}})),
		Answer:      fmt.Sprintf("%d", det),
		Explanation: fmt.Sprintf("det = %d*%d - %d*%d = %d-%d = %d", a, d, b, c, a*d, b*c, det),
	}
}

type det3x3Gen struct{}

func (g *det3x3Gen) Generate(ctx generator.GeneratorContext) generator.Problem {
	scale := int(1 + ctx.Difficulty*5)
	m := make([][]int, 3)
	for i := range m {
		m[i] = make([]int, 3)
		for j := range m[i] {
			m[i][j] = rand.Intn(max(1, scale*2)) - scale
		}
	}
	a, b, c := m[0][0], m[0][1], m[0][2]
	d, e, f := m[1][0], m[1][1], m[1][2]
	gVal, h, i := m[2][0], m[2][1], m[2][2]
	det := a*(e*i-f*h) - b*(d*i-f*gVal) + c*(d*h-e*gVal)
	return generator.Problem{
		Question:    fmt.Sprintf("Find the determinant of \\(%s\\).", formatMatrix(m)),
		Answer:      fmt.Sprintf("%d", det),
		Explanation: fmt.Sprintf("det = %d*(%d*%d-%d*%d) - %d*(%d*%d-%d*%d) + %d*(%d*%d-%d*%d) = %d", a, e, i, f, h, b, d, i, f, gVal, c, d, h, e, gVal, det),
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
	det := a*d - b*c
	if det == 0 {
		det = 1
	}
	askX := rand.Intn(2) == 0
	if askX {
		detX := e1*d - b*e2
		valX := detX / det
		return generator.Problem{
			Question:    fmt.Sprintf("Use Cramer's rule to find \\(x\\): \\(%dx + %dy = %d\\), \\(%dx + %dy = %d\\)", a, b, e1, c, d, e2),
			Answer:      fmt.Sprintf("%d", valX),
			Explanation: fmt.Sprintf("x = det([[%d,%d],[%d,%d]]) / det([[%d,%d],[%d,%d]]) = (%d*%d-%d*%d)/(%d*%d-%d*%d) = %d/%d = %d", e1, b, e2, d, a, b, c, d, e1, d, b, e2, a, d, b, c, detX, det, valX),
		}
	}
	detY := a*e2 - e1*c
	valY := detY / det
	return generator.Problem{
		Question:    fmt.Sprintf("Use Cramer's rule to find \\(y\\): \\(%dx + %dy = %d\\), \\(%dx + %dy = %d\\)", a, b, e1, c, d, e2),
		Answer:      fmt.Sprintf("%d", valY),
		Explanation: fmt.Sprintf("y = det([[%d,%d],[%d,%d]]) / det([[%d,%d],[%d,%d]]) = (%d*%d-%d*%d)/(%d*%d-%d*%d) = %d/%d = %d", a, e1, c, e2, a, b, c, d, a, e2, e1, c, a, d, b, c, detY, det, valY),
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

// Diagonalization: find P and D such that A = PDP⁻¹ for a 2x2 matrix
type diagonalizationGen struct{}

func (g *diagonalizationGen) Grade(expected, userAnswer string) grader.Result {
	// Expect "P=[[a,b],[c,d]] D=[[λ1,0],[0,λ2]]"
	// Allow any order of eigenvalues in D
	return gradeMatrixPair(expected, userAnswer)
}

func (g *diagonalizationGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	// Diagonalizable matrices (real distinct eigenvalues)
	type entry struct {
		a, b, c, d int
		l1, l2     int
	}
	entries := []entry{
		{2, 0, 0, 3, 2, 3},  // diagonal already
		{3, 1, 0, 2, 3, 2},  // triangular
		{1, 2, 2, 1, 3, -1}, // need eigenvectors
		{5, 0, 0, 4, 5, 4},
		{2, 1, 0, 1, 2, 1},
		{3, 0, 0, 5, 3, 5},
		{2, 2, 0, 4, 2, 4},
	}
	e := entries[rand.Intn(len(entries))]
	answer := fmt.Sprintf("P=[[%d,%d],[%d,%d]] D=[[%d,0],[0,%d]]", e.a, e.b, e.c, e.d, e.l1, e.l2)
	return generator.Problem{
		Question:    fmt.Sprintf("Find matrices \\(P\\) and \\(D\\) (diagonal) such that \\(A = PDP^{-1}\\) for \\(A = %s\\).", formatMatrix2([][]int{{e.a, e.b}, {e.c, e.d}})),
		Answer:      answer,
		Explanation: fmt.Sprintf("Eigenvalues are λ=%d and λ=%d. The eigenvectors form the columns of P: %s", e.l1, e.l2, answer),
	}
}

func gradeMatrixPair(expected, userAnswer string) grader.Result {
	// Parse "P=[[...]] D=[[...]]"
	parsePair := func(s string) (pStr, dStr string, ok bool) {
		parts := strings.SplitN(s, " D=", 2)
		if len(parts) != 2 {
			return "", "", false
		}
		pPart := strings.TrimPrefix(parts[0], "P=")
		return pPart, parts[1], true
	}
	eP, eD, ok := parsePair(expected)
	if !ok {
		return grader.Result{Correct: false, Score: 0, Feedback: "Expected format: P=[[...]] D=[[...]]"}
	}
	aP, aD, ok2 := parsePair(userAnswer)
	if !ok2 {
		return grader.Result{Correct: false, Score: 0, Feedback: "Expected format: P=[[...]] D=[[...]]"}
	}
	// Compare matrices — check both orderings in case eigenvalues are swapped
	norm := func(s string) string {
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\n", "")
		return s
	}
	if norm(expected) == norm(userAnswer) {
		return grader.Result{Correct: true, Score: 1}
	}
	// Check swapped: P1 D1 <-> P2 D2
	if norm(eP) == norm(aP) && norm(eD) == norm(aD) {
		return grader.Result{Correct: true, Score: 1}
	}
	if norm(eP) == norm(aD) && norm(eD) == norm(aP) {
		return grader.Result{Correct: true, Score: 1}
	}
	return grader.Result{Correct: false, Score: 0, Feedback: "Incorrect P or D"}
}

// Rank of a matrix
type rankGen struct{}

func (g *rankGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		matrix [][]int
		rank   int
	}
	entries := []entry{
		{[][]int{{1, 0}, {0, 1}}, 2},
		{[][]int{{1, 2}, {2, 4}}, 1},
		{[][]int{{1, 0, 0}, {0, 0, 0}}, 1},
		{[][]int{{1, 2, 3}, {0, 0, 0}}, 1},
		{[][]int{{1, 0}, {0, 0}}, 1},
		{[][]int{{1, 2}, {3, 4}}, 2},
		{[][]int{{1, 0, 0}, {0, 1, 0}}, 2},
		{[][]int{{0, 0}, {0, 0}}, 0},
		{[][]int{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}, 3},
		{[][]int{{1, 1, 1}, {2, 2, 2}, {3, 3, 3}}, 1},
	}
	e := entries[rand.Intn(len(entries))]
	return generator.Problem{
		Question:    fmt.Sprintf("What is the rank of matrix \\(%s\\)?", formatMatrix(e.matrix)),
		Answer:      fmt.Sprintf("%d", e.rank),
		Explanation: fmt.Sprintf("The rank is %d (the number of linearly independent rows/columns).", e.rank),
	}
}

// Cosine similarity between two vectors
type cosineSimilarityGen struct{}

func (g *cosineSimilarityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type entry struct {
		u, v [2]int
		cos  int
	}
	entries := []entry{
		{[2]int{1, 0}, [2]int{1, 0}, 1},
		{[2]int{1, 0}, [2]int{0, 1}, 0},
		{[2]int{1, 1}, [2]int{1, 1}, 1},
		{[2]int{1, 0}, [2]int{-1, 0}, -1},
		{[2]int{3, 4}, [2]int{6, 8}, 1},
		{[2]int{1, 1}, [2]int{1, -1}, 0},
		{[2]int{1, 2}, [2]int{2, 4}, 1},
		{[2]int{1, 0}, [2]int{0, -1}, 0},
		{[2]int{1, 1}, [2]int{-1, 1}, 0},
	}
	e := entries[rand.Intn(len(entries))]
	uStr := fmt.Sprintf("(%d,%d)", e.u[0], e.u[1])
	vStr := fmt.Sprintf("(%d,%d)", e.v[0], e.v[1])
	dot := e.u[0]*e.v[0] + e.u[1]*e.v[1]
	uMag := int(math.Sqrt(float64(e.u[0]*e.u[0] + e.u[1]*e.u[1])))
	vMag := int(math.Sqrt(float64(e.v[0]*e.v[0] + e.v[1]*e.v[1])))
	return generator.Problem{
		Question:    fmt.Sprintf("Compute the cosine similarity between \\(u=%s\\) and \\(v=%s\\).", uStr, vStr),
		Answer:      fmt.Sprintf("%d", e.cos),
		Explanation: fmt.Sprintf("cos(θ) = (u·v)/(|u||v|) = (%d)/(%d×%d) = %d/%d = %d", dot, uMag, vMag, dot, uMag*vMag, e.cos),
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
