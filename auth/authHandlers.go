package main

import (
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

//fuction to pass the publickey to app

func publicKeyHandler(pubKeyPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		pubPEM, err := os.ReadFile(pubKeyPath)
		if err != nil {
			fmt.Println("Failed to read public key:", err)
			c.String(http.StatusInternalServerError, "Failed to load public key")
			return
		}

		c.Data(http.StatusOK, "application/x-pem-file", pubPEM)
	}
}

//logs in a user

func loginHandler(db *sql.DB, privKey *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		username := strings.TrimSpace(req.Username)
		password := req.Password
		if username == "" || password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
			return
		}

		// 1. Look up user safely with a parameterized query
		var userID int64
		var passwordHash string
		err := db.QueryRow(
			"SELECT id, password_hash FROM users WHERE username = $1",
			username,
		).Scan(&userID, &passwordHash)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		if err != nil {
			fmt.Println("Database error during login:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// 2. Compare submitted password against stored bcrypt salt+hash
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// 3. Sign an RS256 JWT with privKey
		now := time.Now()
		claims := UserClaims{
			Username: username,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   fmt.Sprintf("%d", userID),
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signedToken, err := token.SignedString(privKey)
		if err != nil {
			fmt.Println("Failed to sign JWT:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Set cookie with the signed JWT (24 hours = 86400 seconds)
		c.SetCookie(
			"token",     // name
			signedToken, // value (the JWT)
			24*60*60,    // maxAge in seconds
			"/",         // path
			"",          // domain (empty uses current host)
			false,       // secure: set to true over HTTPS in production
			true,        // httpOnly: true helps protect the JWT from JavaScript access (XSS)
		)

		c.JSON(http.StatusOK, gin.H{
			"message":  "Login successful",
			"token":    signedToken,
			"Username": username, // note: using trimmed username is even cleaner than req.Username
		})

	}
}

//Registers a user

func registerHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		username := strings.TrimSpace(req.Username)
		password := req.Password

		// 1. Username rules: non-empty + at least one special character
		if username == "" || !hasSpecialChar(username) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username must contain at least one special character",
			})
			return
		}

		// 2. Password rules: >= 8 chars, 1 lower, 1 upper, 1 digit, 1 special char
		if !isValidPassword(password) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password must be at least 8 characters and include an uppercase letter, lowercase letter, number, and special character",
			})
			return
		}

		// 3. Check for duplicate username using a parameterized query
		var exists bool
		err := db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)",
			username,
		).Scan(&exists)
		if err != nil {
			fmt.Println("Database error checking username:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "The username is already taken"})
			return
		}

		// 4. Hash & salt password with bcrypt
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println("Failed to hash password:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// 5. Insert new user using a parameterized query
		_, err = db.Exec(
			"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
			username,
			string(hashedBytes),
		)
		if err != nil {
			fmt.Println("Failed to insert user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	}
}

func hasSpecialChar(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func isValidPassword(pw string) bool {
	if len(pw) < 8 {
		return false
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range pw {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsSpace(r):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}

// logs out a user
func logoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear cookie if you use one (adjust cookie name, path, domain, and Secure flag to match login)
		c.SetCookie("token", "", -1, "/", "", false, true)

		c.JSON(http.StatusOK, gin.H{
			"message": "Logged out successfully",
		})
	}
}
