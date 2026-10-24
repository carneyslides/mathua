package engine

import (
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/storage"
)

func TestFSRSDurability_DecayDecoupledFromFourteenDays(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("fsrs-durable")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	now := time.Now().UTC()
	stability := 45.0
	difficulty := 4.0
	lastRev := now.AddDate(0, 0, -20)
	due := now.AddDate(0, 0, 25)

	p := &storage.ConceptProgress{
		StudentID:       st.ID,
		ConceptID:       "a",
		Status:          string(mastery.StatusMastered),
		Streak:          5,
		BestStreak:      5,
		AvgResponseTime: 8.0,
		Attempts:        8,
		LastReviewed:    &lastRev,
		NextReviewDue:   &due,
		FSRSStability:   &stability,
		FSRSDifficulty:  &difficulty,
		FSRSLastReview:  &lastRev,
	}

	if err := e.repo.UpsertProgress(p); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}

	// 20 days elapsed is past the old 14-day cutoff.
	// But stability S=45 yields retrievability > 0.90, so it must not decay.
	progressMap, err := e.GetProgress(st.ID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	got := progressMap["a"]
	if got == nil {
		t.Fatal("expected progress for concept a")
	}
	if got.Status != string(mastery.StatusMastered) {
		t.Errorf("status = %q, want %q at 20 days for S=45", got.Status, mastery.StatusMastered)
	}

	// Longitudinal progression: move clock to 50 days (past S=45).
	// Retrievability drops below 0.90 and the concept must now read as DECAYING.
	staleRev := now.AddDate(0, 0, -50)
	p.FSRSLastReview = &staleRev
	p.LastReviewed = &staleRev
	if err := e.repo.UpsertProgress(p); err != nil {
		t.Fatalf("upsert progress stale: %v", err)
	}

	progressMapStale, err := e.GetProgress(st.ID)
	if err != nil {
		t.Fatalf("get progress stale: %v", err)
	}
	gotStale := progressMapStale["a"]
	if gotStale == nil {
		t.Fatal("expected progress for concept a")
	}
	if gotStale.Status != "DECAYING" {
		t.Errorf("status = %q, want DECAYING at 50 days for S=45", gotStale.Status)
	}
}

func TestFSRSDurability_LowStabilityDecaysPromptly(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("fsrs-low-stability")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	now := time.Now().UTC()
	stability := 3.0
	difficulty := 7.0
	lastRev := now.AddDate(0, 0, -5)
	due := now.AddDate(0, 0, -2)

	p := &storage.ConceptProgress{
		StudentID:       st.ID,
		ConceptID:       "a",
		Status:          string(mastery.StatusMastered),
		Streak:          3,
		BestStreak:      3,
		AvgResponseTime: 12.0,
		Attempts:        5,
		LastReviewed:    &lastRev,
		NextReviewDue:   &due,
		FSRSStability:   &stability,
		FSRSDifficulty:  &difficulty,
		FSRSLastReview:  &lastRev,
	}

	if err := e.repo.UpsertProgress(p); err != nil {
		t.Fatalf("upsert progress: %v", err)
	}

	// At 5 days (under the old 14-day threshold), S=3 has already fallen below retention 0.90.
	// It must decay promptly so the learner receives a timely review.
	progressMap, err := e.GetProgress(st.ID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	got := progressMap["a"]
	if got == nil {
		t.Fatal("expected progress for concept a")
	}
	if got.Status != "DECAYING" {
		t.Errorf("status = %q, want DECAYING at 5 days for low stability S=3", got.Status)
	}
}

func TestRemedialHalting_FallbackToWeakestPrerequisiteWhenNoKeyPrerequisites(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{
			ID: "base1", Label: "Base 1", Domain: "math",
			GradingType: "numeric", Prerequisites: []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 2, AvgTimeSeconds: 30},
		},
		{
			ID: "base2", Label: "Base 2", Domain: "math",
			GradingType: "numeric", Prerequisites: []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 2, AvgTimeSeconds: 30},
		},
		{
			ID: "target", Label: "Target Concept", Domain: "math",
			GradingType:   "numeric",
			Prerequisites: []string{"base1", "base2"},
			// Deliberately no KeyPrerequisites defined.
			Variants: []concepts.Variant{
				{Difficulty: 0.2, TimeThresh: 30.0, Label: "easy"},
			},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 2, AvgTimeSeconds: 30},
		},
	})
	if err != nil {
		t.Fatalf("build DAG: %v", err)
	}

	store, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	reg := generator.NewRegistry()
	reg.Register("base1", &testGen{answer: "10"})
	reg.Register("base2", &testGen{answer: "20"})
	reg.Register("target", &testGen{answer: "30"})

	eng := New(store, d, reg, nil, nil)
	st, err := eng.CreateStudent("fallback-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := eng.repo.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// base1 is mastered with streak 4.
	// base2 is mastered with streak 1 (weaker prerequisite).
	now := time.Now().UTC()
	if err := eng.repo.UpsertProgress(&storage.ConceptProgress{
		StudentID: st.ID, ConceptID: "base1",
		Status: string(mastery.StatusMastered), Streak: 4,
		LastAttempted: &now, LastReviewed: &now,
	}); err != nil {
		t.Fatalf("upsert base1: %v", err)
	}
	if err := eng.repo.UpsertProgress(&storage.ConceptProgress{
		StudentID: st.ID, ConceptID: "base2",
		Status: string(mastery.StatusMastered), Streak: 1,
		LastAttempted: &now, LastReviewed: &now,
	}); err != nil {
		t.Fatalf("upsert base2: %v", err)
	}

	// target is now unlocked and served.
	q1, err := eng.NextQuestion(sess.ID, st.ID)
	if err != nil || q1 == nil {
		t.Fatalf("next question 1: %v", err)
	}
	if q1.ConceptID != "target" {
		t.Fatalf("expected target concept, got %s", q1.ConceptID)
	}

	// Miss 1
	res1, err := eng.SubmitAnswer(sess.ID, st.ID, q1.AttemptID, "wrong", 5.0)
	if err != nil {
		t.Fatalf("submit miss 1: %v", err)
	}
	if res1.Halted {
		t.Error("expected no halt after first miss")
	}

	q2, err := eng.NextQuestion(sess.ID, st.ID)
	if err != nil || q2 == nil {
		t.Fatalf("next question 2: %v", err)
	}
	if q2.ConceptID != "target" {
		t.Fatalf("expected target concept on second question, got %s", q2.ConceptID)
	}

	// Miss 2 triggers 2-miss halt
	res2, err := eng.SubmitAnswer(sess.ID, st.ID, q2.AttemptID, "wrong", 5.0)
	if err != nil {
		t.Fatalf("submit answer miss 2: %v", err)
	}
	if !res2.Halted {
		t.Fatal("expected session to halt after 2 misses")
	}

	// Weakest prerequisite is base2 (streak 1 vs base1 streak 4).
	// Verify base2 is queued for targeted remediation.
	remQueue := eng.RemedialQueue(st.ID)
	if len(remQueue) == 0 {
		t.Fatal("expected remedial queue to be populated by fallback")
	}
	if remQueue[0] != "base2" {
		t.Errorf("expected remedial target base2, got %v", remQueue)
	}

	// Next question served must be base2.
	q3, err := eng.NextQuestion(sess.ID, st.ID)
	if err != nil || q3 == nil {
		t.Fatalf("next question 3: %v", err)
	}
	if q3.ConceptID != "base2" {
		t.Errorf("served concept %s, want weakest prerequisite base2", q3.ConceptID)
	}
}

func TestRemedialHalting_DeadlockFreeOnRootConcept(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("deadlock-free")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := e.repo.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	q1, _ := e.NextQuestion(sess.ID, st.ID)
	_, _ = e.SubmitAnswer(sess.ID, st.ID, q1.AttemptID, "wrong", 4.0)

	q2, _ := e.NextQuestion(sess.ID, st.ID)
	res2, _ := e.SubmitAnswer(sess.ID, st.ID, q2.AttemptID, "wrong", 4.0)
	if !res2.Halted {
		t.Fatal("expected halt on root concept")
	}

	// Root concept has no prerequisites in DAG. It must re-queue itself cleanly without hanging.
	q3, err := e.NextQuestion(sess.ID, st.ID)
	if err != nil || q3 == nil {
		t.Fatal("expected question after root concept halt")
	}
	if q3.ConceptID != "a" {
		t.Errorf("served concept = %s, want a", q3.ConceptID)
	}
}
