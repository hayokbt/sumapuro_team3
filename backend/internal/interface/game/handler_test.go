package game

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gomoku-backend/internal/domain/account"
	"gomoku-backend/internal/interface/account/auth"
	"gomoku-backend/internal/interface/httpapi/middleware"
)

type sessions struct{}

func (sessions) Start(context.Context, account.AuthenticatedUser) (string, error) { return "", nil }
func (sessions) Logout(context.Context, string) error                             { return nil }
func (sessions) Verify(_ context.Context, raw string) (account.AuthSession, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return account.AuthSession{}, account.ErrUnauthenticated
	}
	return account.AuthSession{UserID: id}, nil
}
func read(t *testing.T, c *websocket.Conn) state {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	var s state
	if err := c.ReadJSON(&s); err != nil {
		t.Fatal(err)
	}
	return s
}
func TestMatchingMovesAndDisconnect(t *testing.T) {
	h := NewHandler([]string{"http://game.test"})
	s := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
	defer s.Close()
	dial := func(id int, mode, origin string) (*websocket.Conn, *http.Response, error) {
		return websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode="+mode, http.Header{"Origin": {origin}, "Cookie": {fmt.Sprintf("gomoku_auth=%d", id)}})
	}
	_, response, err := dial(1, "cpu", "http://evil.test")
	if err == nil || response.StatusCode != 403 {
		t.Fatal("untrusted origin accepted")
	}
	c1, _, err := dial(1, "online", "http://game.test")
	if err != nil {
		t.Fatal(err)
	}
	defer c1.Close()
	if read(t, c1).Type != "waiting" {
		t.Fatal("not waiting")
	}
	_, response, err = dial(1, "online", "http://game.test")
	if err == nil || response.StatusCode != 409 {
		t.Fatal("duplicate user accepted")
	}
	c2, _, err := dial(2, "online", "http://game.test")
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()
	a, b := read(t, c1), read(t, c2)
	if a.Color+b.Color != 3 {
		t.Fatal("invalid assignment")
	}
	black, white := c1, c2
	if a.Color == 2 {
		black, white = c2, c1
	}
	_ = white.WriteJSON(map[string]any{"type": "move", "index": 0})
	if read(t, white).Type != "error" {
		t.Fatal("out of turn accepted")
	}
	_ = black.WriteJSON(map[string]any{"type": "move", "index": 0})
	for _, c := range []*websocket.Conn{black, white} {
		if x := read(t, c); x.Board[0] != 1 || x.Turn != 2 {
			t.Fatal("board not synchronized")
		}
	}
	black.Close()
	if x := read(t, white); !x.Finished || x.Winner != 2 || x.Reason != "disconnect" {
		t.Fatal("disconnect not resolved")
	}
}

func TestCPUAndResign(t *testing.T) {
	h := NewHandler([]string{"http://game.test"})
	s := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
	defer s.Close()
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode=cpu", http.Header{"Origin": {"http://game.test"}, "Cookie": {"gomoku_auth=1"}})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	x := read(t, c)
	if x.Turn != x.Color {
		t.Fatal("CPU did not make opening move")
	}
	index := 0
	for x.Board[index] != 0 {
		index++
	}
	_ = c.WriteJSON(map[string]any{"type": "move", "index": index})
	x = read(t, c)
	count := 0
	for _, v := range x.Board {
		if v != 0 {
			count++
		}
	}
	expected := 2
	if x.Color == 2 {
		expected = 3
	}
	if count != expected || x.Turn != x.Color {
		t.Fatal("CPU did not reply")
	}
	_ = c.WriteJSON(map[string]any{"type": "resign"})
	x = read(t, c)
	if !x.Finished || x.Winner != 3-x.Color || x.Reason != "resign" {
		t.Fatal("resignation not resolved")
	}
}

func TestOnlineVictory(t *testing.T) {
	h := NewHandler([]string{"http://game.test"})
	s := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
	defer s.Close()
	connect := func(id string) *websocket.Conn {
		t.Helper()
		c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode=online", http.Header{"Origin": {"http://game.test"}, "Cookie": {"gomoku_auth=" + id}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		return c
	}
	a := connect("1")
	read(t, a)
	b := connect("2")
	first := read(t, a)
	read(t, b)
	black, white := a, b
	if first.Color == 2 {
		black, white = b, a
	}
	for i := 0; i < 5; i++ {
		_ = black.WriteJSON(map[string]any{"type": "move", "index": i})
		x, y := read(t, black), read(t, white)
		if i == 4 {
			if !x.Finished || !y.Finished || x.Winner != 1 || y.Winner != 1 {
				t.Fatal("victory not delivered to both players")
			}
			break
		}
		_ = white.WriteJSON(map[string]any{"type": "move", "index": 30 + i*2})
		read(t, black)
		read(t, white)
	}
}

func TestOnlineWaitsUntilAnotherOnlinePlayerJoins(t *testing.T) {
	h := NewHandler([]string{"http://game.test"})
	s := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
	defer s.Close()
	connect := func(id, mode string) *websocket.Conn {
		t.Helper()
		c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode="+mode, http.Header{"Origin": {"http://game.test"}, "Cookie": {"gomoku_auth=" + id}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		return c
	}
	waiting := connect("1", "online")
	var notification struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	_ = waiting.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err := waiting.ReadJSON(&notification); err != nil {
		t.Fatal(err)
	}
	if notification.Type != "waiting" || notification.Message != "対戦相手を待っています。" {
		t.Fatalf("missing waiting notification: %+v", notification)
	}
	cpu := connect("2", "cpu")
	if read(t, cpu).Type != "state" {
		t.Fatal("CPU game did not start")
	}
	// A CPU player must never consume an online waiting player.
	h.mu.Lock()
	stillWaiting := len(h.waiting) == 1 && h.waiting[0].room == nil && h.waiting[0].color == 0
	h.mu.Unlock()
	if !stillWaiting {
		t.Fatal("online player started without an online opponent")
	}
	// Premature moves must keep the player waiting instead of creating a game.
	_ = waiting.WriteJSON(map[string]any{"type": "move", "index": 0})
	if read(t, waiting).Type != "waiting" {
		t.Fatal("waiting player did not remain waiting")
	}
	opponent := connect("3", "online")
	a, b := read(t, waiting), read(t, opponent)
	if a.Type != "state" || b.Type != "state" || a.Color+b.Color != 3 {
		t.Fatal("late opponent did not start the game")
	}
}
