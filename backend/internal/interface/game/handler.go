package game

import (
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
}
type room struct {
	game    *board.Game
	players [2]*client
	cpu     bool
}
type state struct {
	Type string `json:"type"`
	*board.Game
	Color int    `json:"color"`
	Mode  string `json:"mode"`
}
type Hub struct {
	mu       sync.Mutex
	waiting  []*client
	clients  map[int64]*client
	upgrader websocket.Upgrader
}

func NewHandler(origins []string) *Hub {
	allowed := make(map[string]bool)
	for _, origin := range origins {
		allowed[origin] = true
	}
	return &Hub{clients: make(map[int64]*client), upgrader: websocket.Upgrader{
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
	if h.clients[id] != nil {
		h.mu.Unlock()
		http.Error(w, "既に対戦画面が開いています。", 409)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.mu.Unlock()
		return
	}
	c := &client{conn: conn, send: make(chan any, 32), user: id}
	h.clients[id] = c
	go c.write()
	if mode == "cpu" {
		c.color = color
		c.room = &room{game: board.New(), cpu: true}
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
			room := &room{game: board.New(), players: [2]*client{a, b}}
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
func (h *Hub) broadcast(r *room) {
	mode := "online"
	if r.cpu {
		mode = "cpu"
	}
	for _, c := range r.players {
		if c != nil {
			copy := *r.game
			h.send(c, state{Type: "state", Game: &copy, Color: c.color, Mode: mode})
		}
	}
}
func (h *Hub) leave(c *client) {
	delete(h.clients, c.user)
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
