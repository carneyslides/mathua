package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuma-beep/mathua/internal/auth"
	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/engine"
	"github.com/chuma-beep/mathua/internal/generator"
	"github.com/chuma-beep/mathua/internal/lessons"
	"github.com/chuma-beep/mathua/internal/storage"
)

func TestValidRequestID(t *testing.T) {
	validIDs := []string{
		"req-12345678",
		"trace_abc_12345",
		"12345678",
		strings.Repeat("a", 64),
	}
	for _, id := range validIDs {
		if !validRequestID(id) {
			t.Errorf("expected %q to be valid request ID", id)
		}
	}

	invalidIDs := []string{
		"",
		"short",
		"1234567",
		strings.Repeat("a", 65),
		"id with spaces",
		"id;semicolon",
		"id\nnewline",
		"id/slash",
	}
	for _, id := range invalidIDs {
		if validRequestID(id) {
			t.Errorf("expected %q to be rejected as request ID", id)
		}
	}
}

func TestOAuthProviderFromCallbackPath(t *testing.T) {
	cases := []struct {
		path     string
		wantName string
		wantOK   bool
	}{
		{"/api/auth/github/callback", "github", true},
		{"/api/auth/facebook/callback", "facebook", true},
		{"/api/auth/microsoft/callback", "microsoft", true},
		{"/api/auth/apple/callback", "apple", true},
		{"/api/auth/unknown/callback", "", false},
		{"/api/auth/github/login", "", false},
		{"/api/other/path", "", false},
		{"", "", false},
	}

	for _, c := range cases {
		name, ok := oauthProviderFromCallbackPath(c.path)
		if ok != c.wantOK || name != c.wantName {
			t.Errorf("oauthProviderFromCallbackPath(%q) = (%q, %v), want (%q, %v)",
				c.path, name, ok, c.wantName, c.wantOK)
		}
	}
}

func TestLessonBodyRoute_LifecycleAndErrors(t *testing.T) {
	d, err := concepts.Build([]concepts.Concept{
		{ID: "a", Label: "Concept A", Domain: "d", Subdomain: "s", GradingType: "numeric", Prerequisites: []string{},
			MasteryThreshold: concepts.MasteryThreshold{Streak: 3, AvgTimeSeconds: 60}},
	})
	if err != nil {
		t.Fatalf("build DAG: %v", err)
	}

	// 1. Server with nil lesson loader
	store, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	reg := generator.NewRegistry()
	engNoLessons := engine.New(store, d, reg, nil, nil)
	sNoLessons := New(engNoLessons, store, auth.New(store))
	muxNoLessons := http.NewServeMux()
	sNoLessons.Register(muxNoLessons)

	// Method not allowed
	reqPost := httptest.NewRequest("POST", "/api/lessons/body?title=SomeTitle", nil)
	recPost := httptest.NewRecorder()
	muxNoLessons.ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST, got %d", recPost.Code)
	}

	// Unavailable lessons loader (503)
	reqNoLoader := httptest.NewRequest("GET", "/api/lessons/body?title=SomeTitle", nil)
	recNoLoader := httptest.NewRecorder()
	muxNoLessons.ServeHTTP(recNoLoader, reqNoLoader)
	if recNoLoader.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when loader is nil, got %d", recNoLoader.Code)
	}

	// 2. Server with populated lesson loader
	lessonsDir := t.TempDir()
	mappingPath := filepath.Join(lessonsDir, "lessons.json")
	if err := os.WriteFile(mappingPath, []byte(`[{"concept_id":"a","source":"a.md"}]`), 0644); err != nil {
		t.Fatalf("write lessons.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(lessonsDir, "a.md"), []byte("# Power Rule\n\nMultiply by the exponent."), 0644); err != nil {
		t.Fatalf("write a.md: %v", err)
	}

	lessonLoader, err := lessons.Load(lessonsDir)
	if err != nil {
		t.Fatalf("load lessons: %v", err)
	}

	engWithLessons := engine.New(store, d, reg, lessonLoader, nil)
	sWithLessons := New(engWithLessons, store, auth.New(store))
	muxWithLessons := http.NewServeMux()
	sWithLessons.Register(muxWithLessons)

	// Missing title query parameter (400)
	reqEmptyTitle := httptest.NewRequest("GET", "/api/lessons/body?title=%20%20%20", nil)
	recEmptyTitle := httptest.NewRecorder()
	muxWithLessons.ServeHTTP(recEmptyTitle, reqEmptyTitle)
	if recEmptyTitle.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty title, got %d", recEmptyTitle.Code)
	}

	// Lesson not found (404)
	reqNotFound := httptest.NewRequest("GET", "/api/lessons/body?title=NonexistentTitle", nil)
	recNotFound := httptest.NewRecorder()
	muxWithLessons.ServeHTTP(recNotFound, reqNotFound)
	if recNotFound.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown title, got %d", recNotFound.Code)
	}

	// Success case (200)
	reqSuccess := httptest.NewRequest("GET", "/api/lessons/body?title=Power+Rule", nil)
	recSuccess := httptest.NewRecorder()
	muxWithLessons.ServeHTTP(recSuccess, reqSuccess)
	if recSuccess.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid lesson, got %d: %s", recSuccess.Code, recSuccess.Body.String())
	}

	var res struct {
		Title  string          `json:"title"`
		Body   string          `json:"body"`
		Source string          `json:"source"`
		Assets []lessons.Asset `json:"assets"`
	}
	if err := json.Unmarshal(recSuccess.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal lesson body: %v", err)
	}
	if res.Title != "Power Rule" || !strings.Contains(res.Body, "Multiply by the exponent") {
		t.Errorf("unexpected lesson content: %+v", res)
	}
	if res.Assets == nil {
		t.Error("expected non-nil assets slice in response")
	}
}

func TestWeaknessesRoute(t *testing.T) {
	_, mux, store, token, sid := setupTestServerWithPlanner(t)

	// Method not allowed (authenticated request with wrong method)
	reqPost := httptest.NewRequest("POST", "/api/weaknesses", nil)
	reqPost.Header.Set("Authorization", "Bearer "+token)
	recPost := httptest.NewRecorder()
	mux.ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST, got %d", recPost.Code)
	}

	// Unauthenticated request
	reqUnauth := httptest.NewRequest("GET", "/api/weaknesses", nil)
	recUnauth := httptest.NewRecorder()
	mux.ServeHTTP(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated request, got %d", recUnauth.Code)
	}

	// Authenticated request with no attempts yet
	reqAuth := httptest.NewRequest("GET", "/api/weaknesses", nil)
	reqAuth.Header.Set("Authorization", "Bearer "+token)
	recAuth := httptest.NewRecorder()
	mux.ServeHTTP(recAuth, reqAuth)
	if recAuth.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recAuth.Code, recAuth.Body.String())
	}

	var emptyRes struct {
		ByDomain map[string]interface{} `json:"by_domain"`
	}
	if err := json.Unmarshal(recAuth.Body.Bytes(), &emptyRes); err != nil {
		t.Fatalf("decode weaknesses: %v", err)
	}
	if len(emptyRes.ByDomain) != 0 {
		t.Errorf("expected empty by_domain for student with no progress, got %+v", emptyRes.ByDomain)
	}

	// Record a failing progress row
	if err := store.UpsertProgress(&storage.ConceptProgress{
		StudentID: sid,
		ConceptID: "a",
		Status:    "LEARNING",
		Streak:    0,
	}); err != nil {
		t.Fatalf("save progress: %v", err)
	}

	reqProgress := httptest.NewRequest("GET", "/api/weaknesses", nil)
	reqProgress.Header.Set("Authorization", "Bearer "+token)
	recProgress := httptest.NewRecorder()
	mux.ServeHTTP(recProgress, reqProgress)
	if recProgress.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recProgress.Code, recProgress.Body.String())
	}
}

func TestSessionRoute_AuthenticatedAndGuest(t *testing.T) {
	s, _, _, _, sid := setupTestServerWithPlanner(t)

	// Method not allowed
	reqGet := httptest.NewRequest("GET", "/api/session", nil)
	recGet := httptest.NewRecorder()
	s.handleSession(recGet, reqGet)
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET, got %d", recGet.Code)
	}

	// Authenticated student context
	reqAuth := httptest.NewRequest("POST", "/api/session", bytes.NewReader([]byte(`{}`)))
	ctxAuth := context.WithValue(reqAuth.Context(), authStudentKey{}, sid)
	reqAuth = reqAuth.WithContext(ctxAuth)
	recAuth := httptest.NewRecorder()
	s.handleSession(recAuth, reqAuth)
	if recAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 for auth session creation, got %d: %s", recAuth.Code, recAuth.Body.String())
	}

	var authRes startSessionRes
	if err := json.Unmarshal(recAuth.Body.Bytes(), &authRes); err != nil {
		t.Fatalf("unmarshal auth startSessionRes: %v", err)
	}
	if authRes.SessionID == "" || authRes.StudentID == "" {
		t.Errorf("unexpected empty session or student ID: %+v", authRes)
	}

	// Guest with invalid JSON
	reqInvalid := httptest.NewRequest("POST", "/api/session", bytes.NewReader([]byte(`not json`)))
	recInvalid := httptest.NewRecorder()
	s.handleSession(recInvalid, reqInvalid)
	if recInvalid.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", recInvalid.Code)
	}

	// Guest with missing name and student_id
	reqEmpty := httptest.NewRequest("POST", "/api/session", bytes.NewReader([]byte(`{}`)))
	recEmpty := httptest.NewRecorder()
	s.handleSession(recEmpty, reqEmpty)
	if recEmpty.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty guest request, got %d", recEmpty.Code)
	}

	// Guest with name provided
	reqName := httptest.NewRequest("POST", "/api/session", bytes.NewReader([]byte(`{"name":"Guest Learner"}`)))
	recName := httptest.NewRecorder()
	s.handleSession(recName, reqName)
	if recName.Code != http.StatusOK {
		t.Fatalf("expected 200 for guest with name, got %d: %s", recName.Code, recName.Body.String())
	}

	var nameRes startSessionRes
	if err := json.Unmarshal(recName.Body.Bytes(), &nameRes); err != nil {
		t.Fatalf("unmarshal guest startSessionRes: %v", err)
	}
	if nameRes.SessionID == "" || nameRes.StudentID == "" {
		t.Errorf("unexpected empty session or student ID for guest: %+v", nameRes)
	}

	// Guest with existing student_id
	reqID := httptest.NewRequest("POST", "/api/session", bytes.NewReader([]byte(fmt.Sprintf(`{"student_id":%q}`, sid))))
	recID := httptest.NewRecorder()
	s.handleSession(recID, reqID)
	if recID.Code != http.StatusOK {
		t.Fatalf("expected 200 for guest with student_id, got %d: %s", recID.Code, recID.Body.String())
	}
}

func TestActivityRoute_ValidationAndData(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// Method not allowed (authenticated request with wrong method)
	reqPost := httptest.NewRequest("POST", "/api/activity", nil)
	reqPost.Header.Set("Authorization", "Bearer "+token)
	recPost := httptest.NewRecorder()
	mux.ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST, got %d", recPost.Code)
	}

	// Unauthenticated request
	reqUnauth := httptest.NewRequest("GET", "/api/activity", nil)
	recUnauth := httptest.NewRecorder()
	mux.ServeHTTP(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated request, got %d", recUnauth.Code)
	}

	// Invalid days parameters
	for _, badDays := range []string{"abc", "0", "-5", "5000"} {
		reqBad := httptest.NewRequest("GET", "/api/activity?days="+badDays, nil)
		reqBad.Header.Set("Authorization", "Bearer "+token)
		recBad := httptest.NewRecorder()
		mux.ServeHTTP(recBad, reqBad)
		if recBad.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for days=%s, got %d", badDays, recBad.Code)
		}
	}

	// Valid activity request
	reqValid := httptest.NewRequest("GET", "/api/activity?days=30", nil)
	reqValid.Header.Set("Authorization", "Bearer "+token)
	recValid := httptest.NewRecorder()
	mux.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Fatalf("expected 200 for activity, got %d: %s", recValid.Code, recValid.Body.String())
	}

	var activities []storage.DailyActivity
	if err := json.Unmarshal(recValid.Body.Bytes(), &activities); err != nil {
		t.Fatalf("decode activities: %v", err)
	}
}

func TestGoalRoute_Lifecycle(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// Method not allowed (authenticated request with wrong method)
	reqGet := httptest.NewRequest("GET", "/api/goal", nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	recGet := httptest.NewRecorder()
	mux.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET, got %d", recGet.Code)
	}

	// Unauthenticated request
	reqUnauth := httptest.NewRequest("POST", "/api/goal", bytes.NewReader([]byte(`{"concept_ids":["a"]}`)))
	recUnauth := httptest.NewRecorder()
	mux.ServeHTTP(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated POST /api/goal, got %d", recUnauth.Code)
	}

	// Invalid JSON payload
	reqBadJSON := httptest.NewRequest("POST", "/api/goal", bytes.NewReader([]byte(`invalid json`)))
	reqBadJSON.Header.Set("Authorization", "Bearer "+token)
	recBadJSON := httptest.NewRecorder()
	mux.ServeHTTP(recBadJSON, reqBadJSON)
	if recBadJSON.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bad JSON, got %d", recBadJSON.Code)
	}

	// Empty concepts and empty domain
	reqEmpty := httptest.NewRequest("POST", "/api/goal", bytes.NewReader([]byte(`{"concept_ids":[]}`)))
	reqEmpty.Header.Set("Authorization", "Bearer "+token)
	recEmpty := httptest.NewRecorder()
	mux.ServeHTTP(recEmpty, reqEmpty)
	if recEmpty.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty concepts, got %d", recEmpty.Code)
	}

	// Target by concept ID
	reqConcept := httptest.NewRequest("POST", "/api/goal", bytes.NewReader([]byte(`{"concept_ids":["b"]}`)))
	reqConcept.Header.Set("Authorization", "Bearer "+token)
	recConcept := httptest.NewRecorder()
	mux.ServeHTTP(recConcept, reqConcept)
	if recConcept.Code != http.StatusOK {
		t.Fatalf("expected 200 for concept goal, got %d: %s", recConcept.Code, recConcept.Body.String())
	}

	var resConcept struct {
		Concepts []struct {
			ID string `json:"id"`
		} `json:"concepts"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal(recConcept.Body.Bytes(), &resConcept); err != nil {
		t.Fatalf("unmarshal goal response: %v", err)
	}
	if resConcept.Count != 2 || len(resConcept.Concepts) != 2 {
		t.Errorf("unexpected goal path: %+v", resConcept)
	}

	// Target by domain
	reqDomain := httptest.NewRequest("POST", "/api/goal", bytes.NewReader([]byte(`{"domain":"d"}`)))
	reqDomain.Header.Set("Authorization", "Bearer "+token)
	recDomain := httptest.NewRecorder()
	mux.ServeHTTP(recDomain, reqDomain)
	if recDomain.Code != http.StatusOK {
		t.Fatalf("expected 200 for domain goal, got %d: %s", recDomain.Code, recDomain.Body.String())
	}
}

func TestEfficacyRoutes(t *testing.T) {
	_, mux, _, token, _ := setupTestServerWithPlanner(t)

	// GET /api/efficacy method not allowed
	reqPost := httptest.NewRequest("POST", "/api/efficacy", nil)
	reqPost.Header.Set("Authorization", "Bearer "+token)
	recPost := httptest.NewRecorder()
	mux.ServeHTTP(recPost, reqPost)
	if recPost.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST /api/efficacy, got %d", recPost.Code)
	}

	// GET /api/efficacy unauthenticated
	reqUnauth := httptest.NewRequest("GET", "/api/efficacy", nil)
	recUnauth := httptest.NewRecorder()
	mux.ServeHTTP(recUnauth, reqUnauth)
	if recUnauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated /api/efficacy, got %d", recUnauth.Code)
	}

	// GET /api/efficacy authenticated
	reqAuth := httptest.NewRequest("GET", "/api/efficacy", nil)
	reqAuth.Header.Set("Authorization", "Bearer "+token)
	recAuth := httptest.NewRecorder()
	mux.ServeHTTP(recAuth, reqAuth)
	if recAuth.Code != http.StatusOK {
		t.Fatalf("expected 200 for auth /api/efficacy, got %d: %s", recAuth.Code, recAuth.Body.String())
	}

	// GET /api/efficacy/all method not allowed
	reqPostAll := httptest.NewRequest("POST", "/api/efficacy/all", nil)
	recPostAll := httptest.NewRecorder()
	mux.ServeHTTP(recPostAll, reqPostAll)
	if recPostAll.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST /api/efficacy/all, got %d", recPostAll.Code)
	}

	// GET /api/efficacy/all public success
	reqAll := httptest.NewRequest("GET", "/api/efficacy/all", nil)
	recAll := httptest.NewRecorder()
	mux.ServeHTTP(recAll, reqAll)
	if recAll.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/efficacy/all, got %d: %s", recAll.Code, recAll.Body.String())
	}

	// GET /api/efficacy/trend method not allowed
	reqPostTrend := httptest.NewRequest("POST", "/api/efficacy/trend", nil)
	recPostTrend := httptest.NewRecorder()
	mux.ServeHTTP(recPostTrend, reqPostTrend)
	if recPostTrend.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for POST /api/efficacy/trend, got %d", recPostTrend.Code)
	}

	// GET /api/efficacy/trend public success
	reqTrend := httptest.NewRequest("GET", "/api/efficacy/trend", nil)
	recTrend := httptest.NewRecorder()
	mux.ServeHTTP(recTrend, reqTrend)
	if recTrend.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/efficacy/trend, got %d: %s", recTrend.Code, recTrend.Body.String())
	}
}
