package mathutil

import (
	"fmt"
	"math/rand"
)

// Matrix2x2 represents a 2x2 integer matrix:
// [ A  B ]
// [ C  D ]
type Matrix2x2 struct {
	A, B int
	C, D int
}

// Det computes the integer determinant of a 2x2 matrix.
func (m Matrix2x2) Det() int {
	return m.A*m.D - m.B*m.C
}

// Trace computes the trace (sum of diagonal entries).
func (m Matrix2x2) Trace() int {
	return m.A + m.D
}

// Multiply computes the matrix product m * other.
func (m Matrix2x2) Multiply(other Matrix2x2) Matrix2x2 {
	return Matrix2x2{
		A: m.A*other.A + m.B*other.C,
		B: m.A*other.B + m.B*other.D,
		C: m.C*other.A + m.D*other.C,
		D: m.C*other.B + m.D*other.D,
	}
}

// Inverse returns the integer inverse matrix when determinant is 1 or -1.
// If the matrix is not unimodular, it returns false as the second value.
func (m Matrix2x2) Inverse() (Matrix2x2, bool) {
	d := m.Det()
	if d != 1 && d != -1 {
		return Matrix2x2{}, false
	}
	if d == 1 {
		return Matrix2x2{
			A:  m.D,
			B: -m.B,
			C: -m.C,
			D:  m.A,
		}, true
	}
	return Matrix2x2{
		A: -m.D,
		B:  m.B,
		C:  m.C,
		D: -m.A,
	}, true
}

// RandomUnimodular2x2 generates a random 2x2 integer matrix with determinant 1 or -1.
// It constructs the matrix by composing elementary shear operations and reflections.
func RandomUnimodular2x2(rng *rand.Rand, maxShear int) Matrix2x2 {
	if maxShear < 1 {
		maxShear = 1
	}
	// Start with diagonal identity or reflection
	m := Matrix2x2{A: 1, D: 1}
	if rng.Intn(2) == 1 {
		m.A = -1
	}
	if rng.Intn(2) == 1 {
		m.D = -m.D
	}

	// Apply 2 to 4 elementary row and column shears
	shears := 2 + rng.Intn(3)
	for i := 0; i < shears; i++ {
		k := rng.Intn(2*maxShear+1) - maxShear
		if k == 0 {
			k = 1
		}
		switch rng.Intn(4) {
		case 0:
			// Row 1 += k * Row 2
			m = Matrix2x2{A: 1, B: k, C: 0, D: 1}.Multiply(m)
		case 1:
			// Row 2 += k * Row 1
			m = Matrix2x2{A: 1, B: 0, C: k, D: 1}.Multiply(m)
		case 2:
			// Col 1 += k * Col 2
			m = m.Multiply(Matrix2x2{A: 1, B: 0, C: k, D: 1})
		default:
			// Col 2 += k * Col 1
			m = m.Multiply(Matrix2x2{A: 1, B: k, C: 0, D: 1})
		}
	}
	return m
}

// IntegerEigenvalues2x2 constructs a 2x2 integer matrix whose exact eigenvalues are lambda1 and lambda2.
// It computes A = P * diag(lambda1, lambda2) * P^{-1} where P is a random unimodular integer matrix.
func IntegerEigenvalues2x2(rng *rand.Rand, lambda1, lambda2, maxShear int) Matrix2x2 {
	p := RandomUnimodular2x2(rng, maxShear)
	pInv, ok := p.Inverse()
	if !ok {
		// Fallback for safety
		p = Matrix2x2{A: 1, B: 0, C: 0, D: 1}
		pInv = p
	}
	d := Matrix2x2{A: lambda1, B: 0, C: 0, D: lambda2}
	return p.Multiply(d).Multiply(pInv)
}

// LaTeX formats the 2x2 matrix as a LaTeX pmatrix.
func (m Matrix2x2) LaTeX() string {
	return fmt.Sprintf("\\begin{pmatrix} %d & %d \\\\ %d & %d \\end{pmatrix}", m.A, m.B, m.C, m.D)
}
