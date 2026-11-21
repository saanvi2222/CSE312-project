package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

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
	if err := db.Ping(); err != nil {
		fmt.Println("Could not connect to PostgreSQL:", err)
		return
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	// Set up the web application
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.LoadHTMLFiles("templates/index.html")

	// Homepage
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		var currentTime string

		err := db.QueryRow("SELECT NOW()").Scan(&currentTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"message": "Database query failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"database": "connected",
			"database_time": currentTime,
		})
	})

	// Start the server
	fmt.Println("Starting Go application on port 8080...")
	router.Run(":8080")
}
