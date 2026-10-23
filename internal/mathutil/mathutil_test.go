package mathutil

import (
	"math/rand"
	"testing"
)

func TestGCD(t *testing.T) {
	cases := []struct {
		a, b, want int
	}{
		{54, 24, 6},
		{24, 54, 6},
		{48, 18, 6},
		{101, 103, 1},
		{17, 13, 1},
		{8, 9, 1},
		{7, 7, 7},
		{100, 1, 1},
		{1, 100, 1},
		{42, 0, 42},
		{0, 42, 42},
		{12, 36, 12},
		{36, 12, 12},
		{1000, 250, 250},
	}
	for _, tc := range cases {
		if got := GCD(tc.a, tc.b); got != tc.want {
			t.Errorf("GCD(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestGCDProperties(t *testing.T) {
	// Commutativity and common divisor invariants over random positive pairs.
	for i := 0; i < 500; i++ {
		a := 1 + rand.Intn(10000)
		b := 1 + rand.Intn(10000)

		g1 := GCD(a, b)
		g2 := GCD(b, a)
		if g1 != g2 {
			t.Fatalf("GCD is not commutative: GCD(%d, %d) = %d != GCD(%d, %d) = %d", a, b, g1, b, a, g2)
		}
		if g1 <= 0 {
			t.Fatalf("GCD(%d, %d) returned non-positive %d", a, b, g1)
		}
		if a%g1 != 0 || b%g1 != 0 {
			t.Fatalf("GCD(%d, %d) = %d does not divide both numbers", a, b, g1)
		}
	}
}

func TestIntPow(t *testing.T) {
	cases := []struct {
		base, exp, want int
	}{
		{5, 0, 1},
		{0, 0, 1},
		{-3, 0, 1},
		{7, 1, 7},
		{-4, 1, -4},
		{0, 5, 0},
		{1, 50, 1},
		{2, 10, 1024},
		{3, 4, 81},
		{10, 5, 100000},
		{-2, 3, -8},
		{-2, 4, 16},
		{-1, 99, -1},
		{-1, 100, 1},
	}
	for _, tc := range cases {
		if got := IntPow(tc.base, tc.exp); got != tc.want {
			t.Errorf("IntPow(%d, %d) = %d, want %d", tc.base, tc.exp, got, tc.want)
		}
	}
}

func TestAbs(t *testing.T) {
	cases := []struct {
		n, want int
	}{
		{0, 0},
		{1, 1},
		{-1, 1},
		{42, 42},
		{-42, 42},
		{1000000, 1000000},
		{-1000000, 1000000},
	}
	for _, tc := range cases {
		if got := Abs(tc.n); got != tc.want {
			t.Errorf("Abs(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}
