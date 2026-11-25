package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	dbConn := connectDB()
	defer dbConn.Close()
	ensureUsersTable(dbConn)

	//create or load in private key
	privKey, _, err := loadOrCreateRSAKeys("/app/keys")
	if err != nil {
		log.Fatalf("Failed to load or create RSA keys: %v", err)
	}

	router := gin.Default()
	//add endpoints for all auth requests that will come in then send back http response to nginx

	router.POST("/login", loginHandler(dbConn, privKey))
	router.POST("/register", registerHandler(dbConn))
	router.POST("/logout", logoutHandler())
	router.GET("/public-key", publicKeyHandler("/app/keys/public.pem"))

	fmt.Println("Auth server listening on :8081...")
	router.Run(":8081")
}
