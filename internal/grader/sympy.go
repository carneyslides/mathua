package grader

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chuma-beep/mathua/internal/verify"
)

var sympyServicePath string
var findOnce sync.Once

func findSymPyService() string {
	findOnce.Do(func() {
		paths := []string{
			"grading/sympy_service.py",
			"../../grading/sympy_service.py",
			"../grading/sympy_service.py",
			"../../../grading/sympy_service.py",
		}
		if p := os.Getenv("SYMPY_SERVICE_PATH"); p != "" {
			paths = append([]string{p}, paths...)
		}
		if exe, err := os.Executable(); err == nil {
			d := filepath.Dir(exe)
			paths = append(paths,
				filepath.Join(d, "grading", "sympy_service.py"),
				filepath.Join(d, "..", "..", "grading", "sympy_service.py"),
			)
		}
		if wd, err := os.Getwd(); err == nil {
			for _, rel := range []string{"grading/sympy_service.py", "../../grading/sympy_service.py", "../../../grading/sympy_service.py"} {
				paths = append(paths, filepath.Join(wd, rel))
			}
		}
		for _, p := range paths {
			abs, _ := filepath.Abs(p)
			if _, err := os.Stat(abs); err == nil {
				sympyServicePath = abs
				return
			}
		}
	})
	return sympyServicePath
}

type sympyRequest struct {
	ID       string `json:"id"`
	Expected string `json:"expected"`
	Answer   string `json:"answer"`
}

type sympyResponse struct {
	ID       string `json:"id"`
	Correct  bool   `json:"correct"`
	Feedback string `json:"feedback"`
}

// sympyWorker encapsulates a persistent Python subprocess communicating via stdin and stdout json-lines.
type sympyWorker struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
}

func newSympyWorker(path string) (*sympyWorker, error) {
	cmd := exec.Command("python3", path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("pipe error: %w", err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("start error: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	return &sympyWorker{
		cmd:    cmd,
		stdin:  stdin,
		stdout: scanner,
	}, nil
}

func (w *sympyWorker) close() {
	if w.stdin != nil {
		w.stdin.Close()
	}
	if w.cmd != nil && w.cmd.Process != nil {
		w.cmd.Process.Kill()
		w.cmd.Wait()
	}
}

// sympyPool maintains a pool of N long-lived worker processes.
type sympyPool struct {
	mu          sync.Mutex
	path        string
	size        int
	idleWorkers chan *sympyWorker
}

func getWorkerCount() int {
	n := runtime.NumCPU()
	if n < 1 {
		return 1
	}
	if n > 4 {
		return 4
	}
	return n
}

var (
	poolOnce     sync.Once
	poolInstance *sympyPool
)

func getSympyPool() *sympyPool {
	poolOnce.Do(func() {
		path := findSymPyService()
		size := getWorkerCount()
		p := &sympyPool{
			path:        path,
			size:        size,
			idleWorkers: make(chan *sympyWorker, size),
		}
		if path != "" {
			for i := 0; i < size; i++ {
				w, err := newSympyWorker(path)
				if err == nil {
					p.idleWorkers <- w
				}
			}
		}
		poolInstance = p
	})
	return poolInstance
}

func (p *sympyPool) spawnReplacement() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.path == "" {
		return
	}
	w, err := newSympyWorker(p.path)
	if err == nil {
		select {
		case p.idleWorkers <- w:
		default:
			w.close()
		}
	}
}

func (p *sympyPool) Grade(ctx context.Context, expected, answer string) Result {
	select {
	case w := <-p.idleWorkers:
		res, ok := p.execOnWorker(w, expected, answer)
		if ok {
			p.idleWorkers <- w
			return res
		}
		// Worker crashed or timed out: terminate it and replace with fresh worker
		w.close()
		go p.spawnReplacement()
		return res
	case <-ctx.Done():
		return Result{Correct: false, Score: 0, Feedback: "Grading service busy", Unavailable: true}
	}
}

func (p *sympyPool) execOnWorker(w *sympyWorker, expected, answer string) (Result, bool) {
	req := sympyRequest{
		ID:       "1",
		Expected: expected,
		Answer:   answer,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return Result{Correct: false, Score: 0, Feedback: "Internal error"}, true
	}

	type execResult struct {
		line string
		err  error
	}
	ch := make(chan execResult, 1)

	go func() {
		if _, err := w.stdin.Write(append(data, '\n')); err != nil {
			ch <- execResult{err: err}
			return
		}
		if w.stdout.Scan() {
			ch <- execResult{line: w.stdout.Text()}
		} else {
			if err := w.stdout.Err(); err != nil {
				ch <- execResult{err: err}
			} else {
				ch <- execResult{err: io.EOF}
			}
		}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return Result{Correct: false, Score: 0, Feedback: "Grading service error", Unavailable: true}, false
		}
		var resp sympyResponse
		if err := json.Unmarshal([]byte(res.line), &resp); err != nil {
			return Result{Correct: false, Score: 0, Feedback: "Grading service response error", Unavailable: true}, false
		}
		if resp.Correct {
			return Result{Correct: true, Score: 1}, true
		}
		feedback := resp.Feedback
		if feedback == "" {
			feedback = "Incorrect"
		}
		return Result{Correct: false, Score: 0, Feedback: feedback}, true
	case <-time.After(2500 * time.Millisecond):
		return Result{Correct: false, Score: 0, Feedback: "Grading service timed out", Unavailable: true}, false
	}
}

// constEvalNumeric evaluates an expression containing only numbers and
// arithmetic operators (no variables, no functions). `^` is treated as a power.
// Returns false when the expression is not a plain constant.
func constEvalNumeric(expr string) (float64, bool) {
	e := strings.TrimSpace(expr)
	if e == "" {
		return 0, false
	}
	for _, r := range e {
		if !strings.ContainsRune("0123456789.+-*/^() ", r) {
			return 0, false
		}
	}
	e = strings.ReplaceAll(e, "^", "**")
	v, err := verify.Eval(e)
	if err != nil {
		return 0, false
	}
	return v, true
}

func gradeSymPy(expected, answer string) Result {
	// Fast path: exact match after normalization needs no CAS at all, and
	// guarantees that expected == answer is graded correct even if the SymPy
	// runtime is unavailable (the reported "You: 5^4 · Correct: 5^4" ✗ bug).
	if normaliseSymbolic(expected) == normaliseSymbolic(answer) {
		return Result{Correct: true, Score: 1}
	}
	// Constant numeric expressions are evaluated in Go, so numeric-valued
	// symbolic answers (5^4 vs 625) never depend on the subprocess either.
	if ev, ok := constEvalNumeric(expected); ok {
		if av, ok2 := constEvalNumeric(answer); ok2 {
			if math.Abs(ev-av) < 1e-9 {
				return Result{Correct: true, Score: 1}
			}
			return Result{Correct: false, Score: 0, Feedback: "Not equivalent"}
		}
	}

	pool := getSympyPool()
	if pool == nil || pool.path == "" {
		return Result{Correct: false, Score: 0, Feedback: "Grading service not found", Unavailable: true}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	return pool.Grade(ctx, expected, answer)
}
