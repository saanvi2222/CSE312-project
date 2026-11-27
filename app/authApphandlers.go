package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func parseRSAPublicKey(raw string) (*rsa.PublicKey, error) {
	trimmed := strings.TrimSpace(raw)

	if strings.HasPrefix(trimmed, "{") {
		var payload struct {
			PublicKey    string `json:"public_key"`
			PublicKeyAlt string `json:"publicKey"`
			Key          string `json:"key"`
		}
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			for _, candidate := range []string{payload.PublicKey, payload.PublicKeyAlt, payload.Key} {
				if candidate != "" {
					trimmed = strings.TrimSpace(candidate)
					break
				}
			}
		}
	}

	block, _ := pem.Decode([]byte(trimmed))
	if block == nil {
		return nil, errors.New("failed to decode PEM block for public key")
	}

	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("PKIX key is not an RSA public key")
		}
		return rsaPub, nil
	}

	rsaPub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}
	return rsaPub, nil
}

func fetchPublicKeyWithRetry(url string) (*rsa.PublicKey, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	for attempt := 1; attempt <= 15; attempt++ {
		resp, err := client.Get(url)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()

			if readErr == nil && resp.StatusCode == http.StatusOK {
				key, parseErr := parseRSAPublicKey(string(body))
				if parseErr == nil {
					return key, nil
				}
				err = parseErr
			} else if readErr != nil {
				err = readErr
			} else {
				err = fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
			}
		}

		fmt.Printf("Waiting for auth public key at %s (attempt %d/15): %v\n", url, attempt, err)
		time.Sleep(2 * time.Second)
	}

	return nil, errors.New("could not fetch RSA public key from auth service")
}

func verifyToken(tokenString string, pubKey *rsa.PublicKey) (string, error) {
	claims := &AuthClaims{}

	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))

	if err != nil || !parsedToken.Valid {
		return "", errors.New("invalid or expired token")
	}
	if claims.Username == "" {
		return "", errors.New("missing username claim")
	}

	return claims.Username, nil
}

func loginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func registerPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

func homePageHandler(pubKey *rsa.PublicKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenCookie, err := c.Cookie("token")
		if err != nil || tokenCookie == "" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"LoggedIn": false,
				"Username": "",
			})
			return
		}

		username, err := verifyToken(tokenCookie, pubKey)
		if err != nil {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"LoggedIn": false,
				"Username": "",
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"LoggedIn": true,
			"Username": username,
		})
	}
}
