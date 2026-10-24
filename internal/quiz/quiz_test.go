package quiz

import (
	"testing"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
)

type fakeGen struct {
	difficulty float64
}

func (f *fakeGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	f.difficulty = ctx.Difficulty
	return generator.Problem{Question: "q?", Answer: "a", Explanation: "e"}
}

func miniDAG(t *testing.T) *concepts.DAG {
	t.Helper()
	d, err := concepts.Build([]concepts.Concept{
		{ID: "a", Label: "Concept A", Domain: "d", Subdomain: "s1", Prerequisites: []string{}, MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
		{ID: "b", Label: "Concept B", Domain: "d", Subdomain: "s2", Prerequisites: []string{}, MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
		{ID: "c", Label: "Concept C", Domain: "d", Subdomain: "s3", Prerequisites: []string{}, MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return d
}

func TestEngine_NextQuestion_UsesDifficultyMap(t *testing.T) {
	d := miniDAG(t)
	g := &fakeGen{}
	r := generator.NewRegistry()
	for _, id := range []string{"a", "b", "c"} {
		if err := r.Register(id, g); err != nil {
			t.Fatalf("register: %v", err)
		}
	}
	e := NewEngine(d, r)
	sess := e.Start(d.Order())
	// Weak concept "a" → easy difficulty 0.35 (per weakness map); strong "b" → 0.8.
	sess.Difficulties["a"] = 0.35
	sess.Difficulties["b"] = 0.8
	_, cid, err := e.NextQuestion(sess)
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if cid == "" {
		t.Fatal("expected a question")
	}
	if g.difficulty != 0.35 {
		t.Errorf("expected difficulty 0.35 for weak a, got %f", g.difficulty)
	}
	// Fallback when no map entry: 0.8 target.
	sess.Order = []*concepts.Concept{d.Concept("b")}
	sess.Index = 0
	if _, _, err := e.NextQuestion(sess); err != nil {
		t.Fatalf("next b: %v", err)
	}
	if g.difficulty != 0.8 {
		t.Errorf("expected fallback difficulty 0.8, got %f", g.difficulty)
	}
}

func TestPickQuizConcepts_DiverseSubdomains(t *testing.T) {
	d := miniDAG(t)
	picked := PickQuizConcepts(d, map[string]float64{"a": 0.8, "b": 0.7, "c": 0.6})
	if len(picked) != 3 {
		t.Fatalf("expected 3 picked, got %d", len(picked))
	}
	seen := map[string]bool{}
	for _, c := range picked {
		if seen[c.Subdomain] {
			t.Errorf("duplicate subdomain in picked set: %+v", picked)
		}
		seen[c.Subdomain] = true
	}
}

func TestPickQuizConcepts_FallbackOnNoWeak(t *testing.T) {
	d := miniDAG(t)
	picked := PickQuizConcepts(d, map[string]float64{})
	if len(picked) == 0 {
		t.Fatal("expected fallback picks with no weak concepts")
	}
	if len(picked) > 5 {
		t.Errorf("expected at most 5 picks, got %d", len(picked))
	}
}

func TestPickQuizConcepts_DrawsFromKnownConceptsWhenNoWeakness(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{ID: "arith.1", Label: "Counting", Domain: "arithmetic", Subdomain: "count", MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
		{ID: "arith.2", Label: "Addition", Domain: "arithmetic", Subdomain: "add", MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
		{ID: "calc.1", Label: "Derivative", Domain: "calculus", Subdomain: "diff", MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
		{ID: "calc.2", Label: "Integral", Domain: "calculus", Subdomain: "int", MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	// Advanced student has 0 weakness on calculus concepts, 0 arithmetic concepts practiced
	weakness := map[string]float64{"calc.1": 0.0, "calc.2": 0.0}
	picked := PickQuizConcepts(d, weakness)
	if len(picked) != 2 {
		t.Fatalf("expected 2 picked, got %d", len(picked))
	}
	for _, c := range picked {
		if c.Domain != "calculus" {
			t.Errorf("expected calculus concept, got %s (%s)", c.ID, c.Domain)
		}
	}
}

func TestEngine_RecordAnswerAdvances(t *testing.T) {
	d := miniDAG(t)
	r := generator.NewRegistry()
	_ = r.Register("a", &fakeGen{})
	e := NewEngine(d, r)
	sess := e.Start(d.Order())
	e.RecordAnswer(sess, "a", true)
	if sess.Index != 1 {
		t.Errorf("expected index advance to 1, got %d", sess.Index)
	}
	if e.IsComplete(sess) {
		t.Error("expected not complete after 1 of 3")
	}
}

func TestEngine_SkipQuestionAndHasAttempts(t *testing.T) {
	d := miniDAG(t)
	r := generator.NewRegistry()
	_ = r.Register("a", &fakeGen{})
	_ = r.Register("b", &fakeGen{})
	_ = r.Register("c", &fakeGen{})
	e := NewEngine(d, r)
	sess := e.Start(d.Order())

	// Initially no attempts
	if e.HasAttempts(sess) {
		t.Error("expected no attempts initially")
	}

	// Skip question 1
	e.SkipQuestion(sess)
	if sess.Index != 1 {
		t.Errorf("expected index 1 after skip, got %d", sess.Index)
	}
	if e.HasAttempts(sess) {
		t.Error("expected no attempts after skip")
	}

	// Record answer on question 2
	e.RecordAnswer(sess, "b", false)
	if !e.HasAttempts(sess) {
		t.Error("expected HasAttempts true after recording answer")
	}

	// Skip past end
	e.SkipQuestion(sess)
	if !e.IsComplete(sess) {
		t.Error("expected session complete")
	}
	// Extra skip when complete is safe
	e.SkipQuestion(sess)
	if sess.Index != 3 {
		t.Errorf("expected index capped at length 3, got %d", sess.Index)
	}
}

func TestEngine_Start_CapsAtFive(t *testing.T) {
	var cList []concepts.Concept
	for i := 1; i <= 7; i++ {
		id := string(rune('a' + i - 1))
		cList = append(cList, concepts.Concept{
			ID:               id,
			Label:            id,
			Domain:           "d",
			Subdomain:        "s1",
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		})
	}
	d, err := concepts.Build(cList)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	e := NewEngine(d, generator.NewRegistry())

	// Case 1: Start with nil/empty slice uses DAG order capped at 5
	sessDefault := e.Start(nil)
	if len(sessDefault.Order) != 5 {
		t.Errorf("expected default order capped at 5, got %d", len(sessDefault.Order))
	}

	// Case 2: Start with >5 provided concepts caps at 5
	sessProvided := e.Start(d.Order())
	if len(sessProvided.Order) != 5 {
		t.Errorf("expected provided order capped at 5, got %d", len(sessProvided.Order))
	}
}

func TestEngine_NextQuestion_CompletedAndUnregistered(t *testing.T) {
	d := miniDAG(t)
	r := generator.NewRegistry()
	_ = r.Register("a", &fakeGen{})
	e := NewEngine(d, r)
	sess := e.Start([]*concepts.Concept{d.Concept("a")})

	// Question 1 succeeds
	prob, cid, err := e.NextQuestion(sess)
	if err != nil || prob == nil || cid != "a" {
		t.Fatalf("next: %v", err)
	}

	// Advance past end
	sess.Index = 1
	prob, cid, err = e.NextQuestion(sess)
	if err != nil {
		t.Errorf("expected nil error past end, got %v", err)
	}
	if prob != nil || cid != "" {
		t.Errorf("expected nil problem past end, got %v %s", prob, cid)
	}
	if sess.LastProblem != nil {
		t.Error("expected LastProblem cleared past end")
	}

	// Concept missing from registry returns error
	sessMissing := e.Start([]*concepts.Concept{d.Concept("b")})
	_, _, err = e.NextQuestion(sessMissing)
	if err == nil {
		t.Error("expected error for unregistered concept")
	}
}

func TestPickQuizConcepts_SameSubdomainFill(t *testing.T) {
	var cList []concepts.Concept
	for i := 1; i <= 6; i++ {
		id := string(rune('a' + i - 1))
		sub := "s1"
		if i > 3 {
			sub = "s2"
		}
		cList = append(cList, concepts.Concept{
			ID:               id,
			Label:            id,
			Domain:           "d",
			Subdomain:        sub,
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		})
	}
	d, err := concepts.Build(cList)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	weakness := map[string]float64{
		"a": 0.9,
		"b": 0.8,
		"c": 0.7,
		"d": 0.6,
		"e": 0.5,
		"f": 0.4,
	}

	picked := PickQuizConcepts(d, weakness)
	if len(picked) != 5 {
		t.Fatalf("expected 5 picked concepts with fills, got %d", len(picked))
	}

	// Distinct checks
	seen := map[string]bool{}
	for _, c := range picked {
		if seen[c.ID] {
			t.Errorf("duplicate concept %s picked", c.ID)
		}
		seen[c.ID] = true
	}
}

func TestEngine_RecordAnswer_Mismatch(t *testing.T) {
	d := miniDAG(t)
	r := generator.NewRegistry()
	_ = r.Register("a", &fakeGen{})
	e := NewEngine(d, r)
	sess := e.Start(d.Order())

	// Record an answer for a concept that is not the current one ("b" != "a").
	e.RecordAnswer(sess, "b", false)
	if len(sess.Attempts) != 1 {
		t.Fatalf("expected attempt recorded, got %d", len(sess.Attempts))
	}
	if sess.Index != 0 {
		t.Errorf("expected index to stay at 0 on mismatch, got %d", sess.Index)
	}
}
