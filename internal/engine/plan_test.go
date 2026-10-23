package engine

import (
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/planning"
	"github.com/chuma-beep/mathua/internal/storage"
)

func newTestEngineWithPlanner(t *testing.T) (*Engine, *concepts.DAG, storage.Repository) {
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

	planner, err := planning.Load("../../data/courses.json", dag)
	if err != nil {
		t.Fatalf("load courses: %v", err)
	}
	ds, err := planning.LoadDestinations("../../data/destinations.json")
	if err != nil {
		t.Fatalf("load destinations: %v", err)
	}
	planner.SetDestinations(ds)

	e := New(store, dag, generator.NewRegistry(), nil, planner)
	return e, dag, store
}

func TestPaceFor(t *testing.T) {
	e, _, store := newTestEngineWithPlanner(t)
	st, err := store.CreateStudent("Pace Student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	// No activity yet -> returns goal pace
	pace, err := e.PaceFor(st.ID, 20)
	if err != nil {
		t.Fatalf("PaceFor: %v", err)
	}
	if pace.Source != "goal" {
		t.Errorf("pace.Source = %q, want goal", pace.Source)
	}
	if pace.Rate != 20.0 {
		t.Errorf("pace.Rate = %f, want 20.0", pace.Rate)
	}
	if pace.ActiveDays != 0 {
		t.Errorf("pace.ActiveDays = %d, want 0", pace.ActiveDays)
	}

	sess, err := store.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// Record an attempt today so there is daily activity
	attempt := storage.AttemptEntry{
		SessionID: sess.ID,
		StudentID: st.ID,
		ConceptID: "arith.add.single",
		Correct:   true,
		Answer:    "2",
		Timestamp: time.Now().UTC(),
	}
	if _, err := store.RecordAttempt(attempt); err != nil {
		t.Fatalf("RecordAttempt: %v", err)
	}

	pace, err = e.PaceFor(st.ID, 20)
	if err != nil {
		t.Fatalf("PaceFor after activity: %v", err)
	}
	if pace.Source != "measured" {
		t.Errorf("pace.Source = %q, want measured", pace.Source)
	}
	if pace.ActiveDays != 1 {
		t.Errorf("pace.ActiveDays = %d, want 1", pace.ActiveDays)
	}
	if pace.Rate <= 0 {
		t.Errorf("expected positive measured pace rate, got %f", pace.Rate)
	}
}

func TestAttemptStatsFor(t *testing.T) {
	e, _, store := newTestEngineWithPlanner(t)
	st, err := store.CreateStudent("Stats Student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	sess, err := store.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// 1 correct attempt
	_, _ = store.RecordAttempt(storage.AttemptEntry{
		SessionID: sess.ID,
		StudentID: st.ID,
		ConceptID: "arith.add.single",
		Correct:   true,
		Answer:    "4",
		Timestamp: time.Now().UTC(),
	})
	// 1 failed attempt with answer
	_, _ = store.RecordAttempt(storage.AttemptEntry{
		SessionID: sess.ID,
		StudentID: st.ID,
		ConceptID: "arith.add.single",
		Correct:   false,
		Answer:    "5",
		Timestamp: time.Now().UTC(),
	})
	// 1 dont-know attempt (incorrect with empty answer)
	_, _ = store.RecordAttempt(storage.AttemptEntry{
		SessionID: sess.ID,
		StudentID: st.ID,
		ConceptID: "arith.add.single",
		Correct:   false,
		Answer:    "",
		Timestamp: time.Now().UTC(),
	})
	// 1 diagnostic attempt (should be ignored by AttemptStatsFor)
	_, _ = store.RecordAttempt(storage.AttemptEntry{
		SessionID: sess.ID,
		StudentID: st.ID,
		ConceptID: "arith.add.single",
		Source:    "diagnostic",
		Correct:   false,
		Answer:    "9",
		Timestamp: time.Now().UTC(),
	})

	stats, err := e.AttemptStatsFor(st.ID)
	if err != nil {
		t.Fatalf("AttemptStatsFor: %v", err)
	}

	s, ok := stats["arith.add.single"]
	if !ok {
		t.Fatal("expected stats for arith.add.single")
	}
	if s.Correct != 1 {
		t.Errorf("Correct = %d, want 1", s.Correct)
	}
	if s.Failed != 1 {
		t.Errorf("Failed = %d, want 1", s.Failed)
	}
	if s.DontKnow != 1 {
		t.Errorf("DontKnow = %d, want 1", s.DontKnow)
	}
}

func TestDestinations_And_PlanPersistence(t *testing.T) {
	e, _, store := newTestEngineWithPlanner(t)
	st, err := store.CreateStudent("Plan Student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	// Destinations list
	dests, err := e.Destinations(st.ID)
	if err != nil {
		t.Fatalf("Destinations: %v", err)
	}
	if len(dests) == 0 {
		t.Fatal("expected non-empty destinations list")
	}

	firstDest := dests[0]
	if firstDest.Total <= 0 {
		t.Errorf("first destination %s has total = %d, want > 0", firstDest.ID, firstDest.Total)
	}

	// Save plan (persists both choices and baseline snapshot)
	snap, err := e.SavePlan(st.ID, firstDest.ID, 20, 30, 1)
	if err != nil {
		t.Fatalf("SavePlan: %v", err)
	}
	if snap == nil || snap.Destination != firstDest.ID {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}

	loadedPrefs, ok := e.GetPlanPrefs(st.ID)
	if !ok {
		t.Fatal("expected GetPlanPrefs to return true")
	}
	if loadedPrefs.Destination != firstDest.ID || loadedPrefs.DeadlineDays != 30 || loadedPrefs.RestDays != 1 {
		t.Errorf("loadedPrefs mismatch: %+v", loadedPrefs)
	}

	loadedSnap, ok := e.GetPlan(st.ID)
	if !ok {
		t.Fatal("expected GetPlan to return true")
	}
	if loadedSnap.Destination != firstDest.ID {
		t.Errorf("loadedSnap mismatch: %+v", loadedSnap)
	}

	// DestinationEstimate
	est, err := e.DestinationEstimate(st.ID, firstDest.ID, 15, 30, 1, 0.0)
	if err != nil {
		t.Fatalf("DestinationEstimate: %v", err)
	}
	if est.Destination != firstDest.ID {
		t.Errorf("Destination = %q, want %q", est.Destination, firstDest.ID)
	}
	if est.Estimate == nil {
		t.Fatal("expected non-nil WorkloadEstimate")
	}
	if len(est.Probes) == 0 {
		t.Error("expected probes sample in estimate response")
	}
}
