package httpapi

import (
	"fmt"
	"gomoku-backend/internal/domain/account"
	domainGame "gomoku-backend/internal/domain/game"
	"gomoku-backend/internal/infrastructure/password"
	"gomoku-backend/internal/interface/account/profile"
	gameapi "gomoku-backend/internal/interface/game"
	"net/http"

	"gomoku-backend/internal/domain/ranking"
	"gomoku-backend/internal/interface/account/accountcreate"
	"gomoku-backend/internal/interface/account/accountrecovery"
	"gomoku-backend/internal/interface/account/accountverify"
	"gomoku-backend/internal/interface/account/auth"
	"gomoku-backend/internal/interface/account/login"
	"gomoku-backend/internal/interface/httpapi/middleware"
	"gomoku-backend/internal/interface/httpapi/response"
	rankingapi "gomoku-backend/internal/interface/ranking"
	recoveryusecase "gomoku-backend/internal/usecase/account/accountrecovery"
	"gomoku-backend/internal/usecase/account/accountregistration"
	"gomoku-backend/internal/usecase/account/authentication"
)

func NewHandler(
	registration accountregistration.Service,
	loginService authentication.Service,
	recoveryService recoveryusecase.Service,
	sessions *auth.Manager,
	rankingRepository ranking.Repository,
	origins []string,
	profiles account.ProfileRepository,
	results domainGame.ResultRepository,
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
	mux.Handle("/ranking", rankingapi.NewHandler(rankingRepository))
	mux.Handle("/game/ws", middleware.Authentication(sessions, gameapi.NewHandler(origins, results)))
	profileHandler := middleware.Authentication(sessions, profile.Handler{Repository: profiles, Passwords: password.NewPBKDF2(), Auth: sessions})
	mux.Handle("/account", profileHandler)
	mux.Handle("/account/settings", profileHandler)
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
