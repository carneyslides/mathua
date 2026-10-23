package engine

import (
	"strings"
	"testing"
)

func TestEngine_CurriculumDomainAndDepths(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("curriculum_student")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	// Anonymous call (empty student ID)
	currAnon, err := e.CurriculumDomain("", "d")
	if err != nil {
		t.Fatalf("CurriculumDomain anon: %v", err)
	}
	if currAnon.Domain != "d" || len(currAnon.Topics) != 2 {
		t.Fatalf("currAnon = %+v, want 2 topics", currAnon)
	}

	// Concept "a" has no prerequisites (depth 0), "b" requires "a" (depth 1)
	depths := map[string]int{}
	for _, top := range currAnon.Topics {
		depths[top.ConceptID] = top.Depth
	}
	if depths["a"] != 0 {
		t.Errorf("depth[a] = %d, want 0", depths["a"])
	}
	if depths["b"] != 1 {
		t.Errorf("depth[b] = %d, want 1", depths["b"])
	}

	// Authenticated student call
	currAuth, err := e.CurriculumDomain(st.ID, "d")
	if err != nil {
		t.Fatalf("CurriculumDomain auth: %v", err)
	}
	if len(currAuth.Topics) != 2 {
		t.Errorf("len(currAuth.Topics) = %d, want 2", len(currAuth.Topics))
	}

	// Nonexistent domain error
	_, err = e.CurriculumDomain("", "nonexistent_domain")
	if err == nil {
		t.Fatal("expected error for nonexistent domain")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error = %v, want 'not found'", err)
	}
}

func TestEngine_RecommendPaused(t *testing.T) {
	e := testEngine(t)
	rec := e.RecommendPaused("student_paused")
	if rec.Primary == nil {
		t.Fatal("expected primary recommendation")
	}
	if rec.Primary.ID != "paused" {
		t.Errorf("Primary.ID = %q, want 'paused'", rec.Primary.ID)
	}
	if rec.Primary.Action.Href != "/settings" {
		t.Errorf("Primary.Action.Href = %q, want '/settings'", rec.Primary.Action.Href)
	}
	if rec.Primary.Badge != "Paused" {
		t.Errorf("Primary.Badge = %q, want 'Paused'", rec.Primary.Badge)
	}
}

func TestEngine_EfficacyTrend(t *testing.T) {
	e := testEngine(t)

	// Initially empty attempts
	trend, err := e.EfficacyTrend()
	if err != nil {
		t.Fatalf("EfficacyTrend empty: %v", err)
	}
	if len(trend.Weeks) != 0 {
		t.Errorf("expected 0 weeks on fresh engine, got %d", len(trend.Weeks))
	}

	// Answer question to record an attempt
	st, err := e.CreateStudent("efficacy_student")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}
	sess, err := e.repo.CreateSession(st.ID)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	q, err := e.NextQuestion(sess.ID, st.ID)
	if err != nil {
		t.Fatalf("NextQuestion: %v", err)
	}
	_, err = e.SubmitAnswer(sess.ID, st.ID, q.AttemptID, "42", 5.0)
	if err != nil {
		t.Fatalf("SubmitAnswer: %v", err)
	}

	trend, err = e.EfficacyTrend()
	if err != nil {
		t.Fatalf("EfficacyTrend after submit: %v", err)
	}
	if len(trend.Weeks) != 1 {
		t.Errorf("len(trend.Weeks) = %d, want 1", len(trend.Weeks))
	}
}

func TestEngine_StudentAndCatalogGetters(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("getter_student")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	loaded, err := e.GetStudent(st.ID)
	if err != nil {
		t.Fatalf("GetStudent: %v", err)
	}
	if loaded.ID != st.ID || loaded.Name != "getter_student" {
		t.Errorf("loaded = %+v, want student matching %s", loaded, st.ID)
	}

	// When planner is not configured
	if courses := e.PlannerCourses(); courses != nil {
		t.Errorf("PlannerCourses = %v, want nil", courses)
	}
	if cat, err := e.CourseCatalog(st.ID); err != nil || cat != nil {
		t.Errorf("CourseCatalog = (%v, %v), want (nil, nil)", cat, err)
	}
	if pl := e.GetPlanner(); pl != nil {
		t.Errorf("GetPlanner = %v, want nil", pl)
	}
	if gr := e.GetGrader(); gr == nil {
		t.Error("expected non-nil router from GetGrader")
	}
	if path := e.ActivePath(st.ID); path != nil {
		t.Errorf("ActivePath = %v, want nil", path)
	}
}

func TestEngine_CorpusSizeAndClose(t *testing.T) {
	e := testEngine(t)
	cCount, dCount := e.CorpusSize()
	if cCount != 2 || dCount != 1 {
		t.Errorf("CorpusSize = (%d, %d), want (2, 1)", cCount, dCount)
	}

	var nilEng *Engine
	cCount, dCount = nilEng.CorpusSize()
	if cCount != 0 || dCount != 0 {
		t.Errorf("nil CorpusSize = (%d, %d), want (0, 0)", cCount, dCount)
	}

	e.Close()
}

func TestEngine_QuizXPAndDifficulty(t *testing.T) {
	e := testEngine(t)
	st, err := e.CreateStudent("quiz_student")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	xp := e.QuizXP(true, 5.0, 10.0, 1)
	if xp <= 0 {
		t.Errorf("QuizXP = %d, want > 0", xp)
	}

	diff := e.DifficultyFor(st.ID, "a")
	if diff < 0.3 || diff > 1.0 {
		t.Errorf("DifficultyFor = %f, want within [0.3, 1.0]", diff)
	}
}

func TestEngine_ActiveCourseAndCatalog(t *testing.T) {
	e, _, _ := newTestEngineWithPlanner(t)
	st, err := e.CreateStudent("planner_student")
	if err != nil {
		t.Fatalf("CreateStudent: %v", err)
	}

	// Planner courses should not be empty
	courses := e.PlannerCourses()
	if len(courses) == 0 {
		t.Fatal("expected planner courses")
	}

	// Course catalog
	catalog, err := e.CourseCatalog(st.ID)
	if err != nil {
		t.Fatalf("CourseCatalog: %v", err)
	}
	if len(catalog) != len(courses) {
		t.Errorf("len(catalog) = %d, want %d", len(catalog), len(courses))
	}

	// Set active course
	firstCourse := courses[0].ID
	path, err := e.SetActiveCourse(st.ID, firstCourse)
	if err != nil {
		t.Fatalf("SetActiveCourse: %v", err)
	}
	if path == nil || len(path.Concepts) == 0 {
		t.Fatal("expected non-empty path for active course")
	}

	// ActivePath
	active := e.ActivePath(st.ID)
	if len(active) != len(path.Concepts) {
		t.Errorf("len(ActivePath) = %d, want %d", len(active), len(path.Concepts))
	}

	// SetActiveCourse with nil planner returns error
	noPlannerEngine := testEngine(t)
	if _, err := noPlannerEngine.SetActiveCourse(st.ID, firstCourse); err == nil {
		t.Fatal("expected error on SetActiveCourse with nil planner")
	}
}
