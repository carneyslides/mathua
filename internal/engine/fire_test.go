package engine

import (
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/repetition"
	"github.com/chuma-beep/mathua/internal/storage"
)

func fireEngine(t *testing.T) (*Engine, storage.Repository) {
	t.Helper()
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:               "prereq",
			Label:            "Prerequisite Concept",
			Domain:           "algebra",
			GradingType:      "numeric",
			Prerequisites:    []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 60},
		},
		{
			ID:               "adv",
			Label:            "Advanced Concept",
			Domain:           "calculus",
			GradingType:      "numeric",
			Prerequisites:    []string{"prereq"},
			Encompasses:      []string{"prereq"},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 60},
		},
		{
			ID:               "standalone",
			Label:            "Standalone Concept",
			Domain:           "geometry",
			GradingType:      "numeric",
			Prerequisites:    []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 60},
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
	reg.Register("prereq", &testGen{answer: "10"})
	reg.Register("adv", &testGen{answer: "20"})
	reg.Register("standalone", &testGen{answer: "30"})

	return New(store, d, reg, nil, nil), store
}

func TestEngine_FIRe_CompletedAttemptAllocatesImplicitCredit(t *testing.T) {
	e, repo := fireEngine(t)
	st, err := e.CreateStudent("bob")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	// Mark prereq and standalone as mastered so adv is the single available concept
	now := time.Now().UTC()
	err = repo.UpsertProgress(&storage.ConceptProgress{
		StudentID:      st.ID,
		ConceptID:      "prereq",
		Status:         string(mastery.StatusMastered),
		Streak:         3,
		LastReviewed:   &now,
		SM2Interval:    10,
		SM2EFactor:     2.5,
		SM2Repetitions: 2,
	})
	if err != nil {
		t.Fatalf("upsert prereq: %v", err)
	}
	err = repo.UpsertProgress(&storage.ConceptProgress{
		StudentID:      st.ID,
		ConceptID:      "standalone",
		Status:         string(mastery.StatusMastered),
		Streak:         3,
		LastReviewed:   &now,
		SM2Interval:    10,
		SM2EFactor:     2.5,
		SM2Repetitions: 2,
	})
	if err != nil {
		t.Fatalf("upsert standalone: %v", err)
	}

	sess, err := repo.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Fetch question on adv and answer correctly
	q, err := e.NextQuestion(sess.ID, st.ID)
	if err != nil {
		t.Fatalf("next question: %v", err)
	}
	if q.ConceptID != "adv" {
		t.Fatalf("expected adv concept, got %s", q.ConceptID)
	}

	res, err := e.SubmitAnswer(sess.ID, st.ID, q.AttemptID, "20", 5.0)
	if err != nil {
		t.Fatalf("submit answer: %v", err)
	}
	if !res.Correct {
		t.Fatalf("expected correct answer")
	}

	// Verify that implicit credit was allocated and banked for prereq
	cs := repo.(storage.CreditStore)
	acc, err := cs.AccountOf(st.ID, "prereq", 0)
	if err != nil {
		t.Fatalf("account of prereq: %v", err)
	}
	if acc.RetainedUnits != repetition.Micro {
		t.Fatalf("expected %d retained units, got %d", repetition.Micro, acc.RetainedUnits)
	}
	if acc.Pending() != 1 {
		t.Fatalf("expected 1 pending banked ordinal, got %d", acc.Pending())
	}
}

func TestEngine_FIRe_BankedCreditSuppressesExplicitReviewAndAdvancesInterval(t *testing.T) {
	e, repo := fireEngine(t)
	st, err := e.CreateStudent("charlie")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	now := time.Now().UTC()
	past := now.AddDate(0, 0, -5)

	// Prereq is due for review
	err = repo.UpsertProgress(&storage.ConceptProgress{
		StudentID:      st.ID,
		ConceptID:      "prereq",
		Status:         string(mastery.StatusMastered),
		Streak:         3,
		LastReviewed:   &past,
		NextReviewDue:  &past,
		SM2Interval:    5,
		SM2EFactor:     2.5,
		SM2Repetitions: 1,
	})
	if err != nil {
		t.Fatalf("upsert prereq: %v", err)
	}

	// Before credit: review is due
	due, err := e.ReviewsDue(st.ID)
	if err != nil {
		t.Fatalf("reviews due: %v", err)
	}
	if due != 1 {
		t.Fatalf("expected 1 due review before credit, got %d", due)
	}

	sess, err := repo.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	rq, err := e.NextReviewQuestion(sess.ID, st.ID)
	if err != nil {
		t.Fatalf("next review: %v", err)
	}
	if rq == nil || rq.ConceptID != "prereq" {
		t.Fatalf("expected prereq review question before credit, got %+v", rq)
	}

	// Reset active session state for next question test
	sess2, err := repo.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("create session 2: %v", err)
	}

	// Bank 1 implicit repetition ordinal for prereq
	cs := repo.(storage.CreditStore)
	_, _, err = cs.AllocateCredit(st.ID, "attempt:fire:1", 0, []storage.CreditCandidate{
		{ConceptID: "prereq", RawUnits: repetition.Micro},
	})
	if err != nil {
		t.Fatalf("allocate credit: %v", err)
	}

	// After banking credit: calling ReviewsDue or NextReviewQuestion satisfies the review
	dueAfter, err := e.ReviewsDue(st.ID)
	if err != nil {
		t.Fatalf("reviews due after credit: %v", err)
	}
	if dueAfter != 0 {
		t.Fatalf("expected 0 due reviews after credit, got %d", dueAfter)
	}

	rqAfter, err := e.NextReviewQuestion(sess2.ID, st.ID)
	if err != nil {
		t.Fatalf("next review after credit: %v", err)
	}
	if rqAfter != nil {
		t.Fatalf("expected nil review question after implicit satisfaction, got %+v", rqAfter)
	}

	// Verify that concept_progress interval has advanced in the database
	p, err := repo.GetProgress(st.ID, "prereq")
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if p.NextReviewDue == nil || !p.NextReviewDue.After(now) {
		t.Fatalf("expected next review due in future, got %v", p.NextReviewDue)
	}
	if p.SM2Interval <= 5 {
		t.Fatalf("expected SM2 interval to advance beyond 5, got %d", p.SM2Interval)
	}

	// Verify that the credit ordinal was projected
	acc, err := cs.AccountOf(st.ID, "prereq", 0)
	if err != nil {
		t.Fatalf("account of prereq: %v", err)
	}
	if acc.Projected != 1 {
		t.Fatalf("expected 1 projected ordinal, got %d", acc.Projected)
	}
	if acc.Pending() != 0 {
		t.Fatalf("expected 0 pending ordinals, got %d", acc.Pending())
	}
}

func TestEngine_FIRe_NonEncompassedConceptIsNotSuppressed(t *testing.T) {
	e, repo := fireEngine(t)
	st, err := e.CreateStudent("diana")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	now := time.Now().UTC()
	past := now.AddDate(0, 0, -5)

	// Standalone concept has no encompasses relationships
	err = repo.UpsertProgress(&storage.ConceptProgress{
		StudentID:      st.ID,
		ConceptID:      "standalone",
		Status:         string(mastery.StatusMastered),
		Streak:         3,
		LastReviewed:   &past,
		NextReviewDue:  &past,
		SM2Interval:    5,
		SM2EFactor:     2.5,
		SM2Repetitions: 1,
	})
	if err != nil {
		t.Fatalf("upsert standalone: %v", err)
	}

	// Bank credit for standalone
	cs := repo.(storage.CreditStore)
	_, _, err = cs.AllocateCredit(st.ID, "attempt:fire:standalone", 0, []storage.CreditCandidate{
		{ConceptID: "standalone", RawUnits: repetition.Micro},
	})
	if err != nil {
		t.Fatalf("allocate credit: %v", err)
	}

	// Since standalone is not encompassed, explicit review must not be suppressed
	due, err := e.ReviewsDue(st.ID)
	if err != nil {
		t.Fatalf("reviews due: %v", err)
	}
	if due != 1 {
		t.Fatalf("expected 1 due review for non-encompassed concept, got %d", due)
	}

	sess, err := repo.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	rq, err := e.NextReviewQuestion(sess.ID, st.ID)
	if err != nil {
		t.Fatalf("next review: %v", err)
	}
	if rq == nil || rq.ConceptID != "standalone" {
		t.Fatalf("expected standalone review question served, got %+v", rq)
	}
}
