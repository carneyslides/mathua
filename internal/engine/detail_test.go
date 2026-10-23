package engine

import (
	"testing"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/storage"
)

func newTestEngine(t *testing.T) (*Engine, *concepts.DAG, storage.Repository) {
	t.Helper()
	dag, err := concepts.LoadDir("../../data/concepts")
	if err != nil {
		t.Fatalf("load corpus: %v", err)
	}
	store, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	e := New(store, dag, generator.NewRegistry(), nil, nil)
	return e, dag, store
}

func TestConceptDetail_NotFound(t *testing.T) {
	e, _, _ := newTestEngine(t)
	_, err := e.ConceptDetail("", "nonexistent.concept.id")
	if err == nil {
		t.Fatal("expected error for nonexistent concept, got nil")
	}
}

func TestConceptDetail_RootConcept(t *testing.T) {
	e, _, _ := newTestEngine(t)
	// arith.add.single has no prerequisites; it is a root concept
	detail, err := e.ConceptDetail("", "arith.add.single")
	if err != nil {
		t.Fatalf("ConceptDetail failed: %v", err)
	}
	if detail.Concept.ID != "arith.add.single" {
		t.Errorf("ID = %q, want arith.add.single", detail.Concept.ID)
	}
	if len(detail.Prerequisites) != 0 {
		t.Errorf("expected 0 prerequisites for root concept, got %d", len(detail.Prerequisites))
	}
	if !detail.Unlocked {
		t.Error("expected root concept without prerequisites to be unlocked")
	}
	if len(detail.Dependents) == 0 {
		t.Error("expected root concept to have dependents")
	}
}

func TestConceptDetail_WithPrerequisitesAndProgress(t *testing.T) {
	e, _, store := newTestEngine(t)
	st, err := store.CreateStudent("Detail Student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	// arith.add.double has prerequisite arith.add.single
	detail, err := e.ConceptDetail(st.ID, "arith.add.double")
	if err != nil {
		t.Fatalf("ConceptDetail: %v", err)
	}
	if detail.Unlocked {
		t.Error("expected arith.add.double to be locked when prerequisite is unseen")
	}
	if len(detail.Prerequisites) == 0 {
		t.Fatal("expected prerequisites for arith.add.double")
	}
	if detail.Prerequisites[0].Status != string(mastery.StatusUnseen) {
		t.Errorf("status = %q, want unseen", detail.Prerequisites[0].Status)
	}

	// Now record practicing progress for arith.add.single
	singleProg := &storage.ConceptProgress{
		StudentID: st.ID,
		ConceptID: "arith.add.single",
		Status:    string(mastery.StatusPracticing),
		Streak:    1,
	}
	if err := store.UpsertProgress(singleProg); err != nil {
		t.Fatalf("save progress: %v", err)
	}

	detail, err = e.ConceptDetail(st.ID, "arith.add.double")
	if err != nil {
		t.Fatalf("ConceptDetail: %v", err)
	}
	if detail.Unlocked {
		t.Error("expected concept to remain locked while prerequisite is practicing")
	}
	if detail.Prerequisites[0].Status != string(mastery.StatusPracticing) {
		t.Errorf("status = %q, want practicing", detail.Prerequisites[0].Status)
	}
	if detail.Prerequisites[0].MasteryPct <= 0 {
		t.Errorf("mastery_pct = %f, want > 0", detail.Prerequisites[0].MasteryPct)
	}

	// Now mark arith.add.single as mastered
	singleProg.Status = string(mastery.StatusMastered)
	singleProg.Streak = 10
	if err := store.UpsertProgress(singleProg); err != nil {
		t.Fatalf("save progress: %v", err)
	}

	detail, err = e.ConceptDetail(st.ID, "arith.add.double")
	if err != nil {
		t.Fatalf("ConceptDetail: %v", err)
	}
	if !detail.Unlocked {
		t.Error("expected concept to be unlocked once prerequisite is mastered")
	}

	// Now test progress on the concept itself
	doubleProg := &storage.ConceptProgress{
		StudentID: st.ID,
		ConceptID: "arith.add.double",
		Status:    string(mastery.StatusPracticing),
		Streak:    2,
	}
	if err := store.UpsertProgress(doubleProg); err != nil {
		t.Fatalf("save progress: %v", err)
	}

	detail, err = e.ConceptDetail(st.ID, "arith.add.double")
	if err != nil {
		t.Fatalf("ConceptDetail: %v", err)
	}
	if detail.Progress == nil {
		t.Fatal("expected detail.Progress to be populated")
	}
	if detail.Progress.Streak != 2 {
		t.Errorf("streak = %d, want 2", detail.Progress.Streak)
	}
	if detail.Progress.Status != string(mastery.StatusPracticing) {
		t.Errorf("status = %q, want practicing", detail.Progress.Status)
	}
}

func TestLessonPrerequisites_And_Dependents(t *testing.T) {
	e, _, _ := newTestEngine(t)

	// In a lesson covering arith.add.single and arith.add.double:
	// arith.add.double requires arith.add.single, but because arith.add.single
	// is inside the lesson, it must NOT appear in external prerequisites.
	lessonConcepts := []string{"arith.add.single", "arith.add.double"}
	progress := map[string]*storage.ConceptProgress{}

	prereqs := e.LessonPrerequisites(lessonConcepts, progress)
	for _, p := range prereqs {
		if p.ID == "arith.add.single" || p.ID == "arith.add.double" {
			t.Errorf("lesson prerequisite includes concept from inside lesson: %s", p.ID)
		}
	}

	// Test LessonDependents
	deps := e.LessonDependents([]string{"arith.add.single"}, progress)
	if len(deps) == 0 {
		t.Fatal("expected dependents for arith.add.single")
	}
	foundDouble := false
	for _, d := range deps {
		if d.ID == "arith.add.single" {
			t.Errorf("dependents include self: %s", d.ID)
		}
		if d.ID == "arith.add.double" {
			foundDouble = true
		}
	}
	if !foundDouble {
		t.Error("expected arith.add.double in dependents of arith.add.single")
	}

	// Test LessonDependentsByDomain
	byDomain := e.LessonDependentsByDomain([]string{"arith.add.single"}, progress)
	if len(byDomain) == 0 {
		t.Fatal("expected non-empty domain grouping")
	}
	arithDeps, ok := byDomain["arithmetic"]
	if !ok || len(arithDeps) == 0 {
		t.Errorf("expected dependents under domain 'arithmetic', got %v", byDomain)
	}
}
