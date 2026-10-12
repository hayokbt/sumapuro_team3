package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gomoku-backend/internal/infrastructure/database"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Open(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("close database connection: %v", err)
		}
	}()

	var usersCount, scoresCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&usersCount); err != nil {
		log.Fatalf("verify users table: %v", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_scores").Scan(&scoresCount); err != nil {
		log.Fatalf("verify user_scores table: %v", err)
	}

	fmt.Printf("MySQL connection successful; users=%d, user_scores=%d\n", usersCount, scoresCount)
}
