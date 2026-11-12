package database

import (
	"context"
	"fmt"
	"gomoku-backend/internal/domain/account"
	"gomoku-backend/internal/domain/game"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Run against an isolated MySQL database with DB_INTEGRATION=1.
func TestGameResultsAndProfilesIntegration(t *testing.T) {
	if os.Getenv("DB_INTEGRATION") != "1" {
		t.Skip("set DB_INTEGRATION=1 with isolated MySQL")
	}
	ctx := context.Background()
	db, err := Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Schema setup is idempotent, including migration of existing users.
	if err = initializeSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	prefix := fmt.Sprintf("result_test_%d", time.Now().UnixNano())
	var ids []int64
	for i := 0; i < 22; i++ {
		name := fmt.Sprintf("%s_%d", prefix, i)
		res, err := db.Exec(`INSERT INTO users(username,email,password_hash) VALUES(?,?,?)`, name, name+"@example.com", "old-hash")
		if err != nil {
			t.Fatal(err)
		}
		id, _ := res.LastInsertId()
		ids = append(ids, id)
	}
	defer func() {
		for _, id := range ids {
			db.Exec(`DELETE FROM game_results WHERE black_user_id=? OR white_user_id=?`, id, id)
			db.Exec(`DELETE FROM users WHERE id=?`, id)
		}
	}()
	repo := NewGameResultRepository(db)
	result := game.Result{ID: strings.Repeat("a", 64), Mode: "online", Players: [2]int64{ids[0], ids[1]}, Winner: 1, Reason: "five", StartedAt: time.Now(), FinishedAt: time.Now()}
	// A fresh random ID allows rerunning this test in a persistent development DB.
	result.ID = fmt.Sprintf("%064x", time.Now().UnixNano())
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := repo.Save(ctx, result); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	profiles := NewProfileRepository(db)
	a, err := profiles.Get(ctx, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	b, err := profiles.Get(ctx, ids[1])
	if err != nil {
		t.Fatal(err)
	}
	if a.Wins != 1 || a.Rating != 1016 || b.Losses != 1 || b.Rating != 984 {
		t.Fatalf("duplicate updates or wrong Elo: %+v %+v", a, b)
	}
	var count int
	if err = db.QueryRow(`SELECT COUNT(*) FROM game_results WHERE id=?`, result.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("rows %d: %v", count, err)
	}
	var before, after float64
	if err = db.QueryRow(`SELECT black_rating_before,black_rating_after FROM game_results WHERE id=?`, result.ID).Scan(&before, &after); err != nil || before != 1000 || after != 1016 {
		t.Fatalf("rating snapshot: %f %f %v", before, after, err)
	}
	// CPU as black; a white player win is calculated against fixed 1000.
	result.ID = fmt.Sprintf("%064x", time.Now().UnixNano())
	result.Mode = "cpu"
	result.Players = [2]int64{0, ids[2]}
	result.Winner = 2
	result.Reason = "resign"
	if err = repo.Save(ctx, result); err != nil {
		t.Fatal(err)
	}
	p, err := profiles.Get(ctx, ids[2])
	if err != nil || p.Wins != 1 || p.Rating != 1016 {
		t.Fatalf("white CPU result %+v %v", p, err)
	}
	if err = db.QueryRow(`SELECT black_rating_after FROM game_results WHERE id=?`, result.ID).Scan(&after); err != nil || after != 1000 {
		t.Fatalf("CPU rating %f %v", after, err)
	}
	// CPU as white; draw changes the human's rating when ratings differ.
	result.ID = fmt.Sprintf("%064x", time.Now().UnixNano())
	result.Players = [2]int64{ids[2], 0}
	result.Winner = 0
	result.Reason = "draw"
	if err = repo.Save(ctx, result); err != nil {
		t.Fatal(err)
	}
	p, err = profiles.Get(ctx, ids[2])
	want, _ := game.Elo(1016, 1000, .5)
	if err != nil || p.Draws != 1 || p.Rating != want {
		t.Fatalf("CPU draw %+v %v", p, err)
	}
	// Failure cannot leave a result or change a score.
	bad := result
	bad.ID = fmt.Sprintf("%064x", time.Now().UnixNano())
	bad.Players = [2]int64{ids[0], 9223372036854775807}
	if repo.Save(ctx, bad) == nil {
		t.Fatal("accepted missing player")
	}
	if err = db.QueryRow(`SELECT COUNT(*) FROM game_results WHERE id=?`, bad.ID).Scan(&count); err != nil || count != 0 {
		t.Fatal("partial result committed")
	}
	if err = profiles.SaveSettings(ctx, ids[0], account.Settings{BGMVolume: 0, SEVolume: 100}); err != nil {
		t.Fatal(err)
	}
	if err = profiles.Update(ctx, ids[0], prefix+"_renamed", "old-hash", "new-hash"); err != nil {
		t.Fatal(err)
	}
	a, err = profiles.Get(ctx, ids[0])
	if err != nil || a.BGMVolume != 0 || a.SEVolume != 100 || a.PasswordHash != "new-hash" {
		t.Fatalf("profile persistence %+v %v", a, err)
	}
	var version int
	if err = db.QueryRow(`SELECT auth_version FROM users WHERE id=?`, ids[0]).Scan(&version); err != nil || version != 2 {
		t.Fatal("password change did not invalidate sessions")
	}
	if err = profiles.Update(ctx, ids[0], prefix+"_renamed", "old-hash", "stale-hash"); err != account.ErrUnauthenticated {
		t.Fatalf("stale password update: %v", err)
	}
	// All rankings are ordered on the server and capped at twenty.
	for i, id := range ids {
		if _, err = db.Exec(`INSERT INTO user_scores(user_id,rating_score) VALUES(?,?) ON DUPLICATE KEY UPDATE rating_score=?`, id, 3000+i, 3000+i); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := NewRankingRepository(db).List(ctx)
	if err != nil || len(entries) != 20 {
		t.Fatalf("ranking length %d %v", len(entries), err)
	}
	for i, e := range entries {
		if e.Position != i+1 || e.Rating != float64(3021-i) {
			t.Fatalf("ranking entry %+v", e)
		}
	}
}
