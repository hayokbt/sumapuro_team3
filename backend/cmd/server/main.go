package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"gomoku-backend/internal/infrastructure/database"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		fmt.Fprint(w, "This is go backend!!")
	})

	log.Println("Backend listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return fmt.Errorf("start HTTP server: %w", err)
	}
	return nil
}
