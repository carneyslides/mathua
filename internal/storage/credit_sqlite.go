package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/chuma-beep/mathua/internal/repetition"
)

// The credit tables are a shadow ledger: they record what implicit repetition
// *would* justify, and never write next_review_due, sm2_*, streaks, mastery or
// XP. A projection stores its counterfactual date on the ordinal.
//
// Every write serialises on the target row in implicit_credit_targets, which is
// also the epoch holder. SQLite is opened with a plain DEFERRED transaction by
// the rest of the store, which cannot take the write lock until its first write;
// this file therefore uses BEGIN IMMEDIATE on a dedicated connection so two
// allocations for the same learner queue instead of failing the upgrade.
var _ CreditStore = (*SQLiteStore)(nil)

func (s *SQLiteStore) runImmediate(ctx context.Context, fn func(*sql.Conn) error) error {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("credit: begin: %w", err)
	}
	if err := fn(conn); err != nil {
		_, _ = conn.ExecContext(ctx, "ROLLBACK")
		return err
	}
	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return fmt.Errorf("credit: commit: %w", err)
	}
	return nil
}

func (s *SQLiteStore) targetEpoch(conn *sql.Conn, ctx context.Context, studentID, conceptID string) (int64, error) {
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO implicit_credit_targets (student_id, target_concept_id, credit_epoch, created_at)
		VALUES (?, ?, 1, ?)
		ON CONFLICT(student_id, target_concept_id) DO NOTHING
	`, studentID, conceptID, time.Now().UTC().UnixNano()); err != nil {
		return 0, fmt.Errorf("credit: ensure target: %w", err)
	}
	var epoch int64
	if err := conn.QueryRowContext(ctx, `
		SELECT credit_epoch FROM implicit_credit_targets
		WHERE student_id = ? AND target_concept_id = ?
	`, studentID, conceptID).Scan(&epoch); err != nil {
		return 0, fmt.Errorf("credit: read epoch: %w", err)
	}
	return epoch, nil
}

func (s *SQLiteStore) accountTx(ctx context.Context, conn *sql.Conn, studentID, conceptID string, epoch int64) (CreditAccount, error) {
	acc := CreditAccount{ConceptID: conceptID, Epoch: epoch}
	if err := conn.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(retained_units), 0) FROM implicit_credit_events
		WHERE student_id = ? AND target_concept_id = ? AND credit_epoch = ? AND voided_at IS NULL
	`, studentID, conceptID, epoch).Scan(&acc.RetainedUnits); err != nil {
		return acc, fmt.Errorf("credit: sum retained: %w", err)
	}
	if err := conn.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN state = 'projected' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN state = 'stale' THEN 1 ELSE 0 END), 0)
		FROM implicit_credit_ordinals
		WHERE student_id = ? AND target_concept_id = ? AND credit_epoch = ?
	`, studentID, conceptID, epoch).Scan(&acc.Banked, &acc.Projected, &acc.Stale); err != nil {
		return acc, fmt.Errorf("credit: count ordinals: %w", err)
	}
	return acc, nil
}

func (s *SQLiteStore) AllocateCredit(studentID, attemptID string, epoch int64, candidates []CreditCandidate) (CreditAllocation, bool, error) {
	out := CreditAllocation{StudentID: studentID, AttemptID: attemptID, Epoch: epoch}
	if studentID == "" || attemptID == "" {
		return out, false, fmt.Errorf("credit: empty student or attempt")
	}
	raw := make([]repetition.Candidate, 0, len(candidates))
	for _, c := range candidates {
		raw = append(raw, repetition.Candidate{ConceptID: c.ConceptID, RawUnits: c.RawUnits})
	}
	fresh := true
	scaled, err := repetition.Allocate(raw, repetition.Micro)
	if err != nil {
		return out, false, fmt.Errorf("credit: allocate: %w", err)
	}

	err = s.runImmediate(context.Background(), func(conn *sql.Conn) error {
		ctx := context.Background()
		res, err := conn.ExecContext(ctx, `
			INSERT INTO implicit_credit_allocations
				(student_id, source_attempt_id, raw_units, allocated_units, target_count, dropped_count, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(student_id, source_attempt_id) DO NOTHING
		`, studentID, attemptID, sumRaw(raw), sumScaled(scaled), len(scaled), len(raw)-len(scaled), time.Now().UTC().UnixNano())
		if err != nil {
			return fmt.Errorf("credit: insert allocation: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			committed, err := s.replayAllocation(ctx, conn, studentID, attemptID)
			if err != nil {
				return err
			}
			out = committed
			fresh = false
			return nil
		}
		out.RawUnits = sumRaw(raw)
		out.AllocatedUnits = sumScaled(scaled)
		out.TargetCount = len(scaled)
		out.DroppedCount = len(raw) - len(scaled)
		out.Events = make([]CreditEvent, 0, len(scaled))

		for _, a := range scaled {
			targetEpoch, err := s.targetEpoch(conn, ctx, studentID, a.ConceptID)
			if err != nil {
				return err
			}
			if epoch > 0 && targetEpoch != epoch {
				return fmt.Errorf("%w: target %s is at epoch %d, not %d", ErrCreditClosedEpoch, a.ConceptID, targetEpoch, epoch)
			}
			useEpoch := targetEpoch
			if epoch > 0 {
				useEpoch = epoch
			}
			acc, err := s.accountTx(ctx, conn, studentID, a.ConceptID, useEpoch)
			if err != nil {
				return err
			}
			retained := a.Units
			capacity := int64(repetition.MaxBank)*repetition.Micro - acc.RetainedUnits
			if capacity < 0 {
				capacity = 0
			}
			discarded := int64(0)
			reason := ""
			if retained > capacity {
				discarded = retained - capacity
				retained = capacity
				reason = "epoch_bank_limit"
			}
			banked := int64(0)
			if want := (acc.RetainedUnits + retained) / repetition.Micro; want > int64(acc.Banked) {
				banked = want - int64(acc.Banked)
			}
			ev := CreditEvent{
				AttemptID: attemptID, ConceptID: a.ConceptID, Epoch: useEpoch,
				ScaledUnits: a.Units, RetainedUnits: retained, DiscardedUnits: discarded,
				DiscardReason: reason, Banked: int(banked),
			}
			if _, err := conn.ExecContext(ctx, `
				INSERT INTO implicit_credit_events
					(student_id, source_attempt_id, target_concept_id, credit_epoch,
					 scaled_units, retained_units, discarded_units, discard_reason, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(student_id, source_attempt_id, target_concept_id) DO NOTHING
			`, studentID, attemptID, a.ConceptID, useEpoch, a.Units, retained, discarded, reason, time.Now().UTC().UnixNano()); err != nil {
				return fmt.Errorf("credit: insert event: %w", err)
			}
			for i := int64(0); i < banked; i++ {
				if _, err := conn.ExecContext(ctx, `
					INSERT INTO implicit_credit_ordinals
						(student_id, target_concept_id, credit_epoch, ordinal, state, banked_at, attempts)
					VALUES (?, ?, ?, ?, 'banked', ?, 0)
				`, studentID, a.ConceptID, useEpoch, int64(acc.Banked)+i+1, time.Now().UTC().UnixNano()); err != nil {
					return fmt.Errorf("credit: insert ordinal: %w", err)
				}
			}
			out.Events = append(out.Events, ev)
		}
		return nil
	})
	if err != nil {
		return CreditAllocation{}, false, err
	}
	return out, fresh, nil
}

func (s *SQLiteStore) replayAllocation(ctx context.Context, conn *sql.Conn, studentID, attemptID string) (CreditAllocation, error) {
	out := CreditAllocation{StudentID: studentID, AttemptID: attemptID}
	if err := conn.QueryRowContext(ctx, `
		SELECT raw_units, allocated_units, target_count, dropped_count
		FROM implicit_credit_allocations WHERE student_id = ? AND source_attempt_id = ?
	`, studentID, attemptID).Scan(&out.RawUnits, &out.AllocatedUnits, &out.TargetCount, &out.DroppedCount); err != nil {
		return out, fmt.Errorf("credit: replay allocation: %w", err)
	}
	rows, err := conn.QueryContext(ctx, `
		SELECT target_concept_id, credit_epoch, scaled_units, retained_units, discarded_units, discard_reason
		FROM implicit_credit_events
		WHERE student_id = ? AND source_attempt_id = ?
		ORDER BY target_concept_id ASC
	`, studentID, attemptID)
	if err != nil {
		return out, fmt.Errorf("credit: replay events: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ev CreditEvent
		ev.AttemptID = attemptID
		if err := rows.Scan(&ev.ConceptID, &ev.Epoch, &ev.ScaledUnits, &ev.RetainedUnits, &ev.DiscardedUnits, &ev.DiscardReason); err != nil {
			return out, fmt.Errorf("credit: scan replay: %w", err)
		}
		out.Events = append(out.Events, ev)
	}
	if len(out.Events) != out.TargetCount {
		return out, fmt.Errorf("credit: allocation %s records %d targets but has %d events", attemptID, out.TargetCount, len(out.Events))
	}
	return out, rows.Err()
}

func sumRaw(c []repetition.Candidate) int64 {
	var n int64
	for _, x := range c {
		n += x.RawUnits
	}
	return n
}

func sumScaled(a []repetition.Allocation) int64 {
	var n int64
	for _, x := range a {
		n += x.Units
	}
	return n
}

func (s *SQLiteStore) AccountOf(studentID, conceptID string, epoch int64) (CreditAccount, error) {
	ctx := context.Background()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return CreditAccount{}, err
	}
	defer conn.Close()
	if epoch <= 0 {
		var err error
		epoch, err = s.targetEpoch(conn, ctx, studentID, conceptID)
		if err != nil {
			return CreditAccount{}, err
		}
	}
	return s.accountTx(ctx, conn, studentID, conceptID, epoch)
}

func (s *SQLiteStore) ProjectCredit(studentID, conceptID string, epoch int64, stepDays int, now time.Time) ([]CreditOrdinal, error) {
	if stepDays < 1 || stepDays > repetition.MaxStepDays {
		return nil, fmt.Errorf("credit: step %d outside [1,%d]", stepDays, repetition.MaxStepDays)
	}
	var out []CreditOrdinal
	err := s.runImmediate(context.Background(), func(conn *sql.Conn) error {
		ctx := context.Background()
		if epoch <= 0 {
			var err error
			epoch, err = s.targetEpoch(conn, ctx, studentID, conceptID)
			if err != nil {
				return err
			}
		}
		var due sql.NullString
		if err := conn.QueryRowContext(ctx, `
			SELECT next_review_due FROM concept_progress WHERE student_id = ? AND concept_id = ?
		`, studentID, conceptID).Scan(&due); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("credit: read due: %w", err)
		}
		if !due.Valid || due.String == "" {
			return nil
		}
		baseline := due.String
		rows, err := conn.QueryContext(ctx, `
			SELECT ordinal, COALESCE(counterfactual_due, ''), state, attempts
			FROM implicit_credit_ordinals
			WHERE student_id = ? AND target_concept_id = ? AND credit_epoch = ?
			ORDER BY ordinal ASC
		`, studentID, conceptID, epoch)
		if err != nil {
			return fmt.Errorf("credit: read ordinals: %w", err)
		}
		type ordinalState struct {
			ordinal int
			due     string
			state   string
			tries   int
		}
		var all []ordinalState
		for rows.Next() {
			var o ordinalState
			if err := rows.Scan(&o.ordinal, &o.due, &o.state, &o.tries); err != nil {
				rows.Close()
				return err
			}
			all = append(all, o)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		anchor, err := time.Parse(time.RFC3339, baseline)
		if err != nil {
			anchor, err = time.Parse("2006-01-02 15:04:05", baseline)
			if err != nil {
				return fmt.Errorf("credit: parse due %q: %w", baseline, err)
			}
		}
		prev := anchor
		for i := range all {
			o := &all[i]
			if o.state == "stale" {
				return fmt.Errorf("%w: ordinal %d", ErrCreditOrdinalStale, o.ordinal)
			}
			if o.state == "projected" {
				parsed, err := time.Parse(time.RFC3339, o.due)
				if err != nil {
					return fmt.Errorf("credit: parse stored counterfactual: %w", err)
				}
				prev = parsed
				out = append(out, CreditOrdinal{ConceptID: conceptID, Epoch: epoch, Ordinal: o.ordinal, State: o.state, BaselineDue: baseline, CounterfactualDue: o.due, Attempts: o.tries})
				continue
			}
			if o.state != "banked" {
				return fmt.Errorf("%w: ordinal %d is %q", ErrCreditOrdinalStateChanged, o.ordinal, o.state)
			}
			next := prev.AddDate(0, 0, stepDays)
			if floor := now.UTC().AddDate(0, 0, 1); next.Before(floor) {
				next = floor
			}
			res, err := conn.ExecContext(ctx, `
				UPDATE implicit_credit_ordinals
				SET state = 'projected', projected_at = ?, baseline_due = ?, counterfactual_due = ?,
				    attempts = attempts + 1
				WHERE student_id = ? AND target_concept_id = ? AND credit_epoch = ?
				  AND ordinal = ? AND state = 'banked'
			`, time.Now().UTC().UnixNano(), baseline, next.UTC().Format(time.RFC3339), studentID, conceptID, epoch, o.ordinal)
			if err != nil {
				return fmt.Errorf("credit: project ordinal: %w", err)
			}
			if n, _ := res.RowsAffected(); n != 1 {
				return fmt.Errorf("%w: ordinal %d", ErrCreditOrdinalStateChanged, o.ordinal)
			}
			prev = next
			out = append(out, CreditOrdinal{ConceptID: conceptID, Epoch: epoch, Ordinal: o.ordinal, State: "projected", BaselineDue: baseline, CounterfactualDue: next.UTC().Format(time.RFC3339)})
		}

		if s.creditProbe != nil {
			s.creditProbe(func(query string, args ...any) error {
				_, err := conn.ExecContext(ctx, query, args...)
				return err
			})
		}

		// The due date is the baseline the chain was built from. If the explicit
		// scheduler moved it while this transaction waited for the write lock,
		// the whole chain is stale and must not be committed.
		var nowDue sql.NullString
		if err := conn.QueryRowContext(ctx, `
			SELECT next_review_due FROM concept_progress WHERE student_id = ? AND concept_id = ?
		`, studentID, conceptID).Scan(&nowDue); err != nil {
			return fmt.Errorf("credit: reread due: %w", err)
		}
		if !nowDue.Valid || nowDue.String != baseline {
			return fmt.Errorf("%w: was %q, now %q", ErrCreditDueConflict, baseline, nowDue.String)
		}
		return nil
	})
	return out, err
}

func (s *SQLiteStore) ResetCreditEpoch(studentID, conceptID string) (int64, error) {
	var next int64
	err := s.runImmediate(context.Background(), func(conn *sql.Conn) error {
		ctx := context.Background()
		epoch, err := s.targetEpoch(conn, ctx, studentID, conceptID)
		if err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, `
			UPDATE implicit_credit_ordinals SET state = 'stale'
			WHERE student_id = ? AND target_concept_id = ? AND credit_epoch = ? AND state = 'banked'
		`, studentID, conceptID, epoch); err != nil {
			return fmt.Errorf("credit: stale ordinals: %w", err)
		}
		next = epoch + 1
		if _, err := conn.ExecContext(ctx, `
			UPDATE implicit_credit_targets SET credit_epoch = ?
			WHERE student_id = ? AND target_concept_id = ? AND credit_epoch = ?
		`, next, studentID, conceptID, epoch); err != nil {
			return fmt.Errorf("credit: bump epoch: %w", err)
		}
		return nil
	})
	return next, err
}

func (s *SQLiteStore) VoidCreditEvent(studentID, attemptID, conceptID, reason string) error {
	return s.runImmediate(context.Background(), func(conn *sql.Conn) error {
		ctx := context.Background()
		var epoch int64
		if err := conn.QueryRowContext(ctx, `
			SELECT credit_epoch FROM implicit_credit_targets
			WHERE student_id = ? AND target_concept_id = ?
		`, studentID, conceptID).Scan(&epoch); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrCreditOrdinalMissing
			}
			return err
		}
		var evEpoch, retained int64
		var voidedAt sql.NullInt64
		if err := conn.QueryRowContext(ctx, `
			SELECT credit_epoch, retained_units, voided_at FROM implicit_credit_events
			WHERE student_id = ? AND source_attempt_id = ? AND target_concept_id = ?
		`, studentID, attemptID, conceptID).Scan(&evEpoch, &retained, &voidedAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrCreditOrdinalMissing
			}
			return err
		}
		if evEpoch != epoch {
			return fmt.Errorf("%w: event epoch %d, target epoch %d", ErrCreditClosedEpoch, evEpoch, epoch)
		}
		if voidedAt.Valid {
			return nil
		}
		acc, err := s.accountTx(ctx, conn, studentID, conceptID, epoch)
		if err != nil {
			return err
		}
		if acc.RetainedUnits-retained < int64(acc.Banked)*repetition.Micro {
			return fmt.Errorf("%w: retained %d minus %d cannot fund %d ordinals", ErrCreditFundsProjection, acc.RetainedUnits, retained, acc.Banked)
		}
		if _, err := conn.ExecContext(ctx, `
			UPDATE implicit_credit_events
			SET voided_at = ?, voided_by = 'credit', void_reason = ?
			WHERE student_id = ? AND source_attempt_id = ? AND target_concept_id = ?
		`, time.Now().UTC().UnixNano(), reason, studentID, attemptID, conceptID); err != nil {
			return fmt.Errorf("credit: void event: %w", err)
		}
		return nil
	})
}

func (s *SQLiteStore) GetBankedOrdinals(studentID string) (map[string]int, error) {
	ctx := context.Background()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.QueryContext(ctx, `
		SELECT target_concept_id, COUNT(*)
		FROM implicit_credit_ordinals
		WHERE student_id = ? AND state = 'banked'
		GROUP BY target_concept_id
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("credit: get banked ordinals: %w", err)
	}
	defer rows.Close()
	out := make(map[string]int)
	for rows.Next() {
		var cid string
		var count int
		if err := rows.Scan(&cid, &count); err != nil {
			return nil, err
		}
		out[cid] = count
	}
	return out, rows.Err()
}
