package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"gomoku-backend/internal/domain"
)

const maxRequestBodyBytes = 16 * 1024

type JSON struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request must contain one JSON value")
	}
	return nil
}

func WriteRegistrationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRegistration):
		WriteError(w, http.StatusBadRequest, "入力内容を確認してください。パスワードは8〜72バイトで入力してください。")
	case errors.Is(err, domain.ErrAccountExists):
		WriteError(w, http.StatusConflict, "メールアドレスまたはアカウント名は既に使用されています。")
	case errors.Is(err, domain.ErrInvalidCode):
		WriteError(w, http.StatusBadRequest, "確認コードが正しくありません。")
	case errors.Is(err, domain.ErrExpiredCode):
		WriteError(w, http.StatusGone, "確認コードの有効期限が切れています。最初からやり直してください。")
	case errors.Is(err, domain.ErrTooManyAttempts):
		WriteError(w, http.StatusTooManyRequests, "確認コードの入力回数上限に達しました。最初からやり直してください。")
	case errors.Is(err, domain.ErrEmailDelivery):
		log.Printf("send account verification email: %v", err)
		WriteError(w, http.StatusBadGateway, "確認メールを送信できませんでした。時間をおいて再度お試しください。")
	default:
		log.Printf("account registration request failed: %v", err)
		WriteError(w, http.StatusInternalServerError, "アカウント登録を処理できませんでした。")
	}
}

func WriteAuthenticationError(w http.ResponseWriter, err error) {
	log.Printf("account authentication request failed: %v", err)
	WriteError(w, http.StatusInternalServerError, "ログインを処理できませんでした。時間をおいて再度お試しください。")
}

func WriteMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, JSON{Error: message})
}

func WriteJSON(w http.ResponseWriter, status int, value JSON) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write API response: %v", err)
	}
}
