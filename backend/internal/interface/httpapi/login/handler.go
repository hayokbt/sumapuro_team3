package login

import (
	"errors"
	"net/http"

	"gomoku-backend/internal/domain"
	"gomoku-backend/internal/interface/httpapi/response"
	"gomoku-backend/internal/usecase/authentication"
)

type request struct {
	AccountName string `json:"accountName"`
	Password    string `json:"password"`
}

type handler struct {
	authentication authentication.Service
}

func NewHandler(authentication authentication.Service) http.HandlerFunc {
	return (&handler{authentication: authentication}).handle
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
	if err := h.authentication.Authenticate(r.Context(), input.AccountName, input.Password); err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			response.WriteError(w, http.StatusUnauthorized, "アカウント名またはパスワードが正しくありません。")
			return
		}
		response.WriteAuthenticationError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, response.JSON{Message: "ログインしました。"})
}
