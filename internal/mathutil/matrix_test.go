package mathutil

import (
	"math/rand"
	"testing"
)

func TestMathutilMatrix2x2_DetAndTrace(t *testing.T) {
	m := Matrix2x2{A: 3, B: 2, C: 1, D: 4}
	if got := m.Det(); got != 10 {
		t.Errorf("Det() = %d, want 10", got)
	}
	if got := m.Trace(); got != 7 {
		t.Errorf("Trace() = %d, want 7", got)
	}
}

func TestMathutilMatrix2x2_Multiply(t *testing.T) {
	m1 := Matrix2x2{A: 1, B: 2, C: 3, D: 4}
	m2 := Matrix2x2{A: 2, B: 0, C: 1, D: 2}
	prod := m1.Multiply(m2)
	// [1*2 + 2*1, 1*0 + 2*2] = [4, 4]
	// [3*2 + 4*1, 3*0 + 4*2] = [10, 8]
	want := Matrix2x2{A: 4, B: 4, C: 10, D: 8}
	if prod != want {
		t.Errorf("Multiply() = %+v, want %+v", prod, want)
	}
}

func TestMathutilMatrix2x2_Inverse(t *testing.T) {
	m := Matrix2x2{A: 3, B: 5, C: 1, D: 2} // Det = 6 - 5 = 1
	inv, ok := m.Inverse()
	if !ok {
		t.Fatalf("expected unimodular matrix to have integer inverse")
	}
	id := m.Multiply(inv)
	wantId := Matrix2x2{A: 1, B: 0, C: 0, D: 1}
	if id != wantId {
		t.Errorf("m * inv = %+v, want identity %+v", id, wantId)
	}
}

func TestMathutilRandomUnimodular2x2(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		m := RandomUnimodular2x2(rng, 3)
		det := m.Det()
		if det != 1 && det != -1 {
			t.Fatalf("RandomUnimodular2x2 produced det %d, want 1 or -1: %+v", det, m)
		}
		inv, ok := m.Inverse()
		if !ok {
			t.Fatalf("RandomUnimodular2x2 failed to invert: %+v", m)
		}
		id := m.Multiply(inv)
		if id.A != 1 || id.B != 0 || id.C != 0 || id.D != 1 {
			t.Fatalf("m * inv != I: %+v * %+v = %+v", m, inv, id)
		}
	}
}

func TestMathutilIntegerEigenvalues2x2(t *testing.T) {
	rng := rand.New(rand.NewSource(1337))
	for i := 0; i < 200; i++ {
		l1 := rng.Intn(15) - 7
		l2 := rng.Intn(15) - 7
		m := IntegerEigenvalues2x2(rng, l1, l2, 2)

		// Characteristic equation det(A - lambda*I) = 0
		// lambda^2 - Trace(A)*lambda + Det(A) = 0
		trace := m.Trace()
		det := m.Det()
		wantTrace := l1 + l2
		wantDet := l1 * l2
		if trace != wantTrace {
			t.Errorf("Trace() = %d, want %d for eigenvalues %d and %d", trace, wantTrace, l1, l2)
		}
		if det != wantDet {
			t.Errorf("Det() = %d, want %d for eigenvalues %d and %d", det, wantDet, l1, l2)
		}

		// Verify det(A - l1*I) == 0 and det(A - l2*I) == 0
		char1 := Matrix2x2{A: m.A - l1, B: m.B, C: m.C, D: m.D - l1}.Det()
		char2 := Matrix2x2{A: m.A - l2, B: m.B, C: m.C, D: m.D - l2}.Det()
		if char1 != 0 {
			t.Errorf("det(A - l1*I) = %d != 0 for matrix %+v", char1, m)
		}
		if char2 != 0 {
			t.Errorf("det(A - l2*I) = %d != 0 for matrix %+v", char2, m)
		}
	}
}

func TestMathutilMatrix2x2_LaTeX(t *testing.T) {
	m := Matrix2x2{A: 1, B: 2, C: 3, D: 4}
	want := "\\begin{pmatrix} 1 & 2 \\\\ 3 & 4 \\end{pmatrix}"
	if got := m.LaTeX(); got != want {
		t.Errorf("LaTeX() = %q, want %q", got, want)
	}
}
