package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func generateToken(username, role string) (string, error) {
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func parseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

type contextKey string

const userContextKey contextKey = "user"

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[7:]
		claims, err := parseToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUser(r *http.Request) *Claims {
	claims, ok := r.Context().Value(userContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public route - login
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}

		// Simple demo validation
		if req.Username == "admin" && req.Password == "password" {
			token, err := generateToken(req.Username, "admin")
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"token": token})
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	})

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
			user := getUser(r)
			writeJSON(w, http.StatusOK, map[string]string{
				"username": user.Username,
				"role":     user.Role,
				"message":  "Welcome to protected profile!",
			})
		})

		r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
			user := getUser(r)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"data":    []string{"item1", "item2", "item3"},
				"user":    user.Username,
				"message": "This is protected data",
			})
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}
	log.Printf("Chi server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
