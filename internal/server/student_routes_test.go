package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/auth"
	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/engine"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/planning"
	"github.com/chuma-beep/mathua/internal/storage"
)

func setupTestServerWithPlanner(t *testing.T) (*Server, *http.ServeMux, storage.Repository, string, string) {
	t.Helper()
	d, err := concepts.Build([]concepts.Concept{
		{ID: "a", Label: "Concept A", Domain: "d", Subdomain: "s-a", GradingType: "numeric", Prerequisites: []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 60}},
		{ID: "b", Label: "Concept B", Domain: "d", Subdomain: "s-b", GradingType: "numeric", Prerequisites: []string{"a"},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 60}},
	})
	if err != nil {
		t.Fatalf("build DAG: %v", err)
	}

	dir := t.TempDir()
	coursesPath := filepath.Join(dir, "courses.json")
	if err := os.WriteFile(coursesPath, []byte(`[
		{"id":"test-course","name":"Test Course, Advanced","grade":"10","description":"Course with comma","targets":["b"]}
	]`), 0644); err != nil {
		t.Fatalf("write courses: %v", err)
	}
	destPath := filepath.Join(dir, "destinations.json")
	if err := os.WriteFile(destPath, []byte(`[
		{"id":"test-dest","name":"Test Dest","description":"Dest desc","courses":["test-course"]}
	]`), 0644); err != nil {
		t.Fatalf("write destinations: %v", err)
	}

	planner, err := planning.Load(coursesPath, d)
	if err != nil {
		t.Fatalf("load planner: %v", err)
	}
	dests, err := planning.LoadDestinations(destPath)
	if err != nil {
		t.Fatalf("load destinations: %v", err)
	}
	planner.SetDestinations(dests)

	store, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	reg := generator.NewRegistry()
	reg.Register("a", &testGen{})
	reg.Register("b", &testGen{})

	eng := engine.New(store, d, reg, nil, planner)
	authSvc := auth.New(store)
	s := New(eng, store, authSvc)
	mux := http.NewServeMux()
	s.Register(mux)

	token, studentID := quizGuest(t, mux)
	return s, mux, store, token, studentID
}

func TestGzipMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"hello":"world"}`))
	})
	wrapped := GzipMiddleware(handler)

	// Case 1: Client does not accept gzip.
	reqPlain := httptest.NewRequest("GET", "/test", nil)
	recPlain := httptest.NewRecorder()
	wrapped.ServeHTTP(recPlain, reqPlain)
	if recPlain.Header().Get("Content-Encoding") != "" {
		t.Errorf("expected no Content-Encoding without Accept-Encoding, got %q", recPlain.Header().Get("Content-Encoding"))
	}
	if recPlain.Body.String() != `{"hello":"world"}` {
		t.Errorf("expected plain body, got %q", recPlain.Body.String())
	}

	// Case 2: Client accepts gzip.
	reqGzip := httptest.NewRequest("GET", "/test", nil)
	reqGzip.Header.Set("Accept-Encoding", "gzip, deflate")
	recGzip := httptest.NewRecorder()
	wrapped.ServeHTTP(recGzip, reqGzip)
	if recGzip.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding gzip, got %q", recGzip.Header().Get("Content-Encoding"))
	}
	if recGzip.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected default Content-Type application/json, got %q", recGzip.Header().Get("Content-Type"))
	}

	gzReader, err := gzip.NewReader(recGzip.Body)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gzReader.Close()
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("decompress: %v", err)
	}
	if string(decompressed) != `{"hello":"world"}` {
		t.Errorf("expected decompressed body %q, got %q", `{"hello":"world"}`, string(decompressed))
	}
}

func TestCurriculumRoutes(t *testing.T) {
	_, mux, _, _, _ := setupTestServerWithPlanner(t)

	// Method check on domains.
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/curriculum/domains", nil))
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// GET /api/curriculum/domains
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/curriculum/domains", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var domsRes struct {
		Domains []struct {
			ID string `json:"id"`
		} `json:"domains"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &domsRes); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(domsRes.Domains) == 0 || domsRes.Domains[0].ID != "d" {
		t.Errorf("expected domain 'd', got %+v", domsRes.Domains)
	}

	// GET /api/curriculum/domains/d
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/curriculum/domains/d", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200 for domain d, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /api/curriculum/domains/nonexistent
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/curriculum/domains/nonexistent", nil))
	if rec.Code != 404 {
		t.Errorf("expected 404 for nonexistent domain, got %d", rec.Code)
	}
}

func TestSettingsRoutes(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// Unauthenticated GET
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/settings", nil))
	if rec.Code != 401 {
		t.Errorf("expected 401 for unauth settings, got %d", rec.Code)
	}

	// Authenticated GET (empty default)
	req := httptest.NewRequest("GET", "/api/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Invalid PUT
	req = httptest.NewRequest("PUT", "/api/settings", strings.NewReader("not-json"))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("expected 400 for invalid PUT, got %d", rec.Code)
	}

	// Valid PUT
	req = httptest.NewRequest("PUT", "/api/settings", strings.NewReader(`{"theme":"dark","font_size":14}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for valid PUT, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify GET persists settings
	req = httptest.NewRequest("GET", "/api/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &settings); err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	if settings["theme"] != "dark" {
		t.Errorf("expected theme 'dark', got %v", settings["theme"])
	}

	// Method not allowed
	req = httptest.NewRequest("DELETE", "/api/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestSetDailyXPGoal(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// GET method not allowed
	req := httptest.NewRequest("GET", "/api/goals/xp", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Out of bounds: goal too low (< 10)
	req = httptest.NewRequest("POST", "/api/goals/xp", strings.NewReader(`{"goal":5}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("expected 400 for goal < 10, got %d", rec.Code)
	}

	// Out of bounds: goal too high (> 100)
	req = httptest.NewRequest("POST", "/api/goals/xp", strings.NewReader(`{"goal":150}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("expected 400 for goal > 100, got %d", rec.Code)
	}

	// Valid goal
	req = httptest.NewRequest("POST", "/api/goals/xp", strings.NewReader(`{"goal":50}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for valid goal, got %d: %s", rec.Code, rec.Body.String())
	}
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	if res["goal"] != float64(50) {
		t.Errorf("expected goal 50, got %v", res["goal"])
	}
}

func TestAvatarByUsername(t *testing.T) {
	_, mux, store, _, sid := setupTestServerWithPlanner(t)

	// Method check
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/api/avatar/u/alice", nil))
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Empty username
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/avatar/u/", nil))
	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Username with slash
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/avatar/u/alice/sub", nil))
	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	// Learner not found
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/avatar/u/nonexistent_learner", nil))
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	// Create learner with username but no custom avatar
	if err := store.SetUsername(sid, "avatarlearner"); err != nil {
		t.Fatalf("set username: %v", err)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/avatar/u/avatarlearner", nil))
	if rec.Code != 404 {
		t.Errorf("expected 404 for missing avatar image, got %d", rec.Code)
	}

	// Set avatar image
	imgData := []byte("fake-image-bytes")
	if err := store.SetAvatarImage(sid, "image/png", imgData); err != nil {
		t.Fatalf("set avatar: %v", err)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/avatar/u/avatarlearner", nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200 for avatar image, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected Content-Type image/png, got %s", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Errorf("expected Cache-Control header, got %s", rec.Header().Get("Cache-Control"))
	}
	if !bytes.Equal(rec.Body.Bytes(), imgData) {
		t.Errorf("avatar bytes mismatch")
	}
}

func TestCourseAndTranscriptRoutes(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// Courses method check
	req := httptest.NewRequest("POST", "/api/courses", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// GET /api/courses
	req = httptest.NewRequest("GET", "/api/courses", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for courses, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /api/courses/test-course
	req = httptest.NewRequest("GET", "/api/courses/test-course", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for course detail, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /api/courses/nonexistent
	req = httptest.NewRequest("GET", "/api/courses/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Errorf("expected 404 for nonexistent course, got %d", rec.Code)
	}

	// Transcript method check
	req = httptest.NewRequest("POST", "/api/transcript", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// GET /api/transcript (JSON)
	req = httptest.NewRequest("GET", "/api/transcript", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for transcript, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /api/transcript?format=csv
	req = httptest.NewRequest("GET", "/api/transcript?format=csv", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for transcript CSV, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "text/csv; charset=utf-8" {
		t.Errorf("expected text/csv Content-Type, got %s", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "mathua-transcript.csv") {
		t.Errorf("expected attachment Content-Disposition, got %s", rec.Header().Get("Content-Disposition"))
	}
	csvBody := rec.Body.String()
	if !strings.Contains(csvBody, "course,grade,total,mastered,pct,days_remaining,status") {
		t.Errorf("expected CSV header row, got %q", csvBody)
	}
	// Course name has a comma so csvCell should quote it
	if !strings.Contains(csvBody, `"Test Course, Advanced"`) {
		t.Errorf("expected quoted course name in CSV, got %q", csvBody)
	}
}

func TestDestinationsAndPlanRoutes(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// Destinations method check
	req := httptest.NewRequest("POST", "/api/destinations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// GET /api/destinations
	req = httptest.NewRequest("GET", "/api/destinations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for destinations, got %d: %s", rec.Code, rec.Body.String())
	}

	// Malformed destination route
	req = httptest.NewRequest("GET", "/api/destinations/bad_route", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	// GET /api/destinations/test-dest/estimate
	req = httptest.NewRequest("GET", "/api/destinations/test-dest/estimate?daily_goal=20&deadline_days=30&rest_days=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for estimate, got %d: %s", rec.Code, rec.Body.String())
	}

	// POST /api/plans without destination
	req = httptest.NewRequest("POST", "/api/plans", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("expected 400 without destination, got %d", rec.Code)
	}

	// POST /api/plans with valid destination
	req = httptest.NewRequest("POST", "/api/plans", strings.NewReader(`{"destination":"test-dest","daily_goal":20,"deadline_days":30,"rest_days":1}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for plan save, got %d: %s", rec.Code, rec.Body.String())
	}

	// GET /api/plans/current
	req = httptest.NewRequest("GET", "/api/plans/current", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for current plan, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestShareReportRoutes(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// Share method check
	req := httptest.NewRequest("GET", "/api/share", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Invalid body
	req = httptest.NewRequest("POST", "/api/share", strings.NewReader("bad"))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("expected 400 for bad JSON, got %d", rec.Code)
	}

	// Enable share
	req = httptest.NewRequest("POST", "/api/share", strings.NewReader(`{"enabled":true}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 to enable share, got %d: %s", rec.Code, rec.Body.String())
	}
	var shareRes struct {
		Enabled bool   `json:"enabled"`
		Token   string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &shareRes); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !shareRes.Enabled || shareRes.Token == "" {
		t.Fatalf("expected enabled with token, got %+v", shareRes)
	}

	// View share report by token (public endpoint)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/share/"+shareRes.Token, nil))
	if rec.Code != 200 {
		t.Fatalf("expected 200 for share report, got %d: %s", rec.Code, rec.Body.String())
	}

	// View share report by invalid token
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/share/nonexistent_token", nil))
	if rec.Code != 404 {
		t.Errorf("expected 404 for invalid share token, got %d", rec.Code)
	}

	// Disable share
	req = httptest.NewRequest("POST", "/api/share", strings.NewReader(`{"enabled":false}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 to disable share, got %d: %s", rec.Code, rec.Body.String())
	}

	// Share token is now disabled
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/share/"+shareRes.Token, nil))
	if rec.Code != 404 {
		t.Errorf("expected 404 for disabled share token, got %d", rec.Code)
	}
}

func TestDueReviewsRoute(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// Unauthenticated returns 401
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/reviews/due", nil))
	if rec.Code != 401 {
		t.Fatalf("expected 401 for unauthenticated due reviews, got %d", rec.Code)
	}

	// Wrong method authenticated returns 405
	req := httptest.NewRequest("POST", "/api/reviews/due", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	// Authenticated returns count
	req = httptest.NewRequest("GET", "/api/reviews/due", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200 for due reviews, got %d: %s", rec.Code, rec.Body.String())
	}
	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	if _, ok := res["count"]; !ok {
		t.Errorf("expected 'count' field in response, got %v", res)
	}
}

func TestAttemptsRoute(t *testing.T) {
	_, mux, store, token, sid := setupTestServerWithPlanner(t)

	// Unauthenticated
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/attempts", nil))
	if rec.Code != 401 {
		t.Errorf("expected 401 for unauth attempts, got %d", rec.Code)
	}

	// Create session for attempts
	sess, err := store.CreateSession(sid)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	now := time.Now()
	// Insert attempts
	_, err = store.RecordAttempt(storage.AttemptEntry{
		SessionID: sess.ID,
		StudentID: sid,
		ConceptID: "a",
		Answer:    "4",
		Expected:  "4",
		Correct:   true,
		Source:    "practice",
		Timestamp: now.Add(-10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("record attempt 1: %v", err)
	}
	_, err = store.RecordAttempt(storage.AttemptEntry{
		SessionID: sess.ID,
		StudentID: sid,
		ConceptID: "b",
		Answer:    "5",
		Expected:  "4",
		Correct:   false,
		Source:    "quiz",
		Timestamp: now.Add(-5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("record attempt 2: %v", err)
	}

	// Query all attempts
	req := httptest.NewRequest("GET", "/api/attempts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var allRes struct {
		Attempts []struct {
			ConceptID   string `json:"concept_id"`
			ConceptName string `json:"concept_name"`
			Correct     bool   `json:"correct"`
			Source      string `json:"source"`
		} `json:"attempts"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &allRes); err != nil {
		t.Fatalf("decode attempts: %v", err)
	}
	if allRes.Total != 2 {
		t.Errorf("expected total 2, got %d", allRes.Total)
	}

	// Filter by source=quiz
	req = httptest.NewRequest("GET", "/api/attempts?source=quiz", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var quizRes struct {
		Total int `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &quizRes)
	if quizRes.Total != 1 {
		t.Errorf("expected 1 quiz attempt, got %d", quizRes.Total)
	}

	// Filter by incorrect_only=1
	req = httptest.NewRequest("GET", "/api/attempts?incorrect_only=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var incRes struct {
		Total int `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &incRes)
	if incRes.Total != 1 {
		t.Errorf("expected 1 incorrect attempt, got %d", incRes.Total)
	}

	// Pagination: limit=1, offset=1
	req = httptest.NewRequest("GET", "/api/attempts?limit=1&offset=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var pageRes struct {
		Attempts []struct {
			ConceptID string `json:"concept_id"`
		} `json:"attempts"`
		Total int `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &pageRes)
	if pageRes.Total != 2 || len(pageRes.Attempts) != 1 {
		t.Errorf("expected 1 attempt returned with total 2, got len=%d total=%d", len(pageRes.Attempts), pageRes.Total)
	}
}
