package mathutil

import (
	"fmt"
	"math/rand"
	"strings"
)

// Matrix3x3 represents a 3x3 integer matrix:
// [ Entry[0][0]  Entry[0][1]  Entry[0][2] ]
// [ Entry[1][0]  Entry[1][1]  Entry[1][2] ]
// [ Entry[2][0]  Entry[2][1]  Entry[2][2] ]
type Matrix3x3 struct {
	Entry [3][3]int
}

// Identity3x3 returns the 3x3 identity matrix.
func Identity3x3() Matrix3x3 {
	return Matrix3x3{
		Entry: [3][3]int{
			{1, 0, 0},
			{0, 1, 0},
			{0, 0, 1},
		},
	}
}

// Det computes the exact determinant of the 3x3 matrix via Laplace expansion.
func (m Matrix3x3) Det() int {
	a, b, c := m.Entry[0][0], m.Entry[0][1], m.Entry[0][2]
	d, e, f := m.Entry[1][0], m.Entry[1][1], m.Entry[1][2]
	g, h, i := m.Entry[2][0], m.Entry[2][1], m.Entry[2][2]

	return a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
}

// Trace computes the sum of the main diagonal entries.
func (m Matrix3x3) Trace() int {
	return m.Entry[0][0] + m.Entry[1][1] + m.Entry[2][2]
}

// Multiply computes the matrix product m * other.
func (m Matrix3x3) Multiply(other Matrix3x3) Matrix3x3 {
	var res Matrix3x3
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			sum := 0
			for k := 0; k < 3; k++ {
				sum += m.Entry[r][k] * other.Entry[k][c]
			}
			res.Entry[r][c] = sum
		}
	}
	return res
}

// Transpose returns the transpose of the matrix.
func (m Matrix3x3) Transpose() Matrix3x3 {
	return Matrix3x3{
		Entry: [3][3]int{
			{m.Entry[0][0], m.Entry[1][0], m.Entry[2][0]},
			{m.Entry[0][1], m.Entry[1][1], m.Entry[2][1]},
			{m.Entry[0][2], m.Entry[1][2], m.Entry[2][2]},
		},
	}
}

// CofactorMatrix returns the 3x3 cofactor matrix C where C_{ij} = (-1)^{i+j} M_{ij}.
func (m Matrix3x3) CofactorMatrix() Matrix3x3 {
	a, b, c := m.Entry[0][0], m.Entry[0][1], m.Entry[0][2]
	d, e, f := m.Entry[1][0], m.Entry[1][1], m.Entry[1][2]
	g, h, i := m.Entry[2][0], m.Entry[2][1], m.Entry[2][2]

	return Matrix3x3{
		Entry: [3][3]int{
			{e*i - f*h, -(d*i - f*g), d*h - e*g},
			{-(b*i - c*h), a*i - c*g, -(a*h - b*g)},
			{b*f - c*e, -(a*f - c*d), a*e - b*d},
		},
	}
}

// Adjugate returns the classical adjoint (adjugate) matrix: the transpose of the cofactor matrix.
func (m Matrix3x3) Adjugate() Matrix3x3 {
	return m.CofactorMatrix().Transpose()
}

// Inverse returns the exact integer inverse matrix when det(m) is 1 or -1.
// If the matrix is not unimodular, it returns false.
func (m Matrix3x3) Inverse() (Matrix3x3, bool) {
	d := m.Det()
	if d != 1 && d != -1 {
		return Matrix3x3{}, false
	}
	adj := m.Adjugate()
	if d == 1 {
		return adj, true
	}
	// When det is -1, inverse is -adjugate
	var inv Matrix3x3
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			inv.Entry[r][c] = -adj.Entry[r][c]
		}
	}
	return inv, true
}

// IsTriangular returns true if the matrix is strictly upper or lower triangular.
func (m Matrix3x3) IsTriangular() bool {
	upper := m.Entry[1][0] == 0 && m.Entry[2][0] == 0 && m.Entry[2][1] == 0
	lower := m.Entry[0][1] == 0 && m.Entry[0][2] == 0 && m.Entry[1][2] == 0
	return upper || lower
}

// NonTriangularEntries returns the sum of absolute values in the upper and lower off-diagonals.
func (m Matrix3x3) NonTriangularEntries() (upperSum, lowerSum int) {
	upperSum = Abs(m.Entry[0][1]) + Abs(m.Entry[0][2]) + Abs(m.Entry[1][2])
	lowerSum = Abs(m.Entry[1][0]) + Abs(m.Entry[2][0]) + Abs(m.Entry[2][1])
	return upperSum, lowerSum
}

// MaxEntryMagnitude returns the maximum absolute value among all entries.
func (m Matrix3x3) MaxEntryMagnitude() int {
	maxVal := 0
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			v := Abs(m.Entry[r][c])
			if v > maxVal {
				maxVal = v
			}
		}
	}
	return maxVal
}

// CharacteristicPolynomialCoeffs returns [c3, c2, c1, c0] for p(lambda) = det(lambda I - M):
// p(lambda) = lambda^3 - tr(M)*lambda^2 + M_principal*lambda - det(M).
func (m Matrix3x3) CharacteristicPolynomialCoeffs() [4]int {
	tr := m.Trace()
	det := m.Det()
	// Sum of principal 2x2 minors:
	cof := m.CofactorMatrix()
	c1 := cof.Entry[0][0] + cof.Entry[1][1] + cof.Entry[2][2]
	return [4]int{1, -tr, c1, -det}
}

// LaTeX formats the 3x3 matrix as a LaTeX pmatrix.
func (m Matrix3x3) LaTeX() string {
	var rows []string
	for r := 0; r < 3; r++ {
		rows = append(rows, fmt.Sprintf("%d & %d & %d", m.Entry[r][0], m.Entry[r][1], m.Entry[r][2]))
	}
	return fmt.Sprintf("\\begin{pmatrix} %s \\end{pmatrix}", strings.Join(rows, " \\\\ "))
}

// ElementaryShear3x3 constructs an elementary row/column shear E_{row, col}(k):
// Identity with entry (row, col) set to k. (row != col)
func ElementaryShear3x3(row, col, k int) Matrix3x3 {
	m := Identity3x3()
	if row >= 0 && row < 3 && col >= 0 && col < 3 && row != col {
		m.Entry[row][col] = k
	}
	return m
}

// RandomUnimodular3x3 constructs a random 3x3 integer matrix with determinant 1.
// It applies a chain of random elementary shears E_{i,j}(k) to the identity matrix.
// It bounds entry growth to prevent numeric overflow.
func RandomUnimodular3x3(rng *rand.Rand, maxShear int) Matrix3x3 {
	if maxShear < 1 {
		maxShear = 1
	}

	for attempt := 0; attempt < 50; attempt++ {
		m := Identity3x3()
		shears := 3 + rng.Intn(4) // 3 to 6 elementary operations
		for s := 0; s < shears; s++ {
			r := rng.Intn(3)
			c := (r + 1 + rng.Intn(2)) % 3
			k := rng.Intn(2*maxShear+1) - maxShear
			if k == 0 {
				k = 1
				if rng.Intn(2) == 0 {
					k = -1
				}
			}
			shear := ElementaryShear3x3(r, c, k)
			if rng.Intn(2) == 0 {
				m = shear.Multiply(m)
			} else {
				m = m.Multiply(shear)
			}
		}

		if m.Det() == 1 && m.MaxEntryMagnitude() <= 8 && !m.IsTriangular() {
			return m
		}
	}

	// Fallback to a well-behaved unimodular non-triangular matrix if random search exceeds attempts
	return Matrix3x3{
		Entry: [3][3]int{
			{1, 1, 0},
			{1, 2, 1},
			{0, 1, 1},
		},
	}
}

// IntegerEigenvalues3x3 constructs an authentic non-triangular 3x3 integer matrix
// with exact integer eigenvalues l1, l2, l3 using the similarity transform A = P * D * P^{-1}.
func IntegerEigenvalues3x3(rng *rand.Rand, l1, l2, l3 int, maxShear int) (Matrix3x3, Matrix3x3, bool) {
	diag := Matrix3x3{
		Entry: [3][3]int{
			{l1, 0, 0},
			{0, l2, 0},
			{0, 0, l3},
		},
	}

	for attempt := 0; attempt < 250; attempt++ {
		p := RandomUnimodular3x3(rng, maxShear)
		pInv, ok := p.Inverse()
		if !ok {
			continue
		}

		// A = P * D * P^{-1}
		a := p.Multiply(diag).Multiply(pInv)

		// Verification: must be non-triangular on both sides
		uSum, lSum := a.NonTriangularEntries()
		if uSum == 0 || lSum == 0 {
			continue
		}

		// Prevent trivial read-offs: none of the eigenvalues should equal any diagonal entry
		if a.Entry[0][0] == l1 || a.Entry[0][0] == l2 || a.Entry[0][0] == l3 ||
			a.Entry[1][1] == l1 || a.Entry[1][1] == l2 || a.Entry[1][1] == l3 ||
			a.Entry[2][2] == l1 || a.Entry[2][2] == l2 || a.Entry[2][2] == l3 {
			continue
		}

		// Ensure entries are clean integers within bounds
		if a.MaxEntryMagnitude() <= 25 {
			return a, p, true
		}
	}

	// Deterministic non-triangular unimodular fallback with pairwise distinct l1, l2, l3
	p := Matrix3x3{
		Entry: [3][3]int{
			{1, 1, 1},
			{1, 2, 2},
			{1, 2, 3},
		},
	}
	pInv, _ := p.Inverse()
	a := p.Multiply(diag).Multiply(pInv)
	return a, p, true
}

// JordanMatrix3x3 constructs an authentic non-triangular 3x3 matrix with a non-trivial
// Jordan canonical form J = P^{-1} * A * P.
// If singleBlock is true:
//
//	J has a single 3x3 Jordan block: [lambda 1 0; 0 lambda 1; 0 0 lambda]
//
// If singleBlock is false:
//
//	J has a 2x2 Jordan block and a 1x1 block: [lambda 1 0; 0 lambda 0; 0 0 mu]
func JordanMatrix3x3(rng *rand.Rand, lambda, mu int, singleBlock bool, maxShear int) (Matrix3x3, Matrix3x3, bool) {
	var j Matrix3x3
	if singleBlock {
		j = Matrix3x3{
			Entry: [3][3]int{
				{lambda, 1, 0},
				{0, lambda, 1},
				{0, 0, lambda},
			},
		}
	} else {
		j = Matrix3x3{
			Entry: [3][3]int{
				{lambda, 1, 0},
				{0, lambda, 0},
				{0, 0, mu},
			},
		}
	}

	for attempt := 0; attempt < 100; attempt++ {
		p := RandomUnimodular3x3(rng, maxShear)
		pInv, ok := p.Inverse()
		if !ok {
			continue
		}

		// A = P * J * P^{-1}
		a := p.Multiply(j).Multiply(pInv)

		// Must not be triangular
		uSum, lSum := a.NonTriangularEntries()
		if uSum == 0 || lSum == 0 {
			continue
		}

		// Do not leak eigenvalue directly on all diagonal elements
		if a.Entry[0][0] == lambda && a.Entry[1][1] == lambda && a.Entry[2][2] == lambda {
			continue
		}

		if a.MaxEntryMagnitude() <= 25 {
			return a, p, true
		}
	}

	return j, Identity3x3(), false
}
