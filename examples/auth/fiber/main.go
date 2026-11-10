package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

func authMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "missing authorization header"})
	}

	tokenStr := authHeader[7:] // Remove "Bearer "
	claims, err := parseToken(tokenStr)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
	}

	c.Locals("username", claims.Username)
	c.Locals("role", claims.Role)
	return c.Next()
}

func main() {
	app := fiber.New()
	app.Use(logger.New())

	// Public route - login
	app.Post("/login", func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		// Simple demo validation
		if req.Username == "admin" && req.Password == "password" {
			token, err := generateToken(req.Username, "admin")
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "failed to generate token"})
			}
			return c.JSON(fiber.Map{"token": token})
		}
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	})

	// Protected routes
	api := app.Group("/api")
	api.Use(authMiddleware)

	api.Get("/profile", func(c *fiber.Ctx) error {
		username := c.Locals("username").(string)
		role := c.Locals("role").(string)
		return c.JSON(fiber.Map{
			"username": username,
			"role":     role,
			"message":  "Welcome to protected profile!",
		})
	})

	api.Get("/data", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"data":    []string{"item1", "item2", "item3"},
			"user":    c.Locals("username"),
			"message": "This is protected data",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	log.Printf("Fiber server starting on :%s", port)
	log.Fatal(app.Listen(":" + port))
}
