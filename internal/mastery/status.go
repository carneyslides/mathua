package mastery

import "github.com/chuma-beep/mathua/internal/fsrs"

type Status string

const (
	StatusUnseen     Status = "UNSEEN"
	StatusLearning   Status = "LEARNING"
	StatusPracticing Status = "PRACTICING"
	StatusMastered   Status = "MASTERED"
)

// DecayDays is how long a mastered concept stays fresh before it reads as
// DECAYING when no dynamic FSRS stability is present.
//
// This was written four times: as `14` in the engine, as a bare `14` twice in
// the scheduler, and as `30` in the workload estimator. The estimator therefore
// told learners their reviews came round every 30 days while the scheduler
// scheduled them at 14 — a plan built on a review cadence the app would not
// honour, which is the one number a deadline is computed from (ADR-019 prices
// reviews on this path).
//
// One value, owned here beside EffectiveStatus, which is the only place that
// interprets it.
const DecayDays = 14

// EffectiveStatus returns DECAYING if a MASTERED concept hasn't been reviewed
// beyond the decay threshold (in days). Decay is computed at read time —
// the database never stores it.
func EffectiveStatus(current Status, daysSinceLastReview float64, decayThresholdDays float64) Status {
	return EffectiveStatusFSRS(current, daysSinceLastReview, nil, decayThresholdDays)
}

// EffectiveStatusFSRS evaluates whether a MASTERED concept has decayed.
// If stability is present and positive, it evaluates FSRS-6 retrievability:
// R(t, S) = (1 + Factor * t / S)^Decay.
// When R < DefaultRetention (0.90), the concept has decayed to "DECAYING".
// If stability is nil or non-positive, it falls back to decayThresholdDays (or DecayDays).
func EffectiveStatusFSRS(current Status, daysSinceLastReview float64, stability *float64, decayThresholdDays float64) Status {
	if current != StatusMastered {
		return current
	}
	if stability != nil && *stability > 0 {
		r := fsrs.DefaultParameters().Retrievability(daysSinceLastReview, *stability)
		if r < fsrs.DefaultRetention {
			return "DECAYING"
		}
		return StatusMastered
	}
	if decayThresholdDays <= 0 {
		decayThresholdDays = DecayDays
	}
	if daysSinceLastReview >= decayThresholdDays {
		return "DECAYING"
	}
	return current
}
