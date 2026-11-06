package game

import "errors"

const Size = 15

// Black (1) always moves first. Zero denotes an empty intersection.
type Game struct {
	Board    [Size * Size]int `json:"board"`
	Turn     int              `json:"turn"`
	Winner   int              `json:"winner"`
	Finished bool             `json:"finished"`
	Reason   string           `json:"reason"`
	LastMove int              `json:"lastMove"`
}

func New() *Game { return &Game{Turn: 1, LastMove: -1} }

func (g *Game) Move(player, index int) error {
	if g.Finished || player != g.Turn || index < 0 || index >= len(g.Board) || g.Board[index] != 0 {
		return errors.New("この場所には打てません。手番と空きマスを確認してください。")
	}
	g.Board[index] = player
	g.LastMove = index
	if g.hasFive(index, player) {
		g.Finished, g.Winner, g.Reason = true, player, "five"
	} else {
		full := true
		for _, cell := range g.Board {
			if cell == 0 {
				full = false
				break
			}
		}
		if full {
			g.Finished, g.Reason = true, "draw"
		} else {
			g.Turn = 3 - player
		}
	}
	return nil
}

func (g *Game) line(index, player int) int {
	best := 1
	for _, d := range [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}} {
		count := 1
		for _, sign := range []int{-1, 1} {
			r, c := index/Size+d[0]*sign, index%Size+d[1]*sign
			for r >= 0 && r < Size && c >= 0 && c < Size && g.Board[r*Size+c] == player {
				count++
				r += d[0] * sign
				c += d[1] * sign
			}
		}
		if count > best {
			best = count
		}
	}
	return best
}

// Overlines do not win; an exact five in another direction still wins.
func (g *Game) hasFive(index, player int) bool {
	for _, d := range [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}} {
		count := 1
		for _, sign := range []int{-1, 1} {
			r, c := index/Size+d[0]*sign, index%Size+d[1]*sign
			for r >= 0 && r < Size && c >= 0 && c < Size && g.Board[r*Size+c] == player {
				count++
				r += d[0] * sign
				c += d[1] * sign
			}
		}
		if count == 5 {
			return true
		}
	}
	return false
}

// CPU prioritizes a winning move, then blocks an immediate loss, then extends lines.
func (g *Game) CPU() int {
	best, bestScore := -1, -1
	for i, cell := range g.Board {
		if cell != 0 {
			continue
		}
		g.Board[i] = g.Turn
		attack := g.line(i, g.Turn)
		win := g.hasFive(i, g.Turn)
		g.Board[i] = 3 - g.Turn
		defend := g.line(i, 3-g.Turn)
		block := g.hasFive(i, 3-g.Turn)
		g.Board[i] = 0
		score := attack*attack*10 + defend*defend*9
		if block {
			score += 100000
		}
		if win {
			score += 1000000
		}
		// Prefer the center when strategic scores tie.
		r, c := i/Size-Size/2, i%Size-Size/2
		score = score*1000 - r*r - c*c
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	return best
}
