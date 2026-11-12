package game

import (
	"context"
	"math"
	"time"
)

// CPU is represented by a zero user ID and always has rating 1000.
type Result struct {
	ID         string
	Mode       string
	Players    [2]int64
	Winner     int
	Reason     string
	Board      [225]int
	StartedAt  time.Time
	FinishedAt time.Time
}
type ResultRepository interface {
	Save(context.Context, Result) error
}

const InitialRating = 1000.0
const EloK = 32.0

// Elo returns ratings rounded to the database's two decimal places.
func Elo(black, white, blackScore float64) (float64, float64) {
	expected := 1 / (1 + math.Pow(10, (white-black)/400))
	delta := EloK * (blackScore - expected)
	return math.Round((black+delta)*100) / 100, math.Round((white-delta)*100) / 100
}
