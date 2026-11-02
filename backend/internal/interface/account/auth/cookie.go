package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"gomoku-backend/internal/domain/account"
	"gomoku-backend/internal/usecase/account/authsession"
)

const CookieName = "gomoku_auth"

type Sessions interface {
	Start(context.Context, account.AuthenticatedUser) (string, error)
	Verify(context.Context, string) (account.AuthSession, error)
	Logout(context.Context, string) error
}
type Manager struct {
	Sessions Sessions
	Secure   bool
}

func (m *Manager) Set(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: token, Path: "/", HttpOnly: true, Secure: m.Secure, SameSite: http.SameSiteLaxMode, MaxAge: int(authsession.Lifetime.Seconds()), Expires: time.Now().Add(authsession.Lifetime)})
}
func (m *Manager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Path: "/", HttpOnly: true, Secure: m.Secure, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
}
func Raw(r *http.Request) string {
	c, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
func (m *Manager) Current(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", 405)
		return
	}
	s, err := m.Sessions.Verify(r.Context(), Raw(r))
	if err != nil {
		status := 500
		if errors.Is(err, account.ErrUnauthenticated) {
			status = 401
			m.Clear(w)
		}
		http.Error(w, "認証状態を確認できませんでした。", status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"userId": s.UserID})
}
func (m *Manager) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", 405)
		return
	}
	if err := m.Sessions.Logout(r.Context(), Raw(r)); err != nil {
		http.Error(w, "ログアウトできませんでした。", 500)
		return
	}
	m.Clear(w)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"ログアウトしました。"}`))
}
