package profile

import (
	"context"
	"encoding/json"
	"gomoku-backend/internal/domain/account"
	"gomoku-backend/internal/interface/account/auth"
	"gomoku-backend/internal/interface/httpapi/middleware"
	"net/http/httptest"
	"strings"
	"testing"
)

type repository struct {
	p        account.Profile
	updated  bool
	settings bool
}

func (r *repository) Get(context.Context, int64) (account.Profile, error) { return r.p, nil }
func (r *repository) Update(_ context.Context, _ int64, name, old, new string) error {
	r.updated = true
	r.p.AccountName = name
	r.p.PasswordHash = new
	return nil
}
func (r *repository) SaveSettings(_ context.Context, _ int64, s account.Settings) error {
	r.settings = true
	r.p.Settings = s
	return nil
}

type passwords struct{}

func (passwords) Hash(s string) (string, error) { return "hashed-" + s, nil }
func (passwords) Verify(s, hash string) (bool, error) {
	return s == "password123" && hash == "old-hash", nil
}

type sessions struct{}

func (sessions) Start(context.Context, account.AuthenticatedUser) (string, error) { return "", nil }
func (sessions) Logout(context.Context, string) error                             { return nil }
func (sessions) Verify(_ context.Context, s string) (account.AuthSession, error) {
	if s != "valid" {
		return account.AuthSession{}, account.ErrUnauthenticated
	}
	return account.AuthSession{UserID: 1}, nil
}
func TestProfileAndSettings(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body, cookie string
		status                           int
		updated, settings                bool
	}{
		{"unauthenticated", "GET", "/account", "", "", 401, false, false},
		{"read profile", "GET", "/account", "", "valid", 200, false, false},
		{"read settings", "GET", "/account/settings", "", "valid", 200, false, false},
		{"valid settings", "POST", "/account/settings", `{"bgmVolume":0,"seVolume":100}`, "valid", 200, false, true},
		{"missing setting", "POST", "/account/settings", `{"bgmVolume":20}`, "valid", 400, false, false},
		{"null setting", "POST", "/account/settings", `{"bgmVolume":null,"seVolume":0}`, "valid", 400, false, false},
		{"invalid setting", "POST", "/account/settings", `{"bgmVolume":101,"seVolume":0}`, "valid", 400, false, false},
		{"fraction setting", "POST", "/account/settings", `{"bgmVolume":1.5,"seVolume":0}`, "valid", 400, false, false},
		{"forged id", "POST", "/account/settings", `{"bgmVolume":10,"seVolume":0,"userId":2}`, "valid", 400, false, false},
		{"rename", "POST", "/account", `{"accountName":"new","currentPassword":"password123","password":""}`, "valid", 200, true, false},
		{"password change", "POST", "/account", `{"accountName":"new","currentPassword":"password123","password":"newpassword"}`, "valid", 200, true, false},
		{"wrong password", "POST", "/account", `{"accountName":"new","currentPassword":"wrongpassword"}`, "valid", 403, false, false},
		{"invalid name", "POST", "/account", `{"accountName":"  ","currentPassword":"password123"}`, "valid", 400, false, false},
		{"invalid password", "POST", "/account", `{"accountName":"new","currentPassword":"password123","password":"short"}`, "valid", 400, false, false},
		{"method", "DELETE", "/account", "", "valid", 405, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &repository{p: account.Profile{UserID: 1, AccountName: "old", Email: "test@example.com", PasswordHash: "old-hash", Rating: 1000}}
			manager := &auth.Manager{Sessions: sessions{}}
			handler := middleware.Authentication(manager, Handler{Repository: repo, Passwords: passwords{}, Auth: manager})
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.cookie != "" {
				req.Header.Set("Cookie", "gomoku_auth="+tc.cookie)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tc.status || repo.updated != tc.updated || repo.settings != tc.settings {
				t.Fatalf("status %d body %s updated %v settings %v", w.Code, w.Body.String(), repo.updated, repo.settings)
			}
			if strings.Contains(w.Body.String(), "old-hash") {
				t.Fatal("password hash leaked")
			}
			if tc.name == "password change" && len(w.Result().Cookies()) != 1 {
				t.Fatal("password change did not clear cookie")
			}
			if tc.name == "read profile" {
				var profile account.Profile
				if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil || profile.UserID != 1 {
					t.Fatal("invalid profile")
				}
			}
		})
	}
}
