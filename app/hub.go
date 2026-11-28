package main

import (
	"encoding/json"
	"math"
	"sort"
	"sync"
)

const (
	BoardWidth  = 1600
	BoardHeight = 900
	MaxStrokes  = 50000 // cap the history
)

type Hub struct {
	Mu      sync.RWMutex
	Clients map[string]*Client //unique socket id
	Strokes []DrawingMessage
}

func NewHub() *Hub {
	return &Hub{Clients: make(map[string]*Client)}
}

func (h *Hub) Register(client *Client) { //adds a client, sends them the current board, and updates
	h.Mu.Lock()
	h.Clients[client.SocketID] = client
	history := make([]DrawingMessage, len(h.Strokes))
	copy(history, h.Strokes)
	h.Mu.Unlock()

	client.SendJSON(WelcomeMessage{MessageType: "welcome", Username: client.Username})
	client.SendJSON(InitStrokesMessage{MessageType: "init_strokes", Strokes: history})
	h.BroadcastActiveUsers()
}

func (h *Hub) Unregister(socketID string) { // removes a client and tells everyone they left
	h.Mu.Lock()
	client, exists := h.Clients[socketID]
	if exists {
		delete(h.Clients, socketID)
		close(client.Send)
	}
	h.Mu.Unlock()

	if exists {
		h.Broadcast(UserLeftMessage{MessageType: "user_left", Username: client.Username}, "")
		h.BroadcastActiveUsers()
	}
}

func (h *Hub) Broadcast(msg any, exceptID string) { //sends a message to every client except the one with exceptID
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.Mu.RLock()
	defer h.Mu.RUnlock()
	for id, client := range h.Clients {
		if id != exceptID {
			client.Enqueue(payload)
		}
	}
}

func (h *Hub) AddStroke(stroke DrawingMessage) { // saves a segment so people who join later can see it
	h.Mu.Lock()
	if len(h.Strokes) < MaxStrokes {
		h.Strokes = append(h.Strokes, stroke)
	}
	h.Mu.Unlock()
}

func (h *Hub) EraseNear(x, y, radius float64) { //removes every segment that passes within radius of (x, y); board.html runs the same check
	h.Mu.Lock()
	kept := h.Strokes[:0]
	for _, s := range h.Strokes {
		if distanceToSegment(x, y, s.StartX, s.StartY, s.EndX, s.EndY) > radius {
			kept = append(kept, s)
		}
	}
	h.Strokes = kept
	h.Mu.Unlock()
}

func distanceToSegment(px, py, ax, ay, bx, by float64) float64 { //shortest distance from point (px, py) to the segment (ax, ay)-(bx, by)
	dx, dy := bx-ax, by-ay
	lengthSq := dx*dx + dy*dy
	t := 0.0
	if lengthSq > 0 {
		t = ((px-ax)*dx + (py-ay)*dy) / lengthSq
		t = math.Max(0, math.Min(1, t)) // stay on the segment, not the infinite line
	}
	cx, cy := ax+t*dx, ay+t*dy // closest point on the segment
	return math.Hypot(px-cx, py-cy)
}

func (h *Hub) BroadcastActiveUsers() { //sends everyone the list of unique usernames.
	h.Mu.RLock()
	seen := make(map[string]bool)
	users := []string{}
	for _, client := range h.Clients {
		if !seen[client.Username] { // one user with two tabs open only appears once
			seen[client.Username] = true
			users = append(users, client.Username)
		}
	}
	h.Mu.RUnlock()

	sort.Strings(users)
	h.Broadcast(ActiveUsersMessage{MessageType: "active_users_list", Users: users}, "")
}
