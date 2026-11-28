package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func GenerateRandomSocketID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

var guestCounter atomic.Int64 //everyone is a numbered guest

func WebSocketHandler(hub *Hub) gin.HandlerFunc { // upgrades GET /ws to a WebSocket and starts the client
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("websocket upgrade failed:", err)
			return
		}

		client := &Client{
			Conn:     conn,
			Username: fmt.Sprintf("guest-%d", guestCounter.Add(1)),
			SocketID: GenerateRandomSocketID(),
			Send:     make(chan []byte, 256),
		}

		hub.Register(client)
		go client.WritePump()
		client.ReadPump(hub) // blocks until the browser disconnects
	}
}

func BoardPageHandler(c *gin.Context) { // serves the whiteboard page
	c.HTML(http.StatusOK, "board.html", nil)
}
