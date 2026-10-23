package auth

import (
	"errors"
	"testing"
	"time"
)

func TestCheckPassword(t *testing.T) {
	svc := newTestService(t)

	// Non-existent student
	if err := svc.CheckPassword("nonexistent-id", "SomePass123!"); err == nil {
		t.Fatal("expected error for non-existent student, got nil")
	}

	// Passwordless student (OAuth user without PasswordHash)
	st, err := svc.repo.CreateStudent("OAuth User")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if err := svc.CheckPassword(st.ID, "SomePass123!"); err == nil || err.Error() != "no password set on this account" {
		t.Fatalf("expected 'no password set on this account', got: %v", err)
	}

	// Student with password
	_, signedUp, err := svc.Signup("Ada Check", "adacheck", "InitialPass123!")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	if err := svc.CheckPassword(signedUp.ID, "InitialPass123!"); err != nil {
		t.Errorf("expected correct password to succeed, got: %v", err)
	}

	if err := svc.CheckPassword(signedUp.ID, "WrongPassword!"); !errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("expected ErrPasswordMismatch, got: %v", err)
	}
}

func TestChangePassword_EdgeCases(t *testing.T) {
	svc := newTestService(t)

	// Non-existent student
	if err := svc.ChangePassword("nonexistent-id", "Current123!", "NewPass123!"); err == nil {
		t.Fatal("expected error for non-existent student, got nil")
	}

	// Passwordless student
	st, err := svc.repo.CreateStudent("OAuth User 2")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if err := svc.ChangePassword(st.ID, "Old!", "NewPass123!"); err == nil || err.Error() != "no password set — use password reset instead" {
		t.Fatalf("expected 'no password set — use password reset instead', got: %v", err)
	}
}

func TestCompletePasswordReset(t *testing.T) {
	svc := newTestService(t)

	// Empty token
	if _, _, err := svc.CompletePasswordReset("", "NewPass123!"); err == nil || err.Error() != "reset token is required" {
		t.Fatalf("expected 'reset token is required', got: %v", err)
	}

	// Invalid password
	if _, _, err := svc.CompletePasswordReset("some-token", "weak"); err == nil {
		t.Fatal("expected validation error for weak password, got nil")
	}

	// Unknown / non-existent token
	if _, _, err := svc.CompletePasswordReset("nonexistent-token", "ValidPassword123!"); err == nil || err.Error() != "reset link is invalid or expired" {
		t.Fatalf("expected 'reset link is invalid or expired', got: %v", err)
	}

	// Set up valid token in storage
	_, st, err := svc.Signup("Ada Reset", "ada_reset", "OrigPassword123!")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	rawToken, err := newResetToken()
	if err != nil {
		t.Fatalf("newResetToken: %v", err)
	}
	hashed := hashResetToken(rawToken)
	if err := svc.repo.CreatePasswordReset(hashed, st.ID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("CreatePasswordReset: %v", err)
	}

	// Successfully complete reset
	tok, updatedStudent, err := svc.CompletePasswordReset(rawToken, "ResetPassword123!")
	if err != nil {
		t.Fatalf("CompletePasswordReset: %v", err)
	}
	if tok == "" {
		t.Error("expected non-empty JWT token on successful reset")
	}
	if updatedStudent.ID != st.ID {
		t.Errorf("student ID mismatch: got %q, want %q", updatedStudent.ID, st.ID)
	}

	// Verify new password works
	if err := svc.CheckPassword(st.ID, "ResetPassword123!"); err != nil {
		t.Errorf("expected new password to verify, got: %v", err)
	}

	// Verify token is single-use (replay attack fails)
	if _, _, err := svc.CompletePasswordReset(rawToken, "AnotherPassword123!"); err == nil || err.Error() != "reset link is invalid or expired" {
		t.Fatalf("expected single-use token replay to fail, got: %v", err)
	}
}

func TestRequestPasswordReset_Unconfigured(t *testing.T) {
	svc := newTestService(t)

	// When SMTP/SendByte is unconfigured, RequestPasswordReset returns (false, nil)
	// to prevent account enumeration.
	dispatched, err := svc.RequestPasswordReset("nonexistent-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dispatched {
		t.Error("expected dispatched=false for nonexistent user")
	}

	// Student with email on file
	st, err := svc.repo.CreateStudent("Email User")
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if err := svc.repo.SetEmail(st.ID, "user@example.com"); err != nil {
		t.Fatalf("set email: %v", err)
	}

	dispatched, err = svc.RequestPasswordReset("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dispatched {
		t.Error("expected dispatched=false when mail is not configured")
	}
}
