package auth

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/storage"
)

func newTestService(t *testing.T) *AuthService {
	t.Helper()
	repo, err := storage.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	return New(repo)
}

func TestNormalizeUsername(t *testing.T) {
	if got := NormalizeUsername("  Ada_Lovelace "); got != "ada_lovelace" {
		t.Errorf("expected trimmed lowercase, got %q", got)
	}
}

func TestNormalizeEmail(t *testing.T) {
	if got := NormalizeEmail("  Ada@Example.COM "); got != "ada@example.com" {
		t.Errorf("expected trimmed lowercase, got %q", got)
	}
	// Plus-addressing is provider-specific: never folded.
	if got := NormalizeEmail("ada+tag@example.com"); got != "ada+tag@example.com" {
		t.Errorf("expected plus preserved, got %q", got)
	}
}

func TestSignupLogin_CaseInsensitive(t *testing.T) {
	svc := newTestService(t)
	token, st, err := svc.Signup("Ada", "Ada", "Engine!n1")
	if err != nil || token == "" {
		t.Fatalf("signup: %v", err)
	}
	if st.Username != "ada" {
		t.Errorf("expected stored lowercase username, got %q", st.Username)
	}
	for _, variant := range []string{"ada", "ADA", "  Ada  "} {
		tok, _, err := svc.Login(variant, "Engine!n1")
		if err != nil || tok == "" {
			t.Errorf("login(%q): expected success, got %v", variant, err)
		}
	}
}

func TestSignup_DuplicateUsername(t *testing.T) {
	svc := newTestService(t)
	if _, _, err := svc.Signup("Ada", "ada", "Engine!n1"); err != nil {
		t.Fatalf("first signup: %v", err)
	}
	_, _, err := svc.Signup("Other", "ADA", "Engine!n2")
	if err == nil {
		t.Fatal("expected duplicate error for case-variant username")
	}
	if err != ErrUsernameTaken && !storage.IsUniqueViolation(err) {
		t.Errorf("expected taken/unique-violation error, got: %v", err)
	}
}

func TestLogin_GoogleOnlyByEmail(t *testing.T) {
	svc := newTestService(t)
	if _, err := svc.repo.CreateGoogleUser("Gigi", "gigi@example.com", "gid-9", ""); err != nil {
		t.Fatalf("create google user: %v", err)
	}
	_, _, err := svc.Login("gigi@example.com", "whatever")
	if err != ErrGoogleOnly {
		t.Errorf("expected ErrGoogleOnly, got: %v", err)
	}
}

func TestLogin_UnknownEmailSilent(t *testing.T) {
	svc := newTestService(t)
	// Unknown identifiers fail silently (nil error, empty token) so the
	// handler can answer generic 401 without user enumeration.
	if _, _, err := svc.Login("nobody@example.com", "x"); err != nil {
		t.Errorf("expected silent miss (nil error), got: %v", err)
	}
}

func TestValidatePassword_MaxBytes(t *testing.T) {
	long := strings.Repeat("a1!", 25) // 75 bytes
	if err := ValidatePassword(long); err == nil {
		t.Error("expected rejection past 72 bytes")
	}
	ok := strings.Repeat("a1!", 24) // 72 bytes
	if err := ValidatePassword(ok); err != nil {
		t.Errorf("expected 72-byte password accepted, got %v", err)
	}
}

func TestPasswordReset_CompleteAndSingleUse(t *testing.T) {
	svc := newTestService(t)
	_, st, err := svc.Signup("Rita", "rita", "Engine!n1")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	// Seed a token directly (no SMTP in tests).
	raw := "test-reset-token-123"
	if err := svc.repo.CreatePasswordReset(hashResetToken(raw), st.ID, time.Now().UTC().Add(ResetTokenTTL)); err != nil {
		t.Fatalf("seed token: %v", err)
	}
	tok, rst, err := svc.CompletePasswordReset(raw, "N3w!passw")
	if err != nil || tok == "" {
		t.Fatalf("complete: %v", err)
	}
	if rst.ID != st.ID {
		t.Errorf("expected student %q, got %q", st.ID, rst.ID)
	}
	// New password works.
	if _, _, err := svc.Login("rita", "N3w!passw"); err != nil {
		t.Errorf("login with new password: %v", err)
	}
	// Replay burns out.
	if _, _, err := svc.CompletePasswordReset(raw, "Other!n99"); err == nil {
		t.Error("expected replayed token rejected")
	}
}

func TestPasswordReset_ExpiredRejected(t *testing.T) {
	svc := newTestService(t)
	_, st, err := svc.Signup("Rita2", "rita2", "Engine!n1")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	raw := "expired-token-1"
	if err := svc.repo.CreatePasswordReset(hashResetToken(raw), st.ID, time.Now().UTC().Add(-time.Minute)); err != nil {
		t.Fatalf("seed token: %v", err)
	}
	if _, _, err := svc.CompletePasswordReset(raw, "N3w!passw"); err == nil {
		t.Error("expected expired token rejected")
	}
}

func TestPasswordReset_RequestSilent(t *testing.T) {
	svc := newTestService(t)
	// Unknown identifier and missing mailer both look identical: no error.
	if _, err := svc.RequestPasswordReset("nobody-here"); err != nil {
		t.Errorf("expected silent unknown, got %v", err)
	}
	if _, err := svc.RequestPasswordReset(""); err != nil {
		t.Errorf("expected silent empty, got %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	svc := newTestService(t)
	_, st, err := svc.Signup("Cid", "cid", "Engine!n1")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	if err := svc.ChangePassword(st.ID, "Wrong!n9", "N3w!passw"); err == nil {
		t.Error("expected wrong-current rejected")
	}
	if err := svc.ChangePassword(st.ID, "Engine!n1", "weak"); err == nil {
		t.Error("expected weak-new rejected")
	}
	if err := svc.ChangePassword(st.ID, "Engine!n1", "N3w!passw"); err != nil {
		t.Fatalf("change: %v", err)
	}
	if _, _, err := svc.Login("cid", "N3w!passw"); err != nil {
		t.Errorf("login with changed password: %v", err)
	}
}

func TestIssueAndValidateToken(t *testing.T) {
	svc := newTestService(t)

	// Empty student ID rejected
	if _, err := IssueGuestToken(""); err == nil {
		t.Error("expected error issuing guest token for empty student ID")
	}

	// Valid token issuance and validation
	sid := "guest_student_xyz_123"
	tok, err := IssueGuestToken(sid)
	if err != nil || tok == "" {
		t.Fatalf("issue guest token: %v", err)
	}

	gotID, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if gotID != sid {
		t.Errorf("ValidateToken got %q, want %q", gotID, sid)
	}

	// Malformed / invalid tokens
	if _, err := svc.ValidateToken(""); err == nil {
		t.Error("expected error for empty token")
	}
	if _, err := svc.ValidateToken("invalid.jwt.token"); err == nil {
		t.Error("expected error for garbage token")
	}
}

func TestCheckPassword(t *testing.T) {
	svc := newTestService(t)
	_, st, err := svc.Signup("PassUser", "pass_user", "Valid!pass1")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	// Correct password
	if err := svc.CheckPassword(st.ID, "Valid!pass1"); err != nil {
		t.Errorf("expected correct password to succeed, got %v", err)
	}

	// Wrong password
	if err := svc.CheckPassword(st.ID, "Wrong!pass2"); err != ErrPasswordMismatch {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}

	// Passwordless account
	pwless, err := svc.repo.CreateOAuthUser("github", "gh-check", "NoPass", "nopass@example.com", true, "")
	if err != nil {
		t.Fatalf("create oauth user: %v", err)
	}
	if err := svc.CheckPassword(pwless.ID, "Any!pass1"); err == nil {
		t.Error("expected error checking password on passwordless account")
	}

	// Non-existent student
	if err := svc.CheckPassword("non-existent-student-id", "Any!pass1"); err == nil {
		t.Error("expected error for non-existent student ID")
	}
}

func TestOAuth_ProvidersAndCallbacks(t *testing.T) {
	// 1. isRelayAddress
	if !isRelayAddress("user123@privaterelay.appleid.com") {
		t.Error("expected apple private relay detected")
	}
	if !isRelayAddress("User123@PRIVATERELAY.APPLEID.COM") {
		t.Error("expected case-insensitive apple private relay detected")
	}
	if isRelayAddress("user@apple.com") {
		t.Error("apple.com is not private relay")
	}
	if isRelayAddress("user@gmail.com") || isRelayAddress("") {
		t.Error("non-relay address false positive")
	}

	// 2. CallbackURL
	origBase := os.Getenv("OAUTH_REDIRECT_BASE")
	defer os.Setenv("OAUTH_REDIRECT_BASE", origBase)

	os.Setenv("OAUTH_REDIRECT_BASE", "https://mathua.example.com")
	if got := CallbackURL(nil, "google"); got != "https://mathua.example.com/api/auth/google/callback" {
		t.Errorf("CallbackURL with env base got %q", got)
	}

	os.Unsetenv("OAUTH_REDIRECT_BASE")
	if got := CallbackURL(nil, "google"); got != "" {
		t.Errorf("expected empty CallbackURL when no env and no request, got %q", got)
	}

	reqLocal := &http.Request{Host: "localhost:8080"}
	if got := CallbackURL(reqLocal, "github"); got != "http://localhost:8080/api/auth/github/callback" {
		t.Errorf("CallbackURL localhost got %q", got)
	}

	reqFwd := &http.Request{
		Host:   "mathua.org",
		Header: http.Header{"X-Forwarded-Proto": []string{"https"}},
	}
	if got := CallbackURL(reqFwd, "apple"); got != "https://mathua.org/api/auth/apple/callback" {
		t.Errorf("CallbackURL forwarded got %q", got)
	}

	// 3. Drivers map
	drivers := Drivers()
	for _, p := range []string{"github", "facebook", "microsoft", "apple"} {
		d, ok := drivers[p]
		if !ok {
			t.Errorf("missing driver for provider %q", p)
		}
		if d.ClientID == nil || d.LoginURL == nil || d.Exchange == nil {
			t.Errorf("driver for %q has nil functions", p)
		}
	}
}
