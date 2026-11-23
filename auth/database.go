package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func connectDB() *sql.DB {
	host := getEnv("DB_HOST", "db")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB handle: %v", err)
	}

	// Retry briefly in case Postgres is still finishing initialization
	for i := 1; i <= 10; i++ {
		if err = db.Ping(); err == nil {
			log.Println("Connected to PostgreSQL successfully")
			return db
		}
		log.Printf("Waiting for PostgreSQL (attempt %d/10): %v", i, err)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("Could not connect to PostgreSQL after retries: %v", err)
	return nil
}

func ensureUsersTable(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Failed to ensure users table: %v", err)
	}
	log.Println("Ensured users table exists")
}
