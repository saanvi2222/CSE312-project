package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Get database connection details
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	// Create the database connection string
	connectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName,
	)

	// Connect to PostgreSQL
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Check the database connection
	var pingErr error
	for attempt := 1; attempt <= 15; attempt++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		fmt.Printf("Waiting for PostgreSQL (attempt %d/15): %v\n", attempt, pingErr)
		time.Sleep(1 * time.Second)
	}
	if pingErr != nil {
		fmt.Println("Could not connect to PostgreSQL:", pingErr)
		return
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	publicKeyURL := os.Getenv("AUTH_PUBLIC_KEY_URL")
	if publicKeyURL == "" {
		publicKeyURL = "http://auth:8081/public-key"
	}

	pubKey, err := fetchPublicKeyWithRetry(publicKeyURL)
	if err != nil {
		fmt.Println("Public key error:", err)
		return
	}
	fmt.Println("Successfully loaded RS256 public key from auth service!")

	// Set up the web application
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.LoadHTMLFiles(
		"templates/index.html",
		"templates/login.html",
		"templates/register.html",
		"templates/board.html", // NEW
	)

	router.GET("/login", loginPageHandler)
	router.GET("/register", registerPageHandler)
	router.GET("/", homePageHandler(pubKey))

	//ive whiteboard
	hub := NewHub()
	router.GET("/board", BoardPageHandler(pubKey))
	router.GET("/ws", WebSocketHandler(hub, pubKey))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		var currentTime string

		err := db.QueryRow("SELECT NOW()").Scan(&currentTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Database query failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":        "healthy",
			"database":      "connected",
			"database_time": currentTime,
		})
	})

	// Start the server
	fmt.Println("Starting Go application on port 8080...")
	router.Run(":8080")
}
