package mathutil

import (
	"testing"
)

func TestMathutilFraction_ReductionAndSign(t *testing.T) {
	cases := []struct {
		num, den int
		wantNum  int
		wantDen  int
	}{
		{6, 8, 3, 4},
		{4, -8, -1, 2},
		{-6, -9, 2, 3},
		{0, 5, 0, 1},
		{15, 5, 3, 1},
		{-14, 21, -2, 3},
	}

	for _, tc := range cases {
		f := NewFraction(tc.num, tc.den)
		if f.Num != tc.wantNum || f.Den != tc.wantDen {
			t.Errorf("NewFraction(%d, %d) = %d/%d, want %d/%d", tc.num, tc.den, f.Num, f.Den, tc.wantNum, tc.wantDen)
		}
	}
}

func TestMathutilFraction_Arithmetic(t *testing.T) {
	f1 := NewFraction(1, 2)
	f2 := NewFraction(1, 3)

	add := f1.Add(f2) // 1/2 + 1/3 = 5/6
	if add != NewFraction(5, 6) {
		t.Errorf("Add = %v, want 5/6", add)
	}

	sub := f1.Sub(f2) // 1/2 - 1/3 = 1/6
	if sub != NewFraction(1, 6) {
		t.Errorf("Sub = %v, want 1/6", sub)
	}

	mul := f1.Mul(f2) // 1/2 * 1/3 = 1/6
	if mul != NewFraction(1, 6) {
		t.Errorf("Mul = %v, want 1/6", mul)
	}

	div := f1.Div(f2) // (1/2) / (1/3) = 3/2
	if div != NewFraction(3, 2) {
		t.Errorf("Div = %v, want 3/2", div)
	}
}

func TestMathutilFraction_LaTeX(t *testing.T) {
	cases := []struct {
		f    Fraction
		want string
	}{
		{NewFraction(3, 1), "3"},
		{NewFraction(-5, 1), "-5"},
		{NewFraction(2, 3), "\\frac{2}{3}"},
		{NewFraction(-3, 4), "-\\frac{3}{4}"},
	}

	for _, tc := range cases {
		if got := tc.f.LaTeX(); got != tc.want {
			t.Errorf("LaTeX() for %v = %q, want %q", tc.f, got, tc.want)
		}
	}
}

func TestMathutilSimplifyRadical(t *testing.T) {
	cases := []struct {
		n         int
		wantCoeff int
		wantRad   int
	}{
		{12, 2, 3},
		{75, 5, 3},
		{16, 4, 1},
		{7, 1, 7},
		{50, 5, 2},
		{72, 6, 2},
		{1, 1, 1},
	}

	for _, tc := range cases {
		c, r := SimplifyRadical(tc.n)
		if c != tc.wantCoeff || r != tc.wantRad {
			t.Errorf("SimplifyRadical(%d) = (%d, %d), want (%d, %d)", tc.n, c, r, tc.wantCoeff, tc.wantRad)
		}
	}
}

func TestMathutilFormatRadical(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{16, "4"},
		{7, "\\sqrt{7}"},
		{75, "5\\sqrt{3}"},
		{72, "6\\sqrt{2}"},
		{1, "1"},
	}

	for _, tc := range cases {
		if got := FormatRadical(tc.n); got != tc.want {
			t.Errorf("FormatRadical(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}
