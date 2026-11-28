package main

import (
	"encoding/json"
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
