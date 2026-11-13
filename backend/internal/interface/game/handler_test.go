package game

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gomoku-backend/internal/domain/account"
	board "gomoku-backend/internal/domain/game"
	"gomoku-backend/internal/interface/account/auth"
	"gomoku-backend/internal/interface/httpapi/middleware"
)

func TestMatchingThroughFrontendProxy(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node.js is required for the frontend proxy integration test")
	}
	h := NewHandler([]string{"http://localhost:5173"})
	backend := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
	defer backend.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	proxy := exec.CommandContext(ctx, "node", "--input-type=module", "-e", `
import { createServer } from 'node:http';
import { proxyWebSocket } from '../../../../frontend/server/websocket-proxy.mjs';
const server = createServer();
server.on('upgrade', (req, socket, head) => proxyWebSocket(req, socket, head, process.argv[1]));
server.listen(0, '127.0.0.1', () => console.log('http://127.0.0.1:' + server.address().port));
`, backend.URL)
	stdout, err := proxy.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := proxy.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cancel(); _ = proxy.Wait() })
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatal("proxy did not start")
	}
	proxyURL := scanner.Text()
	dial := func(id string) *websocket.Conn {
		t.Helper()
		c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(proxyURL, "http")+"/api/game/ws?mode=online", http.Header{"Origin": {"http://localhost:5173"}, "Cookie": {"gomoku_auth=" + id}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { c.Close() })
		return c
	}
	a := dial("1")
	if read(t, a).Type != "waiting" {
		t.Fatal("first player did not wait")
	}
	// Wait past the proxy's upgrade timeout before joining the second player.
	time.Sleep(11 * time.Second)
	// Two Chrome windows sharing a login cookie must get a useful error,
	// without disconnecting the player already in the queue.
	duplicate := dial("1")
	assertCloseCode(t, duplicate, closeDuplicate)
	if err := a.WriteJSON(map[string]string{"type": "move"}); err != nil {
		t.Fatal(err)
	}
	if read(t, a).Type != "waiting" {
		t.Fatal("duplicate connection evicted waiting player")
	}
	b := dial("2")
	x, y := read(t, a), read(t, b)
	if x.Type != "state" || y.Type != "state" || x.Color+y.Color != 3 {
		t.Fatal("matching failed through frontend proxy")
	}
	black := a
	if x.Color == 2 {
		black = b
	}
	if err := black.WriteJSON(map[string]any{"type": "move", "index": 0}); err != nil {
		t.Fatal(err)
	}
	if read(t, a).Board[0] != 1 || read(t, b).Board[0] != 1 {
		t.Fatal("move was not synchronized")
	}
}

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

func assertCloseCode(t *testing.T, c *websocket.Conn, code int) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err := c.ReadMessage()
	if !websocket.IsCloseError(err, code) {
		t.Fatalf("expected close code %d, got %v", code, err)
	}
}
func TestMatchingMovesAndDisconnect(t *testing.T) {
	h := NewHandler([]string{"http://game.test"})
	s := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
	defer s.Close()
	dial := func(id int, mode, origin string) (*websocket.Conn, *http.Response, error) {
		return websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode="+mode+"&color=black", http.Header{"Origin": {origin}, "Cookie": {fmt.Sprintf("gomoku_auth=%d", id)}})
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
	duplicate, _, err := dial(1, "online", "http://game.test")
	if err != nil {
		t.Fatal(err)
	}
	defer duplicate.Close()
	assertCloseCode(t, duplicate, closeDuplicate)
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
	c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode=cpu&color=black", http.Header{"Origin": {"http://game.test"}, "Cookie": {"gomoku_auth=1"}})
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
		c, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode="+mode+"&color=black", http.Header{"Origin": {"http://game.test"}, "Cookie": {"gomoku_auth=" + id}})
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

func TestCPUColorSelection(t *testing.T) {
	for _, color := range []string{"black", "white", "", "red"} {
		t.Run(color, func(t *testing.T) {
			h := NewHandler([]string{"http://game.test"})
			s := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
			defer s.Close()
			c, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode=cpu&color="+color, http.Header{"Origin": {"http://game.test"}, "Cookie": {"gomoku_auth=1"}})
			if color == "" || color == "red" {
				if err == nil {
					c.Close()
					t.Fatal("missing or invalid selection accepted")
				}
				if response == nil || response.StatusCode != 400 {
					t.Fatal("expected 400 for invalid selection")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			x := read(t, c)
			wantColor, wantStones := 1, 0
			if color == "white" {
				wantColor, wantStones = 2, 1
			}
			stones := 0
			for _, v := range x.Board {
				if v != 0 {
					stones++
					if v != 1 {
						t.Fatal("CPU opening must be black")
					}
				}
			}
			if x.Color != wantColor || x.Turn != wantColor || stones != wantStones {
				t.Fatalf("wrong initial state: color=%d turn=%d stones=%d", x.Color, x.Turn, stones)
			}
			index := 0
			for x.Board[index] != 0 {
				index++
			}
			_ = c.WriteJSON(map[string]any{"type": "move", "index": index})
			next := read(t, c)
			if next.Board[index] != wantColor || next.Turn != wantColor {
				t.Fatal("selected side not used for subsequent move")
			}
			stones = 0
			for _, v := range next.Board {
				if v != 0 {
					stones++
				}
			}
			if stones != wantStones+2 {
				t.Fatal("CPU did not reply with opposite color")
			}
		})
	}
}

type resultRecorder struct{ results chan board.Result }

func (r resultRecorder) Save(_ context.Context, result board.Result) error {
	r.results <- result
	return nil
}
func TestResultsRecordedOnceIncludingDisconnectedPlayer(t *testing.T) {
	for _, reason := range []string{"five", "draw", "resign", "disconnect"} {
		t.Run(reason, func(t *testing.T) {
			recorder := resultRecorder{make(chan board.Result, 4)}
			h := NewHandler(nil, recorder)
			r := newRoom(false)
			r.users = [2]int64{11, 22}
			a := &client{user: 11, color: 1, room: r, send: make(chan any, 32)}
			b := &client{user: 22, color: 2, room: r, send: make(chan any, 32)}
			r.players = [2]*client{a, b}
			h.mu.Lock()
			if reason == "disconnect" {
				h.leave(a)
			} else {
				r.game.Finished = true
				r.game.Winner = 1
				r.game.Reason = reason
				if reason == "draw" {
					r.game.Winner = 0
				}
				h.broadcast(r)
				h.broadcast(r)
			}
			h.mu.Unlock()
			select {
			case got := <-recorder.results:
				if got.Players != r.users || got.Reason != reason || got.ID == "" || got.FinishedAt.Before(got.StartedAt) {
					t.Fatalf("invalid saved result %+v", got)
				}
			case <-time.After(time.Second):
				t.Fatal("result not saved")
			}
			// Synchronize with the saved notification before testing a later disconnect.
			deadline := time.Now().Add(time.Second)
			for {
				h.mu.Lock()
				saved := r.saved
				h.mu.Unlock()
				if saved {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("missing save notification")
				}
				time.Sleep(time.Millisecond)
			}
			h.mu.Lock()
			h.leave(b)
			h.mu.Unlock()
			select {
			case <-recorder.results:
				t.Fatal("result saved twice")
			default:
			}
		})
	}
}

// A network interruption can leave the previous socket registered until its
// heartbeat expires. Explicit retries must work even while that socket is alive.
func TestExplicitRetryReplacesOldConnection(t *testing.T) {
	h := NewHandler([]string{"http://game.test"})
	s := httptest.NewServer(middleware.Authentication(&auth.Manager{Sessions: sessions{}}, h))
	defer s.Close()
	dial := func(query, origin string) (*websocket.Conn, *http.Response, error) {
		return websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http")+"?mode=online"+query, http.Header{"Origin": {origin}, "Cookie": {"gomoku_auth=1"}})
	}
	old, _, err := dial("", "http://game.test")
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	read(t, old)
	// A rejected handshake cannot evict the existing player.
	_, response, err := dial("&replace=1", "http://evil.test")
	if err == nil || response.StatusCode != 403 {
		t.Fatal("invalid origin accepted")
	}
	if err := old.WriteJSON(map[string]string{"type": "move"}); err != nil {
		t.Fatal(err)
	}
	if read(t, old).Type != "waiting" {
		t.Fatal("failed upgrade evicted player")
	}
	next, _, err := dial("&replace=1", "http://game.test")
	if err != nil {
		t.Fatal(err)
	}
	defer next.Close()
	if read(t, next).Type != "waiting" {
		t.Fatal("retry did not join queue")
	}
	assertCloseCode(t, old, closeReplaced)
	// Exercise the old reader's deferred cleanup after replacement, deterministically.
	h.mu.Lock()
	current := h.clients[1]
	if current == nil || len(h.waiting) != 1 || h.waiting[0] != current {
		h.mu.Unlock()
		t.Fatal("replacement removed from registry or queue")
	}
	h.mu.Unlock()
	duplicate, _, err := dial("", "http://game.test")
	if err != nil {
		t.Fatal(err)
	}
	defer duplicate.Close()
	assertCloseCode(t, duplicate, closeDuplicate)
	if err := next.WriteJSON(map[string]string{"type": "move"}); err != nil {
		t.Fatal(err)
	}
	if read(t, next).Type != "waiting" {
		t.Fatal("replacement no longer usable")
	}
}

func TestLeaveOldClientDoesNotRemoveReplacement(t *testing.T) {
	h := NewHandler(nil)
	old := &client{user: 1, send: make(chan any, 1)}
	next := &client{user: 1, send: make(chan any, 1)}
	h.clients[1] = old
	h.waiting = []*client{old}
	h.leave(old)
	h.clients[1] = next
	h.waiting = []*client{next}
	h.leave(old)
	if h.clients[1] != next || len(h.waiting) != 1 || h.waiting[0] != next {
		t.Fatal("deferred cleanup removed replacement")
	}
}
