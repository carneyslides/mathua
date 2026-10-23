package auth

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestValidateEmailAcceptsSubdomains(t *testing.T) {
	cases := []struct {
		email string
		valid bool
	}{
		{"", true},
		{"   ", true},
		{"alice@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"not-an-email", false},
		{"@domain.com", false},
		{"user@", false},
		{"user name@domain.com", false},
	}

	for _, c := range cases {
		err := ValidateEmail(c.email)
		if c.valid && err != nil {
			t.Errorf("expected valid for %q, got error: %v", c.email, err)
		}
		if !c.valid && err == nil {
			t.Errorf("expected error for %q, got nil", c.email)
		}
	}
}

func TestGuestToken_ValidationAndExpiry(t *testing.T) {
	svc := newTestService(t)

	// Empty student ID rejected
	if _, err := IssueGuestToken(""); err == nil {
		t.Error("expected error for empty student ID, got nil")
	}
	if _, err := IssueGuestToken("   "); err == nil {
		t.Error("expected error for whitespace student ID, got nil")
	}

	// Valid guest student ID mints valid token
	tok, err := IssueGuestToken("guest_abcdef123456")
	if err != nil {
		t.Fatalf("issue guest token: %v", err)
	}
	if tok == "" {
		t.Fatal("empty token returned")
	}

	// Validate token resolves student ID
	sid, err := svc.ValidateToken(tok)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if sid != "guest_abcdef123456" {
		t.Errorf("expected sid guest_abcdef123456, got %q", sid)
	}

	// Invalid tokens rejected
	if _, err := svc.ValidateToken(""); err == nil {
		t.Error("expected error for empty token string")
	}
	if _, err := svc.ValidateToken("invalid.jwt.token"); err == nil {
		t.Error("expected error for malformed token string")
	}
}

func TestCheckPasswordRejectsAnEmptyPassword(t *testing.T) {
	svc := newTestService(t)

	// Create user with password
	_, st, err := svc.Signup("PassUser", "passuser", "CorrectHorseBattery9!")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}

	// Valid password check
	if err := svc.CheckPassword(st.ID, "CorrectHorseBattery9!"); err != nil {
		t.Errorf("expected correct password to succeed, got %v", err)
	}

	// Incorrect password check
	if err := svc.CheckPassword(st.ID, "WrongPassword!"); err != ErrPasswordMismatch {
		t.Errorf("expected ErrPasswordMismatch, got %v", err)
	}

	// Empty password check
	if err := svc.CheckPassword(st.ID, ""); err != ErrPasswordMismatch {
		t.Errorf("expected ErrPasswordMismatch for empty pass, got %v", err)
	}

	// Passwordless user (OAuth only)
	oauthUser, err := svc.repo.CreateUser("OAuthUser", "oauthuser", "")
	if err != nil {
		t.Fatalf("create oauth user: %v", err)
	}
	if err := svc.CheckPassword(oauthUser.ID, "AnyPassword!"); err == nil {
		t.Error("expected error checking password on passwordless account, got nil")
	}

	// Nonexistent student
	if err := svc.CheckPassword("nonexistent_id", "pass"); err == nil {
		t.Error("expected error for nonexistent student, got nil")
	}
}

func TestOAuth_CallbackURL(t *testing.T) {
	// Nil request
	if url := CallbackURL(nil, "github"); url != "" {
		t.Errorf("expected empty string for nil request, got %q", url)
	}

	// Localhost request
	reqLocal := httptest.NewRequest("GET", "http://localhost:3000/api/auth/login", nil)
	if url := CallbackURL(reqLocal, "github"); url != "http://localhost:3000/api/auth/github/callback" {
		t.Errorf("expected localhost callback URL, got %q", url)
	}

	// Forwarded proto
	reqRemote := httptest.NewRequest("GET", "http://mathua.org/api/auth/login", nil)
	reqRemote.Header.Set("X-Forwarded-Proto", "https")
	if url := CallbackURL(reqRemote, "microsoft"); url != "https://mathua.org/api/auth/microsoft/callback" {
		t.Errorf("expected https forwarded callback URL, got %q", url)
	}

	// OAUTH_REDIRECT_BASE override
	t.Setenv("OAUTH_REDIRECT_BASE", "https://auth.mathua.com/")
	if url := CallbackURL(reqLocal, "google"); url != "https://auth.mathua.com/api/auth/google/callback" {
		t.Errorf("expected OAUTH_REDIRECT_BASE callback URL, got %q", url)
	}
}

func TestOAuth_ConfiguredProviders(t *testing.T) {
	// Clean all credentials
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GITHUB_CLIENT_ID", "")
	t.Setenv("FACEBOOK_APP_ID", "")
	t.Setenv("FACEBOOK_CLIENT_ID", "")
	t.Setenv("MICROSOFT_CLIENT_ID", "")
	t.Setenv("APPLE_CLIENT_ID", "")

	if len(ConfiguredProviders()) != 0 {
		t.Errorf("expected 0 configured providers, got %v", ConfiguredProviders())
	}
	if ProviderConfigured("github") || ProviderConfigured("facebook") || ProviderConfigured("unknown") {
		t.Error("expected unconfigured providers to report false")
	}

	// Enable GitHub and Apple
	t.Setenv("GITHUB_CLIENT_ID", "gh-client-123")
	t.Setenv("APPLE_CLIENT_ID", "apple-client-456")

	if !ProviderConfigured("github") {
		t.Error("expected github configured")
	}
	if !ProviderConfigured("apple") {
		t.Error("expected apple configured")
	}
	if ProviderConfigured("microsoft") {
		t.Error("expected microsoft unconfigured")
	}

	configured := ConfiguredProviders()
	if len(configured) != 2 || configured[0] != "github" || configured[1] != "apple" {
		t.Errorf("expected [github apple], got %v", configured)
	}

	// Enable Facebook via FACEBOOK_APP_ID
	t.Setenv("FACEBOOK_APP_ID", "fb-app-123")
	if !ProviderConfigured("facebook") {
		t.Error("expected facebook configured via FACEBOOK_APP_ID")
	}

	// Enable Microsoft
	t.Setenv("MICROSOFT_CLIENT_ID", "ms-client-123")
	if !ProviderConfigured("microsoft") {
		t.Error("expected microsoft configured")
	}
}

func TestOAuth_DriversAndLoginURLs(t *testing.T) {
	drivers := Drivers()
	expectedDrivers := []string{"github", "facebook", "microsoft", "apple"}

	for _, name := range expectedDrivers {
		drv, ok := drivers[name]
		if !ok {
			t.Errorf("missing driver %q in Drivers()", name)
			continue
		}
		if drv.LoginURL == nil {
			t.Errorf("driver %q has nil LoginURL", name)
			continue
		}
		if drv.ClientID == nil {
			t.Errorf("driver %q has nil ClientID func", name)
		}
		if drv.Exchange == nil {
			t.Errorf("driver %q has nil Exchange func", name)
		}

		loginURL := drv.LoginURL("test-client-id", "https://example.com/callback", "state123")
		parsed, err := url.Parse(loginURL)
		if err != nil {
			t.Fatalf("driver %q LoginURL %q not valid URL: %v", name, loginURL, err)
		}

		q := parsed.Query()
		if q.Get("client_id") != "test-client-id" {
			t.Errorf("driver %q expected client_id test-client-id, got %q", name, q.Get("client_id"))
		}
		if q.Get("redirect_uri") != "https://example.com/callback" {
			t.Errorf("driver %q expected redirect_uri, got %q", name, q.Get("redirect_uri"))
		}
		if q.Get("state") != "state123" {
			t.Errorf("driver %q expected state state123, got %q", name, q.Get("state"))
		}
	}
}

func TestApple_IsRelayAddress(t *testing.T) {
	cases := []struct {
		email   string
		isRelay bool
	}{
		{"abc@privaterelay.appleid.com", true},
		{"ABC@PRIVATERELAY.APPLEID.COM", true},
		{"  user.hide@privaterelay.appleid.com  ", true},
		{"user@gmail.com", false},
		{"alice@apple.com", false},
		{"privaterelay.appleid.com", false},
		{"", false},
	}

	for _, c := range cases {
		got := isRelayAddress(c.email)
		if got != c.isRelay {
			t.Errorf("isRelayAddress(%q) = %v, want %v", c.email, got, c.isRelay)
		}
	}
}
