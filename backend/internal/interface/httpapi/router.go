package httpapi

import (
	"fmt"
	"net/http"

	"gomoku-backend/internal/interface/httpapi/accountcreate"
	"gomoku-backend/internal/interface/httpapi/accountverify"
	"gomoku-backend/internal/interface/httpapi/login"
	"gomoku-backend/internal/interface/httpapi/response"
	"gomoku-backend/internal/usecase/accountregistration"
	"gomoku-backend/internal/usecase/authentication"
)

func NewHandler(registration accountregistration.Service, loginService authentication.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHealth)
	mux.HandleFunc("/account/create", accountcreate.NewHandler(registration))
	mux.HandleFunc("/account/verify", accountverify.NewHandler(registration))
	mux.HandleFunc("/login", login.NewHandler(loginService))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		response.WriteMethodNotAllowed(w, http.MethodGet)
		return
	}
	_, _ = fmt.Fprint(w, "This is go backend!!")
}
