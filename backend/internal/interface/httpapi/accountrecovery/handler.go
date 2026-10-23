package accountrecovery

import (
	"net/http"

	"gomoku-backend/internal/interface/httpapi/response"
	"gomoku-backend/internal/usecase/accountrecovery"
)

type requestCodeRequest struct {
	Email string `json:"email"`
}

type verifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type resetAccountRequest struct {
	Email       string `json:"email"`
	ResetToken  string `json:"resetToken"`
	AccountName string `json:"accountName"`
	Password    string `json:"password"`
}

func NewRequestHandler(recovery accountrecovery.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteMethodNotAllowed(w, http.MethodPost)
			return
		}
		var input requestCodeRequest
		if err := response.DecodeJSON(w, r, &input); err != nil {
			response.WriteError(w, http.StatusBadRequest, "リクエストの形式が正しくありません。")
			return
		}
		if err := recovery.RequestCode(r.Context(), input.Email); err != nil {
			response.WriteAccountRecoveryError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusAccepted, response.JSON{
			Message: "アカウントが存在する場合、確認コードをメールで送信しました。",
		})
	}
}

func NewVerifyHandler(recovery accountrecovery.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteMethodNotAllowed(w, http.MethodPost)
			return
		}
		var input verifyCodeRequest
		if err := response.DecodeJSON(w, r, &input); err != nil {
			response.WriteError(w, http.StatusBadRequest, "リクエストの形式が正しくありません。")
			return
		}
		token, err := recovery.VerifyCode(r.Context(), input.Email, input.Code)
		if err != nil {
			response.WriteAccountRecoveryError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		response.WriteJSON(w, http.StatusOK, response.JSON{
			Message:    "確認コードを確認しました。",
			ResetToken: token,
		})
	}
}

func NewResetHandler(recovery accountrecovery.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			response.WriteMethodNotAllowed(w, http.MethodPost)
			return
		}
		var input resetAccountRequest
		if err := response.DecodeJSON(w, r, &input); err != nil {
			response.WriteError(w, http.StatusBadRequest, "リクエストの形式が正しくありません。")
			return
		}
		if err := recovery.ResetAccount(
			r.Context(),
			input.Email,
			input.ResetToken,
			input.AccountName,
			input.Password,
		); err != nil {
			response.WriteAccountRecoveryError(w, err)
			return
		}
		response.WriteJSON(w, http.StatusOK, response.JSON{Message: "アカウント情報を更新しました。"})
	}
}
