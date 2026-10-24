package engine

import (
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/storage"
)

type mockScaffoldGen struct {
	steps []generator.ScaffoldStep
}

func (g *mockScaffoldGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	return generator.Problem{
		Question:      "Find the eigenvalues of matrix A.",
		Answer:        "1,4",
		Explanation:   "Characteristic equation yields roots 1 and 4.",
		ScaffoldSteps: g.steps,
	}
}

func (g *mockScaffoldGen) ScaffoldSteps(ctx generator.GeneratorContext) []generator.ScaffoldStep {
	return g.steps
}

func TestCognitiveScaffolding_InteractiveStepDownAndPrerequisiteAttribution(t *testing.T) {
	dag, err := concepts.Build([]concepts.Concept{
		{
			ID:          "linalg.det.2x2",
			Label:       "2x2 Determinant",
			Domain:      "linalg",
			GradingType: "numeric",
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         2,
				AvgTimeSeconds: 30,
			},
		},
		{
			ID:          "alg.quad.formula",
			Label:       "Quadratic Formula",
			Domain:      "algebra",
			GradingType: "numeric",
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         2,
				AvgTimeSeconds: 30,
			},
		},
		{
			ID:          "linalg.eigen.compute",
			Label:       "Compute Eigenvalues",
			Domain:      "linalg",
			GradingType: "tuple",
			Prerequisites: []string{
				"linalg.det.2x2",
				"alg.quad.formula",
			},
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         2,
				AvgTimeSeconds: 45,
			},
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

	scaffoldSteps := []generator.ScaffoldStep{
		{
			Ordinal:       1,
			Prompt:        "Compute the determinant of the 2x2 matrix.",
			Answer:        "4",
			Explanation:   "det(A) = 4",
			PrereqConcept: "linalg.det.2x2",
		},
		{
			Ordinal:       2,
			Prompt:        "Compute the smaller root of the characteristic polynomial.",
			Answer:        "1",
			Explanation:   "Roots are 1 and 4, smaller root is 1.",
			PrereqConcept: "alg.quad.formula",
		},
	}

	reg := generator.NewRegistry()
	reg.Register("linalg.det.2x2", &testGen{answer: "4"})
	reg.Register("alg.quad.formula", &testGen{answer: "1"})
	reg.Register("linalg.eigen.compute", &mockScaffoldGen{steps: scaffoldSteps})

	eng := New(store, dag, reg, nil, nil)
	stu, err := eng.CreateStudent("scaffold_tester")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	// Seed prerequisites as Mastered
	now := time.Now().UTC()
	err = store.UpsertProgress(&storage.ConceptProgress{
		StudentID:    stu.ID,
		ConceptID:    "linalg.det.2x2",
		Status:       string(mastery.StatusMastered),
		Streak:       3,
		LastReviewed: &now,
	})
	if err != nil {
		t.Fatalf("upsert prereq 1: %v", err)
	}
	err = store.UpsertProgress(&storage.ConceptProgress{
		StudentID:    stu.ID,
		ConceptID:    "alg.quad.formula",
		Status:       string(mastery.StatusMastered),
		Streak:       3,
		LastReviewed: &now,
	})
	if err != nil {
		t.Fatalf("upsert prereq 2: %v", err)
	}

	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// 1. Get the parent composite problem
	q1, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("next question 1: %v", err)
	}
	if q1.ConceptID != "linalg.eigen.compute" {
		t.Fatalf("expected linalg.eigen.compute, got %s", q1.ConceptID)
	}

	// 2. Submit wrong answer on composite problem: triggers active step-down!
	res1, err := eng.SubmitAnswer(sess.ID, stu.ID, q1.AttemptID, "wrong_answer", 10.0)
	if err != nil {
		t.Fatalf("submit answer 1: %v", err)
	}
	if res1.Correct {
		t.Fatalf("expected incorrect verdict")
	}
	if !res1.ScaffoldActive {
		t.Fatalf("expected ScaffoldActive=true after missing composite problem")
	}
	if res1.ScaffoldStep != 1 || res1.ScaffoldTotal != 2 {
		t.Fatalf("expected step 1 of 2, got %d of %d", res1.ScaffoldStep, res1.ScaffoldTotal)
	}

	// 3. Next question should serve Subgoal 1
	subgoalQ1, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("next question subgoal 1: %v", err)
	}
	if !subgoalQ1.ScaffoldActive {
		t.Fatalf("expected subgoal question to have ScaffoldActive=true")
	}
	if subgoalQ1.ScaffoldStep != 1 {
		t.Fatalf("expected Subgoal 1, got %d", subgoalQ1.ScaffoldStep)
	}

	// 4. Fail Subgoal 1: verifies Prerequisite Attribution Invariant
	resSub1, err := eng.SubmitAnswer(sess.ID, stu.ID, subgoalQ1.AttemptID, "wrong_det", 8.0)
	if err != nil {
		t.Fatalf("submit subgoal 1: %v", err)
	}
	if resSub1.Correct {
		t.Fatalf("expected incorrect on subgoal 1")
	}

	// Check that prerequisite linalg.det.2x2 was demoted and enqueued
	prereqProg, err := store.GetProgress(stu.ID, "linalg.det.2x2")
	if err != nil {
		t.Fatalf("get prereq progress: %v", err)
	}
	if prereqProg.Streak != 0 {
		t.Fatalf("expected prereq streak reset to 0, got %d", prereqProg.Streak)
	}
	if prereqProg.Status != string(mastery.StatusPracticing) {
		t.Fatalf("expected prereq status demoted to PRACTICING, got %s", prereqProg.Status)
	}

	// Verify prerequisite is in remedial queue
	remQueue, err := store.GetRemedialQueue(stu.ID)
	if err != nil {
		t.Fatalf("get remedial queue: %v", err)
	}
	foundPrereq := false
	for _, id := range remQueue {
		if id == "linalg.det.2x2" {
			foundPrereq = true
			break
		}
	}
	if !foundPrereq {
		t.Fatalf("expected linalg.det.2x2 in remedial queue, got %v", remQueue)
	}

	// 5. Answer Subgoal 1 correctly on retry/re-serve
	subgoalQ1Retry, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("next question subgoal 1 retry: %v", err)
	}
	resSub1Retry, err := eng.SubmitAnswer(sess.ID, stu.ID, subgoalQ1Retry.AttemptID, "4", 5.0)
	if err != nil {
		t.Fatalf("submit subgoal 1 retry: %v", err)
	}
	if !resSub1Retry.Correct {
		t.Fatalf("expected correct on subgoal 1 retry")
	}
	if !resSub1Retry.ScaffoldActive || resSub1Retry.ScaffoldStep != 2 {
		t.Fatalf("expected advancing to Subgoal 2, got active=%v step=%d", resSub1Retry.ScaffoldActive, resSub1Retry.ScaffoldStep)
	}

	// 6. Serve and answer Subgoal 2 correctly
	subgoalQ2, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("next question subgoal 2: %v", err)
	}
	resSub2, err := eng.SubmitAnswer(sess.ID, stu.ID, subgoalQ2.AttemptID, "1", 5.0)
	if err != nil {
		t.Fatalf("submit subgoal 2: %v", err)
	}
	if !resSub2.Correct {
		t.Fatalf("expected correct on subgoal 2")
	}
	// Scaffolding complete!
	if resSub2.ScaffoldActive {
		t.Fatalf("expected ScaffoldActive=false after completing all subgoals")
	}
}
