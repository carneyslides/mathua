package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/concepts"
)

func chdirRepoRoot(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("data/concepts"); os.IsNotExist(err) {
		if _, err := os.Stat("../../data/concepts"); err == nil {
			if err := os.Chdir("../.."); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chdir(orig) })
		}
	}
}

func TestRun_ValidMasteryUnits(t *testing.T) {
	chdirRepoRoot(t)

	var stdout, stderr bytes.Buffer
	err := run(false, concepts.MasteryUnitsPath, "data/concepts", &stdout, &stderr)
	if err != nil {
		t.Fatalf("expected run to succeed, got error: %v, stderr: %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "OK: ") {
		t.Errorf("stdout missing expected 'OK: ', got: %s", out)
	}
	if !strings.Contains(out, "anomaly: singleton_units") {
		t.Errorf("stdout missing anomaly report for singleton_units, got: %s", out)
	}
}

func TestRun_Strict_FailsOnAnomalies(t *testing.T) {
	chdirRepoRoot(t)

	var stdout, stderr bytes.Buffer
	err := run(true, concepts.MasteryUnitsPath, "data/concepts", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected strict run to fail on anomalies, got nil")
	}
	if !strings.Contains(err.Error(), "--strict") {
		t.Errorf("expected error mentioning --strict, got: %v", err)
	}
}

func TestRun_MissingConceptsDir(t *testing.T) {
	chdirRepoRoot(t)

	var stdout, stderr bytes.Buffer
	err := run(false, concepts.MasteryUnitsPath, "nonexistent/dir", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing concepts dir, got nil")
	}
}

func TestRun_MissingPartitionPath(t *testing.T) {
	chdirRepoRoot(t)

	var stdout, stderr bytes.Buffer
	err := run(false, "nonexistent/mastery_units.json", "data/concepts", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing partition file, got nil")
	}
}

func TestRun_ReportsPartitionErrors(t *testing.T) {
	chdirRepoRoot(t)

	// Create an invalid mastery_units.json referencing a fake concept ID.
	invalidMU := &concepts.MasteryUnits{
		Draft:        true,
		GeneratedBy:  "test",
		Rule:         "test",
		BandWidth:    3,
		ConceptCount: 1,
		UnitCount:    1,
		Units: []concepts.MasteryUnit{
			{
				ID:           "test_unit",
				Domain:       "arithmetic",
				Band:         0,
				ConceptCount: 1,
				ConceptIDs:   []string{"nonexistent.concept.id"},
			},
		},
	}
	data, err := json.Marshal(invalidMU)
	if err != nil {
		t.Fatal(err)
	}

	tmpFile := filepath.Join(t.TempDir(), "invalid_units.json")
	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err = run(false, tmpFile, "data/concepts", &stdout, &stderr)
	if err == nil {
		t.Fatal("expected run to fail on invalid partition, got nil")
	}
	if !strings.Contains(err.Error(), "partition error(s)") {
		t.Errorf("expected error message to contain 'partition error(s)', got: %v", err)
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "error: dangling_concept") {
		t.Errorf("stderr missing dangling_concept error, got: %s", errOut)
	}
}
