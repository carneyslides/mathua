package engine

import (
	"testing"

	"github.com/chuma-beep/mathua/internal/diagnostic"
	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/storage"
)

// A conditional completion is a placement. It has to stay distinguishable from
// demonstrated mastery, because three separate rules read it: the fallback that
// demotes a placed concept when a dependent is missed, the prerequisite gate, and
// the mastery check. Before this repair the row was written MASTERED, the machine's
// hysteresis kept it there, and the first correct answer cleared the flag, so one
// answer turned a diagnostic guess into permanent mastery.
func TestPlacementStaysProvisionalUntilEvidenceCertifies(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("placement_student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	report := &diagnostic.DiagnosticReport{ConditionallyCompleted: []string{"Concept A"}}
	if err := e.ApplyDiagnosticReport(st.ID, report); err != nil {
		t.Fatalf("apply diagnostic report: %v", err)
	}

	p, err := e.repo.GetProgress(st.ID, "a")
	if err != nil || p == nil {
		t.Fatalf("get progress: %v", err)
	}
	if p.Status != string(mastery.StatusPracticing) {
		t.Errorf("status = %q, want PRACTICING: a placement must not read as mastery", p.Status)
	}
	if !p.PlacementSeeded || !p.IsProvisional {
		t.Fatalf("seed flags lost: placement=%v provisional=%v", p.PlacementSeeded, p.IsProvisional)
	}

	difficulty := 0.5
	e.applyGradedAttempt(p, gradedAttempt{
		studentID:      st.ID,
		conceptID:      "a",
		requiredStreak: 3,
		timeThreshold:  60,
		elapsed:        30,
		correct:        true,
		difficulty:     &difficulty,
	})
	if err := e.repo.UpsertProgress(p); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	if !p.IsProvisional {
		t.Error("one correct answer retired the placement; only the ladder's demonstration may")
	}
	if !p.PlacementSeeded {
		t.Error("PlacementSeeded cleared without evidence, so the next diagnostic can reseed it")
	}
}

// The guard in ApplyDiagnosticReport keeps demonstrated mastery, and it reads
// PlacementSeeded to tell a seed from real work. Clearing that flag is what makes
// the guard effective, so it has to be part of the same transition.
func TestDiagnosticDoesNotReseedDemonstratedMastery(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("mastered_student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	mastered := &storage.ConceptProgress{
		StudentID:       st.ID,
		ConceptID:       "a",
		Status:          string(mastery.StatusMastered),
		PlacementSeeded: false,
		IsProvisional:   false,
		SM2EFactor:      2.5,
	}
	if err := e.repo.UpsertProgress(mastered); err != nil {
		t.Fatalf("seed mastered row: %v", err)
	}

	report := &diagnostic.DiagnosticReport{ConditionallyCompleted: []string{"Concept A"}}
	if err := e.ApplyDiagnosticReport(st.ID, report); err != nil {
		t.Fatalf("apply diagnostic report: %v", err)
	}

	p, err := e.repo.GetProgress(st.ID, "a")
	if err != nil || p == nil {
		t.Fatalf("get progress: %v", err)
	}
	if p.IsProvisional || p.PlacementSeeded {
		t.Errorf("demonstrated mastery was reseeded: provisional=%v placement=%v", p.IsProvisional, p.PlacementSeeded)
	}
	if p.Status != string(mastery.StatusMastered) {
		t.Errorf("status = %q, want MASTERED untouched", p.Status)
	}
}
