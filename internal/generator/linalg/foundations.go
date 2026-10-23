package linalg

// The linear-algebra concepts the domain was missing: independence, the rank-nullity theorem, the
// inverse of a two-by-two matrix, the two eigenvalue invariants, projection onto a vector, the norm
// and the determinant of a rotation-scaling.
//
// The domain had vectors, matrices, determinants, eigenvalues, systems and decompositions, but
// nothing on when a set of vectors is independent, nothing on dimension or nullity, no inverse and
// no orthogonality at all -- which is most of what a first course is about.
//
// Every answer is an integer by construction: the vectors are built from a known rank, the matrices
// have determinant one or a diagonal form and the projections use a coefficient that divides.
import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mathutil"
)

func registerFoundations(reg *generator.Registry) {
	reg.Register("linalg.independence.count", &independenceCountGen{})
	reg.Register("linalg.rank.nullity", &rankNullityGen{})
	reg.Register("linalg.matrix.inverse_2x2", &inverse2x2Gen{})
	reg.Register("linalg.eigen.trace", &eigenTraceGen{})
	reg.Register("linalg.eigen.det_product", &eigenDetProductGen{})
	reg.Register("linalg.orth.projection_coeff", &projectionCoeffGen{})
	reg.Register("linalg.vec.norm_squared", &normSquaredGen{})
	reg.Register("linalg.transform.det_scale", &detScaleGen{})
}

func smallInt(lo, hi int) int {
	if hi < lo {
		hi = lo
	}
	return lo + rand.Intn(hi-lo+1)
}

// independenceCountGen builds three vectors with a known rank, so the count of independent ones is exact.
type independenceCountGen struct{}

func (g *independenceCountGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	rank := 1 + rand.Intn(3)
	a := 4 + rand.Intn(4+tier*3)
	b := a + 1 + rand.Intn(3+tier*2)
	c := b + 1 + rand.Intn(3+tier*2)
	var vs [3][3]int
	switch rank {
	case 1:
		vs = [3][3]int{{a, b, c}, {2 * a, 2 * b, 2 * c}, {3 * a, 3 * b, 3 * c}}
	case 2:
		vs = [3][3]int{{a, b, c}, {c, a, b}, {a + c, b + a, c + b}}
	default:
		d := 4 + rand.Intn(4+tier*3)
		e := 4 + rand.Intn(4+tier*3)
		f := 4 + rand.Intn(4+tier*3)
		vs = [3][3]int{{a, b, c}, {0, d, e}, {0, 0, f}}
	}
	answer := strconv.Itoa(rank)

	subgoal1 := "Subgoal 1: State governing formula. The number of linearly independent vectors in a set equals the dimension of their span, which is the rank of the matrix formed by placing the vectors as rows."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. Placing vectors v_1 = (%d,%d,%d), v_2 = (%d,%d,%d) and v_3 = (%d,%d,%d) into rows, row echelon reduction reveals %d non-zero pivot rows.",
		vs[0][0], vs[0][1], vs[0][2], vs[1][0], vs[1][1], vs[1][2], vs[2][0], vs[2][1], vs[2][2], rank)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. The row rank is %d, so exactly %s of the vectors are linearly independent.",
		rank, answer)

	return generator.Problem{
		Question: fmt.Sprintf("How many of the vectors \\((%d,%d,%d)\\), \\((%d,%d,%d)\\) and \\((%d,%d,%d)\\) are linearly independent?",
			vs[0][0], vs[0][1], vs[0][2], vs[1][0], vs[1][1], vs[1][2], vs[2][0], vs[2][1], vs[2][2]),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"answer": answer,
		},
	}
}

// rankNullityGen is the rank-nullity theorem read as an arithmetic fact.
type rankNullityGen struct{}

func (g *rankNullityGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var cols, rows, rank, nullity int
	var answer string

	for iter := 0; iter < 20; iter++ {
		cols = 4 + rand.Intn(4+tier*3)
		rows = 3 + rand.Intn(4+tier*3)
		rank = 1 + rand.Intn(cols-1)
		nullity = cols - rank
		answer = strconv.Itoa(nullity)

		// Read-off avoidance: answer must not equal rows, cols or rank
		if nullity != rows && nullity != cols && nullity != rank {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The rank-nullity theorem states that for any m x n matrix, the rank plus the nullity (dimension of the null space) equals the number of columns: rank(A) + nullity(A) = n."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For a matrix with %d columns and rank %d, the relation gives %d + nullity(A) = %d.",
		cols, rank, rank, cols)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Subtracting gives nullity(A) = %d - %d = %s.",
		cols, rank, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A matrix has %d rows and %d columns with rank %d. What is the dimension of its null space?", rows, cols, rank),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"rows": strconv.Itoa(rows), "cols": strconv.Itoa(cols), "rank": strconv.Itoa(rank),
			"answer": answer,
		},
	}
}

// inverse2x2Gen uses matrices with determinant 2, whose inverses have integer entries.
type inverse2x2Gen struct{}

func (g *inverse2x2Gen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, c, d, k, det int
	var answer string

	for iter := 0; iter < 20; iter++ {
		det = 2
		k = 3 + rand.Intn(4+tier*3)
		a = 3 + rand.Intn(4+tier*3)
		if a == k {
			a++
		}
		d = det * k
		b = det
		c = a*k - 1
		answer = strconv.Itoa(k)

		// Read-off avoidance: answer must not match a, b, c, d or det
		if k != a && k != b && k != c && k != d && k != det {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The inverse of an invertible 2x2 matrix A = [[a, b], [c, d]] is given by A^{-1} = (1 / det(A)) [[d, -b], [-c, a]]."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For matrix entries a = %d, b = %d, c = %d, d = %d and determinant det(A) = %d, the top-left entry of the adjugate matrix is d = %d.",
		a, b, c, d, det, d)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing by the determinant gives entry (A^{-1})_{11} = %d / %d = %s.",
		d, det, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("The matrix \\(A = \\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\) has determinant %d, so it is invertible. What is the entry in the first row and first column of \\(A^{-1}\\)?", a, b, c, d, det),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b), "c": strconv.Itoa(c), "d": strconv.Itoa(d),
			"determinant": strconv.Itoa(det), "answer": answer,
		},
	}
}

// eigenTraceGen uses the trace as the sum of the eigenvalues.
type eigenTraceGen struct{}

func (g *eigenTraceGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var l1, l2, trace int
	var answer string

	for iter := 0; iter < 20; iter++ {
		l1 = smallInt(2, 4+tier*3)
		if rand.Float64() < 0.3 && tier > 0 {
			l1 = -l1
		}
		l2 = smallInt(2, 4+tier*3)
		if rand.Float64() < 0.3 && tier > 0 {
			l2 = -l2
		}
		if l2 == l1 {
			l2 += 3
		}
		trace = l1 + l2
		answer = strconv.Itoa(l2)

		absL2 := l2
		if absL2 < 0 {
			absL2 = -absL2
		}
		absL1 := l1
		if absL1 < 0 {
			absL1 = -absL1
		}
		absTrace := trace
		if absTrace < 0 {
			absTrace = -absTrace
		}

		// Read-off avoidance: answer must not match l1 or trace
		if absL2 != absL1 && absL2 != absTrace && absL1 != absTrace && trace != 0 {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The trace of a square matrix equals the sum of its eigenvalues: tr(A) = \\sum \\lambda_i = \\lambda_1 + \\lambda_2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. Given trace tr(A) = %d and first eigenvalue \\lambda_1 = %d, the relation is %d + \\lambda_2 = %d.",
		trace, l1, l1, trace)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Subtracting yields the other eigenvalue \\lambda_2 = %d - %d = %s.",
		trace, l1, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A two-by-two matrix has trace %d and one of its eigenvalues is %d. What is the other eigenvalue?", trace, l1),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"trace": strconv.Itoa(trace), "l1": strconv.Itoa(l1),
			"answer": answer,
		},
	}
}

// eigenDetProductGen uses the determinant as the product of the eigenvalues.
type eigenDetProductGen struct{}

func (g *eigenDetProductGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var l1, l2, det int
	var answer string

	for iter := 0; iter < 20; iter++ {
		l1 = 2 + rand.Intn(4+tier*3)
		l2 = 2 + rand.Intn(4+tier*3)
		if rand.Float64() < 0.3 && tier > 0 {
			l2 = -l2
		}
		if l2 == l1 {
			l2 += 3
		}
		det = l1 * l2
		answer = strconv.Itoa(l2)

		absL2 := l2
		if absL2 < 0 {
			absL2 = -absL2
		}
		absL1 := l1
		if absL1 < 0 {
			absL1 = -absL1
		}
		absDet := det
		if absDet < 0 {
			absDet = -absDet
		}

		// Read-off avoidance: answer must not match l1 or det
		if absL2 != absL1 && absL2 != absDet && absL1 != absDet {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. The determinant of a square matrix equals the product of its eigenvalues: det(A) = \\prod \\lambda_i = \\lambda_1 \\lambda_2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. Given determinant det(A) = %d and known eigenvalue \\lambda_1 = %d, the equation is (%d) * \\lambda_2 = %d.",
		det, l1, l1, det)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing gives other eigenvalue \\lambda_2 = %d / %d = %s.",
		det, l1, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("A two-by-two matrix has determinant %d and one of its eigenvalues is %d. What is the other eigenvalue?", det, l1),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"determinant": strconv.Itoa(det), "l1": strconv.Itoa(l1),
			"answer": answer,
		},
	}
}

// projectionCoeffGen asks for the scalar whose multiple of b is the projection of a onto b.
type projectionCoeffGen struct{}

func (g *projectionCoeffGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var b1, b2, a1, a2, c, bb int
	var answer string

	for iter := 0; iter < 20; iter++ {
		b1 = 2 + rand.Intn(3+tier*2)
		b2 = 2 + rand.Intn(3+tier*2)
		c = 3 + rand.Intn(4+tier*3)
		if rand.Float64() < 0.3 && tier > 0 {
			c = -c
		}
		bb = b1*b1 + b2*b2
		// a = c*b plus a component orthogonal to b, chosen so the coefficient is exactly c
		a1 = c*b1 - b2
		a2 = c*b2 + b1
		answer = strconv.Itoa(c)

		absC := c
		if absC < 0 {
			absC = -absC
		}
		// Read-off avoidance: answer must not match a1, a2, b1 or b2
		collision := false
		for _, v := range []int{a1, a2, b1, b2} {
			absV := v
			if absV < 0 {
				absV = -absV
			}
			if absC == absV {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	dot := a1*b1 + a2*b2
	subgoal1 := "Subgoal 1: State governing formula. The orthogonal projection coefficient c of vector a onto non-zero vector b is given by c = (a \\cdot b) / (b \\cdot b)."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For a = (%d,%d) and b = (%d,%d), the dot product is a \\cdot b = (%d)(%d) + (%d)(%d) = %d and squared norm is b \\cdot b = %d^2 + %d^2 = %d.",
		a1, a2, b1, b2, a1, b1, a2, b2, dot, b1, b2, bb)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Dividing yields scalar c = %d / %d = %s.",
		dot, bb, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("Find the scalar \\(c\\) such that \\(c(%d,%d)\\) is the projection of \\((%d,%d)\\) onto \\((%d,%d)\\).", b1, b2, a1, a2, b1, b2),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"b1": strconv.Itoa(b1), "b2": strconv.Itoa(b2),
			"a1": strconv.Itoa(a1), "a2": strconv.Itoa(a2),
			"answer": answer,
		},
	}
}

// normSquaredGen is the dot product of a vector with itself, which avoids a square root.
type normSquaredGen struct{}

func (g *normSquaredGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	n := 2
	if tier == 2 {
		n = 3
	}
	var v []int
	var sum int
	var answer string

	for iter := 0; iter < 20; iter++ {
		v = make([]int, n)
		sum = 0
		for i := range v {
			v[i] = 2 + rand.Intn(4+tier*3)
			if rand.Float64() < 0.3 && tier > 0 {
				v[i] = -v[i]
			}
			sum += v[i] * v[i]
		}
		answer = strconv.Itoa(sum)

		// Read-off avoidance: answer must not match any |v[i]|
		collision := false
		for _, x := range v {
			absX := x
			if absX < 0 {
				absX = -absX
			}
			if answer == strconv.Itoa(absX) {
				collision = true
				break
			}
		}
		if !collision {
			break
		}
	}

	vStrs := make([]string, len(v))
	sqStrs := make([]string, len(v))
	facts := map[string]string{
		"answer": answer,
	}
	for i := range v {
		vStrs[i] = strconv.Itoa(v[i])
		sqStrs[i] = fmt.Sprintf("(%d)^2", v[i])
		facts[fmt.Sprintf("v%d", i+1)] = strconv.Itoa(v[i])
	}

	subgoal1 := "Subgoal 1: State governing formula. The squared Euclidean norm of a vector v is equal to the sum of its squared components: \\|v\\|^2 = \\sum v_i^2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For vector v = (%s), substituting components gives %s.",
		strings.Join(vStrs, ", "), strings.Join(sqStrs, " + "))
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the sum of squares yields \\|v\\|^2 = %s.",
		answer)

	qStr := fmt.Sprintf("Find the squared length \\(\\|v\\|^2\\) of the vector \\(v = (%s)\\).", strings.Join(vStrs, ", "))

	return generator.Problem{
		Question:    qStr,
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts:       facts,
	}
}

// detScaleGen uses a rotation-scaling matrix, whose determinant is the squared scale plus the squared rotation.
type detScaleGen struct{}

func (g *detScaleGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	tier := mathutil.DifficultyTier(ctx.Difficulty)
	var a, b, det int
	var answer string

	for iter := 0; iter < 20; iter++ {
		a = 2 + rand.Intn(4+tier*3)
		b = 2 + rand.Intn(4+tier*3)
		if rand.Float64() < 0.3 && tier > 0 {
			b = -b
		}
		det = a*a + b*b
		answer = strconv.Itoa(det)

		absA := a
		if absA < 0 {
			absA = -absA
		}
		absB := b
		if absB < 0 {
			absB = -absB
		}
		// Read-off avoidance: answer must not match |a| or |b|
		if det != absA && det != absB {
			break
		}
	}

	subgoal1 := "Subgoal 1: State governing formula. For a 2x2 rotation-scaling matrix A = [[a, -b], [b, a]], the determinant is det(A) = (a)(a) - (-b)(b) = a^2 + b^2."
	subgoal2 := fmt.Sprintf("Subgoal 2: Substitute parameters. For matrix entries a = %d and b = %d, substituting into the determinant gives (%d)^2 + (%d)^2.",
		a, b, a, b)
	subgoal3 := fmt.Sprintf("Subgoal 3: Evaluate and simplify. Evaluating the sum of squares yields det(A) = %d + %d = %s.",
		a*a, b*b, answer)

	return generator.Problem{
		Question:    fmt.Sprintf("Find the determinant of \\(\\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}\\).", a, -b, b, a),
		Answer:      answer,
		Explanation: strings.Join([]string{subgoal1, subgoal2, subgoal3}, " "),
		Facts: map[string]string{
			"a": strconv.Itoa(a), "b": strconv.Itoa(b),
			"answer": answer,
		},
	}
}
