package auth

import (
	"testing"
	"time"
)

func TestVerifyEmail_Scenarios(t *testing.T) {
	svc := newTestService(t)

	// Empty token
	if _, err := svc.VerifyEmail(""); err == nil || err.Error() != "verification token is required" {
		t.Fatalf("expected 'verification token is required', got: %v", err)
	}

	// Unknown / expired token
	if _, err := svc.VerifyEmail("nonexistent-token"); err == nil || err.Error() != "verification link is invalid or expired" {
		t.Fatalf("expected 'verification link is invalid or expired', got: %v", err)
	}

	// Create student with unverified email
	st, err := svc.repo.CreateStudent("Unverified Student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	email := "unverified@example.com"
	if err := svc.repo.SetEmail(st.ID, email); err != nil {
		t.Fatalf("set email: %v", err)
	}

	// Mint token
	rawToken, err := newResetToken()
	if err != nil {
		t.Fatalf("newResetToken: %v", err)
	}
	if err := svc.repo.CreateEmailVerification(hashResetToken(rawToken), st.ID, time.Now().UTC().Add(EmailVerifyTTL)); err != nil {
		t.Fatalf("CreateEmailVerification: %v", err)
	}

	// Verify email succeeds
	verifiedID, err := svc.VerifyEmail(rawToken)
	if err != nil {
		t.Fatalf("VerifyEmail failed: %v", err)
	}
	if verifiedID != st.ID {
		t.Errorf("verified ID = %q, want %q", verifiedID, st.ID)
	}

	// Check repository state: student is now verified
	loaded, err := svc.repo.GetStudent(st.ID)
	if err != nil || loaded == nil {
		t.Fatalf("get student: %v", err)
	}
	if !loaded.EmailVerified {
		t.Error("expected loaded.EmailVerified to be true")
	}

	// Replay must fail (single-use token consumed)
	if _, err := svc.VerifyEmail(rawToken); err == nil || err.Error() != "verification link is invalid or expired" {
		t.Fatalf("expected replay to fail, got: %v", err)
	}
}

func TestLinkToken_Lifecycle(t *testing.T) {
	svc := newTestService(t)

	// Consume empty token
	if _, ok, err := svc.ConsumeLinkToken(""); err != nil || ok {
		t.Fatalf("expected ok=false and nil error for empty link token, got ok=%v, err=%v", ok, err)
	}

	// Consume non-existent token
	if _, ok, err := svc.ConsumeLinkToken("fake-token"); err != nil || ok {
		t.Fatalf("expected ok=false for non-existent token, got ok=%v, err=%v", ok, err)
	}

	// Create link token for student
	st, err := svc.repo.CreateStudent("Link Student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}

	token, err := svc.CreateLinkToken(st.ID)
	if err != nil {
		t.Fatalf("CreateLinkToken: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty link token")
	}

	// Consume token successfully
	studentID, ok, err := svc.ConsumeLinkToken(token)
	if err != nil || !ok {
		t.Fatalf("ConsumeLinkToken failed: ok=%v, err=%v", ok, err)
	}
	if studentID != st.ID {
		t.Errorf("studentID = %q, want %q", studentID, st.ID)
	}

	// Replay must fail
	if _, ok, err := svc.ConsumeLinkToken(token); err != nil || ok {
		t.Fatalf("expected replay of link token to fail, got ok=%v, err=%v", ok, err)
	}
}

func TestRequestEmailVerification_Validation(t *testing.T) {
	svc := newTestService(t)

	// Non-existent student
	if err := svc.RequestEmailVerification("nonexistent-id"); err == nil || err.Error() != "student not found" {
		t.Fatalf("expected 'student not found', got: %v", err)
	}

	// Student with empty email
	st, err := svc.repo.CreateStudent("No Email Student")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if err := svc.RequestEmailVerification(st.ID); err == nil || err.Error() != "no email on file — add one in Settings first" {
		t.Fatalf("expected 'no email on file', got: %v", err)
	}

	// Student with email, but unconfigured mail
	if err := svc.repo.SetEmail(st.ID, "ada@example.com"); err != nil {
		t.Fatalf("set email: %v", err)
	}
	if err := svc.RequestEmailVerification(st.ID); err == nil || err.Error() != "email verification not configured" {
		t.Fatalf("expected 'email verification not configured', got: %v", err)
	}
}
