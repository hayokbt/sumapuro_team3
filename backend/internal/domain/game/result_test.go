package game

import "testing"

func TestElo(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		a, b, score, wantA, wantB float64
	}{
		{"equal win", 1000, 1000, 1, 1016, 984},
		{"equal loss", 1000, 1000, 0, 984, 1016},
		{"equal draw", 1000, 1000, .5, 1000, 1000},
		{"upset", 1000, 1400, 1, 1029.09, 1370.91},
		{"unequal draw", 1000, 1400, .5, 1013.09, 1386.91},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, b := Elo(tc.a, tc.b, tc.score)
			if a != tc.wantA || b != tc.wantB {
				t.Fatalf("got %.2f %.2f, want %.2f %.2f", a, b, tc.wantA, tc.wantB)
			}
		})
	}
}
