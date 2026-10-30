package httpapi

import (
	"fmt"
	"net/http"

	"gomoku-backend/internal/interface/account/accountcreate"
	"gomoku-backend/internal/interface/account/accountrecovery"
	"gomoku-backend/internal/interface/account/accountverify"
	"gomoku-backend/internal/interface/account/auth"
	"gomoku-backend/internal/interface/account/login"
	"gomoku-backend/internal/interface/httpapi/middleware"
	"gomoku-backend/internal/interface/httpapi/response"
	recoveryusecase "gomoku-backend/internal/usecase/account/accountrecovery"
	"gomoku-backend/internal/usecase/account/accountregistration"
	"gomoku-backend/internal/usecase/account/authentication"
)

func NewHandler(
	registration accountregistration.Service,
	loginService authentication.Service,
	recoveryService recoveryusecase.Service,
	sessions *auth.Manager,
	origins []string,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleHealth)
	mux.HandleFunc("/account/create", accountcreate.NewHandler(registration))
	mux.HandleFunc("/account/verify", accountverify.NewHandler(registration))
	mux.HandleFunc("/login", login.NewHandler(loginService, sessions))
	mux.HandleFunc("/account/recovery/request", accountrecovery.NewRequestHandler(recoveryService))
	mux.HandleFunc("/account/recovery/verify", accountrecovery.NewVerifyHandler(recoveryService))
	mux.HandleFunc("/account/recovery/reset", accountrecovery.NewResetHandler(recoveryService))
	mux.HandleFunc("/auth/me", sessions.Current)
	mux.HandleFunc("/logout", sessions.Logout)
	return middleware.CSRF(origins, mux)
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
