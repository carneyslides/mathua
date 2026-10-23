package mathutil

import (
	"testing"
)

func TestMathutilScaleInt(t *testing.T) {
	cases := []struct {
		diff float64
		min  int
		max  int
		want int
	}{
		{0.0, 5, 25, 5},
		{1.0, 5, 25, 25},
		{0.5, 0, 10, 5},
		{-0.2, 10, 20, 10},
		{1.5, 10, 20, 20},
		{0.5, 20, 10, 20}, // max <= min returns min
	}

	for _, tc := range cases {
		got := ScaleInt(tc.diff, tc.min, tc.max)
		if got != tc.want {
			t.Errorf("ScaleInt(%v, %d, %d) = %d, want %d", tc.diff, tc.min, tc.max, got, tc.want)
		}
	}
}

func TestMathutilScaleFloat(t *testing.T) {
	cases := []struct {
		diff float64
		min  float64
		max  float64
		want float64
	}{
		{0.0, 1.0, 5.0, 1.0},
		{1.0, 1.0, 5.0, 5.0},
		{0.5, 2.0, 6.0, 4.0},
		{-0.5, 2.0, 6.0, 2.0},
		{1.5, 2.0, 6.0, 6.0},
	}

	for _, tc := range cases {
		got := ScaleFloat(tc.diff, tc.min, tc.max)
		if got != tc.want {
			t.Errorf("ScaleFloat(%v, %v, %v) = %v, want %v", tc.diff, tc.min, tc.max, got, tc.want)
		}
	}
}

func TestMathutilDifficultyTier(t *testing.T) {
	if got := DifficultyTier(0.1); got != 0 {
		t.Errorf("DifficultyTier(0.1) = %d, want 0", got)
	}
	if got := DifficultyTier(0.5); got != 1 {
		t.Errorf("DifficultyTier(0.5) = %d, want 1", got)
	}
	if got := DifficultyTier(0.9); got != 2 {
		t.Errorf("DifficultyTier(0.9) = %d, want 2", got)
	}
}
