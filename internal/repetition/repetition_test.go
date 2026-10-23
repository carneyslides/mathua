package repetition

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/scheduler"
)

func ids(allocs []Allocation) []string {
	out := make([]string, len(allocs))
	for i, a := range allocs {
		out[i] = a.ConceptID
	}
	return out
}

func TestAllocate_PassesThroughUnderTheCap(t *testing.T) {
	got, err := Allocate([]Candidate{
		{ConceptID: "b", RawUnits: 300_000},
		{ConceptID: "a", RawUnits: 200_000},
		{ConceptID: "c", RawUnits: 0},
	}, Micro)
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if len(got) != 2 || got[0].ConceptID != "a" || got[1].ConceptID != "b" {
		t.Fatalf("allocations = %+v, want a then b with the zero dropped", got)
	}
	if got[0].Units != 200_000 || got[1].Units != 300_000 {
		t.Errorf("units = %d/%d, want unchanged", got[0].Units, got[1].Units)
	}
}

func TestAllocate_ScalesProportionallyAndConservesTheCap(t *testing.T) {
	got, err := Allocate([]Candidate{
		{ConceptID: "a", RawUnits: 300_000},
		{ConceptID: "b", RawUnits: 300_000},
		{ConceptID: "c", RawUnits: 400_000},
	}, 500_000)
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	want := map[string]int64{"a": 150_000, "b": 150_000, "c": 200_000}
	var sum int64
	for _, a := range got {
		if a.Units != want[a.ConceptID] {
			t.Errorf("%s = %d, want %d", a.ConceptID, a.Units, want[a.ConceptID])
		}
		sum += a.Units
	}
	if sum != 500_000 {
		t.Errorf("sum = %d, want exactly the cap 500000", sum)
	}
}

func TestAllocate_TiesGoToTheLowestConceptID(t *testing.T) {
	// Three equal shares of a one-unit cap: every base is zero and every
	// remainder equal, so the single leftover must go to "a".
	got, err := Allocate([]Candidate{
		{ConceptID: "c", RawUnits: 100_000},
		{ConceptID: "b", RawUnits: 100_000},
		{ConceptID: "a", RawUnits: 100_000},
	}, 1)
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if len(got) != 1 || got[0].ConceptID != "a" || got[0].Units != 1 {
		t.Fatalf("allocations = %+v, want only a=1", got)
	}
}

func TestAllocate_IsIndependentOfInputOrder(t *testing.T) {
	forward, err := Allocate([]Candidate{
		{ConceptID: "a", RawUnits: 333_333},
		{ConceptID: "b", RawUnits: 333_333},
		{ConceptID: "c", RawUnits: 333_334},
	}, 700_000)
	if err != nil {
		t.Fatalf("forward: %v", err)
	}
	reverse, err := Allocate([]Candidate{
		{ConceptID: "c", RawUnits: 333_334},
		{ConceptID: "b", RawUnits: 333_333},
		{ConceptID: "a", RawUnits: 333_333},
	}, 700_000)
	if err != nil {
		t.Fatalf("reverse: %v", err)
	}
	if len(forward) != len(reverse) {
		t.Fatalf("lengths differ: %d vs %d", len(forward), len(reverse))
	}
	for i := range forward {
		if forward[i] != reverse[i] {
			t.Errorf("allocation %d differs: %+v vs %+v", i, forward[i], reverse[i])
		}
	}
}

func TestAllocate_ZeroCapEmitsNothing(t *testing.T) {
	got, err := Allocate([]Candidate{{ConceptID: "a", RawUnits: 500_000}}, 0)
	if err != nil {
		t.Fatalf("zero cap: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("zero cap emitted %+v", got)
	}
}

func TestAllocate_RejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		in   []Candidate
		cap  int64
		want error
	}{
		{"negative raw", []Candidate{{ConceptID: "a", RawUnits: -1}}, Micro, ErrNegativeRaw},
		{"raw over a repetition", []Candidate{{ConceptID: "a", RawUnits: Micro + 1}}, Micro, ErrRawTooLarge},
		{"negative cap", []Candidate{{ConceptID: "a", RawUnits: 1}}, -1, ErrNegativeCap},
		{"cap over a repetition", []Candidate{{ConceptID: "a", RawUnits: 1}}, Micro + 1, ErrCapTooLarge},
		{"duplicate target", []Candidate{{ConceptID: "a", RawUnits: 1}, {ConceptID: "a", RawUnits: 2}}, Micro, ErrDuplicateTarget},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Allocate(tc.in, tc.cap); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestAllocate_BoundsTheCandidateCount(t *testing.T) {
	many := make([]Candidate, MaxCandidates+1)
	for i := range many {
		many[i] = Candidate{ConceptID: string(rune('a'+i%26)) + string(rune('a'+i/26)), RawUnits: 1}
	}
	if _, err := Allocate(many, Micro); !errors.Is(err, ErrTooManyCandidates) {
		t.Fatalf("err = %v, want ErrTooManyCandidates", err)
	}
}

func TestAdmit_BankingBoundaries(t *testing.T) {
	var a Account
	next, retained, discarded, err := a.Admit(Micro - 1)
	if err != nil {
		t.Fatalf("admit: %v", err)
	}
	if retained != Micro-1 || discarded != 0 || next.Ordinals.Total() != 0 {
		t.Fatalf("partial admission = %+v retained=%d discarded=%d", next, retained, discarded)
	}
	if next.Balance() != Micro-1 {
		t.Errorf("balance = %d, want %d", next.Balance(), Micro-1)
	}

	next, _, _, err = next.Admit(1)
	if err != nil {
		t.Fatalf("admit one more: %v", err)
	}
	if next.Ordinals.Total() != 1 || next.Balance() != 0 {
		t.Errorf("after a full repetition: ordinals=%d balance=%d, want 1/0", next.Ordinals.Total(), next.Balance())
	}
}

func TestAdmit_CapacityIsLifetimeNotOutstanding(t *testing.T) {
	var a Account
	for i := 0; i < MaxBank; i++ {
		var err error
		a, _, _, err = a.Admit(Micro)
		if err != nil {
			t.Fatalf("admit %d: %v", i, err)
		}
	}
	if a.Ordinals.Total() != MaxBank {
		t.Fatalf("ordinals = %d, want %d", a.Ordinals.Total(), MaxBank)
	}

	// Projecting an ordinal retires it from pending but must not restore the
	// capacity that produced it: the cap is on admission for the epoch.
	projected, err := a.Project()
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if projected.Capacity() != a.Capacity() {
		t.Errorf("capacity restored by projection: %d then %d", a.Capacity(), projected.Capacity())
	}
	next, retained, discarded, err := projected.Admit(Micro)
	if err != nil {
		t.Fatalf("admit at cap: %v", err)
	}
	if retained != 0 || discarded != Micro {
		t.Errorf("retained/discarded = %d/%d, want 0/%d", retained, discarded, Micro)
	}
	if next.RetainedUnits != a.RetainedUnits {
		t.Errorf("credit admitted past the cap: %d", next.RetainedUnits)
	}

	// Staling behaves the same way.
	stale, err := a.Stale()
	if err != nil {
		t.Fatalf("stale: %v", err)
	}
	if stale.Capacity() != 0 {
		t.Errorf("capacity restored by staling: %d", stale.Capacity())
	}
}

func TestAdmit_RejectsOversizedCredit(t *testing.T) {
	var a Account
	if _, _, _, err := a.Admit(Micro + 1); !errors.Is(err, ErrCreditTooLarge) {
		t.Fatalf("err = %v, want ErrCreditTooLarge", err)
	}
	if _, _, _, err := a.Admit(-1); !errors.Is(err, ErrNegativeCredit) {
		t.Fatalf("err = %v, want ErrNegativeCredit", err)
	}

	// Admission on an invalid account fails at initial validation
	bad := Account{RetainedUnits: -1}
	if _, _, _, err := bad.Admit(100); !errors.Is(err, ErrNegativeCredit) {
		t.Errorf("admit on invalid account err = %v, want ErrNegativeCredit", err)
	}
}

func TestValidate_RejectsAFullFractionalBalance(t *testing.T) {
	bad := Account{RetainedUnits: Micro}
	if err := bad.Validate(); !errors.Is(err, ErrBalance) {
		t.Fatalf("err = %v, want ErrBalance", err)
	}
}

func TestStepDays_UsesTheExplicitReviewFormula(t *testing.T) {
	prev := scheduler.SM2{Repetitions: 4, EFactor: 2.5, Interval: 10}
	got, err := StepDays(prev, 1.0)
	if err != nil {
		t.Fatalf("step: %v", err)
	}
	if got != 25 {
		t.Errorf("step = %d, want 25 (10 x 2.5 x 1.0)", got)
	}

	// A NaN speed must not reach math.Round in the scheduler.
	got, err = StepDays(prev, math.NaN())
	if err != nil || got != 25 {
		t.Errorf("NaN speed step = %d err=%v, want 25", got, err)
	}

	// The first repetition has its own fixed step, which is why this must call
	// the real function instead of multiplying by the ease factor.
	got, err = StepDays(scheduler.SM2{Repetitions: 0, EFactor: 2.5}, 1.0)
	if err != nil || got != 1 {
		t.Errorf("first step = %d err=%v, want 1", got, err)
	}
}

func TestChainDue_ChainsAndPreservesTimeOfDay(t *testing.T) {
	baseline := time.Date(2026, 11, 1, 9, 30, 0, 0, time.UTC)
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	got, err := ChainDue(baseline, now, []int{25, 25})
	if err != nil {
		t.Fatalf("chain: %v", err)
	}
	want0 := time.Date(2026, 11, 26, 9, 30, 0, 0, time.UTC)
	want1 := time.Date(2026, 12, 21, 9, 30, 0, 0, time.UTC)
	if !got[0].Equal(want0) || !got[1].Equal(want1) {
		t.Errorf("chain = %s, %s; want %s, %s", got[0], got[1], want0, want1)
	}

	// A due date in the past is pulled to at least one day after now.
	late, err := ChainDue(now.AddDate(0, 0, -40), now, []int{1})
	if err != nil {
		t.Fatalf("late chain: %v", err)
	}
	if !late[0].Equal(now.AddDate(0, 0, 1)) {
		t.Errorf("late due = %s, want %s", late[0], now.AddDate(0, 0, 1))
	}
}

func TestNormalizeSpeed(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{math.NaN(), 1.0},
		{math.Inf(1), 1.0},
		{math.Inf(-1), 1.0},
		{0.0, 1.0},
		{-1.5, 1.0},
		{0.2, 0.5},
		{0.5, 0.5},
		{1.2, 1.2},
		{2.0, 2.0},
		{3.5, 2.0},
	}
	for _, tc := range cases {
		got := NormalizeSpeed(tc.in)
		if got != tc.want {
			t.Errorf("NormalizeSpeed(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestStepDays_Bounds(t *testing.T) {
	// Prev interval so large that next interval exceeds MaxStepDays.
	huge := scheduler.SM2{Interval: 40000, EFactor: 2.5, Repetitions: 5}
	if _, err := StepDays(huge, 1.0); !errors.Is(err, ErrStepRange) {
		t.Errorf("StepDays huge interval err = %v, want ErrStepRange", err)
	}

	// Repetitions >= 2 with zero interval results in 0, triggering MinStepDays clamp.
	zero := scheduler.SM2{Interval: 0, EFactor: 1.3, Repetitions: 2}
	got, err := StepDays(zero, 1.0)
	if err != nil || got != MinStepDays {
		t.Errorf("StepDays zero interval = %d err=%v, want %d", got, err, MinStepDays)
	}
}

func TestChainDue_Errors(t *testing.T) {
	now := time.Now().UTC()
	baseline := now

	// Step less than MinStepDays (1)
	if _, err := ChainDue(baseline, now, []int{0}); !errors.Is(err, ErrStepRange) {
		t.Errorf("ChainDue step 0 err = %v, want ErrStepRange", err)
	}

	// Step greater than MaxStepDays (36500)
	if _, err := ChainDue(baseline, now, []int{36501}); !errors.Is(err, ErrStepRange) {
		t.Errorf("ChainDue step 36501 err = %v, want ErrStepRange", err)
	}

	// Baseline due year exceeding 9000
	distantFuture := time.Date(9001, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := ChainDue(distantFuture, now, []int{10}); !errors.Is(err, ErrDueRange) {
		t.Errorf("ChainDue distant future err = %v, want ErrDueRange", err)
	}
}

func TestAccount_LifecycleAndInvariants(t *testing.T) {
	counts := OrdinalCounts{Banked: 2, Projected: 1, Stale: 0}
	if counts.Pending() != 2 {
		t.Errorf("Pending() = %d, want 2", counts.Pending())
	}
	if counts.Total() != 3 {
		t.Errorf("Total() = %d, want 3", counts.Total())
	}

	// NewAccount with valid state
	acct, err := NewAccount(Micro*3, counts)
	if err != nil {
		t.Fatalf("NewAccount valid err = %v", err)
	}
	if acct.Balance() != 0 {
		t.Errorf("Balance() = %d, want 0", acct.Balance())
	}

	// Negative retained units
	if _, err := NewAccount(-1, counts); !errors.Is(err, ErrNegativeCredit) {
		t.Errorf("NewAccount negative retained err = %v, want ErrNegativeCredit", err)
	}

	// Negative ordinal counts
	if _, err := NewAccount(Micro*2, OrdinalCounts{Banked: -1}); !errors.Is(err, ErrOrdinalTotal) {
		t.Errorf("NewAccount negative banked err = %v, want ErrOrdinalTotal", err)
	}

	// Ordinals exceeding MaxBank
	if _, err := NewAccount(Micro*(MaxBank+1), OrdinalCounts{Banked: int(MaxBank + 1)}); !errors.Is(err, ErrBankLimit) {
		t.Errorf("NewAccount over MaxBank err = %v, want ErrBankLimit", err)
	}

	// Balance >= Micro
	if _, err := NewAccount(Micro*2, OrdinalCounts{Banked: 0}); !errors.Is(err, ErrBalance) {
		t.Errorf("NewAccount balance >= Micro err = %v, want ErrBalance", err)
	}

	// Balance < 0
	if _, err := NewAccount(0, OrdinalCounts{Banked: 1}); !errors.Is(err, ErrBalance) {
		t.Errorf("NewAccount balance < 0 err = %v, want ErrBalance", err)
	}

	// Empty bank Project and Stale return ErrOrdinalTotal
	empty := Account{RetainedUnits: 0, Ordinals: OrdinalCounts{Banked: 0}}
	if _, err := empty.Project(); !errors.Is(err, ErrOrdinalTotal) {
		t.Errorf("empty Project err = %v, want ErrOrdinalTotal", err)
	}
	if _, err := empty.Stale(); !errors.Is(err, ErrOrdinalTotal) {
		t.Errorf("empty Stale err = %v, want ErrOrdinalTotal", err)
	}

	// Capacity clamped to 0 when RetainedUnits exceeds MaxBank*Micro
	overflowAcct := Account{RetainedUnits: int64(MaxBank+1) * Micro}
	if cap := overflowAcct.Capacity(); cap != 0 {
		t.Errorf("Capacity() for overflow = %d, want 0", cap)
	}
}

func TestCheckedArithmetic_Overflow(t *testing.T) {
	// addChecked positive overflow
	if _, err := addChecked(math.MaxInt64, 1); !errors.Is(err, ErrOverflow) {
		t.Errorf("addChecked positive overflow err = %v, want ErrOverflow", err)
	}

	// addChecked negative overflow
	if _, err := addChecked(math.MinInt64, -1); !errors.Is(err, ErrOverflow) {
		t.Errorf("addChecked negative overflow err = %v, want ErrOverflow", err)
	}

	// addChecked normal subtraction
	got, err := addChecked(10, -3)
	if err != nil || got != 7 {
		t.Errorf("addChecked(10, -3) = %d err=%v, want 7", got, err)
	}

	// mulChecked with zero operand
	got, err = mulChecked(0, 100)
	if err != nil || got != 0 {
		t.Errorf("mulChecked(0, 100) = %d err=%v, want 0", got, err)
	}
	got, err = mulChecked(100, 0)
	if err != nil || got != 0 {
		t.Errorf("mulChecked(100, 0) = %d err=%v, want 0", got, err)
	}

	// mulChecked overflow
	if _, err := mulChecked(math.MaxInt64, 2); !errors.Is(err, ErrOverflow) {
		t.Errorf("mulChecked overflow err = %v, want ErrOverflow", err)
	}

	// mulChecked normal multiplication
	got, err = mulChecked(6, 7)
	if err != nil || got != 42 {
		t.Errorf("mulChecked(6, 7) = %d err=%v, want 42", got, err)
	}
}
