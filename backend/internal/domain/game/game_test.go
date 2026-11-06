package game

import "testing"

func TestExactFiveAndOverlines(t *testing.T) {
	for _, d := range [][2]int{{0, 1}, {1, 0}, {1, 1}, {1, -1}} {
		for _, length := range []int{5, 6, 7} {
			g := New()
			start := 3*Size + 7
			gap := 2
			for j := 0; j < length; j++ {
				if j != gap {
					g.Board[start+j*(d[0]*Size+d[1])] = 1
				}
			}
			if err := g.Move(1, start+gap*(d[0]*Size+d[1])); err != nil {
				t.Fatal(err)
			}
			if g.Finished != (length == 5) {
				t.Fatalf("direction %v length %d: finished=%v", d, length, g.Finished)
			}
		}
	}
}

func TestFiveCrossingOverline(t *testing.T) {
	g := New()
	center := 7*Size + 7
	for _, offset := range []int{-3, -2, -1, 1, 2} {
		g.Board[center+offset] = 1
	}
	for _, offset := range []int{-2, -1, 1, 2} {
		g.Board[center+offset*Size] = 1
	}
	if err := g.Move(1, center); err != nil {
		t.Fatal(err)
	}
	if !g.Finished || g.Winner != 1 {
		t.Fatal("exact vertical five must win even with horizontal six")
	}
}

func TestInvalidMoves(t *testing.T) {
	g := New()
	for _, input := range [][2]int{{2, 0}, {1, -1}, {1, 225}} {
		if g.Move(input[0], input[1]) == nil {
			t.Fatal("accepted invalid move")
		}
	}
	_ = g.Move(1, 0)
	if g.Move(2, 0) == nil {
		t.Fatal("accepted occupied intersection")
	}
	g.Finished = true
	if g.Move(2, 1) == nil {
		t.Fatal("accepted move after finish")
	}
}

func TestDraw(t *testing.T) {
	g := New()
	for i := range g.Board {
		g.Board[i] = 1 + (i/Size+i%Size)%2
	}
	g.Board[0] = 0
	if err := g.Move(1, 0); err != nil {
		t.Fatal(err)
	}
	if !g.Finished || g.Winner != 0 || g.Reason != "draw" {
		t.Fatalf("not a draw: %+v", g)
	}
}

func TestCPUWinBlockAndOverline(t *testing.T) {
	for _, opponent := range []bool{false, true} {
		g := New()
		color := 1
		if opponent {
			color = 2
		}
		for i := 0; i < 4; i++ {
			g.Board[i] = color
		}
		if move := g.CPU(); move != 4 {
			t.Fatalf("CPU chose %d, expected winning/blocking move 4", move)
		}
	}
	g := New()
	for _, i := range []int{0, 1, 3, 4, 5} {
		g.Board[i] = 1
	}
	for i := 0; i < 4; i++ {
		g.Board[Size+i] = 1
	}
	if move := g.CPU(); move != Size+4 {
		t.Fatalf("CPU must prefer exact five over six, got %d", move)
	}
}
