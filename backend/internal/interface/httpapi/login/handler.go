package login

import (
	"errors"
	"net/http"

	"gomoku-backend/internal/domain"
	"gomoku-backend/internal/interface/httpapi/auth"
	"gomoku-backend/internal/interface/httpapi/response"
	"gomoku-backend/internal/usecase/authentication"
)

type request struct {
	AccountName string `json:"accountName"`
	Password    string `json:"password"`
}

type handler struct {
	authentication authentication.Service
	sessions       *auth.Manager
}

func NewHandler(authentication authentication.Service, sessions *auth.Manager) http.HandlerFunc {
	return (&handler{authentication: authentication, sessions: sessions}).handle
}

func (h *handler) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.WriteMethodNotAllowed(w, http.MethodPost)
		return
	}

	var input request
	if err := response.DecodeJSON(w, r, &input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "リクエストの形式が正しくありません。")
		return
	}
	user, err := h.authentication.Authenticate(r.Context(), input.AccountName, input.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.WriteError(w, http.StatusUnauthorized, "アカウント名またはパスワードが正しくありません。")
			return
		}
		response.WriteAuthenticationError(w, err)
		return
	}
	token, err := h.sessions.Sessions.Start(r.Context(), user)
	if err != nil {
		response.WriteAuthenticationError(w, err)
		return
	}
	h.sessions.Set(w, token)
	w.Header().Set("Cache-Control", "no-store")
	response.WriteJSON(w, http.StatusOK, response.JSON{Message: "ログインしました。"})
}
