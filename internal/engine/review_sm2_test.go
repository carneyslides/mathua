package engine

import (
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/storage"
)

// TestMasteredReviewAdvancesSM2AndInterval verifies that a successful review of a
// mastered concept advances SM-2 repetitions and interval into the future, and
// does NOT reset to interval 1 day due to streak reset.
func TestMasteredReviewAdvancesSM2AndInterval(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("review-sm2-advances")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	now := time.Now().UTC().Add(-7 * 24 * time.Hour)
	pastDue := now.Add(-24 * time.Hour)
	initialStab := 10.0
	initialDiff := 4.0

	// Concept 'a' has required_streak = 3 in testEngine.
	p := &storage.ConceptProgress{
		StudentID:       st.ID,
		ConceptID:       "a",
		Status:          string(mastery.StatusMastered),
		Streak:          1, // Streak was reset to 1 upon graduation to MASTERED
		AvgResponseTime: 5.0,
		LastReviewed:    &now,
		NextReviewDue:   &pastDue,
		SM2Repetitions:  1,
		SM2Interval:     1,
		SM2EFactor:      2.5,
		FSRSStability:   &initialStab,
		FSRSDifficulty:  &initialDiff,
		FSRSLastReview:  &now,
	}

	// Learner answers 1 review problem correctly in 4 seconds (time threshold is 10s).
	_, _, err = e.applyGradedAttempt(p, gradedAttempt{
		studentID:      st.ID,
		conceptID:      "a",
		requiredStreak: 3,
		timeThreshold:  10.0,
		elapsed:        4.0,
		correct:        true,
	})
	if err != nil {
		t.Fatalf("applyGradedAttempt: %v", err)
	}

	// 1. SM2 Repetitions must advance from 1 to 2
	if p.SM2Repetitions != 2 {
		t.Errorf("SM2Repetitions = %d, want 2 (did it reset on review?)", p.SM2Repetitions)
	}

	// 2. SM2 Interval must advance from 1 to at least 6 days
	if p.SM2Interval < 6 {
		t.Errorf("SM2Interval = %d, want >= 6 (did it reset to 1 day?)", p.SM2Interval)
	}

	// 3. NextReviewDue must be updated into the future (at least 5 days from now)
	if p.NextReviewDue == nil {
		t.Fatalf("NextReviewDue is nil, expected updated future timestamp")
	}
	if p.NextReviewDue.Before(time.Now().UTC().Add(4 * 24 * time.Hour)) {
		t.Errorf("NextReviewDue = %v, expected future date at least 5 days out", p.NextReviewDue)
	}

	// 4. LastReviewed must be updated to current time (within 1 minute)
	if p.LastReviewed == nil || time.Since(*p.LastReviewed) > time.Minute {
		t.Errorf("LastReviewed not updated: %v", p.LastReviewed)
	}
}
