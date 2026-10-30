package httpapi

import (
	"fmt"
	"net/http"

	"gomoku-backend/internal/interface/httpapi/accountcreate"
	"gomoku-backend/internal/interface/httpapi/accountrecovery"
	"gomoku-backend/internal/interface/httpapi/accountverify"
	"gomoku-backend/internal/interface/httpapi/auth"
	"gomoku-backend/internal/interface/httpapi/login"
	"gomoku-backend/internal/interface/httpapi/middleware"
	"gomoku-backend/internal/interface/httpapi/response"
	recoveryusecase "gomoku-backend/internal/usecase/accountrecovery"
	"gomoku-backend/internal/usecase/accountregistration"
	"gomoku-backend/internal/usecase/authentication"
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
