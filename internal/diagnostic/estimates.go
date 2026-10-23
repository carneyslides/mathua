package diagnostic

import (
	"fmt"
	"math"
	"time"
)

// EstimateCompletionDates computes estimated completion date strings under specified daily XP velocities.
// Paces typically include 20, 40 and 60 XP per day.
func EstimateCompletionDates(remainingXP float64, startDate time.Time, velocities []int) map[int]string {
	if len(velocities) == 0 {
		velocities = []int{20, 40, 60}
	}
	out := make(map[int]string, len(velocities))
	for _, pace := range velocities {
		if pace <= 0 {
			continue
		}
		days := math.Ceil(remainingXP / float64(pace))
		if days < 1 {
			days = 1
		}
		if startDate.IsZero() {
			out[pace] = fmt.Sprintf("≈%d days", int(days))
		} else {
			targetDate := startDate.AddDate(0, 0, int(days))
			out[pace] = targetDate.Format("2006-01-02")
		}
	}
	return out
}

// WithCompletionEstimates fills completion estimates under 20, 40 and 60 daily XP velocities.
func (r *DiagnosticReport) WithCompletionEstimates(remainingXP float64, startDate time.Time, velocities ...int) *DiagnosticReport {
	if len(velocities) == 0 {
		velocities = []int{20, 40, 60}
	}
	r.CompletionEstimates = EstimateCompletionDates(remainingXP, startDate, velocities)
	return r
}
