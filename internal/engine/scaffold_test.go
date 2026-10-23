package engine

import (
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/storage"
)

type varTestGen struct {
	lastDiff      float64
	lastKPOrdinal int
}

func (g *varTestGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	g.lastDiff = ctx.Difficulty
	g.lastKPOrdinal = ctx.KPOrdinal
	return generator.Problem{
		Question:    "Solve this question",
		Answer:      "10",
		Explanation: "Step by step solution.",
	}
}

func TestScaffolding_SequentialProgressionAndMastery(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:          "calc.deriv",
			Label:       "Derivatives",
			Domain:      "calculus",
			GradingType: "numeric",
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         3,
				AvgTimeSeconds: 60,
			},
			KnowledgePoints: []concepts.KnowledgePoint{
				{
					Ordinal:       1,
					Title:         "Power Rule",
					WorkedExample: "d/dx(x^2) = 2x",
				},
				{
					Ordinal:       2,
					Title:         "Constant Multiple Rule",
					WorkedExample: "d/dx(c*f) = c*f'",
				},
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

	reg := generator.NewRegistry()
	reg.Register("calc.deriv", &testGen{answer: "10"})

	eng := New(store, d, reg, nil, nil)
	stu, err := eng.CreateStudent("test_student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Question 1: Knowledge Point 1
	q1, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q1: %v", err)
	}
	if q1.KPOrdinal != 1 {
		t.Fatalf("expected q1 KPOrdinal=1, got %d", q1.KPOrdinal)
	}
	if q1.KnowledgePoint == nil || q1.KnowledgePoint.Title != "Power Rule" {
		t.Fatalf("expected Power Rule knowledge point on q1, got %+v", q1.KnowledgePoint)
	}

	// 1st correct answer on KP 1: advances streak to 1, does not advance KP
	res1, err := eng.SubmitAnswer(sess.ID, stu.ID, q1.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit q1: %v", err)
	}
	if !res1.Correct {
		t.Fatalf("expected correct verdict")
	}
	if res1.KPOrdinal != 1 {
		t.Fatalf("expected res1 KPOrdinal=1, got %d", res1.KPOrdinal)
	}
	if res1.KPAdvanced {
		t.Fatalf("expected KPAdvanced=false on 1st correct answer")
	}
	if res1.NewStatus == mastery.StatusMastered {
		t.Fatalf("concept must not be mastered after 1 answer")
	}

	// Question 2: still Knowledge Point 1
	q2, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q2: %v", err)
	}
	if q2.KPOrdinal != 1 {
		t.Fatalf("expected q2 KPOrdinal=1, got %d", q2.KPOrdinal)
	}

	// 2nd consecutive correct answer on KP 1: advances to KP 2
	res2, err := eng.SubmitAnswer(sess.ID, stu.ID, q2.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit q2: %v", err)
	}
	if !res2.Correct {
		t.Fatalf("expected correct verdict")
	}
	if !res2.KPAdvanced {
		t.Fatalf("expected KPAdvanced=true after 2 consecutive correct answers on KP 1")
	}
	if res2.KPOrdinal != 2 {
		t.Fatalf("expected res2 KPOrdinal=2, got %d", res2.KPOrdinal)
	}
	if res2.NewStatus == mastery.StatusMastered {
		t.Fatalf("concept must not be mastered when KP 2 is incomplete")
	}

	// Question 3: Knowledge Point 2
	q3, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q3: %v", err)
	}
	if q3.KPOrdinal != 2 {
		t.Fatalf("expected q3 KPOrdinal=2, got %d", q3.KPOrdinal)
	}
	if q3.KnowledgePoint == nil || q3.KnowledgePoint.Title != "Constant Multiple Rule" {
		t.Fatalf("expected Constant Multiple Rule on q3, got %+v", q3.KnowledgePoint)
	}

	// Incorrect answer on KP 2: streak resets to 0
	res3, err := eng.SubmitAnswer(sess.ID, stu.ID, q3.AttemptID, "wrong", 5.0)
	if err != nil {
		t.Fatalf("submit q3: %v", err)
	}
	if res3.Correct {
		t.Fatalf("expected incorrect verdict")
	}
	if res3.KPOrdinal != 2 {
		t.Fatalf("expected res3 KPOrdinal=2, got %d", res3.KPOrdinal)
	}
	if res3.KPAdvanced {
		t.Fatalf("expected KPAdvanced=false on miss")
	}

	// Question 4: still on Knowledge Point 2
	q4, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q4: %v", err)
	}
	if q4.KPOrdinal != 2 {
		t.Fatalf("expected q4 KPOrdinal=2, got %d", q4.KPOrdinal)
	}

	// 1st correct answer on KP 2
	res4, err := eng.SubmitAnswer(sess.ID, stu.ID, q4.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit q4: %v", err)
	}
	if !res4.Correct {
		t.Fatalf("expected correct verdict")
	}
	if res4.KPAdvanced {
		t.Fatalf("expected KPAdvanced=false after single correct on KP 2")
	}

	// Question 5: still on Knowledge Point 2
	q5, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q5: %v", err)
	}
	if q5.KPOrdinal != 2 {
		t.Fatalf("expected q5 KPOrdinal=2, got %d", q5.KPOrdinal)
	}

	// 2nd consecutive correct answer on KP 2: completes all KPs and achieves mastery
	res5, err := eng.SubmitAnswer(sess.ID, stu.ID, q5.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit q5: %v", err)
	}
	if !res5.Correct {
		t.Fatalf("expected correct verdict")
	}
	if !res5.KPAdvanced {
		t.Fatalf("expected KPAdvanced=true when final KP completes")
	}
	if res5.NewStatus != mastery.StatusMastered {
		t.Fatalf("expected StatusMastered, got %s", res5.NewStatus)
	}

	prog, err := store.GetProgress(stu.ID, "calc.deriv")
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if prog.Status != string(mastery.StatusMastered) {
		t.Fatalf("expected progress status MASTERED, got %s", prog.Status)
	}
	if prog.MasteredAt == nil {
		t.Fatalf("expected MasteredAt timestamp stamped")
	}
}

func TestScaffolding_BackwardsCompatibility(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:          "arith.add",
			Label:       "Addition",
			Domain:      "arithmetic",
			GradingType: "numeric",
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         3,
				AvgTimeSeconds: 60,
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

	reg := generator.NewRegistry()
	reg.Register("arith.add", &testGen{answer: "42"})

	eng := New(store, d, reg, nil, nil)
	stu, err := eng.CreateStudent("bob")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	q, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get question: %v", err)
	}
	if q.KPOrdinal != 0 {
		t.Fatalf("expected KPOrdinal=0 for legacy concept, got %d", q.KPOrdinal)
	}
	if q.KnowledgePoint != nil {
		t.Fatalf("expected nil KnowledgePoint for legacy concept")
	}

	res, err := eng.SubmitAnswer(sess.ID, stu.ID, q.AttemptID, "42", 5.0)
	if err != nil {
		t.Fatalf("submit answer: %v", err)
	}
	if res.KPOrdinal != 0 || res.KPAdvanced {
		t.Fatalf("expected zero KP values for legacy concept, got ordinal=%d adv=%v", res.KPOrdinal, res.KPAdvanced)
	}
}

func TestScaffolding_VariantKeySelection(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:          "alg.quad",
			Label:       "Quadratics",
			Domain:      "algebra",
			GradingType: "numeric",
			Variants: []concepts.Variant{
				{Difficulty: 0.25, TimeThresh: 30, Label: "easy"},
				{Difficulty: 0.85, TimeThresh: 15, Label: "hard"},
			},
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         2,
				AvgTimeSeconds: 30,
			},
			KnowledgePoints: []concepts.KnowledgePoint{
				{
					Ordinal:       1,
					Title:         "Factoring",
					WorkedExample: "x^2 - 1 = (x-1)(x+1)",
					VariantKey:    "hard",
				},
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

	gen := &varTestGen{}
	reg := generator.NewRegistry()
	reg.Register("alg.quad", gen)

	eng := New(store, d, reg, nil, nil)
	stu, err := eng.CreateStudent("charlie")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	_, err = eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get next question: %v", err)
	}
	if gen.lastDiff != 0.85 {
		t.Fatalf("expected difficulty 0.85 from variant key, got %f", gen.lastDiff)
	}
}

func TestScaffolding_RemedialKeyPrerequisites(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:          "alg.linear",
			Label:       "Linear Equations",
			Domain:      "algebra",
			GradingType: "numeric",
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         1,
				AvgTimeSeconds: 10,
			},
		},
		{
			ID:          "alg.system",
			Label:       "Systems of Equations",
			Domain:      "algebra",
			GradingType: "numeric",
			Prerequisites: []string{
				"alg.linear",
			},
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         2,
				AvgTimeSeconds: 20,
			},
			KnowledgePoints: []concepts.KnowledgePoint{
				{
					Ordinal:          1,
					Title:            "Substitution",
					WorkedExample:    "y = 2x, x + y = 6",
					KeyPrerequisites: []string{"alg.linear"},
				},
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

	reg := generator.NewRegistry()
	reg.Register("alg.linear", &testGen{answer: "10"})
	reg.Register("alg.system", &testGen{answer: "20"})

	eng := New(store, d, reg, nil, nil)
	stu, err := eng.CreateStudent("dana")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	// Mark prerequisite as mastered so system is eligible
	now := time.Now().UTC()
	if err := store.UpsertProgress(&storage.ConceptProgress{
		StudentID:    stu.ID,
		ConceptID:    "alg.linear",
		Status:       string(mastery.StatusMastered),
		MasteredAt:   &now,
		LastReviewed: &now,
	}); err != nil {
		t.Fatalf("seed prereq: %v", err)
	}

	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	q1, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q1: %v", err)
	}
	if q1.ConceptID != "alg.system" {
		t.Fatalf("expected alg.system, got %s", q1.ConceptID)
	}

	// First miss
	_, err = eng.SubmitAnswer(sess.ID, stu.ID, q1.AttemptID, "wrong1", 5.0)
	if err != nil {
		t.Fatalf("submit miss 1: %v", err)
	}

	// Second question served
	q2, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q2: %v", err)
	}

	// Second miss triggers halt and remedial queue
	res2, err := eng.SubmitAnswer(sess.ID, stu.ID, q2.AttemptID, "wrong2", 5.0)
	if err != nil {
		t.Fatalf("submit miss 2: %v", err)
	}
	if !res2.Halted {
		t.Fatalf("expected Halted=true after two consecutive misses")
	}

	foundPrereq := false
	for _, rem := range res2.Remedial {
		if rem == "alg.linear" {
			foundPrereq = true
			break
		}
	}
	if !foundPrereq {
		t.Fatalf("expected alg.linear in remedial list, got %v", res2.Remedial)
	}
}

func TestScaffolding_SessionRehydration(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:          "calc.limit",
			Label:       "Limits",
			Domain:      "calculus",
			GradingType: "numeric",
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         2,
				AvgTimeSeconds: 30,
			},
			KnowledgePoints: []concepts.KnowledgePoint{
				{Ordinal: 1, Title: "Limit Definition", WorkedExample: "lim x->0"},
				{Ordinal: 2, Title: "Limit Laws", WorkedExample: "lim (f+g)"},
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

	reg := generator.NewRegistry()
	reg.Register("calc.limit", &testGen{answer: "10"})

	eng := New(store, d, reg, nil, nil)
	stu, err := eng.CreateStudent("elena")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Question 1 on KP 1
	q1, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q1: %v", err)
	}
	// Answer 1 correct
	_, err = eng.SubmitAnswer(sess.ID, stu.ID, q1.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit 1: %v", err)
	}

	// Question 2 on KP 1
	q2, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q2: %v", err)
	}
	// Answer 2 correct -> advances to KP 2
	res2, err := eng.SubmitAnswer(sess.ID, stu.ID, q2.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit 2: %v", err)
	}
	if !res2.KPAdvanced || res2.KPOrdinal != 2 {
		t.Fatalf("expected advance to KP 2, got adv=%v ord=%d", res2.KPAdvanced, res2.KPOrdinal)
	}

	// Question 3 served on KP 2
	q3, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get q3: %v", err)
	}
	if q3.KPOrdinal != 2 {
		t.Fatalf("expected q3 on KP 2, got %d", q3.KPOrdinal)
	}

	// Simulate server restart or session rehydration with fresh Engine instance
	eng2 := New(store, d, reg, nil, nil)
	currentQ, err := eng2.GetCurrentQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("get current question on rehydrated engine: %v", err)
	}
	if currentQ == nil {
		t.Fatalf("expected non-nil question on rehydration")
	}
	if currentQ.KPOrdinal != 2 {
		t.Fatalf("expected rehydrated question to have KPOrdinal=2, got %d", currentQ.KPOrdinal)
	}
	if currentQ.KnowledgePoint == nil || currentQ.KnowledgePoint.Title != "Limit Laws" {
		t.Fatalf("expected KnowledgePoint Limit Laws on rehydrated question, got %+v", currentQ.KnowledgePoint)
	}
}

func TestScaffolding_EngineThreadsKPOrdinalToGenerator(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{
			ID:          "calc.deriv",
			Label:       "Derivatives",
			Domain:      "calculus",
			GradingType: "numeric",
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         3,
				AvgTimeSeconds: 60,
			},
			KnowledgePoints: []concepts.KnowledgePoint{
				{
					Ordinal:       1,
					Title:         "Power Rule",
					WorkedExample: "d/dx(x^2) = 2x",
				},
				{
					Ordinal:       2,
					Title:         "Constant Multiple Rule",
					WorkedExample: "d/dx(c*f) = c*f'",
				},
			},
		},
		{
			ID:          "arith.add",
			Label:       "Addition",
			Domain:      "arithmetic",
			GradingType: "numeric",
			Prerequisites: []string{
				"calc.deriv",
			},
			MasteryThreshold: concepts.MasteryThreshold{
				Streak:         2,
				AvgTimeSeconds: 30,
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

	calcGen := &varTestGen{}
	arithGen := &varTestGen{}
	reg := generator.NewRegistry()
	reg.Register("calc.deriv", calcGen)
	reg.Register("arith.add", arithGen)

	eng := New(store, d, reg, nil, nil)
	stu, err := eng.CreateStudent("kp_test_student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// 1. First question on calc.deriv: should pass KPOrdinal = 1 to generator
	q1, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("next question 1: %v", err)
	}
	if q1.KPOrdinal != 1 {
		t.Fatalf("expected q1 KPOrdinal=1, got %d", q1.KPOrdinal)
	}
	if calcGen.lastKPOrdinal != 1 {
		t.Fatalf("expected generator to receive KPOrdinal=1, got %d", calcGen.lastKPOrdinal)
	}

	// 2. Answer twice correctly to advance to KP 2
	_, err = eng.SubmitAnswer(sess.ID, stu.ID, q1.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit 1: %v", err)
	}
	q2, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("next question 2: %v", err)
	}
	_, err = eng.SubmitAnswer(sess.ID, stu.ID, q2.AttemptID, "10", 5.0)
	if err != nil {
		t.Fatalf("submit 2: %v", err)
	}

	// Next question should now be on KP 2: generator must receive KPOrdinal = 2
	q3, err := eng.NextQuestion(sess.ID, stu.ID)
	if err != nil {
		t.Fatalf("next question 3: %v", err)
	}
	if q3.KPOrdinal != 2 {
		t.Fatalf("expected q3 KPOrdinal=2, got %d", q3.KPOrdinal)
	}
	if calcGen.lastKPOrdinal != 2 {
		t.Fatalf("expected generator to receive KPOrdinal=2, got %d", calcGen.lastKPOrdinal)
	}

	// 3. For a concept without Knowledge Points, generator must receive KPOrdinal = 0
	stu2, err := eng.CreateStudent("plain_student")
	if err != nil {
		t.Fatalf("create student 2: %v", err)
	}
	sess2, err := store.CreateSession(stu2.ID)
	if err != nil {
		t.Fatalf("create session 2: %v", err)
	}
	// Mark calc.deriv mastered so scheduler serves arith.add
	now := time.Now().UTC()
	err = store.UpsertProgress(&storage.ConceptProgress{
		StudentID:    stu2.ID,
		ConceptID:    "calc.deriv",
		Status:       string(mastery.StatusMastered),
		LastReviewed: &now,
	})
	if err != nil {
		t.Fatalf("upsert progress: %v", err)
	}
	qPlain, err := eng.NextQuestion(sess2.ID, stu2.ID)
	if err != nil {
		t.Fatalf("next plain question: %v", err)
	}
	if qPlain.ConceptID != "arith.add" {
		t.Fatalf("expected arith.add, got %s", qPlain.ConceptID)
	}
	if qPlain.KPOrdinal != 0 {
		t.Fatalf("expected KPOrdinal=0 for concept without KPs, got %d", qPlain.KPOrdinal)
	}
	if arithGen.lastKPOrdinal != 0 {
		t.Fatalf("expected generator to receive KPOrdinal=0 for concept without KPs, got %d", arithGen.lastKPOrdinal)
	}
}
