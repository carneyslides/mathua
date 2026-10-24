package mathutil

import (
	"fmt"
)

// IntMatrix represents an m-by-n integer matrix.
type IntMatrix struct {
	Rows int
	Cols int
	Data [][]int
}

// NewIntMatrix creates an m-by-n matrix initialized to zero.
func NewIntMatrix(rows, cols int) IntMatrix {
	data := make([][]int, rows)
	for i := range data {
		data[i] = make([]int, cols)
	}
	return IntMatrix{Rows: rows, Cols: cols, Data: data}
}

// Clone creates a deep copy of the matrix.
func (m IntMatrix) Clone() IntMatrix {
	copied := NewIntMatrix(m.Rows, m.Cols)
	for i := 0; i < m.Rows; i++ {
		copy(copied.Data[i], m.Data[i])
	}
	return copied
}

// Multiply computes the matrix product m * other.
func (m IntMatrix) Multiply(other IntMatrix) (IntMatrix, error) {
	if m.Cols != other.Rows {
		return IntMatrix{}, fmt.Errorf("incompatible dimensions: %dx%d and %dx%d", m.Rows, m.Cols, other.Rows, other.Cols)
	}
	res := NewIntMatrix(m.Rows, other.Cols)
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < other.Cols; j++ {
			sum := 0
			for k := 0; k < m.Cols; k++ {
				sum += m.Data[i][k] * other.Data[k][j]
			}
			res.Data[i][j] = sum
		}
	}
	return res, nil
}

// IsZero returns true if all matrix entries are zero.
func (m IntMatrix) IsZero() bool {
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < m.Cols; j++ {
			if m.Data[i][j] != 0 {
				return false
			}
		}
	}
	return true
}

// SmithNormalForm computes the invariant factors (diagonal elements d_i > 0)
// and the matrix rank over the integers.
func (m IntMatrix) SmithNormalForm() ([]int, int) {
	a := m.Clone()
	rows := a.Rows
	cols := a.Cols
	limit := rows
	if cols < limit {
		limit = cols
	}

	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}

	swapRows := func(r1, r2 int) {
		a.Data[r1], a.Data[r2] = a.Data[r2], a.Data[r1]
	}

	swapCols := func(c1, c2 int) {
		for i := 0; i < rows; i++ {
			a.Data[i][c1], a.Data[i][c2] = a.Data[i][c2], a.Data[i][c1]
		}
	}

	addMultipleRow := func(dest, src, factor int) {
		for j := 0; j < cols; j++ {
			a.Data[dest][j] += factor * a.Data[src][j]
		}
	}

	addMultipleCol := func(dest, src, factor int) {
		for i := 0; i < rows; i++ {
			a.Data[i][dest] += factor * a.Data[i][src]
		}
	}

	for k := 0; k < limit; k++ {
		for {
			// Find non-zero entry with minimal absolute value in submatrix [k..rows-1, k..cols-1]
			pivotR, pivotC := -1, -1
			minVal := 0
			for i := k; i < rows; i++ {
				for j := k; j < cols; j++ {
					v := abs(a.Data[i][j])
					if v > 0 {
						if pivotR == -1 || v < minVal {
							minVal = v
							pivotR = i
							pivotC = j
						}
					}
				}
			}

			// Submatrix is all zero
			if pivotR == -1 {
				break
			}

			// Move pivot to (k, k)
			if pivotR != k {
				swapRows(k, pivotR)
			}
			if pivotC != k {
				swapCols(k, pivotC)
			}
			if a.Data[k][k] < 0 {
				for j := 0; j < cols; j++ {
					a.Data[k][j] = -a.Data[k][j]
				}
			}

			// Try to eliminate column entries
			cleared := true
			for i := k + 1; i < rows; i++ {
				if a.Data[i][k] != 0 {
					q := a.Data[i][k] / a.Data[k][k]
					addMultipleRow(i, k, -q)
					if a.Data[i][k] != 0 {
						cleared = false
						break
					}
				}
			}
			if !cleared {
				continue
			}

			// Try to eliminate row entries
			for j := k + 1; j < cols; j++ {
				if a.Data[k][j] != 0 {
					q := a.Data[k][j] / a.Data[k][k]
					addMultipleCol(j, k, -q)
					if a.Data[k][j] != 0 {
						cleared = false
						break
					}
				}
			}
			if !cleared {
				continue
			}

			// Check divisibility property: a[k][k] must divide all a[i][j] for i > k, j > k
			divisible := true
			for i := k + 1; i < rows; i++ {
				for j := k + 1; j < cols; j++ {
					if a.Data[i][j]%a.Data[k][k] != 0 {
						// Add row i to row k, and restart reduction on position k
						addMultipleRow(k, i, 1)
						divisible = false
						break
					}
				}
				if !divisible {
					break
				}
			}
			if divisible {
				break
			}
		}
	}

	var invariants []int
	rank := 0
	for i := 0; i < limit; i++ {
		v := abs(a.Data[i][i])
		if v > 0 {
			invariants = append(invariants, v)
			rank++
		}
	}
	return invariants, rank
}

// CellularBettiNumbers computes beta0, beta1, beta2 for a 2D CW chain complex:
// 0 -> Z^{c2} --(d2)--> Z^{c1} --(d1)--> Z^{c0} -> 0
// where d1 has shape c0-by-c1 and d2 has shape c1-by-c2.
func CellularBettiNumbers(c0, c1, c2 int, d1, d2 IntMatrix) (int, int, int) {
	_, rank1 := d1.SmithNormalForm()
	_, rank2 := d2.SmithNormalForm()

	beta0 := c0 - rank1
	beta1 := c1 - rank1 - rank2
	beta2 := c2 - rank2

	return beta0, beta1, beta2
}
