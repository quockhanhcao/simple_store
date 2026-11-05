package identity

import (
	"regexp"
	"testing"
)

var encodedHashPattern = regexp.MustCompile(`^\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`)

func TestHashPassword_Format(t *testing.T) {
	t.Parallel()

	for _, password := range []string{"Demo123!", "", "pässwörd 🔑"} {
		encoded, err := HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword(%q) error: %v", password, err)
		}
		if !encodedHashPattern.MatchString(encoded) {
			t.Errorf("HashPassword(%q) = %q, does not match %s", password, encoded, encodedHashPattern)
		}
	}
}

func TestHashPassword_UniqueSalt(t *testing.T) {
	t.Parallel()

	first, err := HashPassword("Demo123!")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("Demo123!")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Errorf("hashing the same password twice produced the same output %q; salt is not random", first)
	}
}

func TestVerifyPassword_RoundTrip(t *testing.T) {
	t.Parallel()

	const password = "Demo123!"
	encoded, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{name: "correct password", password: password, want: true},
		{name: "wrong password", password: "Demo123?", want: false},
		{name: "different case", password: "demo123!", want: false},
		{name: "empty password", password: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := VerifyPassword(tt.password, encoded)
			if err != nil {
				t.Fatalf("VerifyPassword error: %v", err)
			}
			if got != tt.want {
				t.Errorf("VerifyPassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestVerifyPassword_UsesParamsFromHash(t *testing.T) {
	t.Parallel()

	// Hash produced with weaker params (m=4096,t=1,p=1) than defaultParams;
	// verification must read params from the string, not from defaultParams.
	const encoded = "$argon2id$v=19$m=4096,t=1,p=1$c29tZXNhbHRzb21lc2FsdA$2gwotFrXd0+zAFJqh3DHaDnAxcUpU8hrA/whR9a0Ls0"

	got, err := VerifyPassword("password", encoded)
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if !got {
		t.Error("VerifyPassword = false, want true")
	}
}

func TestVerifyPassword_MalformedHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		encoded string
	}{
		{name: "empty", encoded: ""},
		{name: "too few parts", encoded: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA"},
		{name: "wrong variant", encoded: "$argon2i$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA"},
		{name: "wrong version", encoded: "$argon2id$v=16$m=65536,t=3,p=4$c2FsdA$aGFzaA"},
		{name: "bad params", encoded: "$argon2id$v=19$m=x,t=3,p=4$c2FsdA$aGFzaA"},
		{name: "bad salt base64", encoded: "$argon2id$v=19$m=65536,t=3,p=4$!!!$aGFzaA"},
		{name: "bad hash base64", encoded: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$!!!"},
		{name: "zero iterations", encoded: "$argon2id$v=19$m=65536,t=0,p=4$c2FsdA$aGFzaA"},
		{name: "zero parallelism", encoded: "$argon2id$v=19$m=65536,t=3,p=0$c2FsdA$aGFzaA"},
		{name: "memory too large", encoded: "$argon2id$v=19$m=2097152,t=3,p=4$c2FsdA$aGFzaA"},
		{name: "empty salt", encoded: "$argon2id$v=19$m=65536,t=3,p=4$$aGFzaA"},
		{name: "empty hash", encoded: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := VerifyPassword("Demo123!", tt.encoded)
			if err == nil {
				t.Errorf("VerifyPassword(%q) error = nil, want error", tt.encoded)
			}
			if got {
				t.Errorf("VerifyPassword(%q) = true, want false", tt.encoded)
			}
		})
	}
}
