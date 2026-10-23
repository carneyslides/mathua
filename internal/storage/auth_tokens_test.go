package storage

import (
	"testing"
	"time"
)

func TestPasswordReset_Lifecycle(t *testing.T) {
	s := newTestStore(t)
	user, err := s.CreateUser("Alice", "alice_reset", "initial_hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// 1. Valid token consumption.
	tokenHash := "valid_token_hash_1"
	expires := time.Now().UTC().Add(30 * time.Minute)
	if err := s.CreatePasswordReset(tokenHash, user.ID, expires); err != nil {
		t.Fatalf("create password reset: %v", err)
	}

	gotID, ok, err := s.ConsumePasswordReset(tokenHash)
	if err != nil {
		t.Fatalf("consume password reset: %v", err)
	}
	if !ok || gotID != user.ID {
		t.Fatalf("consume returned ok=%v, gotID=%q, want %q", ok, gotID, user.ID)
	}

	// 2. Replay prevention: already burned token.
	gotID, ok, err = s.ConsumePasswordReset(tokenHash)
	if err != nil {
		t.Fatalf("re-consume error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected burned token to fail, got ok=%v, id=%q", ok, gotID)
	}

	// 3. Expired token rejection.
	expiredTokenHash := "expired_token_hash"
	pastExpires := time.Now().UTC().Add(-15 * time.Minute)
	if err := s.CreatePasswordReset(expiredTokenHash, user.ID, pastExpires); err != nil {
		t.Fatalf("create expired reset: %v", err)
	}
	gotID, ok, err = s.ConsumePasswordReset(expiredTokenHash)
	if err != nil {
		t.Fatalf("consume expired reset error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected expired token to fail, got ok=%v, id=%q", ok, gotID)
	}

	// 4. Non-existent token rejection.
	gotID, ok, err = s.ConsumePasswordReset("nonexistent_reset_token")
	if err != nil {
		t.Fatalf("consume missing token error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected missing token to fail, got ok=%v, id=%q", ok, gotID)
	}

	// 5. Update password hash.
	newHash := "updated_secure_hash"
	if err := s.SetPasswordHash(user.ID, newHash); err != nil {
		t.Fatalf("set password hash: %v", err)
	}
	refreshed, err := s.GetStudent(user.ID)
	if err != nil {
		t.Fatalf("get student: %v", err)
	}
	if refreshed.PasswordHash != newHash {
		t.Fatalf("expected password hash %q, got %q", newHash, refreshed.PasswordHash)
	}
}

func TestEmailVerification_Lifecycle(t *testing.T) {
	s := newTestStore(t)
	user, err := s.CreateUser("Bob", "bob_verify", "hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// 1. Initial email verification status.
	student, err := s.GetStudent(user.ID)
	if err != nil {
		t.Fatalf("get student: %v", err)
	}
	if student.EmailVerified {
		t.Fatalf("expected email_verified=false initially, got true")
	}

	// 2. Valid email verification consumption.
	tokenHash := "email_verify_hash_1"
	expires := time.Now().UTC().Add(1 * time.Hour)
	if err := s.CreateEmailVerification(tokenHash, user.ID, expires); err != nil {
		t.Fatalf("create email verification: %v", err)
	}

	gotID, ok, err := s.ConsumeEmailVerification(tokenHash)
	if err != nil {
		t.Fatalf("consume email verification: %v", err)
	}
	if !ok || gotID != user.ID {
		t.Fatalf("consume email verification returned ok=%v, id=%q, want %q", ok, gotID, user.ID)
	}

	// 3. Replay prevention.
	gotID, ok, err = s.ConsumeEmailVerification(tokenHash)
	if err != nil {
		t.Fatalf("re-consume error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected used email verification token to fail, got ok=%v, id=%q", ok, gotID)
	}

	// 4. Expired token rejection.
	expiredHash := "expired_email_verify_hash"
	if err := s.CreateEmailVerification(expiredHash, user.ID, time.Now().UTC().Add(-10*time.Minute)); err != nil {
		t.Fatalf("create expired verification: %v", err)
	}
	gotID, ok, err = s.ConsumeEmailVerification(expiredHash)
	if err != nil {
		t.Fatalf("consume expired verification error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected expired verification token to fail, got ok=%v, id=%q", ok, gotID)
	}

	// 5. SetEmailVerified update.
	if err := s.SetEmailVerified(user.ID, true); err != nil {
		t.Fatalf("set email verified true: %v", err)
	}
	student, err = s.GetStudent(user.ID)
	if err != nil {
		t.Fatalf("get student: %v", err)
	}
	if !student.EmailVerified {
		t.Fatalf("expected email_verified=true after update")
	}

	if err := s.SetEmailVerified(user.ID, false); err != nil {
		t.Fatalf("set email verified false: %v", err)
	}
	student, err = s.GetStudent(user.ID)
	if err != nil {
		t.Fatalf("get student: %v", err)
	}
	if student.EmailVerified {
		t.Fatalf("expected email_verified=false after second update")
	}
}

func TestLinkToken_Lifecycle(t *testing.T) {
	s := newTestStore(t)
	user, err := s.CreateUser("Charlie", "charlie_link", "hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// 1. Valid link token consumption.
	tokenHash := "link_token_hash_1"
	expires := time.Now().UTC().Add(20 * time.Minute)
	if err := s.CreateLinkToken(tokenHash, user.ID, expires); err != nil {
		t.Fatalf("create link token: %v", err)
	}

	gotID, ok, err := s.ConsumeLinkToken(tokenHash)
	if err != nil {
		t.Fatalf("consume link token: %v", err)
	}
	if !ok || gotID != user.ID {
		t.Fatalf("consume link token returned ok=%v, id=%q, want %q", ok, gotID, user.ID)
	}

	// 2. Replay prevention.
	gotID, ok, err = s.ConsumeLinkToken(tokenHash)
	if err != nil {
		t.Fatalf("re-consume error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected used link token to fail, got ok=%v, id=%q", ok, gotID)
	}

	// 3. Expired token rejection.
	expiredHash := "expired_link_token"
	if err := s.CreateLinkToken(expiredHash, user.ID, time.Now().UTC().Add(-5*time.Minute)); err != nil {
		t.Fatalf("create expired link token: %v", err)
	}
	gotID, ok, err = s.ConsumeLinkToken(expiredHash)
	if err != nil {
		t.Fatalf("consume expired link token error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected expired link token to fail, got ok=%v, id=%q", ok, gotID)
	}

	// 4. Missing token rejection.
	gotID, ok, err = s.ConsumeLinkToken("missing_link_token")
	if err != nil {
		t.Fatalf("consume missing link token error: %v", err)
	}
	if ok || gotID != "" {
		t.Fatalf("expected missing link token to fail, got ok=%v, id=%q", ok, gotID)
	}
}

func TestIdentities_Lifecycle(t *testing.T) {
	s := newTestStore(t)
	user, err := s.CreateUser("Dana", "dana_id", "hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	// 1. Initially no identities.
	ids, err := s.ListIdentities(user.ID)
	if err != nil {
		t.Fatalf("list identities: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected 0 identities, got %d", len(ids))
	}

	// 2. Add Google and Apple identities.
	if err := s.CreateIdentity("google", "gid-12345", user.ID, "dana@gmail.com", true); err != nil {
		t.Fatalf("create google identity: %v", err)
	}
	if err := s.CreateIdentity("apple", "aid-67890", user.ID, "dana@privaterelay.appleid.com", true); err != nil {
		t.Fatalf("create apple identity: %v", err)
	}

	// 3. List identities: ordered by provider ("apple", "google").
	ids, err = s.ListIdentities(user.ID)
	if err != nil {
		t.Fatalf("list identities after creation: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 identities, got %d", len(ids))
	}
	if ids[0].Provider != "apple" || ids[0].ProviderID != "aid-67890" {
		t.Errorf("expected first identity apple, got %+v", ids[0])
	}
	if ids[1].Provider != "google" || ids[1].ProviderID != "gid-12345" {
		t.Errorf("expected second identity google, got %+v", ids[1])
	}

	// 4. FindStudentByIdentity.
	byGoogle, err := s.FindStudentByIdentity("google", "gid-12345")
	if err != nil || byGoogle == nil || byGoogle.ID != user.ID {
		t.Fatalf("find by google identity: got %+v, err=%v", byGoogle, err)
	}

	byApple, err := s.FindStudentByIdentity("apple", "aid-67890")
	if err != nil || byApple == nil || byApple.ID != user.ID {
		t.Fatalf("find by apple identity: got %+v, err=%v", byApple, err)
	}

	// 5. Lookups with empty parameters or non-existent identities return nil.
	emptyLookup, err := s.FindStudentByIdentity("", "gid-12345")
	if err != nil || emptyLookup != nil {
		t.Fatalf("expected nil for empty provider, got %+v, err=%v", emptyLookup, err)
	}
	emptyLookup, err = s.FindStudentByIdentity("google", "")
	if err != nil || emptyLookup != nil {
		t.Fatalf("expected nil for empty providerID, got %+v, err=%v", emptyLookup, err)
	}
	missingLookup, err := s.FindStudentByIdentity("google", "nonexistent_uid")
	if err != nil || missingLookup != nil {
		t.Fatalf("expected nil for non-existent identity, got %+v, err=%v", missingLookup, err)
	}

	// 6. Delete identity: delete Apple identity, ensure Google identity persists.
	if err := s.DeleteIdentity("apple", user.ID); err != nil {
		t.Fatalf("delete apple identity: %v", err)
	}
	ids, err = s.ListIdentities(user.ID)
	if err != nil {
		t.Fatalf("list identities after deletion: %v", err)
	}
	if len(ids) != 1 || ids[0].Provider != "google" {
		t.Fatalf("expected 1 google identity after deletion, got %+v", ids)
	}

	byApple, err = s.FindStudentByIdentity("apple", "aid-67890")
	if err != nil || byApple != nil {
		t.Fatalf("expected apple identity to be gone, got %+v, err=%v", byApple, err)
	}
	byGoogle, err = s.FindStudentByIdentity("google", "gid-12345")
	if err != nil || byGoogle == nil || byGoogle.ID != user.ID {
		t.Fatalf("expected google identity to remain valid, got %+v, err=%v", byGoogle, err)
	}
}

func TestCreateGoogleUser(t *testing.T) {
	s := newTestStore(t)
	user, err := s.CreateGoogleUser("Eve Google", "eve@example.com", "google-sub-777", "https://example.com/avatar.png")
	if err != nil {
		t.Fatalf("create google user: %v", err)
	}

	if user.ID == "" {
		t.Fatal("expected non-empty user ID")
	}
	if !user.EmailVerified {
		t.Error("expected EmailVerified to be true for Google users")
	}
	if user.GoogleID != "google-sub-777" {
		t.Errorf("expected GoogleID google-sub-777, got %q", user.GoogleID)
	}
	if user.AvatarURL != "https://example.com/avatar.png" {
		t.Errorf("expected AvatarURL https://example.com/avatar.png, got %q", user.AvatarURL)
	}

	// Verify identity lookup.
	found, err := s.FindStudentByIdentity("google", "google-sub-777")
	if err != nil || found == nil || found.ID != user.ID {
		t.Fatalf("find google user by identity: got %+v, err=%v", found, err)
	}

	// Verify listed identities.
	ids, err := s.ListIdentities(user.ID)
	if err != nil || len(ids) != 1 || ids[0].ProviderID != "google-sub-777" {
		t.Fatalf("list identities for google user: got %+v, err=%v", ids, err)
	}
}

func TestFindByUsername(t *testing.T) {
	s := newTestStore(t)
	user, err := s.CreateUser("Ada", "ada_lovelace", "secret_hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := s.FindByUsername("ada_lovelace")
	if err != nil || found == nil || found.ID != user.ID {
		t.Fatalf("find by username: got %+v, err=%v", found, err)
	}

	missing, err := s.FindByUsername("nonexistent_username")
	if err != nil || missing != nil {
		t.Fatalf("expected nil for missing username, got %+v, err=%v", missing, err)
	}
}
