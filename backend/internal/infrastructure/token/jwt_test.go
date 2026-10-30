package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gomoku-backend/internal/domain/account"
)

func TestJWTValidation(t *testing.T) {
	key := "01234567890123456789012345678901"
	codec, err := NewJWT(key)
	if err != nil {
		t.Fatal(err)
	}
	s := account.AuthSession{ID: "0123456789012345678901234567890123456789012345678901234567890123", UserID: 42, Version: 2, ExpiresAt: time.Now().Add(time.Minute)}
	raw, err := codec.Issue(s)
	if err != nil {
		t.Fatal(err)
	}
	got, err := codec.Verify(raw)
	if err != nil || got.UserID != 42 || got.Version != 2 {
		t.Fatalf("round trip: %v %v", got, err)
	}
	other, _ := NewJWT("abcdefghijklmnopqrstuvwxyz123456")
	if _, err := other.Verify(raw); err == nil {
		t.Fatal("accepted wrong signature")
	}
	s.ExpiresAt = time.Now().Add(-time.Minute)
	expired, _ := codec.Issue(s)
	if _, err := codec.Verify(expired); err == nil {
		t.Fatal("accepted expired token")
	}
	for _, method := range []jwt.SigningMethod{jwt.SigningMethodHS384, jwt.SigningMethodHS256} {
		c := claims{SessionID: s.ID, Version: 2, RegisteredClaims: jwt.RegisteredClaims{Subject: "42", Issuer: "gomoku-backend", Audience: jwt.ClaimStrings{"gomoku-web"}}}
		invalid, _ := jwt.NewWithClaims(method, c).SignedString([]byte(key))
		if _, err := codec.Verify(invalid); err == nil {
			t.Fatal("accepted algorithm mismatch or missing expiry")
		}
	}
	if _, err := NewJWT("short"); err == nil {
		t.Fatal("accepted short key")
	}
}
