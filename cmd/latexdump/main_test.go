package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createSampleLessonsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	lesson1 := `# Introduction
Here is an inline formula: \(x + y = z\).
And an equation:
\[
E = mc^2
\]
`
	if err := os.WriteFile(filepath.Join(dir, "lesson1.md"), []byte(lesson1), 0644); err != nil {
		t.Fatalf("write lesson1: %v", err)
	}

	subDir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	lesson2 := `# Advanced Topic
The Pythagorean theorem states that $a^2 + b^2 = c^2$.
`
	if err := os.WriteFile(filepath.Join(subDir, "lesson2.md"), []byte(lesson2), 0644); err != nil {
		t.Fatalf("write lesson2: %v", err)
	}

	// Non-markdown file should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a lesson"), 0644); err != nil {
		t.Fatalf("write notes.txt: %v", err)
	}

	return dir
}

func TestRun_DefaultBodiesMap(t *testing.T) {
	dir := createSampleLessonsDir(t)
	var stdout, stderr bytes.Buffer

	code := run([]string{dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	var bodies map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &bodies); err != nil {
		t.Fatalf("decode stdout JSON: %v", err)
	}

	if len(bodies) != 2 {
		t.Errorf("expected 2 lesson bodies, got %d: %+v", len(bodies), bodies)
	}

	if _, ok := bodies["lesson1.md"]; !ok {
		t.Errorf("expected lesson1.md in output bodies")
	}
	if _, ok := bodies["sub/lesson2.md"]; !ok {
		t.Errorf("expected sub/lesson2.md in output bodies")
	}
	if _, ok := bodies["notes.txt"]; ok {
		t.Errorf("notes.txt should not be present in output bodies")
	}
}

func TestRun_SpansJSONL(t *testing.T) {
	dir := createSampleLessonsDir(t)
	var stdout, stderr bytes.Buffer

	code := run([]string{"-spans", dir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr: %s", code, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		t.Fatal("expected math spans output, got empty stdout")
	}

	var spans []span
	for i, line := range lines {
		var s span
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			t.Fatalf("line %d invalid JSON: %v", i+1, err)
		}
		if s.File == "" || s.Kind == "" || s.Content == "" {
			t.Errorf("span on line %d missing fields: %+v", i+1, s)
		}
		spans = append(spans, s)
	}

	if len(spans) < 3 {
		t.Errorf("expected at least 3 math spans, got %d", len(spans))
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "latexdump: 2 files,") {
		t.Errorf("expected file count in stderr summary, got %q", stderrStr)
	}
	if !strings.Contains(stderrStr, "math spans") {
		t.Errorf("expected math spans in stderr summary, got %q", stderrStr)
	}
}

func TestRun_NonexistentDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	nonexistent := filepath.Join(t.TempDir(), "nonexistent-dir")

	code := run([]string{nonexistent}, &stdout, &stderr)
	if code != 1 {
		t.Errorf("expected exit code 1 for nonexistent dir, got %d", code)
	}

	if !strings.Contains(stderr.String(), "latexdump:") {
		t.Errorf("expected error prefix in stderr, got %q", stderr.String())
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"-unknownflag"}, &stdout, &stderr)
	if code != 2 {
		t.Errorf("expected exit code 2 for flag parse error, got %d", code)
	}
}
