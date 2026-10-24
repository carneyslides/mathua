package mathutil

import (
	"math/rand"
	"testing"
)

func TestMatrix3x3_DetAndInverse(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 500; i++ {
		p := RandomUnimodular3x3(rng, 2)
		d := p.Det()
		if d != 1 {
			t.Fatalf("draw %d: expected det(P) == 1, got %d", i, d)
		}

		pInv, ok := p.Inverse()
		if !ok {
			t.Fatalf("draw %d: expected p.Inverse() to succeed for unimodular matrix", i)
		}

		prod := p.Multiply(pInv)
		ident := Identity3x3()
		if prod != ident {
			t.Fatalf("draw %d: expected P * P^{-1} == I, got %v", i, prod)
		}
	}
}

func TestMatrix3x3_IntegerEigenvalues(t *testing.T) {
	rng := rand.New(rand.NewSource(123))

	for i := 0; i < 500; i++ {
		l1 := rng.Intn(7) - 3 // -3 to 3
		l2 := l1 + 1 + rng.Intn(4)
		l3 := l2 + 1 + rng.Intn(4)

		a, _, ok := IntegerEigenvalues3x3(rng, l1, l2, l3, 2)
		if !ok {
			t.Fatalf("draw %d: failed to construct IntegerEigenvalues3x3", i)
		}

		// Verify non-triangular
		if a.IsTriangular() {
			t.Fatalf("draw %d: matrix is triangular", i)
		}
		uSum, lSum := a.NonTriangularEntries()
		if uSum == 0 || lSum == 0 {
			t.Fatalf("draw %d: expected non-triangular entries on both sides, got u=%d, l=%d", i, uSum, lSum)
		}

		// Verify eigenvalues via characteristic polynomial det(lambda I - A) == 0
		for _, eig := range []int{l1, l2, l3} {
			shifted := a
			shifted.Entry[0][0] -= eig
			shifted.Entry[1][1] -= eig
			shifted.Entry[2][2] -= eig
			if det := shifted.Det(); det != 0 {
				t.Fatalf("draw %d: lambda=%d is not an eigenvalue, det(A - lambda I) = %d", i, eig, det)
			}
		}

		// Verify trace is sum of eigenvalues
		if tr := a.Trace(); tr != (l1 + l2 + l3) {
			t.Fatalf("draw %d: expected trace %d, got %d", i, l1+l2+l3, tr)
		}

		// Verify determinant is product of eigenvalues
		if det := a.Det(); det != (l1 * l2 * l3) {
			t.Fatalf("draw %d: expected det %d, got %d", i, l1*l2*l3, det)
		}
	}
}

func TestMatrix3x3_JordanMatrix(t *testing.T) {
	rng := rand.New(rand.NewSource(456))

	for i := 0; i < 500; i++ {
		lambda := rng.Intn(7) - 3
		mu := lambda + 1 + rng.Intn(4)
		singleBlock := (i%2 == 0)

		a, p, ok := JordanMatrix3x3(rng, lambda, mu, singleBlock, 2)
		if !ok {
			t.Fatalf("draw %d: JordanMatrix3x3 failed", i)
		}

		pInv, pOk := p.Inverse()
		if !pOk {
			t.Fatalf("draw %d: P must be invertible", i)
		}

		// J = P^{-1} * A * P
		j := pInv.Multiply(a).Multiply(p)

		if singleBlock {
			if j.Entry[0][0] != lambda || j.Entry[1][1] != lambda || j.Entry[2][2] != lambda {
				t.Fatalf("draw %d: expected diagonal lambda=%d in single-block J, got %v", i, lambda, j)
			}
			if j.Entry[0][1] != 1 || j.Entry[1][2] != 1 {
				t.Fatalf("draw %d: expected superdiagonal 1s in single-block J, got %v", i, j)
			}
		} else {
			if j.Entry[0][0] != lambda || j.Entry[1][1] != lambda || j.Entry[2][2] != mu {
				t.Fatalf("draw %d: expected diagonal in two-block J, got %v", i, j)
			}
			if j.Entry[0][1] != 1 {
				t.Fatalf("draw %d: expected superdiagonal 1 in block 1, got %v", i, j)
			}
		}

		// Geometric multiplicity of lambda must be strictly less than algebraic multiplicity
		shifted := a
		shifted.Entry[0][0] -= lambda
		shifted.Entry[1][1] -= lambda
		shifted.Entry[2][2] -= lambda

		// det(A - lambda I) must be 0
		if det := shifted.Det(); det != 0 {
			t.Fatalf("draw %d: det(A - lambda I) = %d != 0", i, det)
		}
	}
}
