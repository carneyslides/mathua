package combinatorics

import (
	"testing"
)

func TestPow(t *testing.T) {
	if got := pow(3, 0); got != 1 {
		t.Errorf("pow(3, 0) = %d, want 1", got)
	}
	if got := pow(2, 5); got != 32 {
		t.Errorf("pow(2, 5) = %d, want 32", got)
	}
}

func TestOntoCount(t *testing.T) {
	// Onto functions from {1, 2, 3} to {A, B}: 2^3 - 2 = 6
	if got := ontoCount(3, 2); got != 6 {
		t.Errorf("ontoCount(3, 2) = %d, want 6", got)
	}
}

func TestStirling2(t *testing.T) {
	cases := []struct {
		m, n int
		want int
	}{
		{0, 0, 0},
		{3, 0, 0},
		{2, 3, 0},
		{4, 1, 1},
		{4, 4, 1},
		{4, 2, 7},
		{4, 3, 6},
	}
	for _, tc := range cases {
		if got := stirling2(tc.m, tc.n); got != tc.want {
			t.Errorf("stirling2(%d, %d) = %d, want %d", tc.m, tc.n, got, tc.want)
		}
	}
}

func TestTechniquesGenerators(t *testing.T) {
	fuzzGen(t, &vandermondeGen{})
	fuzzGen(t, &doubleCountingGen{})
	fuzzGen(t, &surjectionGen{})
	fuzzGen(t, &stirlingSecondGen{})
	fuzzGen(t, &bellGen{})
	fuzzGen(t, &integerPartitionGen{})
	fuzzGen(t, &latticePathGen{})
	fuzzGen(t, &compositionGen{})
}
