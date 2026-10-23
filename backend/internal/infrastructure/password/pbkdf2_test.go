package password

import "testing"

func TestPBKDF2HashAndVerify(t *testing.T) {
	hasher := NewPBKDF2()
	encodedHash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	matches, err := hasher.Verify("correct horse battery staple", encodedHash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !matches {
		t.Fatal("Verify() did not match the correct password")
	}

	matches, err = hasher.Verify("wrong password", encodedHash)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if matches {
		t.Fatal("Verify() matched an incorrect password")
	}
}

func TestPBKDF2VerifyRejectsInvalidHashes(t *testing.T) {
	for _, encodedHash := range []string{
		"",
		"unsupported$600000$MDEyMzQ1Njc4OWFiY2RlZg$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"pbkdf2-sha256$999999999$MDEyMzQ1Njc4OWFiY2RlZg$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"pbkdf2-sha256$600000$invalid$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	} {
		if _, err := NewPBKDF2().Verify("password", encodedHash); err == nil {
			t.Errorf("Verify(%q) returned no error for an invalid hash", encodedHash)
		}
	}
}
