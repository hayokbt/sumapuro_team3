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
	"gomoku-backend/internal/interface/httpapi"
	"gomoku-backend/internal/usecase/accountregistration"
	"gomoku-backend/internal/usecase/authentication"
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
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewHandler(registration, loginService),
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
