package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
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

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		tokenStr := authHeader[7:] // Remove "Bearer "
		claims, err := parseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func main() {
	r := gin.Default()

	// Public route - login
	r.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Simple demo validation
		if req.Username == "admin" && req.Password == "password" {
			token, err := generateToken(req.Username, "admin")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"token": token})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	})

	// Protected routes
	api := r.Group("/api")
	api.Use(authMiddleware())

	api.GET("/profile", func(c *gin.Context) {
		username := c.MustGet("username").(string)
		role := c.MustGet("role").(string)
		c.JSON(http.StatusOK, gin.H{
			"username": username,
			"role":     role,
			"message":  "Welcome to protected profile!",
		})
	})

	api.GET("/data", func(c *gin.Context) {
		username := c.MustGet("username").(string)
		c.JSON(http.StatusOK, gin.H{
			"data":    []string{"item1", "item2", "item3"},
			"user":    username,
			"message": "This is protected data",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	log.Printf("Gin server starting on :%s", port)
	log.Fatal(r.Run(":" + port))
}
