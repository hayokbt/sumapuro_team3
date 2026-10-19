package accountverify

import (
	"net/http"

	"gomoku-backend/internal/interface/httpapi/response"
	"gomoku-backend/internal/usecase/accountregistration"
)

type request struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type handler struct {
	registration accountregistration.Service
}

func NewHandler(registration accountregistration.Service) http.HandlerFunc {
	return (&handler{registration: registration}).handle
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
	if err := h.registration.Verify(r.Context(), input.Email, input.Code); err != nil {
		response.WriteRegistrationError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, response.JSON{Message: "アカウントを作成しました。"})
}
