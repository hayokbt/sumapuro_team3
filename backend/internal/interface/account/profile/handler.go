package profile

import (
	"encoding/json"
	"errors"
	"gomoku-backend/internal/domain/account"
	"gomoku-backend/internal/interface/account/auth"
	"gomoku-backend/internal/interface/httpapi/middleware"
	"gomoku-backend/internal/interface/httpapi/response"
	"log"
	"net/http"
)

type Passwords interface {
	Hash(string) (string, error)
	Verify(string, string) (bool, error)
}
type Handler struct {
	Repository account.ProfileRepository
	Passwords  Passwords
	Auth       *auth.Manager
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	id, ok := middleware.UserID(r.Context())
	if !ok {
		response.WriteError(w, 401, "ログインが必要です。")
		return
	}
	if r.Method != "GET" && r.Method != "POST" {
		response.WriteMethodNotAllowed(w, "GET, POST")
		return
	}
	p, err := h.Repository.Get(r.Context(), id)
	if err != nil {
		h.failure(w, err)
		return
	}
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.URL.Path == "/account/settings" {
			json.NewEncoder(w).Encode(p.Settings)
		} else {
			json.NewEncoder(w).Encode(p)
		}
		return
	}
	if r.URL.Path == "/account/settings" {
		var input struct {
			BGMVolume *int `json:"bgmVolume"`
			SEVolume  *int `json:"seVolume"`
		}
		if response.DecodeJSON(w, r, &input) != nil || input.BGMVolume == nil || input.SEVolume == nil {
			response.WriteError(w, 400, "BGMと効果音の音量を指定してください。")
			return
		}
		s := account.Settings{BGMVolume: *input.BGMVolume, SEVolume: *input.SEVolume}
		if s.BGMVolume < 0 || s.BGMVolume > 100 || s.SEVolume < 0 || s.SEVolume > 100 {
			response.WriteError(w, 400, "音量は0〜100で指定してください。")
			return
		}
		if err = h.Repository.SaveSettings(r.Context(), id, s); err != nil {
			h.failure(w, err)
			return
		}
	} else {
		var input struct {
			AccountName     string `json:"accountName"`
			CurrentPassword string `json:"currentPassword"`
			Password        string `json:"password"`
		}
		if response.DecodeJSON(w, r, &input) != nil || len(input.CurrentPassword) < 8 || len(input.CurrentPassword) > 72 {
			response.WriteError(w, 400, "入力内容と現在のパスワードを確認してください。")
			return
		}
		password := input.Password
		if password == "" {
			password = input.CurrentPassword
		}
		registration, err := account.NewRegistration(p.Email, input.AccountName, password)
		if err != nil {
			response.WriteError(w, 400, "名前は1〜50文字、パスワードは8〜72バイトで入力してください。")
			return
		}
		valid, err := h.Passwords.Verify(input.CurrentPassword, p.PasswordHash)
		if err != nil {
			h.failure(w, err)
			return
		}
		if !valid {
			response.WriteError(w, 403, "現在のパスワードが正しくありません。")
			return
		}
		hash := p.PasswordHash
		if input.Password != "" {
			hash, err = h.Passwords.Hash(input.Password)
			if err != nil {
				h.failure(w, err)
				return
			}
		}
		if err = h.Repository.Update(r.Context(), id, registration.AccountName, p.PasswordHash, hash); err != nil {
			h.failure(w, err)
			return
		}
		if input.Password != "" {
			h.Auth.Clear(w)
		}
	}
	response.WriteJSON(w, 200, response.JSON{Message: "保存しました。"})
}
func (h Handler) failure(w http.ResponseWriter, err error) {
	if errors.Is(err, account.ErrAccountExists) {
		response.WriteError(w, 409, "そのアカウント名は使用されています。")
		return
	}
	if errors.Is(err, account.ErrUnauthenticated) {
		response.WriteError(w, 401, "再ログインしてください。")
		return
	}
	log.Printf("account profile: %v", err)
	response.WriteError(w, 500, "アカウント情報を処理できませんでした。")
}
