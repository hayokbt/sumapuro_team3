package authsession

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"gomoku-backend/internal/domain"
)

const Lifetime = 30 * time.Minute

type Token interface {
	Issue(domain.AuthSession) (string, error)
	Verify(string) (domain.AuthSession, error)
}
type Service struct {
	repository domain.AuthSessionRepository
	token      Token
}

func New(repository domain.AuthSessionRepository, token Token) *Service {
	return &Service{repository, token}
}
func (s *Service) Start(ctx context.Context, u domain.AuthenticatedUser) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	session := domain.AuthSession{ID: hex.EncodeToString(b), UserID: u.ID, Version: u.Version, ExpiresAt: time.Now().UTC().Add(Lifetime).Truncate(time.Second)}
	token, err := s.token.Issue(session)
	if err != nil {
		return "", err
	}
	if err = s.repository.Create(ctx, session); err != nil {
		return "", err
	}
	return token, nil
}
func (s *Service) Verify(ctx context.Context, token string) (domain.AuthSession, error) {
	session, err := s.token.Verify(token)
	if err != nil {
		return session, domain.ErrUnauthenticated
	}
	valid, err := s.repository.Valid(ctx, session)
	if err != nil {
		return session, err
	}
	if !valid {
		return session, domain.ErrUnauthenticated
	}
	return session, nil
}
func (s *Service) Logout(ctx context.Context, token string) error {
	session, err := s.token.Verify(token)
	if err != nil {
		return nil
	}
	return s.repository.Revoke(ctx, session.ID)
}
