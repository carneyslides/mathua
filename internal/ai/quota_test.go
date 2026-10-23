package ai

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type failStore struct{ err error }

func (f failStore) ReserveAIUsage(context.Context, string, []AIUsageLimit) (bool, error) {
	return false, f.err
}
func (f failStore) AIUsage(context.Context, string, string) (int, error) {
	return 0, f.err
}

func TestQuotaReserveCapsAtLimit(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 3, 0)
	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		ok, err := q.Reserve(ctx, "user-1")
		if err != nil || !ok {
			t.Fatalf("reserve %d: ok=%v err=%v", i, ok, err)
		}
	}
	ok, err := q.Reserve(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("fourth reserve must be denied")
	}
	// A denied request spends nothing, so the counter stays at the ceiling
	// rather than growing past it. The refusal is what costs the caller, and the
	// write limiter is what bounds the refusal attempts.
	used, err := q.Used(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if used != 3 {
		t.Errorf("used = %d, want 3", used)
	}
}

func TestQuotaIsPerIdentity(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 1, 0)
	ctx := context.Background()
	if ok, _ := q.Reserve(ctx, "a"); !ok {
		t.Fatal("first identity should have its own budget")
	}
	if ok, _ := q.Reserve(ctx, "b"); !ok {
		t.Fatal("second identity should have its own budget")
	}
	if ok, _ := q.Reserve(ctx, "a"); ok {
		t.Error("first identity must be capped")
	}
}

// TestQuotaGlobalCeilingStopsEveryone is the reason the global cap exists: a
// per-identity cap bounds one learner, not the shared free tier, so the third
// distinct identity is refused even though it has never made a request.
func TestQuotaGlobalCeilingStopsEveryone(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 10, 3)
	ctx := context.Background()
	for i, id := range []string{"a", "b", "c"} {
		if ok, err := q.Reserve(ctx, id); !ok || err != nil {
			t.Fatalf("identity %s (%d) should be within the shared ceiling: ok=%v err=%v", id, i, ok, err)
		}
	}
	if ok, err := q.Reserve(ctx, "d"); ok || err != nil {
		t.Errorf("the shared ceiling must refuse a fresh identity: ok=%v err=%v", ok, err)
	}
	// The refusal must not have spent identity "d"'s own budget.
	if used, _ := q.Used(ctx, "d"); used != 0 {
		t.Errorf("refused identity spent %d of its own budget, want 0", used)
	}
}

// TestQuotaGlobalCounterIsMaintained pins that the shared ceiling is a real
// counter, not a derived guess: every admitted request moves it.
func TestQuotaGlobalCounterIsMaintained(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 10, 10)
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		if ok, err := q.Reserve(ctx, "user"); !ok || err != nil {
			t.Fatalf("reserve %d: ok=%v err=%v", i, ok, err)
		}
	}
	if used, _ := q.Used(ctx, GlobalQuotaKey); used != 4 {
		t.Errorf("global counter = %d, want 4", used)
	}
}

func TestQuotaConcurrentReservations(t *testing.T) {
	const cap = 20
	const goroutines = 200
	q := NewQuota(NewMemUsageStore(), cap, 0)

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := q.Reserve(context.Background(), "shared")
			if err != nil {
				return
			}
			if ok {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != cap {
		t.Errorf("concurrent allowed = %d, want exactly %d", allowed, cap)
	}
}

func TestQuotaUnlimitedWhenCapsZero(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 0, 0)
	for i := 0; i < 1000; i++ {
		ok, err := q.Reserve(context.Background(), "user")
		if err != nil || !ok {
			t.Fatalf("caps of zero must allow every request: ok=%v err=%v", ok, err)
		}
	}
}

func TestQuotaGlobalCapAppliesWithNoUserCap(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 0, 2)
	ctx := context.Background()
	if ok, _ := q.Reserve(ctx, "a"); !ok {
		t.Fatal("first request should pass")
	}
	if ok, _ := q.Reserve(ctx, "b"); !ok {
		t.Fatal("second request should pass")
	}
	if ok, _ := q.Reserve(ctx, "c"); ok {
		t.Error("the global cap must apply even when the per-identity cap is disabled")
	}
}

func TestQuotaNilIsUnlimited(t *testing.T) {
	var q *Quota
	ok, err := q.Reserve(context.Background(), "user")
	if err != nil || !ok {
		t.Fatalf("nil quota must allow: ok=%v err=%v", ok, err)
	}
}

func TestQuotaEmptyKeyDenied(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 5, 5)
	ok, err := q.Reserve(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("an empty identity must be denied, not silently allowed")
	}
}

func TestQuotaStoreErrorFailsClosed(t *testing.T) {
	q := NewQuota(failStore{err: errors.New("db down")}, 5, 5)
	ok, err := q.Reserve(context.Background(), "user")
	if err == nil {
		t.Fatal("want error")
	}
	if ok {
		t.Error("a store failure must deny, not allow")
	}
}

func TestQuotaRollsOverByDay(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 1, 0)
	ctx := context.Background()

	day := time.Date(2026, 10, 16, 23, 0, 0, 0, time.UTC)
	q.now = func() time.Time { return day }
	if ok, _ := q.Reserve(ctx, "user"); !ok {
		t.Fatal("first request of the day should pass")
	}
	if ok, _ := q.Reserve(ctx, "user"); ok {
		t.Fatal("second request of the same day should be denied")
	}

	// A day later the budget is fresh, even though UTC has only rolled two
	// hours forward from the 23:00 start.
	q.now = func() time.Time { return day.Add(2 * time.Hour) }
	if ok, _ := q.Reserve(ctx, "user"); !ok {
		t.Error("a new UTC day must restore the budget")
	}
}

func TestQuotaCapGettersAndUsedEdgeCases(t *testing.T) {
	q := NewQuota(NewMemUsageStore(), 7, 42)
	if q.Cap() != 7 {
		t.Errorf("Cap = %d, want 7", q.Cap())
	}
	if q.GlobalCap() != 42 {
		t.Errorf("GlobalCap = %d, want 42", q.GlobalCap())
	}

	var nilQ *Quota
	if nilQ.Cap() != 0 {
		t.Errorf("nil Cap = %d, want 0", nilQ.Cap())
	}
	if nilQ.GlobalCap() != 0 {
		t.Errorf("nil GlobalCap = %d, want 0", nilQ.GlobalCap())
	}
	used, err := nilQ.Used(context.Background(), "user")
	if err != nil || used != 0 {
		t.Errorf("nil Used = (%d, %v), want (0, nil)", used, err)
	}
	used, err = q.Used(context.Background(), "")
	if err != nil || used != 0 {
		t.Errorf("empty key Used = (%d, %v), want (0, nil)", used, err)
	}
}
