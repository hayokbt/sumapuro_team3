package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gomoku-backend/internal/infrastructure/database"
	"gomoku-backend/internal/infrastructure/email"
	"gomoku-backend/internal/infrastructure/password"
	"gomoku-backend/internal/infrastructure/token"
	"gomoku-backend/internal/interface/httpapi"
	"gomoku-backend/internal/interface/httpapi/auth"
	"gomoku-backend/internal/usecase/accountrecovery"
	"gomoku-backend/internal/usecase/accountregistration"
	"gomoku-backend/internal/usecase/authentication"
	"gomoku-backend/internal/usecase/authsession"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	db, err := database.Open(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database connection: %v", err)
		}
	}()

	mailer, err := email.NewSMTP(
		envOrDefault("SMTP_HOST", "localhost"),
		envOrDefault("SMTP_PORT", "1025"),
		envOrDefault("SMTP_FROM", "no-reply@example.com"),
	)
	if err != nil {
		return fmt.Errorf("configure SMTP: %w", err)
	}

	registrationRepository := database.NewRegistrationRepository(db)
	registration := accountregistration.NewAccountRegistration(
		registrationRepository,
		mailer,
		password.NewPBKDF2(),
	)
	loginService := authentication.NewAuthentication(
		database.NewAuthenticationRepository(db),
		password.NewPBKDF2(),
	)
	recoveryService := accountrecovery.NewAccountRecovery(
		database.NewAccountRecoveryRepository(db),
		mailer,
		password.NewPBKDF2(),
	)
	secret, err := loadJWTSecret()
	if err != nil {
		return err
	}
	jwtToken, err := token.NewJWT(secret)
	if err != nil {
		return err
	}
	sessions := &auth.Manager{Sessions: authsession.New(database.NewAuthSessionRepository(db), jwtToken), Secure: os.Getenv("COOKIE_SECURE") == "true"}
	origins := strings.Split(envOrDefault("APP_ORIGINS", "http://localhost:5173,http://localhost:3000"), ",")
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewHandler(registration, loginService, recoveryService, sessions, origins),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Backend listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("start HTTP server: %w", err)
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
