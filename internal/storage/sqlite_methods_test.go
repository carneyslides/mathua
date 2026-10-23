package storage

import (
	"bytes"
	"testing"
	"time"
)

func TestSQLiteStore_CreateOAuthUser(t *testing.T) {
	store := newTestStore(t)

	// User with display name and verified email
	u1, err := store.CreateOAuthUser("github", "gh-100", "Alice Smith", "alice@example.com", true, "https://example.com/alice.png")
	if err != nil {
		t.Fatalf("create oauth user 1: %v", err)
	}
	if u1.ID == "" {
		t.Error("expected non-empty student ID")
	}
	if u1.Name != "Alice Smith" {
		t.Errorf("got name %q, want %q", u1.Name, "Alice Smith")
	}
	if !u1.EmailVerified {
		t.Error("expected email to be verified")
	}
	if u1.AvatarURL != "https://example.com/alice.png" {
		t.Errorf("got avatar url %q", u1.AvatarURL)
	}

	// User without display name falls back to email
	u2, err := store.CreateOAuthUser("google", "goog-200", "", "bob@example.com", false, "")
	if err != nil {
		t.Fatalf("create oauth user 2: %v", err)
	}
	if u2.Name != "bob@example.com" {
		t.Errorf("got name %q, want email fallback %q", u2.Name, "bob@example.com")
	}
	if u2.EmailVerified {
		t.Error("expected email not to be verified")
	}
}

func TestSQLiteStore_GetStudentByShareToken(t *testing.T) {
	store := newTestStore(t)

	// Empty token returns nil without error
	got, err := store.GetStudentByShareToken("")
	if err != nil {
		t.Fatalf("empty token: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil student for empty token, got %+v", got)
	}

	// Unknown token returns nil without error
	got, err = store.GetStudentByShareToken("missing-token")
	if err != nil {
		t.Fatalf("missing token: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil student for missing token, got %+v", got)
	}

	// Established share token lookup
	stu, err := store.CreateStudent("token-user")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	token := "valid-share-token-xyz"
	if _, err := store.db.Exec("UPDATE students SET share_token = ? WHERE id = ?", token, stu.ID); err != nil {
		t.Fatalf("update share token: %v", err)
	}

	found, err := store.GetStudentByShareToken(token)
	if err != nil {
		t.Fatalf("get by share token: %v", err)
	}
	if found == nil || found.ID != stu.ID {
		t.Errorf("got student %+v, want ID %q", found, stu.ID)
	}
}

func TestSQLiteStore_SetCourseID(t *testing.T) {
	store := newTestStore(t)
	stu, err := store.CreateStudent("course-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	if err := store.SetCourseID(stu.ID, "calc-101"); err != nil {
		t.Fatalf("set course ID: %v", err)
	}

	loaded, err := store.GetStudent(stu.ID)
	if err != nil {
		t.Fatalf("get student: %v", err)
	}
	if loaded.CourseID != "calc-101" {
		t.Errorf("got course ID %q, want %q", loaded.CourseID, "calc-101")
	}
}

func TestSQLiteStore_SetDailyXPGoal(t *testing.T) {
	store := newTestStore(t)
	stu, err := store.CreateStudent("xp-goal-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	if err := store.SetDailyXPGoal(stu.ID, 75); err != nil {
		t.Fatalf("set daily xp goal: %v", err)
	}

	loaded, err := store.GetStudent(stu.ID)
	if err != nil {
		t.Fatalf("get student: %v", err)
	}
	if loaded.DailyXPGoal != 75 {
		t.Errorf("got daily xp goal %d, want 75", loaded.DailyXPGoal)
	}
}

func TestSQLiteStore_ListStudentsMissingUsernames(t *testing.T) {
	store := newTestStore(t)
	stu1, err := store.CreateStudent("no-username-1")
	if err != nil {
		t.Fatalf("create student 1: %v", err)
	}
	stu2, err := store.CreateStudent("has-username")
	if err != nil {
		t.Fatalf("create student 2: %v", err)
	}
	if err := store.SetUsername(stu2.ID, "user2"); err != nil {
		t.Fatalf("set username: %v", err)
	}

	missing, err := store.ListStudentsMissingUsernames()
	if err != nil {
		t.Fatalf("list missing usernames: %v", err)
	}

	found := false
	for _, id := range missing {
		if id == stu1.ID {
			found = true
		}
		if id == stu2.ID {
			t.Errorf("student with username %q was unexpectedly included in missing list", stu2.ID)
		}
	}
	if !found {
		t.Errorf("student %q was expected in missing usernames list", stu1.ID)
	}
}

func TestSQLiteStore_SettingsLifecycle(t *testing.T) {
	store := newTestStore(t)

	// Non-existent student returns default empty json
	s, err := store.GetSettings("missing-student-id")
	if err != nil {
		t.Fatalf("get settings missing: %v", err)
	}
	if s != "{}" {
		t.Errorf("got settings %q, want %q", s, "{}")
	}

	stu, err := store.CreateStudent("settings-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	initSettings, err := store.GetSettings(stu.ID)
	if err != nil {
		t.Fatalf("get settings initial: %v", err)
	}
	if initSettings != "{}" {
		t.Errorf("got %q, want %q", initSettings, "{}")
	}

	custom := `{"theme":"nord","volume":0.8}`
	if err := store.UpdateSettings(stu.ID, custom); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	updated, err := store.GetSettings(stu.ID)
	if err != nil {
		t.Fatalf("get updated settings: %v", err)
	}
	if updated != custom {
		t.Errorf("got settings %q, want %q", updated, custom)
	}
}

func TestSQLiteStore_AvatarImageLifecycle(t *testing.T) {
	store := newTestStore(t)

	// Missing student avatar lookup
	_, _, found, err := store.GetAvatarImage("missing-student")
	if err != nil {
		t.Fatalf("get avatar missing: %v", err)
	}
	if found {
		t.Error("expected found to be false for missing avatar")
	}

	stu, err := store.CreateStudent("avatar-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	payload := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := store.SetAvatarImage(stu.ID, "image/png", payload); err != nil {
		t.Fatalf("set avatar image: %v", err)
	}

	ct, data, found, err := store.GetAvatarImage(stu.ID)
	if err != nil {
		t.Fatalf("get avatar image: %v", err)
	}
	if !found {
		t.Fatal("expected avatar image to be found")
	}
	if ct != "image/png" {
		t.Errorf("got content type %q, want image/png", ct)
	}
	if !bytes.Equal(data, payload) {
		t.Errorf("got avatar data %v, want %v", data, payload)
	}

	// Update avatar image
	newPayload := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	if err := store.SetAvatarImage(stu.ID, "image/jpeg", newPayload); err != nil {
		t.Fatalf("update avatar image: %v", err)
	}
	ct, data, found, err = store.GetAvatarImage(stu.ID)
	if err != nil || !found || ct != "image/jpeg" || !bytes.Equal(data, newPayload) {
		t.Fatalf("unexpected updated avatar: ct=%q found=%v data=%v", ct, found, data)
	}

	// Clear avatar image
	if err := store.ClearAvatarImage(stu.ID); err != nil {
		t.Fatalf("clear avatar image: %v", err)
	}
	_, _, found, err = store.GetAvatarImage(stu.ID)
	if err != nil {
		t.Fatalf("get avatar after clear: %v", err)
	}
	if found {
		t.Error("expected avatar image to be cleared")
	}
}

func TestSQLiteStore_GetSession(t *testing.T) {
	store := newTestStore(t)

	// Missing session lookup returns nil
	got, err := store.GetSession("non-existent-session-id")
	if err != nil {
		t.Fatalf("get missing session: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil session for non-existent id, got %+v", got)
	}

	stu, err := store.CreateStudent("session-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	created, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	loaded, err := store.GetSession(created.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected session, got nil")
	}
	if loaded.ID != created.ID {
		t.Errorf("got ID %q, want %q", loaded.ID, created.ID)
	}
	if loaded.StudentID != stu.ID {
		t.Errorf("got student ID %q, want %q", loaded.StudentID, stu.ID)
	}
	if loaded.StartedAt.IsZero() {
		t.Error("expected non-zero started_at timestamp")
	}
}

func TestSQLiteStore_DeleteActiveSession(t *testing.T) {
	store := newTestStore(t)
	stu, err := store.CreateStudent("active-sess-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	active := &ActiveSession{
		SessionID:      sess.ID,
		StudentID:      stu.ID,
		ConceptID:      "ode.test",
		ConceptName:    "ODE Test",
		ExpectedAnswer: "42",
		AttemptID:      "att-1",
		Question:       "What is 6 times 7?",
	}
	if err := store.UpsertActiveSession(active); err != nil {
		t.Fatalf("upsert active session: %v", err)
	}

	got, err := store.GetActiveSession(sess.ID)
	if err != nil || got == nil {
		t.Fatalf("expected active session: err=%v got=%+v", err, got)
	}

	if err := store.DeleteActiveSession(sess.ID); err != nil {
		t.Fatalf("delete active session: %v", err)
	}

	afterDelete, err := store.GetActiveSession(sess.ID)
	if err != nil {
		t.Fatalf("get active session after delete: %v", err)
	}
	if afterDelete != nil {
		t.Errorf("expected active session to be deleted, got %+v", afterDelete)
	}
}

func TestSQLiteStore_AttemptsQueryAndCorrectness(t *testing.T) {
	store := newTestStore(t)
	stu, err := store.CreateStudent("attempts-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	now := time.Now().UTC()
	diff := 0.5
	att1ID, err := store.RecordAttempt(AttemptEntry{
		SessionID:      sess.ID,
		StudentID:      stu.ID,
		ConceptID:      "calc.limits",
		Answer:         "4",
		Expected:       "4",
		Correct:        true,
		ElapsedSeconds: 5,
		Timestamp:      now.Add(-2 * time.Minute),
		Question:       "Evaluate limit",
		Source:         "generator",
		Explanation:    "Direct substitution",
		Difficulty:     &diff,
	})
	if err != nil {
		t.Fatalf("record attempt 1: %v", err)
	}

	att2ID, err := store.RecordAttempt(AttemptEntry{
		SessionID:      sess.ID,
		StudentID:      stu.ID,
		ConceptID:      "calc.limits",
		Answer:         "5",
		Expected:       "4",
		Correct:        false,
		ElapsedSeconds: 8,
		Timestamp:      now.Add(-1 * time.Minute),
		Question:       "Evaluate limit 2",
		Source:         "generator",
		Explanation:    "L'Hopital rule",
		Difficulty:     &diff,
	})
	if err != nil {
		t.Fatalf("record attempt 2: %v", err)
	}

	// Recent attempts query with limit
	recent, err := store.GetRecentAttemptsForConcept(stu.ID, "calc.limits", 1)
	if err != nil {
		t.Fatalf("get recent attempts: %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("got %d attempts, want 1", len(recent))
	}
	if recent[0].ID != att2ID {
		t.Errorf("got attempt ID %d, want most recent %d", recent[0].ID, att2ID)
	}

	// Update attempt correctness
	if err := store.UpdateAttemptCorrect(att2ID, true); err != nil {
		t.Fatalf("update attempt correct: %v", err)
	}
	recent2, err := store.GetRecentAttemptsForConcept(stu.ID, "calc.limits", 2)
	if err != nil {
		t.Fatalf("get recent attempts after update: %v", err)
	}
	if len(recent2) != 2 || !recent2[1].Correct {
		t.Errorf("expected updated attempt to be correct, got %+v", recent2)
	}

	// All attempts query
	all, err := store.GetAllAttempts()
	if err != nil {
		t.Fatalf("get all attempts: %v", err)
	}
	if len(all) < 2 {
		t.Errorf("expected at least 2 attempts in GetAllAttempts, got %d", len(all))
	}
	_ = att1ID
}

func TestSQLiteStore_QuestionsImportAndPurge(t *testing.T) {
	store := newTestStore(t)
	conceptID := "ode.verify.purge"

	count, err := store.GetQuestionCount(conceptID)
	if err != nil {
		t.Fatalf("get question count initial: %v", err)
	}
	if count != 0 {
		t.Errorf("got initial question count %d, want 0", count)
	}

	questions := []Question{
		{
			ConceptID:   conceptID,
			Question:    "Sample Q1",
			Answer:      "A1",
			Explanation: "Exp1",
			Source:      "test-batch",
			Difficulty:  0.4,
		},
		{
			ConceptID:   conceptID,
			Question:    "Sample Q2",
			Answer:      "A2",
			Explanation: "Exp2",
			Source:      "test-batch",
			Difficulty:  0.4,
		},
	}

	if err := store.ImportQuestions(questions); err != nil {
		t.Fatalf("import questions: %v", err)
	}

	count, err = store.GetQuestionCount(conceptID)
	if err != nil {
		t.Fatalf("get question count after import: %v", err)
	}
	if count != 2 {
		t.Errorf("got question count %d, want 2", count)
	}

	fetched, err := store.GetQuestions(conceptID, 1)
	if err != nil {
		t.Fatalf("get questions: %v", err)
	}
	if len(fetched) != 1 {
		t.Errorf("got %d fetched questions, want 1", len(fetched))
	}

	// Purge questions by concept ID map
	purged, err := store.PurgeGeneratedQuestions(map[string]bool{conceptID: true})
	if err != nil {
		t.Fatalf("purge generated questions: %v", err)
	}
	if purged != 2 {
		t.Errorf("got %d purged questions, want 2", purged)
	}

	countAfter, err := store.GetQuestionCount(conceptID)
	if err != nil {
		t.Fatalf("get question count after purge: %v", err)
	}
	if countAfter != 0 {
		t.Errorf("got question count %d after purge, want 0", countAfter)
	}

	// Empty purge map is no-op
	purgedEmpty, err := store.PurgeGeneratedQuestions(nil)
	if err != nil || purgedEmpty != 0 {
		t.Errorf("expected 0 purged for empty map: err=%v purged=%d", err, purgedEmpty)
	}
}

func TestSQLiteStore_DailyActivity(t *testing.T) {
	store := newTestStore(t)
	stu, err := store.CreateStudent("daily-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	sess, err := store.CreateSession(stu.ID)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	now := time.Now().UTC()
	diff := 0.3
	_, err = store.RecordAttempt(AttemptEntry{
		SessionID:      sess.ID,
		StudentID:      stu.ID,
		ConceptID:      "calc.integral",
		Answer:         "x",
		Expected:       "x",
		Correct:        true,
		ElapsedSeconds: 4,
		Timestamp:      now,
		Question:       "int dx",
		Source:         "gen",
		Difficulty:     &diff,
	})
	if err != nil {
		t.Fatalf("record attempt: %v", err)
	}

	activities, err := store.GetDailyActivity(stu.ID, 7)
	if err != nil {
		t.Fatalf("get daily activity: %v", err)
	}
	if len(activities) != 1 {
		t.Fatalf("got %d activity entries, want 1", len(activities))
	}
	if activities[0].Questions != 1 || activities[0].Correct != 1 {
		t.Errorf("unexpected activity entry: %+v", activities[0])
	}
	if len(activities[0].Concepts) != 1 || activities[0].Concepts[0] != "calc.integral" {
		t.Errorf("got concepts %v, want [calc.integral]", activities[0].Concepts)
	}
}

func TestSQLiteStore_GetAdminUser(t *testing.T) {
	store := newTestStore(t)

	// Missing student returns nil
	adm, err := store.GetAdminUser("missing-admin-student")
	if err != nil {
		t.Fatalf("get admin user missing: %v", err)
	}
	if adm != nil {
		t.Errorf("expected nil for missing admin user, got %+v", adm)
	}

	stu, err := store.CreateStudent("admin-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	adm, err = store.GetAdminUser(stu.ID)
	if err != nil {
		t.Fatalf("get admin user existing: %v", err)
	}
	if adm == nil {
		t.Fatal("expected admin user projection, got nil")
	}
	if adm.ID != stu.ID {
		t.Errorf("got ID %q, want %q", adm.ID, stu.ID)
	}
	if adm.Name != "admin-student" {
		t.Errorf("got name %q, want %q", adm.Name, "admin-student")
	}
	if adm.Role != RoleStudent {
		t.Errorf("got role %q, want %q", adm.Role, RoleStudent)
	}
}

func TestSQLiteStore_DiagnosticSessionDelete(t *testing.T) {
	store := newTestStore(t)
	stu, err := store.CreateStudent("diag-student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	sessID := "diag-sess-1"
	payload := []byte(`{"step":1}`)
	if err := store.SaveDiagnosticSession(sessID, stu.ID, payload); err != nil {
		t.Fatalf("save diagnostic session: %v", err)
	}

	data, found, err := store.LoadDiagnosticSession(sessID)
	if err != nil || !found || !bytes.Equal(data, payload) {
		t.Fatalf("unexpected load diagnostic session: found=%v err=%v data=%s", found, err, string(data))
	}

	// Delete by session ID
	if err := store.DeleteDiagnosticSession(sessID); err != nil {
		t.Fatalf("delete diagnostic session: %v", err)
	}
	_, found, err = store.LoadDiagnosticSession(sessID)
	if err != nil {
		t.Fatalf("load after delete: %v", err)
	}
	if found {
		t.Error("expected diagnostic session to be deleted")
	}

	// Delete by student ID
	sessID2 := "diag-sess-2"
	if err := store.SaveDiagnosticSession(sessID2, stu.ID, payload); err != nil {
		t.Fatalf("save diagnostic session 2: %v", err)
	}
	if err := store.DeleteDiagnosticSessionsForStudent(stu.ID); err != nil {
		t.Fatalf("delete student diagnostic sessions: %v", err)
	}
	_, found, err = store.LoadDiagnosticSession(sessID2)
	if err != nil {
		t.Fatalf("load after student delete: %v", err)
	}
	if found {
		t.Error("expected diagnostic session 2 to be deleted by student ID wipe")
	}

	// Empty student ID is no-op
	if err := store.DeleteDiagnosticSessionsForStudent(""); err != nil {
		t.Fatalf("delete empty student ID error: %v", err)
	}
}
