package grader

import (
	"sync"
	"testing"
)

func TestSympyWorkerPool20Concurrent(t *testing.T) {
	if !sympyAvailable(t) {
		t.Skip("SymPy not available")
	}

	testCases := []struct {
		expected, answer string
	}{
		{"x^2+2x+1", "(x+1)^2"},
		{"6x^2", "6*x^2"},
		{"(x+3)(x+4)", "x^2+7x+12"},
		{"e^x", "exp(x)"},
		{"2e^(2x)", "2*exp(2*x)"},
		{"cos(x)", "cos(x)"},
		{"3(2x+1)^2", "12x^2+12x+3"},
		{"y = (z - 2x)/3", "(z-2*x)/3"},
		{"xe^x-e^x+C", "e^x*(x-1)+C"},
		{"(x+5)(x-5)", "x^2-25"},
		{"(x+1)(x+2)", "x^2+3x+2"},
		{"x^3+3x^2+3x+1", "(x+1)^3"},
		{"sin(x)^2+cos(x)^2", "1"},
		{"tan(x)", "sin(x)/cos(x)"},
		{"(2x+3)(2x-3)", "4x^2-9"},
		{"x^2-4x+4", "(x-2)^2"},
		{"2*(x+1)", "2x+2"},
		{"(x+2)(x+3)", "x^2+5x+6"},
		{"x*(x+1)", "x^2+x"},
		{"(x-1)*(x+1)", "x^2-1"},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(testCases))

	for i, tc := range testCases {
		wg.Add(1)
		go func(idx int, expected, answer string) {
			defer wg.Done()
			res := gradeSymPy(expected, answer)
			if !res.Correct {
				t.Errorf("worker %d: expected %q ≈ %q to be correct, got feedback: %q", idx, expected, answer, res.Feedback)
			}
		}(i, tc.expected, tc.answer)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

func TestSympyWorkerPoolRecoversFromWorkerCrash(t *testing.T) {
	if !sympyAvailable(t) {
		t.Skip("SymPy not available")
	}

	pool := getSympyPool()
	if pool == nil {
		t.Fatal("pool is nil")
	}

	// Acquire one worker and forcibly kill its process
	var w *sympyWorker
	select {
	case w = <-pool.idleWorkers:
	default:
		t.Fatal("no worker available in pool")
	}

	if w.cmd != nil && w.cmd.Process != nil {
		w.cmd.Process.Kill()
		w.cmd.Wait()
	}

	// Return dead worker back to pool to trigger crash detection on next call
	pool.idleWorkers <- w

	// Grade call should notice error on dead worker, spawn a replacement, and succeed on retry
	res := gradeSymPy("x^2+2x+1", "(x+1)^2")
	// The call on the dead worker detects failure, replaces worker, and returns fail-soft or recovers
	if !res.Correct && !res.Unavailable {
		t.Errorf("expected correct or fail-soft unavailable, got feedback: %q", res.Feedback)
	}

	// Next call must succeed with the replacement worker
	res2 := gradeSymPy("x^2+2x+1", "(x+1)^2")
	if !res2.Correct {
		t.Errorf("expected replacement worker to successfully grade, got feedback: %q", res2.Feedback)
	}
}

func BenchmarkSympyWorkerPool20Concurrent(b *testing.B) {
	if findSymPyService() == "" {
		b.Skip("SymPy service not found")
	}

	testCases := []struct {
		expected, answer string
	}{
		{"x^2+2x+1", "(x+1)^2"},
		{"(x+3)(x+4)", "x^2+7x+12"},
		{"3(2x+1)^2", "12x^2+12x+3"},
		{"(x+5)(x-5)", "x^2-25"},
		{"(x+1)(x+2)", "x^2+3x+2"},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tc := testCases[i%len(testCases)]
			i++
			res := gradeSymPy(tc.expected, tc.answer)
			if !res.Correct {
				b.Errorf("expected %q ≈ %q, got %q", tc.expected, tc.answer, res.Feedback)
			}
		}
	})
}
