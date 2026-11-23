package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {

	gin.SetMode(gin.ReleaseMode)
	dbConn := connectDB()
	defer dbConn.Close()
	ensureUsersTable(dbConn)
	router := gin.Default()
	//add endpoints for all auth requests that will come in then send back http response to nginx

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "auth",
			"status":  "up",
		})
	})

	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"handled_by": "auth",
			"route":      "/login",
			"message":    "Login route reached the auth service successfully.",
		})
	})

	fmt.Println("Auth server listening on :8081...")
	router.Run(":8081")
}
