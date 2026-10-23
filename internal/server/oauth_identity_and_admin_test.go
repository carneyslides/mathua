package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chuma-beep/mathua/internal/concepts"
	"github.com/chuma-beep/mathua/internal/engine"
	"github.com/chuma-beep/mathua/internal/generator"
)

func TestOAuth_LinkToken(t *testing.T) {
	env := newAdminEnv(t)

	// Unauthenticated POST must return 401.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/link-token", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated link token code: %d, want 401", rec.Code)
	}

	// Method not allowed for GET.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/link-token", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET link token code: %d, want 405", rec.Code)
	}

	// Authenticated POST creates link token.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/link-token", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authenticated link token code: %d, want 200", rec.Code)
	}

	var res struct {
		LinkToken string `json:"link_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode link token response: %v", err)
	}
	if strings.TrimSpace(res.LinkToken) == "" {
		t.Fatalf("expected non-empty link_token")
	}
}

func TestOAuth_IdentitiesLifecycle(t *testing.T) {
	env := newAdminEnv(t)

	// Unauthenticated GET must return 401.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/identities", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated identities code: %d, want 401", rec.Code)
	}

	// Method not allowed for POST.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/identities", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST identities code: %d, want 405", rec.Code)
	}

	// Authenticated GET returns empty slice initially.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/identities", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("authenticated identities code: %d, want 200", rec.Code)
	}

	var out []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode identities: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected 0 identities, got %d", len(out))
	}

	// Seed one identity directly in repository.
	if err := env.store.CreateIdentity("google", "sub-12345", env.studentID, "ada@example.com", true); err != nil {
		t.Fatalf("create identity: %v", err)
	}

	// Listing identities now returns the created provider.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/identities", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("identities after creation code: %d, want 200", rec.Code)
	}

	out = nil
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode identities after creation: %v", err)
	}
	if len(out) != 1 || out[0]["provider"] != "google" {
		t.Fatalf("expected 1 google identity, got %+v", out)
	}

	// Unauthenticated DELETE must return 401.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/auth/identities/google", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated delete identity code: %d, want 401", rec.Code)
	}

	// Unknown provider DELETE must return 404.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/auth/identities/invalid-provider", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete unknown provider code: %d, want 404", rec.Code)
	}

	// Wrong HTTP method on provider endpoint returns 405.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/identities/google", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET provider identity code: %d, want 405", rec.Code)
	}

	// Disconnecting connected identity succeeds because user has password credential.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/auth/identities/google", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete google identity code: %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	// List identities is now empty again.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/identities", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	out = nil
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if len(out) != 0 {
		t.Fatalf("expected 0 identities after disconnect, got %d", len(out))
	}
}

func TestOAuth_EmailVerificationRoutes(t *testing.T) {
	env := newAdminEnv(t)

	// Unauthenticated email request returns 401.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/email/request", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated email request code: %d, want 401", rec.Code)
	}

	// Method not allowed for GET email request.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/email/request", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET email request code: %d, want 405", rec.Code)
	}

	// Authenticated email request in test environment returns 400 (unconfigured mailer).
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/email/request", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("email request unconfigured code: %d, want 400", rec.Code)
	}

	// Method not allowed for GET email verify.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/email/verify", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET email verify code: %d, want 405", rec.Code)
	}

	// Invalid JSON payload returns 400.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/email/verify", strings.NewReader("bad-json"))
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid json verify code: %d, want 400", rec.Code)
	}

	// Unknown or invalid token returns 400.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/email/verify", strings.NewReader(`{"token":"missing-token"}`))
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown token verify code: %d, want 400", rec.Code)
	}

	// Seed valid verification token in repository.
	rawToken := "valid-verify-token-12345"
	h := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(h[:])
	if err := env.store.CreateEmailVerification(tokenHash, env.studentID, time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("seed email verification: %v", err)
	}

	// Valid token verification returns 200 ok.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/email/verify", strings.NewReader(`{"token":"valid-verify-token-12345"}`))
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid verify code: %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	// Replay must fail with 400 since token was consumed.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/email/verify", strings.NewReader(`{"token":"valid-verify-token-12345"}`))
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("replayed token verify code: %d, want 400", rec.Code)
	}
}

func TestOAuth_GoogleConnectRoute(t *testing.T) {
	env := newAdminEnv(t)

	// Unauthenticated connect returns 401.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/google/connect", strings.NewReader(`{"id_token":"token"}`))
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated google connect code: %d, want 401", rec.Code)
	}

	// Method not allowed for GET.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/google/connect", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET google connect code: %d, want 405", rec.Code)
	}

	// Invalid JSON body returns 400.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/google/connect", strings.NewReader("{not-valid-json}"))
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed json google connect code: %d, want 400", rec.Code)
	}

	// Invalid or unconfigured id_token returns 401.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/google/connect", strings.NewReader(`{"id_token":"dummy.token.payload"}`))
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid id_token google connect code: %d, want 401, body: %s", rec.Code, rec.Body.String())
	}
}

func TestOAuth_LoginAndCallbackHandling(t *testing.T) {
	env := newAdminEnv(t)

	// Unknown provider login returns 404.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/nonexistent/login", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown provider login code: %d, want 404", rec.Code)
	}

	// POST login returns 405.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/github/login", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST oauth login code: %d, want 405", rec.Code)
	}

	// Unconfigured client ID returns 500.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/github/login", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unconfigured oauth login code: %d, want 500", rec.Code)
	}

	// Unknown provider callback returns 404.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/nonexistent/callback", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown provider callback code: %d, want 404", rec.Code)
	}

	// Unsupported method returns 405.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/auth/github/callback", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("PUT oauth callback code: %d, want 405", rec.Code)
	}

	// Error parameter triggers redirect to profile with error code.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?error=access_denied", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("oauth callback error code: %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "github_denied") {
		t.Fatalf("expected redirect with github_denied, got: %s", loc)
	}

	// Missing state cookie returns 400.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state=state123", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oauth callback missing cookie code: %d, want 400", rec.Code)
	}

	// Valid state cookie but missing code returns 400.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/auth/github/callback?state=state123", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "state123"})
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("oauth callback missing code: %d, want 400", rec.Code)
	}
}

func TestAdmin_AdminsRouteAndRequireAdmin(t *testing.T) {
	env := newAdminEnv(t)

	// Unauthenticated GET returns 401.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/admins", nil)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated admin admins code: %d, want 401", rec.Code)
	}

	// Student GET returns 403.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/admin/admins", nil)
	req.Header.Set("Authorization", "Bearer "+env.studentTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("student admin admins code: %d, want 403", rec.Code)
	}

	// Admin POST returns 405.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/admins", nil)
	req.Header.Set("Authorization", "Bearer "+env.adminTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("admin admins POST code: %d, want 405", rec.Code)
	}

	// Admin GET returns 200 with staff list.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/admin/admins?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+env.adminTok)
	env.mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("admin admins GET code: %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		Admins []map[string]interface{} `json:"admins"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode admin admins response: %v", err)
	}
	if len(res.Admins) == 0 {
		t.Fatalf("expected at least 1 admin in staff list")
	}

	// Verify requireAdmin method wrapper directly.
	d, _ := concepts.Build([]concepts.Concept{})
	reg := generator.NewRegistry()
	srv := New(engine.New(env.store, d, reg, nil, nil), env.store, nil)

	// When auth is disabled, requireAdmin returns 401.
	called := false
	handler := srv.requireAdmin(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	handler(rec, req)
	if rec.Code != http.StatusUnauthorized || called {
		t.Fatalf("requireAdmin with nil auth code: %d, called: %v", rec.Code, called)
	}
}

func TestAdmin_LimitAndOffsetClamping(t *testing.T) {
	// adminLimit tests
	if got := adminLimit("10", 5, 20); got != 10 {
		t.Fatalf("adminLimit(10) = %d, want 10", got)
	}
	if got := adminLimit("100", 5, 20); got != 20 {
		t.Fatalf("adminLimit(100) = %d, want 20", got)
	}
	if got := adminLimit("0", 5, 20); got != 5 {
		t.Fatalf("adminLimit(0) = %d, want 5", got)
	}
	if got := adminLimit("-4", 5, 20); got != 5 {
		t.Fatalf("adminLimit(-4) = %d, want 5", got)
	}
	if got := adminLimit("invalid", 5, 20); got != 5 {
		t.Fatalf("adminLimit(invalid) = %d, want 5", got)
	}

	// adminOffset tests
	if got := adminOffset("15"); got != 15 {
		t.Fatalf("adminOffset(15) = %d, want 15", got)
	}
	if got := adminOffset("-5"); got != 0 {
		t.Fatalf("adminOffset(-5) = %d, want 0", got)
	}
	if got := adminOffset("not-a-number"); got != 0 {
		t.Fatalf("adminOffset(not-a-number) = %d, want 0", got)
	}
}
