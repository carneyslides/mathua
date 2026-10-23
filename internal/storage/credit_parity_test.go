package storage

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/repetition"
)

// creditStores returns every backend under test. Postgres is skipped when
// TEST_POSTGRES_DSN is unset, and the CI parity job sets it so a skip there
// would be visible as a missing test rather than a pass.
func creditStores(t *testing.T) map[string]Repository {
	t.Helper()
	out := map[string]Repository{"sqlite": newCreditStore(t)}
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		pg, err := NewPostgresStore(dsn)
		if err != nil {
			t.Fatalf("open postgres: %v", err)
		}
		t.Cleanup(func() { pg.Close() })
		out["postgres"] = pg
	}
	return out
}

func seedDueRepo(t *testing.T, repo Repository, studentID, conceptID string, due time.Time) {
	t.Helper()
	stmt := `INSERT INTO concept_progress (student_id, concept_id, status, next_review_due)
	         VALUES (?, ?, 'PRACTICING', ?)
	         ON CONFLICT(student_id, concept_id) DO UPDATE SET next_review_due = excluded.next_review_due`
	switch s := repo.(type) {
	case *SQLiteStore:
		if _, err := s.db.Exec(stmt, studentID, conceptID, due.UTC().Format("2006-01-02 15:04:05")); err != nil {
			t.Fatalf("seed due: %v", err)
		}
	case *PostgresStore:
		pg := `INSERT INTO concept_progress (student_id, concept_id, status, next_review_due)
		       VALUES ($1, $2, 'PRACTICING', $3)
		       ON CONFLICT (student_id, concept_id) DO UPDATE SET next_review_due = excluded.next_review_due`
		if _, err := s.db.Exec(pg, studentID, conceptID, due.UTC().Format("2006-01-02 15:04:05")); err != nil {
			t.Fatalf("seed due: %v", err)
		}
	default:
		t.Fatalf("seed due: unsupported store %T", repo)
	}
}

func readDueRepo(t *testing.T, repo Repository, studentID, conceptID string) string {
	t.Helper()
	var due string
	var err error
	switch s := repo.(type) {
	case *SQLiteStore:
		err = s.db.QueryRow(`SELECT next_review_due FROM concept_progress WHERE student_id = ? AND concept_id = ?`, studentID, conceptID).Scan(&due)
	case *PostgresStore:
		err = s.db.QueryRow(`SELECT next_review_due FROM concept_progress WHERE student_id = $1 AND concept_id = $2`, studentID, conceptID).Scan(&due)
	}
	if err != nil {
		t.Fatalf("read due: %v", err)
	}
	return due
}

// TestCreditParity runs the accounting contract against both backends: the same
// inputs must produce the same allocation, the same banking, the same replay,
// the same projection and the same refusal to underfund an ordinal.
func TestCreditParity(t *testing.T) {
	for tag, repo := range creditStores(t) {
		// Shadowed because the goroutines below outlive the loop iteration; a
		// captured range variable is how the wrong store tag reaches a worker.
		tag, repo := tag, repo
		t.Run(tag, func(t *testing.T) {
			credit := repo.(CreditStore)
			stu, err := repo.CreateStudent("parity credit " + tag)
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			due := time.Date(2026, 11, 1, 9, 30, 0, 0, time.UTC)
			seedDueRepo(t, repo, stu.ID, "a", due)

			alloc, fresh, err := credit.AllocateCredit(stu.ID, tag+":1", 1, []CreditCandidate{{ConceptID: "a", RawUnits: 1_000_000}})
			if err != nil || !fresh {
				t.Fatalf("allocate: %v fresh=%v", err, fresh)
			}
			if alloc.AllocatedUnits != repetition.Micro || alloc.Events[0].Banked != 1 {
				t.Fatalf("allocation = %+v", alloc)
			}
			replay, fresh, err := credit.AllocateCredit(stu.ID, tag+":1", 1, []CreditCandidate{{ConceptID: "a", RawUnits: 1_000_000}})
			if err != nil || fresh {
				t.Fatalf("replay: %v fresh=%v", err, fresh)
			}
			if replay.AllocatedUnits != alloc.AllocatedUnits {
				t.Errorf("replay = %d, want %d", replay.AllocatedUnits, alloc.AllocatedUnits)
			}
			acc, err := credit.AccountOf(stu.ID, "a", 1)
			if err != nil || acc.Banked != 1 || acc.RetainedUnits != repetition.Micro {
				t.Fatalf("account = %+v err=%v", acc, err)
			}

			now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
			ords, err := credit.ProjectCredit(stu.ID, "a", 1, 25, now)
			if err != nil || len(ords) != 1 || ords[0].CounterfactualDue != "2026-11-26T09:30:00Z" {
				t.Fatalf("project = %+v err=%v", ords, err)
			}
			if got := readDueRepo(t, repo, stu.ID, "a"); got != "2026-11-01 09:30:00" {
				t.Errorf("production due moved to %q", got)
			}
			if _, err := credit.ProjectCredit(stu.ID, "a", 1, 25, now); err != nil {
				t.Errorf("projection retry: %v", err)
			}

			if _, _, err := credit.AllocateCredit(stu.ID, tag+":2", 1, []CreditCandidate{{ConceptID: "a", RawUnits: 500_000}}); err != nil {
				t.Fatalf("second allocate: %v", err)
			}
			if err := credit.VoidCreditEvent(stu.ID, tag+":1", "a", "test"); err == nil {
				t.Error("voided the only event funding a banked ordinal")
			}
			if err := credit.VoidCreditEvent(stu.ID, tag+":2", "a", "test"); err != nil {
				t.Errorf("void of unfunded-by-ordinal event: %v", err)
			}

			next, err := credit.ResetCreditEpoch(stu.ID, "a")
			if err != nil || next != 2 {
				t.Fatalf("reset = %d err=%v", next, err)
			}
			// Projection already consumed the only whole repetition, so reset has
			// no banked ordinal to stale: history keeps it as projected and the
			// old epoch must not gain a second usable repetition.
			old, _ := credit.AccountOf(stu.ID, "a", 1)
			if old.Banked != 1 || old.Balance() != 0 {
				t.Errorf("old epoch = %+v balance=%d", old, old.Balance())
			}
		})
	}
}

func TestCreditParity_ConcurrentDuplicatesInsertOnce(t *testing.T) {
	for tag, repo := range creditStores(t) {
		tag := tag
		repo := repo
		t.Run(tag, func(t *testing.T) {
			credit := repo.(CreditStore)
			stu, err := repo.CreateStudent("concurrent credit " + tag)
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			const n = 6
			var wg sync.WaitGroup
			errs := make([]error, n)
			fresh := make([]bool, n)
			for i := 0; i < n; i++ {
				wg.Add(1)
				go func(i int, attemptKey string) {
					defer wg.Done()
					_, fresh[i], errs[i] = credit.AllocateCredit(stu.ID, attemptKey, 1,
						[]CreditCandidate{{ConceptID: "a", RawUnits: 1_000_000}})
				}(i, tag+":dup")
			}
			wg.Wait()
			winners := 0
			for i := range errs {
				if errs[i] != nil {
					t.Fatalf("worker %d: %v", i, errs[i])
				}
				if fresh[i] {
					winners++
				}
			}
			if winners != 1 {
				t.Errorf("%d fresh allocations, want 1", winners)
			}
			acc, err := credit.AccountOf(stu.ID, "a", 1)
			if err != nil || acc.Banked != 1 {
				t.Errorf("account = %+v err=%v, want one ordinal", acc, err)
			}
		})
	}
}
