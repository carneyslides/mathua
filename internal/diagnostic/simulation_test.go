package diagnostic

import (
	"fmt"
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/generator"
)

// buildCurriculumDAG builds a DAG with two parallel branches (algebra and geometry).
func buildCurriculumDAG(t *testing.T, countPerBranch int) (*concepts.DAG, *generator.Registry) {
	t.Helper()
	var raw []concepts.Concept
	reg := generator.NewRegistry()

	// Root foundational concepts
	for i := 0; i < 4; i++ {
		id := fmt.Sprintf("arith_%02d", i)
		var prereqs []string
		if i > 0 {
			prereqs = []string{fmt.Sprintf("arith_%02d", i-1)}
		}
		raw = append(raw, concepts.Concept{
			ID:               id,
			Label:            "Arithmetic " + id,
			Domain:           "arithmetic",
			Subdomain:        "arithmetic.basics",
			Prerequisites:    prereqs,
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		})
		if err := reg.Register(id, &fakeGen{}); err != nil {
			t.Fatalf("register: %v", err)
		}
	}

	// Branch 1: Algebra
	prevID := "arith_03"
	for i := 0; i < countPerBranch; i++ {
		id := fmt.Sprintf("alg_%02d", i)
		raw = append(raw, concepts.Concept{
			ID:               id,
			Label:            "Algebra " + id,
			Domain:           "algebra",
			Subdomain:        "algebra.linear",
			Prerequisites:    []string{prevID},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		})
		if err := reg.Register(id, &fakeGen{}); err != nil {
			t.Fatalf("register: %v", err)
		}
		prevID = id
	}

	// Branch 2: Geometry
	prevID = "arith_03"
	for i := 0; i < countPerBranch; i++ {
		id := fmt.Sprintf("geo_%02d", i)
		raw = append(raw, concepts.Concept{
			ID:               id,
			Label:            "Geometry " + id,
			Domain:           "geometry",
			Subdomain:        "geometry.triangles",
			Prerequisites:    []string{prevID},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
		})
		if err := reg.Register(id, &fakeGen{}); err != nil {
			t.Fatalf("register: %v", err)
		}
		prevID = id
	}

	dag, err := concepts.Build(raw)
	if err != nil {
		t.Fatalf("build curriculum dag: %v", err)
	}
	return dag, reg
}

func TestSimulation_SyntheticStudents(t *testing.T) {
	dag, reg := buildCurriculumDAG(t, 10)
	eng := NewEngine(dag, reg)

	t.Run("StrongStudent", func(t *testing.T) {
		s := eng.Start()
		// Answers all correctly and quickly.
		for i := 0; i < 60 && !eng.IsComplete(s); i++ {
			_, cid, err := eng.NextQuestion(s)
			if err != nil || cid == "" {
				break
			}
			eng.RecordAnswer(s, cid, true, true)
		}
		rep := eng.Report(s)
		if rep.FrontierIdx < len(s.order)-4 {
			t.Errorf("expected strong student frontier near curriculum top, got index %d of %d",
				rep.FrontierIdx, len(s.order))
		}
		if rep.AvgConfidence < 0.60 {
			t.Errorf("expected high average confidence for strong student, got %f", rep.AvgConfidence)
		}
	})

	t.Run("WeakStudent", func(t *testing.T) {
		s := eng.Start()
		// Fails all questions.
		for i := 0; i < 60 && !eng.IsComplete(s); i++ {
			_, cid, err := eng.NextQuestion(s)
			if err != nil || cid == "" {
				break
			}
			eng.RecordAnswer(s, cid, false, false)
		}
		rep := eng.Report(s)
		if rep.FrontierIdx > 3 {
			t.Errorf("expected weak student frontier at base, got index %d", rep.FrontierIdx)
		}
		if len(rep.GapsByDomain["arithmetic"]) == 0 && len(rep.GapsByDomain["algebra"]) == 0 {
			t.Error("expected foundational gaps recorded for weak student")
		}
	})

	t.Run("AverageStudent", func(t *testing.T) {
		s := eng.Start()
		// Knows foundational arithmetic (arith_*) and early algebra (alg_00, alg_01).
		// Struggles on advanced algebra and all geometry.
		knows := map[string]bool{
			"arith_00": true,
			"arith_01": true,
			"arith_02": true,
			"arith_03": true,
			"alg_00":   true,
			"alg_01":   true,
		}
		for i := 0; i < 60 && !eng.IsComplete(s); i++ {
			_, cid, err := eng.NextQuestion(s)
			if err != nil || cid == "" {
				break
			}
			correct := knows[cid]
			eng.RecordAnswer(s, cid, correct, correct)
		}
		rep := eng.Report(s)
		// Frontier concept must be among the known concepts or early frontier.
		if rep.FrontierIdx < 0 {
			t.Fatal("expected non-negative frontier index for average student")
		}
		frontierConcept := s.order[rep.FrontierIdx]
		if !knows[frontierConcept.ID] && frontierConcept.ID != "alg_02" {
			t.Errorf("expected frontier concept to be near knowledge boundary, got %s", frontierConcept.ID)
		}
	})

	t.Run("UnevenStudent", func(t *testing.T) {
		s := eng.Start()
		// Strong on arithmetic and algebra, completely unlearned in geometry.
		for i := 0; i < 60 && !eng.IsComplete(s); i++ {
			_, cid, err := eng.NextQuestion(s)
			if err != nil || cid == "" {
				break
			}
			c := dag.Concept(cid)
			correct := c != nil && (c.Domain == "arithmetic" || c.Domain == "algebra")
			eng.RecordAnswer(s, cid, correct, correct)
		}
		rep := eng.Report(s)
		if len(rep.GapsByDomain["geometry"]) == 0 {
			t.Errorf("expected geometry gaps for uneven student, got gaps: %v", rep.GapsByDomain)
		}
	})
}

func TestSimulation_BidirectionalEvidencePropagationWithDecay(t *testing.T) {
	dag := chainDAG(t, 6) // c00 <- c01 <- c02 <- c03 <- c04 <- c05
	reg := chainRegistry(t, 6)
	eng := NewEngine(dag, reg)

	t.Run("CorrectAnswerPropagatesDecayingPositiveEvidence", func(t *testing.T) {
		s := eng.Start()
		// Answer correctly on c04. Ancestors c03 (dist 1), c02 (dist 2), c01 (dist 3) and c00 (dist 4)
		// must all receive positive evidence with topological distance decay.
		eng.RecordAnswer(s, "c04", true, false)

		b4 := s.beliefs["c04"]
		b3 := s.beliefs["c03"]
		b2 := s.beliefs["c02"]
		b1 := s.beliefs["c01"]
		b0 := s.beliefs["c00"]

		if b4 <= 0.6 {
			t.Errorf("expected probed concept c04 belief > 0.6, got %f", b4)
		}
		if b3 <= 0.5 || b2 <= 0.5 || b1 <= 0.5 || b0 <= 0.5 {
			t.Errorf("expected all ancestors boosted above 0.5: b3=%f b2=%f b1=%f b0=%f", b3, b2, b1, b0)
		}
		if !(b3 > b2 && b2 > b1 && b1 > b0) {
			t.Errorf("expected topological decay ordering b3 > b2 > b1 > b0, got b3=%f b2=%f b1=%f b0=%f",
				b3, b2, b1, b0)
		}
	})

	t.Run("IncorrectAnswerPropagatesDecayingNegativeEvidence", func(t *testing.T) {
		s := eng.Start()
		// Answer incorrectly on c01. Descendants c02 (dist 1), c03 (dist 2) and c04 (dist 3)
		// must all receive negative evidence with topological distance decay.
		eng.RecordAnswer(s, "c01", false, false)

		b1 := s.beliefs["c01"]
		b2 := s.beliefs["c02"]
		b3 := s.beliefs["c03"]
		b4 := s.beliefs["c04"]

		if b1 >= 0.5 {
			t.Errorf("expected missed concept c01 belief < 0.5, got %f", b1)
		}
		if b2 >= 0.5 || b3 >= 0.5 || b4 >= 0.5 {
			t.Errorf("expected all descendants lowered below 0.5: b2=%f b3=%f b4=%f", b2, b3, b4)
		}
		if !(b2 < b3 && b3 < b4) {
			t.Errorf("expected topological decay ordering b2 < b3 < b4, got b2=%f b3=%f b4=%f",
				b2, b3, b4)
		}
	})
}

func TestSimulation_SupplementalProbeOnLowConfidenceFrontier(t *testing.T) {
	dag := chainDAG(t, 10)
	reg := chainRegistry(t, 10)
	eng := NewEngine(dag, reg)
	s := eng.Start()

	// Settle concepts with single probes leaving confidence at 0.5 (< 0.70).
	for i := 0; i < 5; i++ {
		cid := chainID(i)
		eng.RecordAnswer(s, cid, true, true)
		s.doneSet[cid] = true
	}
	s.Lock()
	target := eng.supplementalDiagnostic(s)
	s.Unlock()

	if target == "" {
		t.Fatal("expected supplemental diagnostic to pick target when confidence < 0.70")
	}
}

func TestSimulation_DiagnosticReportTimelines(t *testing.T) {
	estimates := EstimateCompletionDates(300, time.Time{}, []int{20, 40, 60})
	if estimates[20] != "≈15 days" {
		t.Errorf("expected 20 pace to be ≈15 days, got %s", estimates[20])
	}
	if estimates[40] != "≈8 days" {
		t.Errorf("expected 40 pace to be ≈8 days, got %s", estimates[40])
	}
	if estimates[60] != "≈5 days" {
		t.Errorf("expected 60 pace to be ≈5 days, got %s", estimates[60])
	}

	startDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	dateEstimates := EstimateCompletionDates(200, startDate, []int{20, 40})
	if dateEstimates[20] != "2026-10-11" {
		t.Errorf("expected 20 pace date 2026-10-11, got %s", dateEstimates[20])
	}
	if dateEstimates[40] != "2026-10-06" {
		t.Errorf("expected 40 pace date 2026-10-06, got %s", dateEstimates[40])
	}
}
