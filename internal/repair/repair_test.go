package repair

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/engine"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/grader"
	"github.com/chuma-beep/mathua/internal/storage"
)

type unavailGen struct{}

func (g *unavailGen) Generate(ctx generator.GeneratorContext) generator.Problem {
	return generator.Problem{Question: "q", Answer: "42", Explanation: "e"}
}

func (g *unavailGen) Grade(expected, answer string) grader.Result {
	if expected == "unavail" {
		return grader.Result{Unavailable: true}
	}
	if expected == answer && answer != "" {
		return grader.Result{Correct: true, Score: 1}
	}
	return grader.Result{Correct: false}
}

func testRepo(t *testing.T) (*storage.SQLiteStore, *concepts.DAG, string, string) {
	t.Helper()
	dag, err := concepts.Build([]concepts.Concept{{
		ID: "t", Label: "T", Domain: "d", GradingType: "numeric",
		MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
	}})
	if err != nil {
		t.Fatalf("build dag: %v", err)
	}
	repo, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	stu, err := repo.CreateStudent("S")
	if err != nil {
		t.Fatalf("student: %v", err)
	}
	sess, err := repo.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	return repo, dag, stu.ID, sess.ID
}

func TestRepair_FlipsFalseNegative(t *testing.T) {
	repo, dag, sid, sess := testRepo(t)
	base := time.Now().UTC().Add(-time.Hour)

	// False negative: the old numeric grader rejected the spoken mixed number.
	must(t, func() error {
		_, err := repo.RecordAttempt(storage.AttemptEntry{
			SessionID: sess, StudentID: sid, ConceptID: "t",
			Answer: "4 and 1/10", Expected: "4 1/10", Correct: false,
			ElapsedSeconds: 5, Timestamp: base,
		})
		return err
	}())
	// Genuine miss: must stay incorrect after repair.
	must(t, func() error {
		_, err := repo.RecordAttempt(storage.AttemptEntry{
			SessionID: sess, StudentID: sid, ConceptID: "t",
			Answer: "5", Expected: "4 1/10", Correct: false,
			ElapsedSeconds: 9, Timestamp: base.Add(time.Minute),
		})
		return err
	}())

	reg := generator.NewRegistry()
	eng := engine.New(repo, dag, reg, nil, nil)
	defer eng.Close()

	var buf bytes.Buffer
	sum, err := Run(eng, repo, dag, true, &buf)
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	if sum.Repaired != 1 {
		t.Fatalf("dry repaired=%d, want 1\n%s", sum.Repaired, buf.String())
	}
	if all, _ := repo.GetAllAttempts(); all[0].Correct {
		t.Fatal("dry run must not write")
	}

	buf.Reset()
	sum, err = Run(eng, repo, dag, false, &buf)
	if err != nil {
		t.Fatalf("repair: %v", err)
	}
	if sum.Repaired != 1 || sum.ProgressWrites != 1 {
		t.Fatalf("repair repaired=%d writes=%d, want 1/1\n%s", sum.Repaired, sum.ProgressWrites, buf.String())
	}

	all, err := repo.GetAllAttempts()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	var flipped, wrong int
	for _, a := range all {
		if a.Answer == "4 and 1/10" && a.Correct {
			flipped++
		}
		if a.Answer == "5" && !a.Correct {
			wrong++
		}
	}
	if flipped != 1 {
		t.Errorf("false negative not flipped: %+v", all)
	}
	if wrong != 1 {
		t.Errorf("genuine miss was changed: %+v", all)
	}

	// Rebuilt progress reflects the corrected history (attempts counted).
	p, err := repo.GetProgress(sid, "t")
	if err != nil {
		t.Fatalf("progress: %v", err)
	}
	if p == nil || p.Attempts != 2 {
		t.Fatalf("progress not rebuilt: %+v", p)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func TestRepair_NothingToRepair(t *testing.T) {
	repo, dag, sid, sess := testRepo(t)
	reg := generator.NewRegistry()
	eng := engine.New(repo, dag, reg, nil, nil)
	defer eng.Close()

	// Empty repo: nothing to repair
	var buf bytes.Buffer
	sum, err := Run(eng, repo, dag, false, &buf)
	if err != nil {
		t.Fatalf("run empty: %v", err)
	}
	if sum.Repaired != 0 || sum.Attempts != 0 {
		t.Errorf("expected 0 repaired and 0 attempts, got %+v", sum)
	}
	if !strings.Contains(buf.String(), "nothing to repair") {
		t.Errorf("output expected 'nothing to repair', got %q", buf.String())
	}

	// Genuine miss and correct attempt: both not repairable
	base := time.Now().UTC()
	must(t, func() error {
		_, err := repo.RecordAttempt(storage.AttemptEntry{
			SessionID: sess, StudentID: sid, ConceptID: "t",
			Answer: "5", Expected: "4 1/10", Correct: false,
			ElapsedSeconds: 9, Timestamp: base,
		})
		return err
	}())
	must(t, func() error {
		_, err := repo.RecordAttempt(storage.AttemptEntry{
			SessionID: sess, StudentID: sid, ConceptID: "t",
			Answer: "4 1/10", Expected: "4 1/10", Correct: true,
			ElapsedSeconds: 5, Timestamp: base.Add(time.Minute),
		})
		return err
	}())

	buf.Reset()
	sum, err = Run(eng, repo, dag, false, &buf)
	if err != nil {
		t.Fatalf("run with misses: %v", err)
	}
	if sum.Repaired != 0 || sum.Attempts != 2 || sum.Incorrect != 1 {
		t.Errorf("expected 0 repaired, 2 attempts, 1 incorrect; got %+v", sum)
	}
	if !strings.Contains(buf.String(), "nothing to repair") {
		t.Errorf("output expected 'nothing to repair', got %q", buf.String())
	}
}

func TestRepair_GraderUnavailableAndOverflowShown(t *testing.T) {
	dag, err := concepts.Build([]concepts.Concept{
		{ID: "num", Label: "Num", Domain: "d", GradingType: "numeric", MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
		{ID: "unavail", Label: "Unavail", Domain: "d", GradingType: "numeric", MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10}},
	})
	if err != nil {
		t.Fatalf("dag: %v", err)
	}
	repo, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	defer repo.Close()
	stu, err := repo.CreateStudent("S2")
	if err != nil {
		t.Fatalf("student: %v", err)
	}
	sess, err := repo.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("session: %v", err)
	}

	reg := generator.NewRegistry()
	reg.Register("unavail", &unavailGen{})
	eng := engine.New(repo, dag, reg, nil, nil)
	defer eng.Close()

	base := time.Now().UTC().Add(-2 * time.Hour)

	// Record an attempt on unavail concept: custom GradedGenerator returns Unavailable
	must(t, func() error {
		_, err := repo.RecordAttempt(storage.AttemptEntry{
			SessionID: sess.ID, StudentID: stu.ID, ConceptID: "unavail",
			Answer: "wrong", Expected: "unavail", Correct: false,
			ElapsedSeconds: 5, Timestamp: base,
		})
		return err
	}())

	// Record 10 attempts on numeric concept that will be repaired (> 8 shown limit)
	for i := 0; i < 10; i++ {
		must(t, func() error {
			_, err := repo.RecordAttempt(storage.AttemptEntry{
				SessionID: sess.ID, StudentID: stu.ID, ConceptID: "num",
				Answer: "4 and 1/10", Expected: "4 1/10", Correct: false,
				ElapsedSeconds: 5, Timestamp: base.Add(time.Duration(i+1) * time.Minute),
			})
			return err
		}())
	}

	var buf bytes.Buffer
	sum, err := Run(eng, repo, dag, false, &buf)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if sum.Unavailable != 1 {
		t.Errorf("sum.Unavailable = %d, want 1", sum.Unavailable)
	}
	if sum.Repaired != 10 {
		t.Errorf("sum.Repaired = %d, want 10", sum.Repaired)
	}
	if !strings.Contains(buf.String(), "… and 2 more") {
		t.Errorf("expected '… and 2 more' in output, got:\n%s", buf.String())
	}
}

func TestRepair_ReplayMasteryAndReviewDue(t *testing.T) {
	repo, dag, sid, sess := testRepo(t)
	reg := generator.NewRegistry()
	eng := engine.New(repo, dag, reg, nil, nil)
	defer eng.Close()

	// Seed topic speed
	must(t, repo.UpsertTopicSpeed(&storage.TopicSpeed{
		StudentID: sid, ConceptID: "t", LearningSpeed: 1.5, EFactor: 2.5, Interval: 1, Repetitions: 1,
	}))

	base := time.Now().UTC().Add(-time.Hour)
	// Record 3 consecutive false negatives for concept "t" (required streak 3, avg time 10s)
	for i := 0; i < 3; i++ {
		must(t, func() error {
			_, err := repo.RecordAttempt(storage.AttemptEntry{
				SessionID: sess, StudentID: sid, ConceptID: "t",
				Answer: "4 and 1/10", Expected: "4 1/10", Correct: false,
				ElapsedSeconds: 4.0, Timestamp: base.Add(time.Duration(i) * time.Minute),
			})
			return err
		}())
	}

	var buf bytes.Buffer
	sum, err := Run(eng, repo, dag, false, &buf)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if sum.Repaired != 3 {
		t.Fatalf("repaired = %d, want 3", sum.Repaired)
	}

	p, err := repo.GetProgress(sid, "t")
	if err != nil {
		t.Fatalf("progress: %v", err)
	}
	if p.WeaknessScore != 0.0 {
		t.Errorf("WeaknessScore = %f, want 0.0", p.WeaknessScore)
	}
	if p.Streak != 3 || p.BestStreak != 3 {
		t.Errorf("Streak/BestStreak = %d/%d, want 3/3", p.Streak, p.BestStreak)
	}
	if p.LastReviewed == nil {
		t.Error("LastReviewed expected non-nil")
	}
	if p.NextReviewDue == nil {
		t.Error("NextReviewDue expected non-nil")
	}
	if p.SM2Interval <= 0 {
		t.Errorf("SM2Interval = %d, want > 0", p.SM2Interval)
	}
}

func TestRepair_ReplayConceptFallbackAndWeaknessCap(t *testing.T) {
	repo, dag, sid, _ := testRepo(t)
	now := time.Now().UTC()

	// Replay on unmapped concept with 10 consecutive misses
	atts := make([]storage.AttemptEntry, 10)
	ok := make([]bool, 10)
	for i := 0; i < 10; i++ {
		atts[i] = storage.AttemptEntry{
			StudentID: sid, ConceptID: "unmapped",
			ElapsedSeconds: 15.0, Timestamp: now.Add(time.Duration(i) * time.Minute),
		}
		ok[i] = false
	}

	p := replay(dag, repo, sid, "unmapped", atts, ok)
	if p.ConceptID != "unmapped" {
		t.Errorf("ConceptID = %q, want unmapped", p.ConceptID)
	}
	if p.WeaknessScore > 1.0 {
		t.Errorf("WeaknessScore = %f, want capped at 1.0", p.WeaknessScore)
	}
	if p.Streak != 0 {
		t.Errorf("Streak = %d, want 0", p.Streak)
	}
}

func TestRepair_RepoError(t *testing.T) {
	repo, dag, _, _ := testRepo(t)
	reg := generator.NewRegistry()
	eng := engine.New(repo, dag, reg, nil, nil)
	defer eng.Close()

	repo.Close() // close store to trigger error

	var buf bytes.Buffer
	if _, err := Run(eng, repo, dag, false, &buf); err == nil {
		t.Error("expected error from Run with closed repository")
	}
}
func TestRepair_ReplayFallbackWhenConceptAbsentOrZeroThreshold(t *testing.T) {
	repo, dag, sid, sess := testRepo(t)
	base := time.Now().UTC().Add(-time.Hour)

	// Case 1: Concept absent from DAG entirely.
	// 6 consistent attempts in window with distinct instances to achieve mastery.
	attsAbsent := make([]storage.AttemptEntry, 6)
	correctAbsent := make([]bool, 6)
	diff := 0.8
	for i := 0; i < 6; i++ {
		attsAbsent[i] = storage.AttemptEntry{
			SessionID:      sess,
			StudentID:      sid,
			ConceptID:      "missing_concept",
			Question:       time.Duration(i).String(),
			Answer:         "ans",
			Correct:        true,
			ElapsedSeconds: 3.0,
			Difficulty:     &diff,
			Timestamp:      base.Add(time.Duration(i) * time.Minute),
		}
		correctAbsent[i] = true
	}
	pAbsent := replay(dag, repo, sid, "missing_concept", attsAbsent, correctAbsent)
	if pAbsent == nil {
		t.Fatal("expected non-nil progress for absent concept")
	}
	if pAbsent.Attempts != 6 || pAbsent.BestStreak != 5 || pAbsent.Streak != 1 {
		t.Errorf("absent concept: attempts=%d bestStreak=%d streak=%d, want 6/5/1", pAbsent.Attempts, pAbsent.BestStreak, pAbsent.Streak)
	}
	// Fallback required streak of 3 and threshold of 10s should transition to mastered after full evidence window
	if pAbsent.Status != "MASTERED" {
		t.Errorf("absent concept status=%s, want MASTERED", pAbsent.Status)
	}
	if pAbsent.MasteredAt == nil {
		t.Errorf("absent concept expected MasteredAt to be set")
	}

	// Case 2: Concept with zero streak and zero time threshold in raw DAG struct.
	// Since concepts.Build rejects non-positive streak, construct a Concept directly in DAG.
	cZero := &concepts.Concept{
		ID:          "zero_thresh",
		Label:       "Zero Thresh",
		Domain:      "d",
		GradingType: "numeric",
		MasteryThreshold: concepts.MasteryThreshold{
			Streak:         0,
			AvgTimeSeconds: 0,
		},
	}
	dagZero, err := concepts.Build([]concepts.Concept{{
		ID: "dummy", Label: "Dummy", Domain: "d", GradingType: "numeric",
		MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 10},
	}})
	if err != nil {
		t.Fatalf("build dagZero base: %v", err)
	}
	// Inject the zero threshold concept directly into DAG to test replay fallback guard
	dagZero.Concepts()["zero_thresh"] = cZero

	attsZero := make([]storage.AttemptEntry, 6)
	correctZero := make([]bool, 6)
	for i := 0; i < 6; i++ {
		attsZero[i] = storage.AttemptEntry{
			SessionID:      sess,
			StudentID:      sid,
			ConceptID:      "zero_thresh",
			Question:       time.Duration(i).String(),
			Answer:         "ans",
			Correct:        true,
			ElapsedSeconds: 2.5,
			Difficulty:     &diff,
			Timestamp:      base.Add(time.Duration(i) * time.Minute),
		}
		correctZero[i] = true
	}
	pZero := replay(dagZero, repo, sid, "zero_thresh", attsZero, correctZero)
	if pZero == nil {
		t.Fatal("expected non-nil progress for zero threshold concept")
	}
	if pZero.Status != "MASTERED" {
		t.Errorf("zero threshold concept status=%s, want MASTERED", pZero.Status)
	}
	if pZero.WeaknessScore != 0.0 {
		t.Errorf("weakness score=%f, want 0.0 after mastery", pZero.WeaknessScore)
	}
	if pZero.NextReviewDue == nil {
		t.Errorf("expected NextReviewDue to be scheduled for mastered concept")
	}
}

func TestRepair_ReplayWithTopicSpeed(t *testing.T) {
	repo, dag, sid, sess := testRepo(t)
	base := time.Now().UTC().Add(-time.Hour)

	must(t, repo.UpsertTopicSpeed(&storage.TopicSpeed{
		StudentID:     sid,
		ConceptID:     "t",
		LearningSpeed: 1.5,
	}))

	atts := make([]storage.AttemptEntry, 6)
	correct := make([]bool, 6)
	diff := 0.7
	for i := 0; i < 6; i++ {
		atts[i] = storage.AttemptEntry{
			SessionID:      sess,
			StudentID:      sid,
			ConceptID:      "t",
			Question:       time.Duration(i).String(),
			Answer:         "1",
			Correct:        true,
			ElapsedSeconds: 2.0,
			Difficulty:     &diff,
			Timestamp:      base.Add(time.Duration(i) * time.Minute),
		}
		correct[i] = true
	}
	p := replay(dag, repo, sid, "t", atts, correct)
	if p == nil {
		t.Fatal("expected non-nil progress")
	}
	if p.Status != "MASTERED" {
		t.Errorf("status=%s, want MASTERED", p.Status)
	}
	if p.SM2Repetitions < 3 {
		t.Errorf("SM2Repetitions=%d, want >= 3", p.SM2Repetitions)
	}
	if p.SM2Interval <= 0 {
		t.Errorf("SM2Interval=%d, want > 0", p.SM2Interval)
	}
}
