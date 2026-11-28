package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"log"
	"net/http"

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

func usernameFromRequest(c *gin.Context, pubKey *rsa.PublicKey) (string, bool) { // reads the token cookie and verifies the JWT with the auth server's public key (no auth server or database call)
	tokenCookie, err := c.Cookie("token")
	if err != nil || tokenCookie == "" {
		return "", false
	}
	username, err := verifyToken(tokenCookie, pubKey) // rejects bad signatures, other algorithms, and expired tokens
	if err != nil {
		return "", false
	}
	return username, true
}

func WebSocketHandler(hub *Hub, pubKey *rsa.PublicKey) gin.HandlerFunc { // upgrades GET /ws to a WebSocket and starts the client
	return func(c *gin.Context) {
		username, ok := usernameFromRequest(c, pubKey) // check the JWT before accepting the connection
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("websocket upgrade failed:", err)
			return
		}

		client := &Client{
			Conn:     conn,
			Username: username, // from the verified JWT, so a user can never act as someone else
			SocketID: GenerateRandomSocketID(),
			Send:     make(chan []byte, 256),
		}

		hub.Register(client)
		go client.WritePump()
		client.ReadPump(hub) // blocks until the browser disconnects
	}
}

func BoardPageHandler(pubKey *rsa.PublicKey) gin.HandlerFunc { // serves the whiteboard page, only to logged-in users
	return func(c *gin.Context) {
		if _, ok := usernameFromRequest(c, pubKey); !ok {
			c.Redirect(http.StatusFound, "/login")
			return
		}
		c.HTML(http.StatusOK, "board.html", nil)
	}
}
