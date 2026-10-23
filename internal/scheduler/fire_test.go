package scheduler

import (
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/mastery"
)

func fireDAG(t *testing.T) *concepts.DAG {
	t.Helper()
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:               "prereq",
			Label:            "Prerequisite Topic",
			Domain:           "algebra",
			Subdomain:        "foundations",
			Prerequisites:    []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		},
		{
			ID:               "adv",
			Label:            "Advanced Topic",
			Domain:           "calculus",
			Subdomain:        "derivatives",
			Prerequisites:    []string{"prereq"},
			Encompasses:      []string{"prereq"},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		},
		{
			ID:               "standalone",
			Label:            "Standalone Topic",
			Domain:           "geometry",
			Subdomain:        "shapes",
			Prerequisites:    []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		},
	})
	if err != nil {
		t.Fatalf("build fire DAG: %v", err)
	}
	return d
}

func TestFIRe_BankedImplicitCreditSuppressesExplicitReview(t *testing.T) {
	d := fireDAG(t)
	s := New(d)
	now := time.Now().UTC()
	past := now.AddDate(0, 0, -5)

	// Concept prereq is mastered but due for explicit review
	snapWithoutCredit := map[string]*ConceptSnapshot{
		"prereq": {
			Status:         mastery.StatusMastered,
			LastReviewed:   past,
			NextReviewDue:  &past,
			RequiredStreak: 3,
			Interval:       5,
			EFactor:        2.5,
			BankedOrdinals: 0,
		},
		"adv": {
			Status:         mastery.StatusLearning,
			RequiredStreak: 3,
		},
	}

	// Without banked credit, prereq must be selected as a due review
	reviewWithout := s.NextReview(snapWithoutCredit, "")
	if reviewWithout == nil || reviewWithout.Concept.ID != "prereq" {
		t.Fatalf("expected prereq due for review without credit, got %+v", reviewWithout)
	}

	// Now test with 1 banked implicit repetition ordinal
	snapWithCredit := map[string]*ConceptSnapshot{
		"prereq": {
			Status:         mastery.StatusMastered,
			LastReviewed:   past,
			NextReviewDue:  &past,
			RequiredStreak: 3,
			Interval:       5,
			EFactor:        2.5,
			BankedOrdinals: 1,
			LearningSpeed:  1.0,
		},
		"adv": {
			Status:         mastery.StatusLearning,
			RequiredStreak: 3,
		},
	}

	// Banked implicit credit must suppress redundant review
	reviewWith := s.NextReview(snapWithCredit, "")
	if reviewWith != nil {
		t.Fatalf("expected review suppressed by implicit credit, got concept %s", reviewWith.Concept.ID)
	}

	// The interval must have advanced and the ordinal must be consumed
	prereqSnap := snapWithCredit["prereq"]
	if prereqSnap.BankedOrdinals != 0 {
		t.Fatalf("expected banked ordinals decremented to 0, got %d", prereqSnap.BankedOrdinals)
	}
	if prereqSnap.NextReviewDue == nil || !prereqSnap.NextReviewDue.After(now) {
		t.Fatalf("expected next review due pushed into future, got %v", prereqSnap.NextReviewDue)
	}
	if prereqSnap.Interval <= 5 {
		t.Fatalf("expected interval to advance beyond 5 days, got %d", prereqSnap.Interval)
	}
}

func TestFIRe_NextSmart_SuppressesEncompassedReviewTask(t *testing.T) {
	d := fireDAG(t)
	s := New(d)
	now := time.Now().UTC()
	past := now.AddDate(0, 0, -20)

	snap := map[string]*ConceptSnapshot{
		"prereq": {
			Status:         mastery.StatusMastered,
			LastReviewed:   past,
			NextReviewDue:  &past,
			RequiredStreak: 3,
			Interval:       6,
			EFactor:        2.5,
			BankedOrdinals: 1,
			LearningSpeed:  1.0,
		},
		"adv": {
			Status:         mastery.StatusLearning,
			RequiredStreak: 3,
		},
	}

	// NextSmart with high reviewCount to trigger review preference
	cands := s.NextSmart(snap, nil, 10, 0, nil)
	for _, c := range cands {
		if c.Concept.ID == "prereq" && c.IsReview {
			t.Fatalf("expected prereq review suppressed in NextSmart, but got review candidate")
		}
	}
}

func TestFIRe_ComputePriority_SuppressesDecayBonus(t *testing.T) {
	now := time.Now().UTC()
	past := now.AddDate(0, 0, -15)

	snapWithCredit := &ConceptSnapshot{
		Status:         mastery.StatusMastered,
		LastReviewed:   past,
		LastAttempted:  past,
		NextReviewDue:  &past,
		RequiredStreak: 3,
		BankedOrdinals: 1,
		Encompassed:    true,
	}

	snapWithoutCredit := &ConceptSnapshot{
		Status:         mastery.StatusMastered,
		LastReviewed:   past,
		LastAttempted:  past,
		NextReviewDue:  &past,
		RequiredStreak: 3,
		BankedOrdinals: 0,
		Encompassed:    true,
	}

	pWith := computePriority(snapWithCredit, now, true)
	pWithout := computePriority(snapWithoutCredit, now, true)

	diff := pWithout - pWith
	if diff < 4.99 || diff > 5.01 {
		t.Fatalf("expected 5.0 decay priority difference when credit suppresses review, got %f", diff)
	}
}

func TestFIRe_BackwardsCompatibility_NonEncompassedConcept(t *testing.T) {
	d := fireDAG(t)
	s := New(d)
	now := time.Now().UTC()
	past := now.AddDate(0, 0, -10)

	// Concept standalone has no encompassing parents in the DAG
	snap := map[string]*ConceptSnapshot{
		"standalone": {
			Status:         mastery.StatusMastered,
			LastReviewed:   past,
			NextReviewDue:  &past,
			RequiredStreak: 3,
			Interval:       3,
			EFactor:        2.5,
			BankedOrdinals: 1,
			Encompassed:    false,
		},
	}

	// Since standalone is not encompassed, implicit credit must not suppress review
	review := s.NextReview(snap, "")
	if review == nil || review.Concept.ID != "standalone" {
		t.Fatalf("expected non-encompassed concept to remain due for review, got %+v", review)
	}
}

func TestFIRe_ChainedOrdinalsAdvance(t *testing.T) {
	d := fireDAG(t)
	s := New(d)
	now := time.Now().UTC()
	past := now.AddDate(0, 0, -10)

	snap := map[string]*ConceptSnapshot{
		"prereq": {
			Status:         mastery.StatusMastered,
			LastReviewed:   past,
			NextReviewDue:  &past,
			RequiredStreak: 3,
			Interval:       4,
			EFactor:        2.5,
			BankedOrdinals: 2,
			LearningSpeed:  1.0,
		},
		"adv": {
			Status:         mastery.StatusLearning,
			RequiredStreak: 3,
		},
	}

	// First cycle: consumes 1 ordinal
	s.NextReview(snap, "")
	prereq := snap["prereq"]
	if prereq.BankedOrdinals != 1 {
		t.Fatalf("expected 1 ordinal remaining, got %d", prereq.BankedOrdinals)
	}
	firstDue := *prereq.NextReviewDue
	if !firstDue.After(now) {
		t.Fatalf("expected first due date in future, got %v", firstDue)
	}

	// Simulate time passing until the advanced review becomes due again
	secondPast := now.Add(-time.Hour)
	snap["prereq"].NextReviewDue = &secondPast
	snap["prereq"].LastReviewed = secondPast

	// Second cycle: consumes second ordinal and advances further
	s.NextReview(snap, "")
	if prereq.BankedOrdinals != 0 {
		t.Fatalf("expected 0 ordinals remaining after second cycle, got %d", prereq.BankedOrdinals)
	}
	if !prereq.NextReviewDue.After(firstDue) {
		t.Fatalf("expected second due date after first due date, got %v <= %v", prereq.NextReviewDue, firstDue)
	}
}

func TestFIRe_EligibilityStateRemainsMastered(t *testing.T) {
	d := fireDAG(t)
	now := time.Now().UTC()
	past := now.AddDate(0, 0, -15)

	snapWithCredit := map[string]*ConceptSnapshot{
		"prereq": {
			Status:         mastery.StatusMastered,
			LastReviewed:   past,
			NextReviewDue:  &past,
			RequiredStreak: 3,
			BankedOrdinals: 1,
			Encompassed:    true,
		},
	}

	c := d.Concept("prereq")
	el := EligibilityOf(snapWithCredit, now, c)
	if el.State != StateMastered {
		t.Fatalf("expected StateMastered when credit suppresses review, got %s", el.State)
	}

	snapWithoutCredit := map[string]*ConceptSnapshot{
		"prereq": {
			Status:         mastery.StatusMastered,
			LastReviewed:   past,
			NextReviewDue:  &past,
			RequiredStreak: 3,
			BankedOrdinals: 0,
			Encompassed:    true,
		},
	}
	elNoCredit := EligibilityOf(snapWithoutCredit, now, c)
	if elNoCredit.State != StateDueForReview {
		t.Fatalf("expected StateDueForReview without credit, got %s", elNoCredit.State)
	}
}
