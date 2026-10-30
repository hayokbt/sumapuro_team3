package token

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gomoku-backend/internal/domain"
)

type JWT struct{ key []byte }
type claims struct {
	SessionID string `json:"sid"`
	Version   int64  `json:"ver"`
	jwt.RegisteredClaims
}

func NewJWT(key string) (*JWT, error) {
	if len(key) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	return &JWT{[]byte(key)}, nil
}
func (j *JWT) Issue(s domain.AuthSession) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims{s.ID, s.Version, jwt.RegisteredClaims{Subject: strconv.FormatInt(s.UserID, 10), Issuer: "gomoku-backend", Audience: jwt.ClaimStrings{"gomoku-web"}, IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(s.ExpiresAt)}}).SignedString(j.key)
}
func (j *JWT) Verify(raw string) (domain.AuthSession, error) {
	c := &claims{}
	_, err := jwt.ParseWithClaims(raw, c, func(*jwt.Token) (any, error) { return j.key, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuer("gomoku-backend"), jwt.WithAudience("gomoku-web"))
	if err != nil {
		return domain.AuthSession{}, err
	}
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil || id <= 0 || len(c.SessionID) != 64 || c.Version < 1 {
		return domain.AuthSession{}, domain.ErrUnauthenticated
	}
	return domain.AuthSession{ID: c.SessionID, UserID: id, Version: c.Version, ExpiresAt: c.ExpiresAt.Time}, nil
}
