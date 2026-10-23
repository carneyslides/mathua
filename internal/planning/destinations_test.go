package planning

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chuma-beep/mathua/internal/storage"
)

func TestLoadDestinations_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "destinations.json")
	content := `[
		{"id":"bundle-b","name":"Bundle B","description":"Desc B","courses":["c2"]},
		{"id":"bundle-a","name":"Bundle A","description":"Desc A","courses":["c1"]}
	]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	ds, err := LoadDestinations(path)
	if err != nil {
		t.Fatalf("LoadDestinations: %v", err)
	}
	if len(ds) != 2 {
		t.Fatalf("expected 2 destinations, got %d", len(ds))
	}
	if ds[0].ID != "bundle-b" || ds[1].ID != "bundle-a" {
		t.Errorf("unexpected destinations loaded: %+v", ds)
	}
}

func TestLoadDestinations_Errors(t *testing.T) {
	dir := t.TempDir()

	// Missing file
	if _, err := LoadDestinations(filepath.Join(dir, "missing.json")); err == nil {
		t.Error("expected error for missing file")
	}

	// Malformed JSON
	badJSON := filepath.Join(dir, "bad.json")
	os.WriteFile(badJSON, []byte(`not json`), 0644)
	if _, err := LoadDestinations(badJSON); err == nil {
		t.Error("expected error for bad json")
	}

	// Empty ID
	emptyID := filepath.Join(dir, "empty_id.json")
	os.WriteFile(emptyID, []byte(`[{"id":"","name":"A","courses":["c1"]}]`), 0644)
	if _, err := LoadDestinations(emptyID); err == nil {
		t.Error("expected error for empty destination id")
	}

	// Duplicate ID
	dupID := filepath.Join(dir, "dup_id.json")
	os.WriteFile(dupID, []byte(`[
		{"id":"same","name":"A","courses":["c1"]},
		{"id":"same","name":"B","courses":["c2"]}
	]`), 0644)
	if _, err := LoadDestinations(dupID); err == nil {
		t.Error("expected error for duplicate destination id")
	}

	// Empty courses
	emptyCourses := filepath.Join(dir, "empty_courses.json")
	os.WriteFile(emptyCourses, []byte(`[{"id":"nodest","name":"A","courses":[]}]`), 0644)
	if _, err := LoadDestinations(emptyCourses); err == nil {
		t.Error("expected error for empty courses")
	}
}

func TestPlanner_DestinationsRegistry(t *testing.T) {
	p := testPlanner(t)

	bundles := []*Destination{
		{ID: "dest-2", Name: "Two", Description: "Two desc", Courses: []string{"test"}},
		{ID: "dest-1", Name: "One", Description: "One desc", Courses: []string{"test"}},
	}

	p.SetDestinations(bundles)

	all := p.Destinations()
	if len(all) != 2 {
		t.Fatalf("expected 2 destinations, got %d", len(all))
	}
	if all[0].ID != "dest-1" || all[1].ID != "dest-2" {
		t.Errorf("expected sorted destinations [dest-1 dest-2], got [%s %s]", all[0].ID, all[1].ID)
	}

	if d := p.Destination("dest-1"); d == nil || d.Name != "One" {
		t.Errorf("expected dest-1, got %+v", d)
	}
	if d := p.Destination("unknown"); d != nil {
		t.Errorf("expected nil for unknown destination, got %+v", d)
	}
}

func TestPlanner_PathForDestination(t *testing.T) {
	p := testPlanner(t)

	bundles := []*Destination{
		{ID: "valid-dest", Name: "Valid", Description: "Desc", Courses: []string{"test"}},
		{ID: "unknown-course", Name: "Bad Course", Description: "Desc", Courses: []string{"nonexistent-course"}},
	}
	p.SetDestinations(bundles)

	// Valid path resolution
	path, err := p.PathForDestination("valid-dest")
	if err != nil {
		t.Fatalf("PathForDestination: %v", err)
	}
	if len(path.Concepts) != 4 {
		t.Errorf("expected 4 concepts in destination path, got %d", len(path.Concepts))
	}

	// Nonexistent destination
	if _, err := p.PathForDestination("nonexistent-dest"); err == nil {
		t.Error("expected error for nonexistent destination")
	}

	// Destination referencing unknown course
	if _, err := p.PathForDestination("unknown-course"); err == nil {
		t.Error("expected error for unknown course reference")
	}
}

func TestPlanner_UnmasteredWeakAreasReadiness(t *testing.T) {
	p := testPlanner(t)
	path, err := p.PathForCourse("test")
	if err != nil {
		t.Fatalf("path: %v", err)
	}

	// Case 1: Empty progress
	unmastered := p.Unmastered(path, nil)
	if len(unmastered) != 4 {
		t.Errorf("expected 4 unmastered, got %d", len(unmastered))
	}
	weak := p.WeakAreas(path, nil)
	if len(weak) != 4 {
		t.Errorf("expected 4 weak areas, got %d", len(weak))
	}
	readiness := p.Readiness(path, nil)
	if readiness != 0.0 {
		t.Errorf("expected readiness 0.0, got %f", readiness)
	}

	// Case 2: Partial progress (2 mastered out of 4)
	partial := map[string]*storage.ConceptProgress{
		"a": {ConceptID: "a", Status: "MASTERED"},
		"b": {ConceptID: "b", Status: "MASTERED"},
		"c": {ConceptID: "c", Status: "PRACTICED"},
	}
	unmasteredPartial := p.Unmastered(path, partial)
	if len(unmasteredPartial) != 2 {
		t.Errorf("expected 2 unmastered, got %d", len(unmasteredPartial))
	}
	readinessPartial := p.Readiness(path, partial)
	if readinessPartial != 0.5 {
		t.Errorf("expected readiness 0.5, got %f", readinessPartial)
	}

	// Case 3: Empty path readiness
	emptyPath := &Path{}
	if r := p.Readiness(emptyPath, partial); r != 1.0 {
		t.Errorf("expected readiness 1.0 for empty path, got %f", r)
	}
}

func TestPlanner_SortCourses(t *testing.T) {
	p := testPlanner(t)
	courses := []*Course{
		{ID: "z-course", Name: "Z"},
		{ID: "a-course", Name: "A"},
		{ID: "m-course", Name: "M"},
	}

	p.SortCourses(courses)

	if courses[0].ID != "a-course" || courses[1].ID != "m-course" || courses[2].ID != "z-course" {
		t.Errorf("expected sorted courses [a m z], got [%s %s %s]", courses[0].ID, courses[1].ID, courses[2].ID)
	}
}
