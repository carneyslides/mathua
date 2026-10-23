package storage

import (
	"errors"
	"time"

	"github.com/chuma-beep/mathua/internal/repetition"
)

// Shadow credit accounting. Nothing here writes a production scheduling column:
// a projection stores a counterfactual due date on the ordinal and leaves
// concept_progress.next_review_due exactly as the explicit scheduler set it.
// That is what makes the whole feature reversible by dropping tables.
var (
	ErrCreditOrdinalMissing      = errors.New("credit: ordinal not found")
	ErrCreditOrdinalStale        = errors.New("credit: ordinal is stale")
	ErrCreditOrdinalStateChanged = errors.New("credit: ordinal state changed")
	ErrCreditClosedEpoch         = errors.New("credit: epoch is closed")
	ErrCreditFundsProjection     = errors.New("credit: void would leave a banked ordinal unfunded")
	ErrCreditDueConflict         = errors.New("credit: the due date moved during projection")
)

// CreditCandidate is one target the attempt is believed to exercise, with the
// raw credit before the per-attempt cap is applied.
type CreditCandidate struct {
	ConceptID string
	RawUnits  int64
}

// CreditEvent is one target's share of one attempt's credit, after the cap.
type CreditEvent struct {
	AttemptID      string
	ConceptID      string
	Epoch          int64
	ScaledUnits    int64
	RetainedUnits  int64
	DiscardedUnits int64
	DiscardReason  string
	Banked         int
}

// CreditAllocation is the committed record of one attempt. Raw and allocated
// units are different numbers on purpose: raw is the pre-cap sum, allocated is
// what was actually emitted, so a replay can be checked without recomputing.
type CreditAllocation struct {
	StudentID      string
	AttemptID      string
	Epoch          int64
	RawUnits       int64
	AllocatedUnits int64
	TargetCount    int
	DroppedCount   int
	Events         []CreditEvent
}

// CreditOrdinal is one banked or projected repetition.
type CreditOrdinal struct {
	ConceptID         string
	Epoch             int64
	Ordinal           int
	State             string
	BaselineDue       string
	CounterfactualDue string
	Attempts          int
}

// CreditAccount is the derived balance for one target in one epoch. Retained is
// the sum of non-voided retained units; Banked counts every ordinal in the
// epoch, including projected and stale ones, because MaxBank is a lifetime cap.
type CreditAccount struct {
	ConceptID     string
	Epoch         int64
	RetainedUnits int64
	Banked        int
	Projected     int
	Stale         int
}

// Balance is the fractional credit left after banking: always in [0, Micro).
func (a CreditAccount) Balance() int64 {
	return a.RetainedUnits - int64(a.Banked)*repetition.Micro
}

// Pending returns the number of unprojected banked ordinals waiting to satisfy reviews.
func (a CreditAccount) Pending() int {
	p := a.Banked - a.Projected - a.Stale
	if p < 0 {
		return 0
	}
	return p
}

// CreditStore is the slice of the repository this stage needs. Declared
// separately so both backends implement exactly these operations and no
// caller can reach a production scheduling column through this door.
type CreditStore interface {
	// AllocateCredit applies the per-attempt cap, admits what capacity allows,
	// banks whole repetitions and commits everything in one transaction. The
	// bool is false when this attempt was already allocated, in which case the
	// committed rows are returned unchanged: a replay never banks, stales or
	// projects.
	AllocateCredit(studentID, attemptID string, epoch int64, candidates []CreditCandidate) (CreditAllocation, bool, error)
	// AccountOf recomputes the balance from the ledger rather than trusting a
	// stored counter.
	AccountOf(studentID, conceptID string, epoch int64) (CreditAccount, error)
	// ProjectCredit turns banked ordinals into counterfactual due dates chained
	// from the real one. It never moves the real due date.
	ProjectCredit(studentID, conceptID string, epoch int64, stepDays int, now time.Time) ([]CreditOrdinal, error)
	// ResetCreditEpoch stales banked ordinals and returns the new epoch.
	ResetCreditEpoch(studentID, conceptID string) (int64, error)
	// VoidCreditEvent removes one event's credit without deleting the row.
	VoidCreditEvent(studentID, attemptID, conceptID, reason string) error
	// GetBankedOrdinals returns the count of unprojected banked ordinals per concept for a student.
	GetBankedOrdinals(studentID string) (map[string]int, error)
}
