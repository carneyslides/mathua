package engine

import (
	"testing"

	"github.com/chuma-beep/mathua/internal/diagnostic"
	"github.com/chuma-beep/mathua/internal/scheduler"
	"github.com/chuma-beep/mathua/internal/storage"
)

func TestEligibilityFor_FreshLearner(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("placement-learner")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	root, err := e.EligibilityFor(st.ID, "a")
	if err != nil {
		t.Fatalf("eligibility a: %v", err)
	}
	if !root.Eligible || root.State != scheduler.StateUnlocked {
		t.Errorf("a = %+v, want eligible unlocked (no prerequisites)", root)
	}

	child, err := e.EligibilityFor(st.ID, "b")
	if err != nil {
		t.Fatalf("eligibility b: %v", err)
	}
	if child.Eligible || child.State != scheduler.StateLocked {
		t.Errorf("b = %+v, want locked (a unseen)", child)
	}
	if child.Reason != "missing_prerequisites" {
		t.Errorf("b.reason = %q, want missing_prerequisites", child.Reason)
	}
	if len(child.Prerequisites) != 1 || child.Prerequisites[0].Met {
		t.Errorf("b.prerequisites = %+v, want a marked unmet", child.Prerequisites)
	}
}

// A MASTERED row with no review timestamp (legacy or freshly seeded) reads
// as DECAYING, never as fresh mastery: GetProgress applies decay on read,
// and a missing date is the stalest possible signal.
func TestGetProgress_NilLastReviewedReadsDecaying(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("legacy-mastered")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if err := e.repo.UpsertProgress(&storage.ConceptProgress{
		StudentID: st.ID, ConceptID: "a", Status: "MASTERED", Streak: 5,
	}); err != nil {
		t.Fatalf("seed mastery: %v", err)
	}
	got, err := e.GetProgress(st.ID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if got["a"].Status != "DECAYING" {
		t.Errorf("status = %q, want DECAYING", got["a"].Status)
	}
}

// Placement seeds reachability, never mastery: a diagnostic-correct concept
// unlocks its successor for starting while still reporting zero mastered
// concepts. This is the property the whole post-diagnostic flow rests on.
func TestApplyGoalResults_PlacementUnlocksWithoutMastery(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("placed-learner")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	session := &diagnostic.Session{
		StudentID: st.ID,
		Attempts:  []diagnostic.Attempt{{ConceptID: "a", Correct: true, Fast: true}},
	}
	if err := e.ApplyGoalResults(st.ID, session); err != nil {
		t.Fatalf("apply placement: %v", err)
	}

	// The seeded row is placement, not mastery.
	progress, err := e.repo.GetAllProgress(st.ID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	ap := progress["a"]
	if ap == nil || !ap.PlacementSeeded {
		t.Fatalf("a progress = %+v, want PlacementSeeded", ap)
	}
	if ap.Status == "MASTERED" {
		t.Errorf("placement wrote mastery: status = %q", ap.Status)
	}

	// The successor is now startable.
	child, err := e.EligibilityFor(st.ID, "b")
	if err != nil {
		t.Fatalf("eligibility b: %v", err)
	}
	if !child.Eligible || child.State != scheduler.StateUnlocked {
		t.Errorf("b = %+v, want eligible unlocked via placement", child)
	}
}

// Conditionally completed concepts from the diagnostic report seed reachability
// as PRACTICING, unlocking their successors at the frontier. A struggle on the
// successor demotes the provisional prerequisite to LEARNING, which relocks
// dependent progress until remedial practice certifies it.
func TestApplyDiagnosticReport_SuccessorIsEligibleUntilFallbackDemotion(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("placed-diagnostic-learner")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	report := &diagnostic.DiagnosticReport{
		ConditionallyCompleted: []string{"Concept A"},
	}
	if err := e.ApplyDiagnosticReport(st.ID, report); err != nil {
		t.Fatalf("apply diagnostic report: %v", err)
	}

	// Prerequisite a is seeded PRACTICING with placement and provisional flags.
	pa, err := e.repo.GetProgress(st.ID, "a")
	if err != nil || pa == nil {
		t.Fatalf("get progress a: %v", err)
	}
	if pa.Status != "PRACTICING" || !pa.PlacementSeeded || !pa.IsProvisional {
		t.Fatalf("a progress = %+v, want PRACTICING with placement and provisional flags", pa)
	}

	// Successor b is unlocked and startable.
	child, err := e.EligibilityFor(st.ID, "b")
	if err != nil {
		t.Fatalf("eligibility b: %v", err)
	}
	if !child.Eligible || child.State != scheduler.StateUnlocked {
		t.Fatalf("b = %+v, want eligible unlocked via placement", child)
	}

	// Learner struggles on b: provisional fallback demotes prerequisite a.
	demoted, err := e.CheckProvisionalFallback(st.ID, "b")
	if err != nil {
		t.Fatalf("check fallback: %v", err)
	}
	if len(demoted) == 0 || demoted[0] != "a" {
		t.Fatalf("expected demoted a, got %v", demoted)
	}

	// Demoted prerequisite a is now LEARNING with elevated weakness.
	paAfter, err := e.repo.GetProgress(st.ID, "a")
	if err != nil || paAfter == nil {
		t.Fatalf("get progress a after fallback: %v", err)
	}
	if paAfter.Status != "LEARNING" || paAfter.WeaknessScore != 0.8 {
		t.Fatalf("a progress after demotion = %+v, want LEARNING and weakness 0.8", paAfter)
	}

	// Successor b is now relocked until a is remediated.
	childAfter, err := e.EligibilityFor(st.ID, "b")
	if err != nil {
		t.Fatalf("eligibility b after fallback: %v", err)
	}
	if childAfter.Eligible || childAfter.State == scheduler.StateUnlocked {
		t.Fatalf("b after fallback = %+v, want locked due to demoted prerequisite", childAfter)
	}
}
