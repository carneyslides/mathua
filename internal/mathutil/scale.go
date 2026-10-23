package mathutil

import (
	"math"
)

// ScaleInt interpolates an integer range based on difficulty in [0.0, 1.0].
// At difficulty 0.0 it returns a value near minVal; at 1.0 it returns maxVal.
func ScaleInt(difficulty float64, minVal, maxVal int) int {
	if difficulty < 0.0 {
		difficulty = 0.0
	}
	if difficulty > 1.0 {
		difficulty = 1.0
	}
	if maxVal <= minVal {
		return minVal
	}
	val := float64(minVal) + difficulty*float64(maxVal-minVal)
	return int(math.Round(val))
}

// ScaleFloat interpolates a floating point range based on difficulty in [0.0, 1.0].
func ScaleFloat(difficulty float64, minVal, maxVal float64) float64 {
	if difficulty < 0.0 {
		difficulty = 0.0
	}
	if difficulty > 1.0 {
		difficulty = 1.0
	}
	return minVal + difficulty*(maxVal-minVal)
}

// DifficultyTier categorizes difficulty in [0.0, 1.0] into three pedagogical tiers:
// 0 = foundational (0.0 to 0.35)
// 1 = intermediate (0.35 to 0.70)
// 2 = advanced (0.70 to 1.00)
func DifficultyTier(difficulty float64) int {
	if difficulty < 0.35 {
		return 0
	}
	if difficulty < 0.70 {
		return 1
	}
	return 2
}
