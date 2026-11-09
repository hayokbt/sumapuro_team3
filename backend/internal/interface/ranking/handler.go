package ranking

import (
	"encoding/json"
	"log"
	"net/http"

	"gomoku-backend/internal/domain/ranking"
	"gomoku-backend/internal/interface/httpapi/response"
)

func NewHandler(repository ranking.Repository) http.Handler {
	return handler{repository: repository}
}

type handler struct {
	repository ranking.Repository
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.WriteMethodNotAllowed(w, http.MethodGet)
		return
	}

	entries, err := h.repository.List(r.Context())
	if err != nil {
		log.Printf("load ranking: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "ランキングを取得できませんでした。")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("write ranking response: %v", err)
	}
}
