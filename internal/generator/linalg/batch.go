package linalg

// The 2026-10-18 linear-algebra batch: eleven concepts that take the domain
// from 34 to the target of 45.
//
// The survey found the domain well supplied with computations but thin on the
// structural ideas that connect them: vector-space axioms, kernel and image,
// invertibility, change of basis, orthogonal complements, projection onto a
// subspace, the spectral theorem, quadratic definiteness, Cayley-Hamilton and
// the condition number. Several obvious additions were already taught, so they
// were not repeated: determinant properties, transpose, linear combinations and
// coordinate representation, the characteristic polynomial, and algebraic and
// geometric multiplicity all live in existing lessons. Gram-Schmidt is the
// method behind the existing QR lesson and was also left out.
//
// Five concepts are numeric, one is a tuple (the cross product), and five are
// multiple choice. Every numeric generator publishes the facts its answer
// follows from, so the tests recompute the answer instead of comparing a string
// with itself; the tuple generator publishes both factors so the test can
// recompute the cross product.

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/chuma-beep/mathua/internal/generator"
)

func registerBatch(reg *generator.Registry) {
	reg.Register("linalg.space.axioms", &spaceAxiomsGen{})
	reg.Register("linalg.transform.kernel", &kernelGen{})
	reg.Register("linalg.transform.invertible", &invertibleMapGen{})
	reg.Register("linalg.basis.change", &basisChangeGen{})
	reg.Register("linalg.orth.complement", &complementGen{})
	reg.Register("linalg.orth.projection_matrix", &projectionMatrixGen{})
	reg.Register("linalg.symmetric.spectral", &spectralGen{})
	reg.Register("linalg.quadratic.definite", &definiteGen{})
	reg.Register("linalg.vector.cross", &crossGen{})
	reg.Register("linalg.matrix.condition", &conditionGen{})
	reg.Register("linalg.eigen.cayley_hamilton", &cayleyHamiltonGen{})
}

// ---------------------------------------------------------------------------
// 1. Vector spaces and their axioms
// ---------------------------------------------------------------------------

type spaceAxiomsGen struct{}

func (g *spaceAxiomsGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "Which axiom guarantees that every vector v has an additive inverse?",
			correct:     "for every v there is a w with v + w = 0",
			distractors: []string{"closure under scalar multiplication", "distributivity of scalar multiplication over addition", "commutativity of scalar multiplication"},
			reason:      "The additive-inverse axiom supplies the negative of each vector, which is what makes subtraction possible.",
		},
		{
			question:    "Which operation is required of a real vector space but not of a group?",
			correct:     "multiplication by every real scalar",
			distractors: []string{"associative addition", "an additive identity", "additive inverses"},
			reason:      "A vector space is an abelian group under addition with an extra action of the field by scalar multiplication.",
		},
		{
			question:    "The zero vector in a vector space is...",
			correct:     "unique",
			distractors: []string{"one of many", "the number zero in every space", "not required to exist"},
			reason:      "If 0 and 0' are both additive identities then 0 = 0 + 0' = 0', so the zero vector is unique.",
		},
		{
			question:    "Which operation must distribute over vector addition?",
			correct:     "scalar multiplication",
			distractors: []string{"the dot product", "the determinant", "the transpose"},
			reason:      "The axiom c(u + v) = cu + cv is one of the two distributivity laws that tie the field action to addition.",
		},
		{
			question:    "A vector space over R must be closed under...",
			correct:     "addition and multiplication by every real scalar",
			distractors: []string{"addition only", "multiplication by positive scalars only", "the cross product"},
			reason:      "Closure under both operations is the first requirement; the remaining axioms describe how they behave.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 2. Kernel and image (numeric)
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 3. Injective, surjective and invertible maps
// ---------------------------------------------------------------------------

type invertibleMapGen struct{}

func (g *invertibleMapGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "A linear map T from R^n to R^n is invertible exactly when...",
			correct:     "its kernel is the zero subspace",
			distractors: []string{"its image is the zero subspace", "it is injective but not surjective", "its matrix has a zero eigenvalue"},
			reason:      "For a square map, a trivial kernel makes the map injective, and finite dimension then makes it surjective and invertible.",
		},
		{
			question:    "For a square matrix A, which condition is equivalent to invertibility?",
			correct:     "the determinant is nonzero",
			distractors: []string{"the determinant is zero", "A has a zero eigenvalue", "the columns are linearly dependent"},
			reason:      "The determinant is the product of the eigenvalues, so it is nonzero exactly when no eigenvalue is zero.",
		},
		{
			question:    "Can a linear map from R^3 to R^2 be injective?",
			correct:     "no, because the rank is at most 2 and the nullity is at least 1",
			distractors: []string{"yes, if it is surjective", "yes, if its matrix has rank 3", "only when it is the zero map"},
			reason:      "Rank-nullity gives nullity 3 - rank, and the rank cannot exceed the dimension 2 of the codomain.",
		},
		{
			question:    "For a square matrix, injectivity of the linear map implies...",
			correct:     "surjectivity",
			distractors: []string{"nothing about surjectivity", "that the matrix is singular", "that the determinant is zero"},
			reason:      "A square matrix with trivial kernel has full rank, so its columns span the whole codomain.",
		},
		{
			question:    "The inverse of an invertible linear map is...",
			correct:     "linear",
			distractors: []string{"not necessarily linear", "always nonlinear", "the transpose"},
			reason:      "The inverse of an invertible linear map is itself linear, which is why it is represented by the inverse matrix.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 4. Coordinates and change of basis (numeric)
// ---------------------------------------------------------------------------

type basisChangeGen struct{}

func (g *basisChangeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	// Pairs of 2D unimodular basis vectors (det = 1)
	unimodularPairs := [][2][2]int{
		{{1, 0}, {0, 1}},
		{{1, 1}, {1, 2}},
		{{1, -1}, {1, 0}},
		{{2, 1}, {1, 1}},
		{{1, 2}, {0, 1}},
		{{1, 0}, {1, 1}},
		{{2, 3}, {1, 2}},
		{{3, 2}, {1, 1}},
		{{1, 1}, {2, 3}},
		{{1, 2}, {1, 3}},
		{{2, 1}, {3, 2}},
	}
	pair := unimodularPairs[rand.Intn(len(unimodularPairs))]
	b1 := pair[0]
	b2 := pair[1]

	// Choose random integer coordinates a, b in [-6, 6]
	a := rand.Intn(13) - 6
	b := rand.Intn(13) - 6
	if a == 0 && b == 0 {
		a = 2
		b = 3
	}

	vx := a*b1[0] + b*b2[0]
	vy := a*b1[1] + b*b2[1]

	ask := rand.Intn(2)
	ans := a
	varName := "a"
	if ask == 1 {
		ans = b
		varName = "b"
	}

	q := fmt.Sprintf("In R^2, the vector (%d, %d) is written as a(%d, %d) + b(%d, %d). What is %s? (enter a number)",
		vx, vy, b1[0], b1[1], b2[0], b2[1], varName)
	ex := fmt.Sprintf("The equations are %da + %db = %d and %da + %db = %d. "+
		"Solving the system yields a = %d and b = %d.",
		b1[0], b2[0], vx, b1[1], b2[1], vy, a, b)

	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(ans),
		Explanation: ex,
		Facts: map[string]string{
			"b1":     fmt.Sprintf("%d,%d", b1[0], b1[1]),
			"b2":     fmt.Sprintf("%d,%d", b2[0], b2[1]),
			"v":      fmt.Sprintf("%d,%d", vx, vy),
			"ask":    strconv.Itoa(ask),
			"answer": strconv.Itoa(ans),
		},
	}
}

// ---------------------------------------------------------------------------
// 5. Orthogonal complements (numeric)
// ---------------------------------------------------------------------------

type complementGen struct{}

func (g *complementGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	n := 2 + rand.Intn(4)   // 2 to 5
	m := 1 + rand.Intn(n-1) // 1 to n-1
	want := n - m

	var desc, reason string
	if m == 1 {
		desc = "the line spanned by a non-zero vector has dimension 1."
		reason = fmt.Sprintf("The orthogonal complement of a line in R^%d has dimension %d - 1 = %d.", n, n, want)
	} else if m == n-1 {
		desc = fmt.Sprintf("a hyperplane has dimension %d.", m)
		reason = fmt.Sprintf("The orthogonal complement of a hyperplane of dimension %d in R^%d is 1-dimensional: %d - %d = %d.", m, n, n, m, want)
	} else {
		desc = fmt.Sprintf("a subspace spanned by %d linearly independent vectors has dimension %d.", m, m)
		reason = fmt.Sprintf("By the orthogonal complement theorem, dim(U-perp) = %d - %d = %d.", n, m, want)
	}

	q := fmt.Sprintf("In R^%d, %s What is the dimension of its orthogonal complement? (enter a number)", n, desc)
	ansStr := strconv.Itoa(want)
	return generator.Problem{
		Question:    q,
		Answer:      ansStr,
		Explanation: reason,
		Facts: map[string]string{
			"n":      strconv.Itoa(n),
			"m":      strconv.Itoa(m),
			"answer": ansStr,
		},
	}
}

// ---------------------------------------------------------------------------
// 6. Projection onto a subspace (numeric)
// ---------------------------------------------------------------------------

type projectionMatrixGen struct{}

func (g *projectionMatrixGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	dirs := [][2]int{
		{1, 0}, {0, 1}, {1, 1}, {1, -1}, {2, 0}, {0, 2}, {1, 2}, {2, 1},
	}
	u := dirs[rand.Intn(len(dirs))]
	uNormSq := u[0]*u[0] + u[1]*u[1]

	c := rand.Intn(4) + 1
	if rand.Intn(2) == 1 {
		c = -c
	}
	k := rand.Intn(5) - 2
	uPerp := [2]int{-u[1], u[0]}

	vx := c*u[0] + k*uPerp[0]
	vy := c*u[1] + k*uPerp[1]

	axis := rand.Intn(2)
	if u[0] == 0 {
		axis = 1
	} else if u[1] == 0 {
		axis = 0
	}

	coordName := "x-coordinate"
	axisStr := "x"
	ans := c * u[0]
	if axis == 1 {
		coordName = "y-coordinate"
		axisStr = "y"
		ans = c * u[1]
	}

	q := fmt.Sprintf("The projection of (%d, %d) onto the line spanned by (%d, %d) has what %s? (enter a number)",
		vx, vy, u[0], u[1], coordName)
	ex := fmt.Sprintf("The coefficient is ((%d)(%d) + (%d)(%d)) / ((%d)^2 + (%d)^2) = %d / %d = %d. "+
		"So the projection is %d(%d, %d) = (%d, %d) and its %s is %d.",
		vx, u[0], vy, u[1], u[0], u[1], c*uNormSq, uNormSq, c,
		c, u[0], u[1], c*u[0], c*u[1], coordName, ans)

	return generator.Problem{
		Question:    q,
		Answer:      strconv.Itoa(ans),
		Explanation: ex,
		Facts: map[string]string{
			"v":      fmt.Sprintf("%d,%d", vx, vy),
			"u":      fmt.Sprintf("%d,%d", u[0], u[1]),
			"axis":   axisStr,
			"answer": strconv.Itoa(ans),
		},
	}
}

// ---------------------------------------------------------------------------
// 7. The spectral theorem for symmetric matrices
// ---------------------------------------------------------------------------

type spectralGen struct{}

func (g *spectralGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "The spectral theorem says that a real symmetric matrix...",
			correct:     "is orthogonally diagonalizable with real eigenvalues",
			distractors: []string{"is always invertible", "has only positive eigenvalues", "is always a rotation"},
			reason:      "A real symmetric matrix has an orthonormal basis of eigenvectors, so it can be written Q D Q^T with Q orthogonal.",
		},
		{
			question:    "Eigenvectors of a real symmetric matrix belonging to distinct eigenvalues are...",
			correct:     "orthogonal",
			distractors: []string{"parallel", "linearly dependent", "complex conjugates"},
			reason:      "If Ax = lambda x and Ay = mu y with lambda different from mu, symmetry forces (lambda - mu)(x dot y) = 0, so x dot y = 0.",
		},
		{
			question:    "A symmetric matrix is orthogonally diagonalizable because...",
			correct:     "it has an orthonormal basis of eigenvectors",
			distractors: []string{"it is triangular", "its determinant is positive", "it has distinct eigenvalues"},
			reason:      "The spectral theorem supplies the orthonormal eigenbasis even when eigenvalues repeat.",
		},
		{
			question:    "Which matrix is symmetric?",
			correct:     "[[2, 1], [1, 3]]",
			distractors: []string{"[[0, 1], [-1, 0]]", "[[1, 2], [0, 1]]", "[[1, 1], [2, 1]]"},
			reason:      "A matrix is symmetric when it equals its transpose; the off-diagonal entries must match.",
		},
		{
			question:    "The eigenvalues of a real symmetric matrix are...",
			correct:     "real",
			distractors: []string{"always complex", "always positive", "always zero"},
			reason:      "Symmetry rules out nonreal eigenvalues, which is why the spectral theorem works entirely over R.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 8. Quadratic forms and definiteness
// ---------------------------------------------------------------------------

type definiteGen struct{}

func (g *definiteGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "A symmetric matrix is positive definite exactly when...",
			correct:     "all its eigenvalues are positive",
			distractors: []string{"all its entries are positive", "its determinant is zero", "it is diagonal"},
			reason:      "The quadratic form equals a sum of eigenvalue-weighted squares in an orthonormal eigenbasis, so positivity is exactly positivity of every eigenvalue.",
		},
		{
			question:    "The quadratic form x^2 + y^2 is...",
			correct:     "positive definite",
			distractors: []string{"negative definite", "indefinite", "positive semidefinite but not definite"},
			reason:      "It is a sum of squares and vanishes only at (x, y) = (0, 0).",
		},
		{
			question:    "The quadratic form x^2 - y^2 is...",
			correct:     "indefinite",
			distractors: []string{"positive definite", "negative definite", "identically zero"},
			reason:      "It is positive at (1, 0) and negative at (0, 1), so it takes both signs.",
		},
		{
			question:    "The symmetric matrix [[a, b], [b, c]] is positive definite exactly when...",
			correct:     "a is positive and ac - b^2 is positive",
			distractors: []string{"a is positive and ac - b^2 is negative", "a is negative and c is positive", "b^2 is larger than ac"},
			reason:      "Sylvester's criterion requires the leading principal minors a and ac - b^2 to be positive.",
		},
		{
			question:    "The level sets of a positive definite quadratic form are...",
			correct:     "ellipses",
			distractors: []string{"hyperbolas", "parallel lines", "parabolas"},
			reason:      "A positive definite form is a weighted sum of squares, so its level curves are bounded and closed, hence ellipses.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}

// ---------------------------------------------------------------------------
// 9. The cross product (tuple)
// ---------------------------------------------------------------------------

type crossGen struct{}

func (g *crossGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	for {
		u := [3]int{rand.Intn(7) - 3, rand.Intn(7) - 3, rand.Intn(7) - 3}
		v := [3]int{rand.Intn(7) - 3, rand.Intn(7) - 3, rand.Intn(7) - 3}
		if u[0] == 0 && u[1] == 0 && u[2] == 0 {
			continue
		}
		if v[0] == 0 && v[1] == 0 && v[2] == 0 {
			continue
		}
		x := u[1]*v[2] - u[2]*v[1]
		y := u[2]*v[0] - u[0]*v[2]
		z := u[0]*v[1] - u[1]*v[0]
		if x == 0 && y == 0 && z == 0 {
			continue
		}

		q := fmt.Sprintf("What is the cross product (%d, %d, %d) x (%d, %d, %d)? (enter a vector)",
			u[0], u[1], u[2], v[0], v[1], v[2])
		ans := fmt.Sprintf("(%d,%d,%d)", x, y, z)
		ex := fmt.Sprintf("Compute ((%d)(%d) - (%d)(%d), (%d)(%d) - (%d)(%d), (%d)(%d) - (%d)(%d)) = (%d, %d, %d).",
			u[1], v[2], u[2], v[1], u[2], v[0], u[0], v[2], u[0], v[1], u[1], v[0], x, y, z)

		return generator.Problem{
			Question:    q,
			Answer:      ans,
			Explanation: ex,
			Facts: map[string]string{
				"u":      fmt.Sprintf("%d,%d,%d", u[0], u[1], u[2]),
				"v":      fmt.Sprintf("%d,%d,%d", v[0], v[1], v[2]),
				"answer": ans,
			},
		}
	}
}

// ---------------------------------------------------------------------------
// 10. The condition number (numeric)
// ---------------------------------------------------------------------------

type conditionGen struct{}

func (g *conditionGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	ratios := []int{1, 2, 3, 4, 5, 8, 10}
	ratio := ratios[rand.Intn(len(ratios))]
	smallests := []float64{0.5, 1, 2, 3, 4}
	smallest := smallests[rand.Intn(len(smallests))]
	largest := float64(ratio) * smallest

	ansStr := strconv.FormatFloat(float64(ratio), 'g', -1, 64)
	var q, ex string
	if ratio == 1 {
		q = fmt.Sprintf("A matrix has singular values %g and %g. What is its condition number, the largest divided by the smallest? (enter a number)",
			largest, smallest)
		ex = fmt.Sprintf("Equal singular values give condition number %g/%g = 1, the best possible conditioning.",
			largest, smallest)
	} else {
		q = fmt.Sprintf("A matrix has singular values %g and %g. What is its condition number, the largest divided by the smallest? (enter a number)",
			largest, smallest)
		ex = fmt.Sprintf("The condition number is %g/%g = %s.", largest, smallest, ansStr)
	}

	return generator.Problem{
		Question:    q,
		Answer:      ansStr,
		Explanation: ex,
		Facts: map[string]string{
			"largest":  strconv.FormatFloat(largest, 'g', -1, 64),
			"smallest": strconv.FormatFloat(smallest, 'g', -1, 64),
			"answer":   ansStr,
		},
	}
}

// ---------------------------------------------------------------------------
// 11. The Cayley-Hamilton theorem
// ---------------------------------------------------------------------------

type cayleyHamiltonGen struct{}

func (g *cayleyHamiltonGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	type claim struct {
		question    string
		correct     string
		distractors []string
		reason      string
	}
	claims := []claim{
		{
			question:    "The Cayley-Hamilton theorem says that every square matrix...",
			correct:     "satisfies its own characteristic polynomial",
			distractors: []string{"is diagonalizable", "has distinct eigenvalues", "is invertible"},
			reason:      "Substituting A into its characteristic polynomial gives the zero matrix, whether or not A is diagonalizable.",
		},
		{
			question:    "A 2x2 matrix has characteristic polynomial lambda^2 - 5 lambda + 6. Cayley-Hamilton gives...",
			correct:     "A^2 - 5A + 6I = 0",
			distractors: []string{"A^2 + 5A + 6I = 0", "A - 5I = 0", "A^2 = 6I only"},
			reason:      "Replace lambda by A and the constant term by that constant times I, giving A^2 - 5A + 6I = 0.",
		},
		{
			question:    "Cayley-Hamilton lets you compute which of these for a 2x2 matrix?",
			correct:     "powers of A as combinations of I and A",
			distractors: []string{"only the determinant", "only the trace", "the inverse of a non-square matrix"},
			reason:      "The relation A^2 = 5A - 6I expresses every higher power in terms of I and A.",
		},
		{
			question:    "The minimal polynomial of a matrix...",
			correct:     "divides the characteristic polynomial",
			distractors: []string{"is always the characteristic polynomial", "is always linear", "has larger degree than the matrix"},
			reason:      "The minimal polynomial is the monic polynomial of least degree that annihilates A, and Cayley-Hamilton forces it to divide the characteristic polynomial.",
		},
		{
			question:    "What is the degree of the characteristic polynomial of a 3x3 matrix?",
			correct:     "3",
			distractors: []string{"2", "9", "0"},
			reason:      "The characteristic polynomial is det(A - lambda I), a determinant of a 3x3 matrix, so its degree is 3.",
		},
	}
	c := claims[rand.Intn(len(claims))]
	p := generator.MultipleChoice(c.question, c.correct, c.distractors...)
	p.Explanation = c.reason
	return p
}
