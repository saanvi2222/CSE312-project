package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//add endpoints for all auth requests that will come in then send back http response to nginx 

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "auth",
			"status":  "up",
		})
	})

	fmt.Println("Auth server listening on :8081...")
	router.Run(":8081")
}