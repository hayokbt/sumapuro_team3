package database

import (
	"context"
	"database/sql"
	"fmt"

	"gomoku-backend/internal/domain/ranking"
)

type RankingRepository struct {
	db *sql.DB
}

func NewRankingRepository(db *sql.DB) *RankingRepository {
	return &RankingRepository{db: db}
}

func (r *RankingRepository) List(ctx context.Context) ([]ranking.Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT users.username, user_scores.rating_score
		FROM user_scores
		INNER JOIN users ON users.id = user_scores.user_id
		ORDER BY user_scores.rating_score DESC, user_scores.wins DESC, user_scores.user_id ASC
		LIMIT 20
	`)
	if err != nil {
		return nil, fmt.Errorf("query ranking: %w", err)
	}
	defer rows.Close()

	entries := make([]ranking.Entry, 0)
	for rows.Next() {
		var entry ranking.Entry
		entry.Position = len(entries) + 1
		if err := rows.Scan(&entry.AccountName, &entry.Rating); err != nil {
			return nil, fmt.Errorf("scan ranking entry: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ranking: %w", err)
	}
	return entries, nil
}
