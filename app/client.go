package main

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second    // max time to write one message
	pongWait       = 60 * time.Second    // drop clients that stop answering pings
	pingPeriod     = (pongWait * 9) / 10 // how often we ping (heartbeat)
	maxMessageSize = 4096
)

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type Client struct { // one open WebSocket connection (one browser tab)
	Conn     *websocket.Conn
	Username string
	SocketID string
	Send     chan []byte // outgoing messages wait here for WritePump
}

func (c *Client) Enqueue(payload []byte) { // queues a message without blocking
	select {
	case c.Send <- payload:
	default:
	}
}

func (c *Client) SendJSON(msg any) {
	if payload, err := json.Marshal(msg); err == nil {
		c.Enqueue(payload)
	}
}

func (c *Client) ReadPump(hub *Hub) { // reads messages from this browser and acts on them
	defer func() {
		hub.Unregister(c.SocketID)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.MessageType {
		case "cursor":
			if !onBoard(msg.X, msg.Y) {
				continue
			}
			hub.Broadcast(CursorMessage{
				MessageType: "cursor",
				Username:    c.Username, // from the server, never the client
				X:           msg.X,
				Y:           msg.Y,
			}, c.SocketID)

		case "drawing":
			if !onBoard(msg.StartX, msg.StartY) || !onBoard(msg.EndX, msg.EndY) ||
				!colorPattern.MatchString(msg.Color) {
				continue
			}
			stroke := DrawingMessage{
				MessageType: "drawing",
				Username:    c.Username,
				StartX:      msg.StartX,
				StartY:      msg.StartY,
				EndX:        msg.EndX,
				EndY:        msg.EndY,
				Color:       msg.Color,
			}
			hub.AddStroke(stroke)
			hub.Broadcast(stroke, c.SocketID)
		}
		// Unknown message types are ignored
	}
}

func (c *Client) WritePump() { // the only goroutine that writes to this connection
	ticker := time.NewTicker(pingPeriod) // it sends queued messages and a ping every pingPeriod
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case payload, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { // hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func onBoard(x, y float64) bool {
	return x >= 0 && x <= BoardWidth && y >= 0 && y <= BoardHeight
}
