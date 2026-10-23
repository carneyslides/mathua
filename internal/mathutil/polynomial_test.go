package mathutil

import (
	"testing"
)

func TestMathutilPolyFromRoots(t *testing.T) {
	cases := [][]int{
		{2, 3},
		{-1, 4},
		{0, 5},
		{-2, -3, 1},
		{3, 3},
		{-4, 0, 4},
	}

	for _, roots := range cases {
		p := PolyFromRoots(roots...)
		// Degree should equal number of roots
		if p.Degree() != len(roots) {
			t.Errorf("Degree() = %d, want %d for roots %v", p.Degree(), len(roots), roots)
		}
		// Leading coefficient must be 1 (monic)
		if p.Coeffs[len(p.Coeffs)-1] != 1 {
			t.Errorf("Leading coefficient = %d, want 1", p.Coeffs[len(p.Coeffs)-1])
		}
		// Evaluating at each root must produce 0
		for _, r := range roots {
			if val := p.Eval(r); val != 0 {
				t.Errorf("P(%d) = %d != 0 for roots %v", r, val, roots)
			}
		}
	}
}

func TestMathutilPolynomial_Derivative(t *testing.T) {
	// P(x) = 6 - 5x + x^2 => P'(x) = -5 + 2x
	p := NewPolynomial(6, -5, 1)
	dp := p.Derivative()
	if len(dp.Coeffs) != 2 || dp.Coeffs[0] != -5 || dp.Coeffs[1] != 2 {
		t.Errorf("Derivative() = %+v, want [-5, 2]", dp.Coeffs)
	}

	// Constant polynomial derivative is 0
	c := NewPolynomial(42)
	dc := c.Derivative()
	if len(dc.Coeffs) != 1 || dc.Coeffs[0] != 0 {
		t.Errorf("Constant derivative = %+v, want [0]", dc.Coeffs)
	}
}

func TestMathutilPolynomial_Format(t *testing.T) {
	cases := []struct {
		coeffs []int
		varName string
		want   string
	}{
		{[]int{6, -5, 1}, "x", "x^2 - 5x + 6"},
		{[]int{-4, 0, 1}, "t", "t^2 - 4"},
		{[]int{0, 3, 2}, "x", "2x^2 + 3x"},
		{[]int{7}, "x", "7"},
		{[]int{-12}, "y", "-12"},
		{[]int{0, 0, 1}, "x", "x^2"},
		{[]int{1, -1}, "x", "-x + 1"},
		{[]int{-8, 12, -6, 1}, "x", "x^3 - 6x^2 + 12x - 8"},
	}

	for _, tc := range cases {
		p := NewPolynomial(tc.coeffs...)
		got := p.Format(tc.varName)
		if got != tc.want {
			t.Errorf("Format(%q) for %v = %q, want %q", tc.varName, tc.coeffs, got, tc.want)
		}
	}
}
