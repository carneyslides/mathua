// Package repair fixes grading false negatives in persisted data.
//
// A grader infrastructure fault (or a parser gap) recorded correct answers as
// misses, which reset streaks, inflated weakness, and pushed diagnostic
// frontiers backwards. This pass re-grades every persisted incorrect attempt
// through the canonical grader (Engine.GradeAnswer) and, where an answer now
// grades correct, flips the attempt and rebuilds the affected concept's
// progress from its corrected history.
//
// It is intentionally conservative: attempts the grader cannot decide
// (Result.Unavailable) are left untouched, and XP is not rewritten (replaying
// task/streak/rush XP exactly is not reliable). Run with dryRun first.
package repair

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/engine"
	"github.com/chuma-beep/mathua/internal/mastery"
	"github.com/chuma-beep/mathua/internal/scheduler"
	"github.com/chuma-beep/mathua/internal/storage"
)

type Summary struct {
	Attempts       int
	Incorrect      int
	Repaired       int
	Unavailable    int
	AffectedPairs  int
	AffectedUsers  int
	ProgressWrites int
}

type pairKey struct{ student, concept string }

// Run re-grades and (unless dryRun) repairs. It writes a human-readable report
// to out and returns a summary.
func Run(eng *engine.Engine, repo storage.Repository, dag *concepts.DAG, dryRun bool, out io.Writer) (Summary, error) {
	var sum Summary

	all, err := repo.GetAllAttempts()
	if err != nil {
		return sum, fmt.Errorf("load attempts: %w", err)
	}
	sum.Attempts = len(all)

	corrected := make([]bool, len(all))
	repaired := make([]bool, len(all))
	pairs := map[pairKey]bool{}
	users := map[string]bool{}

	for i, a := range all {
		corrected[i] = a.Correct
		if a.Correct {
			continue
		}
		sum.Incorrect++
		res := eng.GradeAnswer(a.ConceptID, a.Expected, a.Answer)
		if res.Unavailable {
			sum.Unavailable++
			continue
		}
		if res.Correct {
			repaired[i] = true
			corrected[i] = true
			sum.Repaired++
			pairs[pairKey{a.StudentID, a.ConceptID}] = true
			users[a.StudentID] = true
		}
	}
	sum.AffectedPairs = len(pairs)
	sum.AffectedUsers = len(users)

	fmt.Fprintf(out, "scanned %d attempts · %d incorrect · %d now grade correct · %d grader-unavailable (skipped)\n",
		sum.Attempts, sum.Incorrect, sum.Repaired, sum.Unavailable)
	if sum.Repaired == 0 {
		fmt.Fprintln(out, "nothing to repair")
		return sum, nil
	}

	shown := 0
	for i, a := range all {
		if repaired[i] && shown < 8 {
			fmt.Fprintf(out, "  repair %s %s: expected=%q answer=%q\n", a.StudentID, a.ConceptID, a.Expected, a.Answer)
			shown++
		}
	}
	if sum.Repaired > shown {
		fmt.Fprintf(out, "  … and %d more\n", sum.Repaired-shown)
	}

	if dryRun {
		fmt.Fprintf(out, "dry run: would flip %d attempts and rebuild %d progress rows across %d student(s)\n",
			sum.Repaired, sum.AffectedPairs, sum.AffectedUsers)
		return sum, nil
	}

	for i, a := range all {
		if !repaired[i] {
			continue
		}
		if err := repo.UpdateAttemptCorrect(a.ID, true); err != nil {
			return sum, fmt.Errorf("update attempt %d: %w", a.ID, err)
		}
	}

	// Group attempts by pair once, then replay each affected pair.
	group := map[pairKey][]int{}
	for i, a := range all {
		k := pairKey{a.StudentID, a.ConceptID}
		if !pairs[k] {
			continue
		}
		group[k] = append(group[k], i)
	}
	for k, idxs := range group {
		sort.Slice(idxs, func(x, y int) bool { return all[idxs[x]].Timestamp.Before(all[idxs[y]].Timestamp) })
		atts := make([]storage.AttemptEntry, len(idxs))
		ok := make([]bool, len(idxs))
		for j, i := range idxs {
			atts[j] = all[i]
			atts[j].Correct = corrected[i]
			ok[j] = corrected[i]
		}
		p := replay(dag, repo, k.student, k.concept, atts, ok)
		if err := repo.UpsertProgress(p); err != nil {
			return sum, fmt.Errorf("rebuild progress %s/%s: %w", k.student, k.concept, err)
		}
		sum.ProgressWrites++
	}

	fmt.Fprintf(out, "repaired %d attempts and rebuilt %d progress rows across %d student(s)\n",
		sum.Repaired, sum.ProgressWrites, sum.AffectedUsers)
	return sum, nil
}

// replay reconstructs a concept's progress from its (corrected) attempt history
// using the same mastery/SM-2 primitives the engine uses. Best-effort: it does
// not reproduce halt/rush XP or per-topic learning-speed evolution exactly, and
// concepts absent from the DAG or carrying a non-positive required streak or time
// threshold fall back to required_streak = 3 and time_threshold = 10.0s.
func replay(dag *concepts.DAG, repo storage.Repository, studentID, conceptID string, atts []storage.AttemptEntry, correct []bool) *storage.ConceptProgress {
	reqStreak := 3
	thresh := 10.0
	if c := dag.Concept(conceptID); c != nil {
		reqStreak = c.MasteryThreshold.Streak
		thresh = c.MasteryThreshold.AvgTimeSeconds
	}
	if reqStreak <= 0 {
		reqStreak = 3
	}
	if thresh <= 0 {
		thresh = 10.0
	}
	speed := 1.0
	if ts, err := repo.GetTopicSpeed(studentID, conceptID); err == nil && ts != nil && ts.LearningSpeed > 0 {
		speed = ts.LearningSpeed
	}

	machine := &mastery.Machine{}
	p := &storage.ConceptProgress{StudentID: studentID, ConceptID: conceptID, SM2EFactor: 2.5}
	status := mastery.StatusUnseen
	var lastReviewAt *time.Time
	sm2 := scheduler.SM2{Repetitions: 0, EFactor: 2.5, Interval: 0}

	for i, a := range atts {
		p.Attempts++
		ts := a.Timestamp
		p.LastAttempted = &ts
		p.AvgResponseTime = (p.AvgResponseTime*float64(p.Attempts-1) + a.ElapsedSeconds) / float64(p.Attempts)

		if correct[i] {
			p.Streak++
			if p.Streak > p.BestStreak {
				p.BestStreak = p.Streak
			}
		} else {
			p.Streak = 0
		}

		if correct[i] {
			p.WeaknessScore *= 0.5
			if a.ElapsedSeconds <= thresh {
				p.WeaknessScore *= 0.3
			}
		} else {
			p.WeaknessScore += 0.2
			if p.WeaknessScore > 1.0 {
				p.WeaknessScore = 1.0
			}
		}

		var evAtts []mastery.Attempt
		start := 0
		if i+1 > mastery.EvidenceWindow {
			start = i + 1 - mastery.EvidenceWindow
		}
		for k := start; k <= i; k++ {
			inst := atts[k].Question
			if inst == "" {
				inst = atts[k].Answer
			}
			evAtts = append(evAtts, mastery.Attempt{
				Correct:    correct[k],
				Elapsed:    atts[k].ElapsedSeconds,
				Difficulty: atts[k].Difficulty,
				Instance:   inst,
			})
		}
		evidence := mastery.BuildEvidence(evAtts, mastery.EvidenceOptions{TimeThreshold: thresh})

		next := machine.Next(status, mastery.TransitionCtx{
			Evidence:          evidence,
			Streak:            p.Streak,
			RequiredStreak:    reqStreak,
			AvgResponseTime:   p.AvgResponseTime,
			ResponseThreshold: thresh,
		})
		advanced := next != status
		if next == mastery.StatusMastered && status != mastery.StatusMastered {
			mt := a.Timestamp
			p.MasteredAt = &mt
			p.WeaknessScore = 0.0
		}
		// The live engine resets the streak on a tier advance before it computes
		// SM-2 quality, so an advancing attempt is graded at quality 0 and the
		// repaired interval matches what a real run would have stored.
		if advanced && correct[i] {
			p.Streak = 1
		}
		status = next

		q := mastery.SM2Quality(correct[i] && p.Streak >= reqStreak, p.AvgResponseTime/thresh)
		sm2 = scheduler.ComputeSM2WithSpeed(sm2, q, speed)

		// The decay clock starts at the last attempt that advanced a tier, or
		// that met the streak and time threshold. Tracking the timestamp here is
		// what keeps LastReviewed honest: the final attempt may have advanced
		// nothing, and then it is not the review date.
		if advanced || (p.Streak >= reqStreak && p.AvgResponseTime <= thresh) {
			ts := a.Timestamp
			lastReviewAt = &ts
		}
	}

	p.Status = string(status)
	p.SM2Repetitions = sm2.Repetitions
	p.SM2EFactor = sm2.EFactor
	p.SM2Interval = sm2.Interval
	if lastReviewAt != nil {
		p.LastReviewed = lastReviewAt
		due := lastReviewAt.AddDate(0, 0, sm2.Interval)
		p.NextReviewDue = &due
	}
	return p
}
