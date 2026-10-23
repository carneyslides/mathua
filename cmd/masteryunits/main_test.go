package main

import (
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

func TestSizesOf_MedianOf_CountWhere(t *testing.T) {
	mu := &concepts.MasteryUnits{
		Units: []concepts.MasteryUnit{
			{ID: "u1", ConceptCount: 5},
			{ID: "u2", ConceptCount: 12},
			{ID: "u3", ConceptCount: 2},
			{ID: "u4", ConceptCount: 8},
			{ID: "u5", ConceptCount: 1},
		},
	}

	sizes := sizesOf(mu)
	wantSizes := []int{1, 2, 5, 8, 12}
	for i, v := range sizes {
		if v != wantSizes[i] {
			t.Errorf("sizes[%d] = %d, want %d", i, v, wantSizes[i])
		}
	}

	med := medianOf(sizes)
	if med != 5 {
		t.Errorf("medianOf(sizes) = %d, want 5", med)
	}

	if medianOf(nil) != 0 {
		t.Errorf("medianOf(nil) = %d, want 0", medianOf(nil))
	}

	singletons := countWhere(sizes, func(n int) bool { return n == 1 })
	if singletons != 1 {
		t.Errorf("countWhere(n == 1) = %d, want 1", singletons)
	}

	ge5 := countWhere(sizes, func(n int) bool { return n >= 5 })
	if ge5 != 3 {
		t.Errorf("countWhere(n >= 5) = %d, want 3", ge5)
	}
}

func TestRun_GeneratesPartitionAndReport(t *testing.T) {
	chdirRepoRoot(t)

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "mastery_units.json")
	reportPath := filepath.Join(tmpDir, "mastery_units.report.md")

	if err := run(3, outPath, reportPath); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	// Verify partition file exists and validates.
	mu, err := concepts.LoadMasteryUnits(outPath)
	if err != nil {
		t.Fatalf("load generated units: %v", err)
	}
	if mu.BandWidth != 3 {
		t.Errorf("mu.BandWidth = %d, want 3", mu.BandWidth)
	}
	if mu.ConceptCount == 0 || mu.UnitCount == 0 {
		t.Errorf("empty partition: concepts=%d units=%d", mu.ConceptCount, mu.UnitCount)
	}

	// Verify report file was generated and carries expected sections.
	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	reportStr := string(reportBytes)
	for _, header := range []string{
		"# Mastery units: draft boundary review",
		"## Size distribution",
		"## Proposed units",
		"## Backward band edges",
		"## Band width sweep",
		"## Why depth is measured per concept",
	} {
		if !strings.Contains(reportStr, header) {
			t.Errorf("report missing section %q", header)
		}
	}
}
