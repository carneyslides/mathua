package auth

import "testing"

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name string
		pw   string
		want string // empty means valid
	}{
		{"valid", "correct9!", ""},
		{"exactly 8 with all rules", "abcd12#$", ""},
		{"too short", "ab1!", "at least 8"},
		{"no digit", "abcdefgh!", "at least one number"},
		{"no special", "abcd1234", "at least one special"},
		{"empty", "", "at least 8"},
		{"unicode letter counts as letter", "λambda12!", ""},
		{"space counts as special", "abcd 123", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.pw)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected valid, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		email   string
		wantErr bool
	}{
		{"", false},
		{"   ", false},
		{"user@example.com", false},
		{"first.last+tag@sub.domain.org", false},
		{"plainaddress", true},
		{"@missinguser.com", true},
		{"missingdomain@", true},
	}
	for _, tc := range cases {
		err := ValidateEmail(tc.email)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateEmail(%q) err=%v, wantErr=%v", tc.email, err, tc.wantErr)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	cases := []struct {
		name    string
		u       string
		wantErr bool
	}{
		{"valid 3 chars", "ada", false},
		{"valid with underscore and dot", "ada_lovelace.01", false},
		{"valid numbers and letters", "student123", false},
		{"too short", "ab", true},
		{"too long", "a_very_long_username_exceeding_twenty", true},
		{"invalid char space", "user name", true},
		{"invalid char at", "user@name", true},
		{"invalid char hash", "user#1", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUsername(tc.u)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateUsername(%q) err=%v, wantErr=%v", tc.u, err, tc.wantErr)
			}
		})
	}
}


