package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"gomoku-backend/internal/domain/game"
	"sort"
)

type GameResultRepository struct{ db *sql.DB }

func NewGameResultRepository(db *sql.DB) *GameResultRepository { return &GameResultRepository{db} }

// The unique game ID makes retries safe, including a lost COMMIT response.
func (r *GameResultRepository) Save(ctx context.Context, result game.Result) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	board, err := json.Marshal(result.Board)
	if err != nil {
		return err
	}
	var players [2]any
	for i, id := range result.Players {
		if id != 0 {
			players[i] = id
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO game_results
 (id,mode,black_user_id,white_user_id,winner,reason,board,started_at,finished_at,
 black_rating_before,white_rating_before,black_rating_after,white_rating_after)
 VALUES(?,?,?,?,?,?,?,?,?,1000,1000,1000,1000)`, result.ID, result.Mode, players[0], players[1], result.Winner, result.Reason, board, result.StartedAt.UTC(), result.FinishedAt.UTC())
	if err != nil {
		var duplicate *mysql.MySQLError
		if errors.As(err, &duplicate) && duplicate.Number == 1062 {
			return nil
		}
		return fmt.Errorf("insert game result: %w", err)
	}
	// Lock players in ID order so concurrent results cannot lose updates or deadlock.
	ids := []int64{}
	for _, id := range result.Players {
		if id != 0 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	ratings := map[int64]float64{0: game.InitialRating}
	for _, id := range ids {
		if _, err = tx.ExecContext(ctx, `INSERT INTO user_scores(user_id) VALUES(?) ON DUPLICATE KEY UPDATE user_id=user_id`, id); err != nil {
			return err
		}
		var rating float64
		if err = tx.QueryRowContext(ctx, `SELECT rating_score FROM user_scores WHERE user_id=? FOR UPDATE`, id).Scan(&rating); err != nil {
			return err
		}
		ratings[id] = rating
	}
	score := 0.5
	if result.Winner == 1 {
		score = 1
	} else if result.Winner == 2 {
		score = 0
	}
	before := [2]float64{ratings[result.Players[0]], ratings[result.Players[1]]}
	a, b := game.Elo(before[0], before[1], score)
	after := [2]float64{a, b}
	for i, id := range result.Players {
		if id == 0 {
			after[i] = game.InitialRating
			continue
		}
		win, loss, draw := 0, 0, 0
		if result.Winner == 0 {
			draw = 1
		} else if result.Winner == i+1 {
			win = 1
		} else {
			loss = 1
		}
		if _, err = tx.ExecContext(ctx, `UPDATE user_scores SET wins=wins+?,losses=losses+?,draws=draws+?,rating_score=? WHERE user_id=?`, win, loss, draw, after[i], id); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE game_results SET black_rating_before=?,white_rating_before=?,black_rating_after=?,white_rating_after=? WHERE id=?`, before[0], before[1], after[0], after[1], result.ID); err != nil {
		return err
	}
	return tx.Commit()
}
