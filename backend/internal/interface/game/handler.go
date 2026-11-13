package game

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	board "gomoku-backend/internal/domain/game"
	"gomoku-backend/internal/interface/httpapi/middleware"
)

type client struct {
	conn  *websocket.Conn
	send  chan any
	user  int64
	color int
	room  *room
	left  bool
}

const (
	closeDuplicate = 4409
	closeReplaced  = 4410
)

type room struct {
	id      string
	users   [2]int64
	started time.Time
	saved   bool
	saving  bool
	game    *board.Game
	players [2]*client
	cpu     bool
}
type state struct {
	ResultSaved bool   `json:"resultSaved"`
	Type        string `json:"type"`
	*board.Game
	Color int    `json:"color"`
	Mode  string `json:"mode"`
}
type Hub struct {
	results  board.ResultRepository
	mu       sync.Mutex
	waiting  []*client
	clients  map[int64]*client
	upgrader websocket.Upgrader
}

func NewHandler(origins []string, repositories ...board.ResultRepository) *Hub {
	var results board.ResultRepository
	if len(repositories) > 0 {
		results = repositories[0]
	}
	allowed := make(map[string]bool)
	for _, origin := range origins {
		allowed[origin] = true
	}
	return &Hub{results: results, clients: make(map[int64]*client), upgrader: websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin:      func(r *http.Request) bool { return allowed[r.Header.Get("Origin")] },
	}}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode != "cpu" && mode != "online" {
		http.Error(w, "invalid mode", 400)
		return
	}
	color := 0
	if mode == "cpu" {
		switch r.URL.Query().Get("color") {
		case "black":
			color = 1
		case "white":
			color = 2
		default:
			http.Error(w, "先手（black）か後手（white）を選んでください。", http.StatusBadRequest)
			return
		}
	}
	id, ok := middleware.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", 401)
		return
	}
	h.mu.Lock()
	previous := h.clients[id]
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.mu.Unlock()
		return
	}
	if previous != nil && r.URL.Query().Get("replace") != "1" {
		h.mu.Unlock()
		// HTTP handshake errors are hidden from browser JavaScript. Use a
		// dedicated close code so the UI can explain shared login cookies.
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeDuplicate, "account_already_connected"), time.Now().Add(time.Second))
		conn.Close()
		return
	}
	// An explicit retry takes over a stale connection immediately, without
	// waiting for its read deadline. Failed upgrades must leave it untouched.
	if previous != nil {
		_ = previous.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(closeReplaced, "connection_replaced"), time.Now().Add(time.Second))
		h.leave(previous)
		previous.conn.Close()
	}
	c := &client{conn: conn, send: make(chan any, 32), user: id}
	h.clients[id] = c
	go c.write()
	if mode == "cpu" {
		c.color = color
		c.room = newRoom(true)
		c.room.users[c.color-1] = c.user
		c.room.players[c.color-1] = c
		if c.color == 2 {
			_ = c.room.game.Move(1, c.room.game.CPU())
		}
		h.broadcast(c.room)
	} else {
		h.waiting = append(h.waiting, c)
		if len(h.waiting) >= 2 {
			a, b := h.waiting[0], h.waiting[1]
			h.waiting = h.waiting[2:]
			if rand.IntN(2) == 1 {
				a, b = b, a
			}
			room := newRoom(false)
			room.players = [2]*client{a, b}
			room.users = [2]int64{a.user, b.user}
			a.room, b.room, a.color, b.color = room, room, 1, 2
			h.broadcast(room)
		} else {
			h.sendWaiting(c)
		}
	}
	h.mu.Unlock()
	defer func() { h.mu.Lock(); h.leave(c); h.mu.Unlock(); conn.Close() }()
	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(70 * time.Second)) })
	for {
		var msg struct {
			Type  string `json:"type"`
			Index *int   `json:"index"`
		}
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}
		h.mu.Lock()
		if c.left {
			h.mu.Unlock()
			return
		}
		room := c.room
		if room == nil {
			h.sendWaiting(c)
		} else if msg.Type == "resign" && !room.game.Finished {
			room.game.Finished, room.game.Winner, room.game.Reason = true, 3-c.color, "resign"
			h.broadcast(room)
		} else if msg.Type == "move" && msg.Index != nil {
			if err := room.game.Move(c.color, *msg.Index); err != nil {
				h.send(c, map[string]string{"type": "error", "message": err.Error()})
			} else {
				if room.cpu && !room.game.Finished {
					_ = room.game.Move(3-c.color, room.game.CPU())
				}
				h.broadcast(room)
			}
		} else {
			h.send(c, map[string]string{"type": "error", "message": "無効な操作です。"})
		}
		h.mu.Unlock()
	}
}

// All state and queue operations run under the hub lock. Socket writes use a single writer.
func (h *Hub) sendWaiting(c *client) {
	h.send(c, map[string]string{"type": "waiting", "message": "対戦相手を待っています。"})
}

func (h *Hub) send(c *client, message any) {
	select {
	case c.send <- message:
	default:
		c.conn.Close()
	}
}
func newRoom(cpu bool) *room {
	var id [32]byte
	if _, err := cryptorand.Read(id[:]); err != nil {
		panic(err)
	}
	return &room{id: hex.EncodeToString(id[:]), game: board.New(), cpu: cpu, started: time.Now().UTC()}
}

func (h *Hub) broadcast(r *room) {
	if r.game.Finished && !r.saving && h.results != nil {
		r.saving = true
		mode := "online"
		if r.cpu {
			mode = "cpu"
		}
		result := board.Result{ID: r.id, Mode: mode, Players: r.users, Winner: r.game.Winner, Reason: r.game.Reason, Board: r.game.Board, StartedAt: r.started, FinishedAt: time.Now().UTC()}
		go h.saveResult(r, result)
	}

	mode := "online"
	if r.cpu {
		mode = "cpu"
	}
	for _, c := range r.players {
		if c != nil {
			copy := *r.game
			h.send(c, state{ResultSaved: r.saved, Type: "state", Game: &copy, Color: c.color, Mode: mode})
		}
	}
}
func (h *Hub) leave(c *client) {
	if c.left {
		return
	}
	c.left = true
	if h.clients[c.user] == c {
		delete(h.clients, c.user)
	}
	for i, p := range h.waiting {
		if p == c {
			h.waiting = append(h.waiting[:i], h.waiting[i+1:]...)
			break
		}
	}
	if r := c.room; r != nil {
		r.players[c.color-1] = nil
		if !r.game.Finished {
			r.game.Finished, r.game.Winner, r.game.Reason = true, 3-c.color, "disconnect"
			h.broadcast(r)
		}
	}
	close(c.send)
}
func (c *client) write() {
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	defer c.conn.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if c.conn.WriteJSON(message) != nil {
				return
			}
		case <-ticker.C:
			if c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)) != nil {
				return
			}
		}
	}
}

// Keep retrying transient database failures; one game ID is applied at most once.
func (h *Hub) saveResult(r *room, result board.Result) {
	for attempt := 0; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := h.results.Save(ctx, result)
		cancel()
		if err == nil {
			h.mu.Lock()
			r.saved = true
			h.broadcast(r)
			h.mu.Unlock()
			return
		}
		log.Printf("save game %s (attempt %d): %v", result.ID, attempt+1, err)
		if attempt == 0 {
			h.mu.Lock()
			for _, c := range r.players {
				if c != nil {
					h.send(c, map[string]string{"type": "error", "message": "対戦結果の保存を再試行しています。ランキング反映までお待ちください。"})
				}
			}
			h.mu.Unlock()
		}
		time.Sleep(5 * time.Second)
	}
}
